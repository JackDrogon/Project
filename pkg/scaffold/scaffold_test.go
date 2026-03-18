package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestListLangs(t *testing.T) {
	fsys := fstest.MapFS{
		"go/Makefile":  {Data: []byte("build:")},
		"cpp/Makefile": {Data: []byte("build:")},
	}

	c := NewCreator(fsys, &bytes.Buffer{})
	langs, err := c.ListLangs()
	if err != nil {
		t.Fatalf("ListLangs() error = %v", err)
	}

	if len(langs) != 2 {
		t.Fatalf("ListLangs() returned %d languages, want 2", len(langs))
	}

	// fstest.MapFS returns entries sorted alphabetically
	want := []string{"cpp", "go"}
	for i, got := range langs {
		if got != want[i] {
			t.Errorf("ListLangs()[%d] = %q, want %q", i, got, want[i])
		}
	}
}

func TestCheckDestDir(t *testing.T) {
	c := NewCreator(fstest.MapFS{}, &bytes.Buffer{})

	t.Run("missing destination is allowed", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "newproj")
		if err := c.checkDestDir(Options{ProjectName: "demo", TargetDir: dest}); err != nil {
			t.Fatalf("checkDestDir() error = %v", err)
		}
	})

	t.Run("existing file is rejected", func(t *testing.T) {
		tmp := t.TempDir()
		dest := filepath.Join(tmp, "existing")
		if err := os.WriteFile(dest, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}

		err := c.checkDestDir(Options{ProjectName: "demo", TargetDir: dest})
		if err == nil {
			t.Fatal("checkDestDir() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "not a directory") {
			t.Fatalf("checkDestDir() error = %v, want contains %q", err, "not a directory")
		}
	})

	t.Run("existing empty directory without allow is rejected", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "existing")
		if err := os.Mkdir(dest, 0755); err != nil {
			t.Fatal(err)
		}

		err := c.checkDestDir(Options{ProjectName: "demo", TargetDir: dest})
		if err == nil {
			t.Fatal("checkDestDir() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("checkDestDir() error = %v, want contains %q", err, "already exists")
		}
	})

	t.Run("existing empty directory can be reused when allowed", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "existing")
		if err := os.Mkdir(dest, 0755); err != nil {
			t.Fatal(err)
		}

		if err := c.checkDestDir(Options{ProjectName: "demo", TargetDir: dest, AllowExistingEmptyDir: true}); err != nil {
			t.Fatalf("checkDestDir() error = %v", err)
		}
	})

	t.Run("existing non-empty directory is rejected when only empty reuse is allowed", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "existing")
		if err := os.Mkdir(dest, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dest, "old.txt"), []byte("old"), 0644); err != nil {
			t.Fatal(err)
		}

		err := c.checkDestDir(Options{ProjectName: "demo", TargetDir: dest, AllowExistingEmptyDir: true})
		if err == nil {
			t.Fatal("checkDestDir() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "not empty") {
			t.Fatalf("checkDestDir() error = %v, want contains %q", err, "not empty")
		}
	})

	t.Run("existing directory with force is removed", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "existing")
		if err := os.MkdirAll(filepath.Join(dest, "nested"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dest, "nested", "old.txt"), []byte("old"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := c.checkDestDir(Options{ProjectName: "demo", TargetDir: dest, Force: true}); err != nil {
			t.Fatalf("checkDestDir() error = %v", err)
		}

		if _, err := os.Stat(dest); !os.IsNotExist(err) {
			t.Fatalf("destination should be removed, stat err = %v", err)
		}
	})

	t.Run("force refuses to remove current directory", func(t *testing.T) {
		cwd := withTempWorkingDir(t)
		if err := os.WriteFile(filepath.Join(cwd, "old.txt"), []byte("old"), 0644); err != nil {
			t.Fatal(err)
		}

		err := c.checkDestDir(Options{ProjectName: "demo", TargetDir: ".", Force: true})
		if err == nil {
			t.Fatal("checkDestDir() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "refusing to remove current directory") {
			t.Fatalf("checkDestDir() error = %v, want contains %q", err, "refusing to remove current directory")
		}
	})
}
