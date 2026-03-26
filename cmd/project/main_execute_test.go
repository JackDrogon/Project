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

type exitPanic struct {
	code int
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	fn()
	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	return string(data)
}

func TestMain_PrintsVersion(t *testing.T) {
	oldArgs := os.Args
	oldTag := buildinfo.Tag
	os.Args = []string{"project", "version"}
	buildinfo.Tag = "main-test-tag"
	t.Cleanup(func() {
		os.Args = oldArgs
		buildinfo.Tag = oldTag
	})

	output := captureStdout(t, func() {
		main()
	})
	if !strings.Contains(output, "main-test-tag") {
		t.Fatalf("output = %q, want contains test tag", output)
	}
}

func TestExecute_ExitsOnError(t *testing.T) {
	oldArgs := os.Args
	oldExit := exitFunc
	oldStderr := stderrWriter
	os.Args = []string{"project", "new"}
	var stderr bytes.Buffer
	exitFunc = func(code int) {
		panic(exitPanic{code: code})
	}
	stderrWriter = &stderr
	t.Cleanup(func() {
		os.Args = oldArgs
		exitFunc = oldExit
		stderrWriter = oldStderr
	})

	creator := appcreate.NewCreator(fstest.MapFS{}, io.Discard)

	defer func() {
		r := recover()
		panicValue, ok := r.(exitPanic)
		if !ok {
			t.Fatalf("recover() = %#v, want exit panic", r)
		}
		if panicValue.code != 1 {
			t.Fatalf("exit code = %d, want 1", panicValue.code)
		}
		if !strings.Contains(stderr.String(), "accepts 1 arg") {
			t.Fatalf("stderr = %q, want arg error", stderr.String())
		}
	}()

	Execute(creator)
}

func TestExecute_PropagatesMalformedDiscoveredConfigToStderr(t *testing.T) {
	tempConfigHome := t.TempDir()
	configDir := filepath.Join(tempConfigHome, "project")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("version = 1\n[version]\nverbose = \"nope\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	oldArgs := os.Args
	oldExit := exitFunc
	oldStderr := stderrWriter
	oldConfigService := newConfigService
	os.Args = []string{"project", "version"}
	var stderr bytes.Buffer
	exitFunc = func(code int) {
		panic(exitPanic{code: code})
	}
	stderrWriter = &stderr
	newConfigService = func() *appconfig.Service {
		return appconfig.NewServiceWithDeps(appconfig.Dependencies{
			UserConfigDir: func() (string, error) { return tempConfigHome, nil },
		})
	}
	t.Cleanup(func() {
		os.Args = oldArgs
		exitFunc = oldExit
		stderrWriter = oldStderr
		newConfigService = oldConfigService
	})

	creator := appcreate.NewCreator(fstest.MapFS{}, io.Discard)

	defer func() {
		r := recover()
		panicValue, ok := r.(exitPanic)
		if !ok {
			t.Fatalf("recover() = %#v, want exit panic", r)
		}
		if panicValue.code != 1 {
			t.Fatalf("exit code = %d, want 1", panicValue.code)
		}
		if !strings.Contains(stderr.String(), configPath) {
			t.Fatalf("stderr = %q, want contains config path %q", stderr.String(), configPath)
		}
		if !strings.Contains(stderr.String(), "cannot decode TOML string") {
			t.Fatalf("stderr = %q, want contains decode failure detail", stderr.String())
		}
	}()

	Execute(creator)
}

func TestExecute_PropagatesMalformedExplicitConfigToStderr(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("version = 1\n[completion]\nshell = 42\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	oldArgs := os.Args
	oldExit := exitFunc
	oldStderr := stderrWriter
	os.Args = []string{"project", "--config", configPath, "version"}
	var stderr bytes.Buffer
	exitFunc = func(code int) {
		panic(exitPanic{code: code})
	}
	stderrWriter = &stderr
	t.Cleanup(func() {
		os.Args = oldArgs
		exitFunc = oldExit
		stderrWriter = oldStderr
	})

	creator := appcreate.NewCreator(fstest.MapFS{}, io.Discard)

	defer func() {
		r := recover()
		panicValue, ok := r.(exitPanic)
		if !ok {
			t.Fatalf("recover() = %#v, want exit panic", r)
		}
		if panicValue.code != 1 {
			t.Fatalf("exit code = %d, want 1", panicValue.code)
		}
		if !strings.Contains(stderr.String(), configPath) {
			t.Fatalf("stderr = %q, want contains config path %q", stderr.String(), configPath)
		}
		if !strings.Contains(stderr.String(), "cannot decode TOML integer") {
			t.Fatalf("stderr = %q, want contains decode failure detail", stderr.String())
		}
	}()

	Execute(creator)
}
