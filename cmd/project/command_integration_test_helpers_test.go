package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

func requireSubcommand(t *testing.T, creator *appcreate.Creator, key commandKey) *cobra.Command {
	t.Helper()

	for _, provider := range registeredCommandProviders() {
		if provider.key() == key {
			return provider.buildCommand(commandDependencies{creator: creator})
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
