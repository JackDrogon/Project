package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	appcreate "github.com/JackDrogon/project/internal/app/create"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

// Scaffolding must never delete user files. These tests pin that guarantee for
// every route that can turn on --force, including a replay file supplying it.

const destinationSentinelName = "keep-me.txt"

func templateFSForDestinationTest() fstest.MapFS {
	return fstest.MapFS{
		"go/go.mod.tmpl": {Data: []byte("module {{.ModulePath}}\n")},
	}
}

// seedDestination creates dir, and for a non-empty sentinel also writes a file
// whose survival proves nothing was removed.
func seedDestination(t *testing.T, dir, sentinel string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", dir, err)
	}
	if sentinel == "" {
		return
	}
	if err := os.WriteFile(filepath.Join(dir, destinationSentinelName), []byte(sentinel), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", destinationSentinelName, err)
	}
}

func requireSentinelIntact(t *testing.T, dir, want string) {
	t.Helper()

	got, err := os.ReadFile(filepath.Join(dir, destinationSentinelName))
	if err != nil {
		t.Fatalf("pre-existing file was removed: ReadFile error = %v", err)
	}
	if string(got) != want {
		t.Fatalf("pre-existing file = %q, want %q", string(got), want)
	}
}

func TestNewCmd_ForceRejectsNonEmptyDirectoryAndKeepsFiles(t *testing.T) {
	workDir := withTempWorkingDir(t, "workspace")
	target := filepath.Join(workDir, "demo")
	seedDestination(t, target, "user data")

	creator := appcreate.NewCreator(templateFSForDestinationTest(), &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "--module", "example.com/demo", "--force", "demo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for non-empty destination, got nil")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("Execute() error = %v, want non-empty destination error", err)
	}

	requireSentinelIntact(t, target, "user data")
}

func TestNewCmd_ReplayForceCannotDeleteNonEmptyDirectory(t *testing.T) {
	workDir := withTempWorkingDir(t, "workspace")
	target := filepath.Join(workDir, "demo")
	seedDestination(t, target, "user data")

	replayPath := writeReplayTOMLForTest(t, protocoltoml.Replay{
		Version:  protocoltoml.ReplayVersion,
		Mode:     string(appcreate.CommandNew),
		Template: protocoltoml.ReplayTemplate{Lang: "go"},
		Project: protocoltoml.ReplayProject{
			Name:       "demo",
			TargetDir:  "demo",
			ModulePath: "example.com/demo",
		},
		Git:     protocoltoml.ReplayGit{Mode: domain.GitModeNone},
		Options: protocoltoml.ReplayOptions{Force: true},
		Inputs:  map[string]string{"module_path": "example.com/demo"},
	})

	creator := appcreate.NewCreator(templateFSForDestinationTest(), &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--replay", replayPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for non-empty destination, got nil")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("Execute() error = %v, want non-empty destination error", err)
	}

	requireSentinelIntact(t, target, "user data")
}

func TestNewCmd_ForceScaffoldsIntoExistingEmptyDirectory(t *testing.T) {
	workDir := withTempWorkingDir(t, "workspace")
	target := filepath.Join(workDir, "demo")
	seedDestination(t, target, "")

	creator := appcreate.NewCreator(templateFSForDestinationTest(), &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "--module", "example.com/demo", "--force", "demo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if string(goMod) != "module example.com/demo\n" {
		t.Fatalf("go.mod = %q, want %q", string(goMod), "module example.com/demo\n")
	}
}

func TestNewCmd_RejectsExistingEmptyDirectoryWithoutForce(t *testing.T) {
	workDir := withTempWorkingDir(t, "workspace")
	seedDestination(t, filepath.Join(workDir, "demo"), "")

	creator := appcreate.NewCreator(templateFSForDestinationTest(), &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "--module", "example.com/demo", "demo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for existing directory, got nil")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Fatalf("Execute() error = %v, want hint to pass --force", err)
	}
}
