//go:build unix

package templatefs

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"testing/fstest"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

func TestMaterialize_CorrectsForUmaskOnPOSIX(t *testing.T) {
	t.Parallel()

	oldUmask := syscall.Umask(0o077)
	defer syscall.Umask(oldUmask)

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

	info, err := os.Stat(filepath.Join(dest, "script.sh"))
	if err != nil {
		t.Fatalf("Stat(script.sh) error = %v", err)
	}
	if got, want := info.Mode().Perm(), fs.FileMode(0o755); got != want {
		t.Fatalf("script.sh mode = %o, want %o", got, want)
	}
}
