package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/JackDrogon/project/internal/adapters/buildinfo"
	appconfig "github.com/JackDrogon/project/internal/app/config"
	appcreate "github.com/JackDrogon/project/internal/app/create"
)

func newRunTestDependencies() dependencies {
	return newTestDependencies(appcreate.NewCreator(fstest.MapFS{}, io.Discard))
}

func TestRun_PrintsVersion(t *testing.T) {
	oldTag := buildinfo.Tag
	buildinfo.Tag = "main-test-tag"
	t.Cleanup(func() {
		buildinfo.Tag = oldTag
	})

	var stdout, stderr bytes.Buffer
	code := run(newRunTestDependencies(), []string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("run() = %d, want 0 (stderr = %q)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "main-test-tag") {
		t.Fatalf("stdout = %q, want contains test tag", stdout.String())
	}
}

func TestRun_ExitsWithFailureCodeOnError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(newRunTestDependencies(), []string{"new"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "accepts 1 arg") {
		t.Fatalf("stderr = %q, want arg error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want failures to stay off stdout", stdout.String())
	}
}

// The error used to be printed by cobra and then again by the exit path.
func TestRun_PrintsErrorOnlyOnce(t *testing.T) {
	var stdout, stderr bytes.Buffer
	run(newRunTestDependencies(), []string{"new"}, &stdout, &stderr)

	const message = "accepts 1 arg(s), received 0"
	if got := strings.Count(stderr.String(), message); got != 1 {
		t.Fatalf("stderr contains %q %d times, want exactly 1\nstderr = %q", message, got, stderr.String())
	}
}

func TestRun_PropagatesMalformedDiscoveredConfigToStderr(t *testing.T) {
	tempConfigHome := t.TempDir()
	configDir := filepath.Join(tempConfigHome, "project")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("version = 1\n[version]\nverbose = \"nope\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	deps := newRunTestDependencies()
	deps.newConfigService = func() *appconfig.Service {
		return appconfig.NewServiceWithDeps(appconfig.Dependencies{
			UserConfigDir: func() (string, error) { return tempConfigHome, nil },
		})
	}

	var stdout, stderr bytes.Buffer
	code := run(deps, []string{"version"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), configPath) {
		t.Fatalf("stderr = %q, want contains config path %q", stderr.String(), configPath)
	}
	if !strings.Contains(stderr.String(), "cannot decode TOML string") {
		t.Fatalf("stderr = %q, want contains decode failure detail", stderr.String())
	}
}

func TestRun_PropagatesMalformedExplicitConfigToStderr(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("version = 1\n[completion]\nshell = 42\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run(newRunTestDependencies(), []string{"--config", configPath, "version"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), configPath) {
		t.Fatalf("stderr = %q, want contains config path %q", stderr.String(), configPath)
	}
	if !strings.Contains(stderr.String(), "cannot decode TOML integer") {
		t.Fatalf("stderr = %q, want contains decode failure detail", stderr.String())
	}
}
