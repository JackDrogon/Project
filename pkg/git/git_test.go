package git

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestGitRunHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_GIT_HELPER_PROCESS=1")
	return cmd
}

func TestRun(t *testing.T) {
	oldExecCommand := execCommand
	execCommand = fakeExecCommand
	t.Cleanup(func() {
		execCommand = oldExecCommand
	})

	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("GIT_HELPER_MODE", "success")
		t.Setenv("GIT_EXPECTED_DIR", dir)

		if err := Run(dir, "status"); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	})

	t.Run("failure", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("GIT_HELPER_MODE", "fail")
		t.Setenv("GIT_EXPECTED_DIR", dir)

		err := Run(dir, "commit")
		if err == nil {
			t.Fatal("Run() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "git commit failed") || !strings.Contains(err.Error(), "helper failure") {
			t.Fatalf("Run() error = %v, want wrapped command failure", err)
		}
	})
}

func TestGitRunHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_GIT_HELPER_PROCESS") != "1" {
		return
	}

	if cwd, err := os.Getwd(); err != nil || cwd != os.Getenv("GIT_EXPECTED_DIR") {
		os.Stderr.WriteString("unexpected working directory")
		os.Exit(1)
	}

	if os.Getenv("GIT_HELPER_MODE") == "fail" {
		os.Stderr.WriteString("helper failure")
		os.Exit(1)
	}

	os.Exit(0)
}
