package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/JackDrogon/project/pkg/scaffold"
)

func withTempWorkingDir(t *testing.T, baseName string) string {
	t.Helper()

	parent := t.TempDir()
	dir := filepath.Join(parent, baseName)
	if err := os.MkdirAll(dir, 0755); err != nil {
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

func TestInitCmd_DefaultsToCurrentDirectory(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "demo")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newInitCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--no-git"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(workDir, "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(got), `const Name = "demo"`) {
		t.Fatalf("main.go content = %q, want rendered project name", string(got))
	}
}

func TestInitCmd_UsesExplicitTargetDirectory(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newInitCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--no-git", filepath.Join("nested", "demo")})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(workDir, "nested", "demo", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(got), `const Name = "demo"`) {
		t.Fatalf("main.go content = %q, want rendered project name", string(got))
	}
}

func TestInitCmd_RejectsNonEmptyDirectory(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	targetDir := filepath.Join(workDir, "demo")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", targetDir, err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "old.txt"), []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile(old.txt) error = %v", err)
	}

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newInitCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--no-git", "demo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("Execute() error = %v, want contains %q", err, "not empty")
	}
}

func TestInitCmd_RequiresLang(t *testing.T) {
	creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newInitCmd(creator)
	cmd.SetArgs([]string{"--git", "none"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "required flag") {
		t.Fatalf("Execute() error = %v, want required flag error", err)
	}
}

func TestInitCmd_RejectsTooManyArgs(t *testing.T) {
	creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newInitCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "one", "two"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "accepts between 0 and 1 arg") {
		t.Fatalf("Execute() error = %v, want arg count error", err)
	}
}

func TestInitCmd_PropagatesGitOptionConflict(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	withTempWorkingDir(t, "workspace")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newInitCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--no-git", "--git", "init-only"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "conflicting git options") {
		t.Fatalf("Execute() error = %v, want git conflict error", err)
	}
}

func TestInitCmd_DryRunRejectsNonEmptyDirectory(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	targetDir := filepath.Join(workDir, "demo")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", targetDir, err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "old.txt"), []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile(old.txt) error = %v", err)
	}

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newInitCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "--dry-run", "demo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("Execute() error = %v, want contains %q", err, "not empty")
	}
	if _, err := os.Stat(filepath.Join(targetDir, "main.go")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create files, stat err = %v", err)
	}
}

func TestProjectNameFromTargetDirFailsWhenCWDIsMissing(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir(%q) error = %v", tmp, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
	if err := os.RemoveAll(tmp); err != nil {
		t.Fatalf("RemoveAll(%q) error = %v", tmp, err)
	}

	_, err = projectNameFromTargetDir(".")
	if err == nil {
		t.Fatal("projectNameFromTargetDir() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to resolve target directory") {
		t.Fatalf("projectNameFromTargetDir() error = %v, want resolution error", err)
	}
}

func TestInitCmd_ReportsProjectNameResolutionError(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir(%q) error = %v", tmp, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
	if err := os.RemoveAll(tmp); err != nil {
		t.Fatalf("RemoveAll(%q) error = %v", tmp, err)
	}

	creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newInitCmd(creator)
	cmd.SetArgs([]string{"--lang", "go"})

	err = cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to resolve target directory") {
		t.Fatalf("Execute() error = %v, want resolution error", err)
	}
}
