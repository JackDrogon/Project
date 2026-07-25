package templatefs

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

func TestMaterialize_PreservesFileModesOnPOSIX(t *testing.T) {
	if runtime.GOOS == goosWindows {
		t.Skip("POSIX mode contract does not apply on Windows")
	}

	fsys := fstest.MapFS{
		"lang/dev-tools":           {Mode: fs.ModeDir | 0o755},
		"lang/dev-tools/script.sh": {Data: []byte("#!/bin/sh\nexit 0\n")},
		"lang/README.md":           {Data: []byte("# README\n")},
	}
	modeMap := map[string]fs.FileMode{
		"lang/dev-tools":           0o755,
		"lang/dev-tools/script.sh": 0o755,
		"lang/README.md":           0o644,
	}
	resolveMode := func(sourcePath string, isDir bool) (fs.FileMode, bool) {
		mode, ok := modeMap[sourcePath]
		return mode, ok
	}

	dest := filepath.Join(t.TempDir(), "demo")
	vars := domain.NewTemplateVars("Demo", "example.com/demo", "1.25", "alice", 2030)
	if err := Materialize(&bytes.Buffer{}, fsys, "lang", dest, vars, resolveMode); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}

	info, err := os.Stat(filepath.Join(dest, "dev-tools", "script.sh"))
	if err != nil {
		t.Fatalf("Stat(script.sh) error = %v", err)
	}
	if got, want := info.Mode().Perm(), fs.FileMode(0o755); got != want {
		t.Fatalf("script.sh mode = %o, want %o", got, want)
	}

	info, err = os.Stat(filepath.Join(dest, "README.md"))
	if err != nil {
		t.Fatalf("Stat(README.md) error = %v", err)
	}
	if got, want := info.Mode().Perm(), fs.FileMode(0o644); got != want {
		t.Fatalf("README.md mode = %o, want %o", got, want)
	}
}

func TestMaterialize_PreservesDirectoryModesOnPOSIX(t *testing.T) {
	if runtime.GOOS == goosWindows {
		t.Skip("POSIX mode contract does not apply on Windows")
	}

	fsys := fstest.MapFS{
		"lang/bin":         {Mode: fs.ModeDir | 0o750},
		"lang/bin/tool.sh": {Data: []byte("#!/bin/sh\nexit 0\n")},
	}
	resolveMode := func(sourcePath string, isDir bool) (fs.FileMode, bool) {
		switch sourcePath {
		case "lang/bin":
			return 0o750, true
		case "lang/bin/tool.sh":
			return 0o755, true
		default:
			return 0, false
		}
	}

	dest := filepath.Join(t.TempDir(), "demo")
	vars := domain.NewTemplateVars("Demo", "example.com/demo", "1.25", "alice", 2030)
	if err := Materialize(&bytes.Buffer{}, fsys, "lang", dest, vars, resolveMode); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}

	info, err := os.Stat(filepath.Join(dest, "bin"))
	if err != nil {
		t.Fatalf("Stat(bin) error = %v", err)
	}
	if got, want := info.Mode().Perm(), fs.FileMode(0o750); got != want {
		t.Fatalf("bin mode = %o, want %o", got, want)
	}
}

func TestMaterialize_NonPOSIXGracefulDegradation(t *testing.T) {
	if runtime.GOOS != goosWindows {
		t.Skip("non-POSIX degradation contract is only exercised on Windows")
	}

	fsys := fstest.MapFS{
		"lang/script.sh": {Data: []byte("#!/bin/sh\nexit 0\n")},
	}
	resolveMode := func(sourcePath string, isDir bool) (fs.FileMode, bool) {
		if sourcePath == "lang/script.sh" {
			return 0o755, true
		}
		return 0, false
	}

	dest := filepath.Join(t.TempDir(), "demo")
	vars := domain.NewTemplateVars("Demo", "example.com/demo", "1.25", "alice", 2030)
	if err := Materialize(&bytes.Buffer{}, fsys, "lang", dest, vars, resolveMode); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "script.sh")); err != nil {
		t.Fatalf("Stat(script.sh) error = %v", err)
	}
}
