package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	appconfig "github.com/JackDrogon/project/internal/app/config"
	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

// newTestDependencies starts from the production wiring and pins the creator,
// so a test only states what it actually wants to be different.
func newTestDependencies(creator *appcreate.Creator) dependencies {
	deps := newDependencies()
	deps.newCreator = func(io.Writer) *appcreate.Creator { return creator }
	return deps
}

func requireSubcommand(t *testing.T, creator *appcreate.Creator, key commandKey) *cobra.Command {
	t.Helper()

	return requireSubcommandWithDeps(t, newTestDependencies(creator), key)
}

// requireSubcommandWithDeps returns a parentless subcommand so tests can run it
// directly without cobra rerouting Execute through the root command.
func requireSubcommandWithDeps(t *testing.T, deps dependencies, key commandKey) *cobra.Command {
	t.Helper()

	return requireSubcommandWithConfig(t, deps, &appconfig.Resolved{}, key)
}

// requireSubcommandWithConfig is requireSubcommandWithDeps for tests that need
// the subcommand to see the config the root command would have resolved.
func requireSubcommandWithConfig(t *testing.T, deps dependencies, config *appconfig.Resolved, key commandKey) *cobra.Command {
	t.Helper()

	for _, cmd := range subcommands(deps, config) {
		if cmd.Name() == string(key) {
			return cmd
		}
	}

	t.Fatalf("subcommand %q not found", key)
	return nil
}

func requireOrderedSubstrings(t *testing.T, got string, want []string) {
	t.Helper()

	searchFrom := 0
	for _, fragment := range want {
		idx := strings.Index(got[searchFrom:], fragment)
		if idx == -1 {
			t.Fatalf("output = %q, want contains %q", got, fragment)
		}
		searchFrom += idx + len(fragment)
	}
}

func writeReplayTOMLForTest(t *testing.T, replay protocoltoml.Replay) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "replay.toml")
	if err := protocoltoml.WriteReplay(path, replay); err != nil {
		t.Fatalf("WriteReplay(%q) error = %v", path, err)
	}

	return path
}

func withTempWorkingDir(t *testing.T, baseName string) string {
	t.Helper()

	parent := t.TempDir()
	dir := filepath.Join(parent, baseName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", dir, err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%q) error = %v", dir, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	return dir
}
