package gitexec

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func fakeExecCommand(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestGitRunHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_GIT_HELPER_PROCESS=1")
	return cmd
}

func TestNew(t *testing.T) {
	adapter := New()
	if adapter == nil {
		t.Fatal("New() = nil")
	}
}

func TestRun(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = fakeExecCommand
	t.Cleanup(func() {
		execCommand = oldExecCommand
	})

	adapter := New()

	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("GIT_HELPER_MODE", "success")
		t.Setenv("GIT_EXPECTED_DIR", dir)

		if err := adapter.Run(t.Context(), dir, "status"); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	t.Run("failure", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("GIT_HELPER_MODE", "fail")
		t.Setenv("GIT_EXPECTED_DIR", dir)

		err := adapter.Run(t.Context(), dir, "commit")
		if err == nil {
			t.Fatal("Run() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "git commit failed") || !strings.Contains(err.Error(), "helper failure") {
			t.Fatalf("Run() error = %v, want wrapped command failure", err)
		}
	})

	t.Run("canceled context reports cancellation", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("GIT_HELPER_MODE", "success")
		t.Setenv("GIT_EXPECTED_DIR", dir)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := adapter.Run(ctx, dir, "commit")
		if err == nil {
			t.Fatal("Run() expected error, got nil")
		}
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run() error = %v, want context.Canceled", err)
		}
		if !strings.Contains(err.Error(), "git commit canceled") {
			t.Fatalf("Run() error = %v, want cancellation message", err)
		}
	})
}

func TestGitRunHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_GIT_HELPER_PROCESS") != "1" {
		return
	}

	if cwd, err := os.Getwd(); err != nil || cwd != os.Getenv("GIT_EXPECTED_DIR") {
		_, _ = os.Stderr.WriteString("unexpected working directory")
		os.Exit(1)
	}

	if os.Getenv("GIT_HELPER_MODE") == "fail" {
		_, _ = os.Stderr.WriteString("helper failure")
		os.Exit(1)
	}

	os.Exit(0)
}
