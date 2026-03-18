package scaffold

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

type failingReadDirFS struct {
	err error
}

func (f failingReadDirFS) Open(name string) (fs.File, error) {
	return nil, f.err
}

func (f failingReadDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, f.err
}

func stubScaffoldOSFuncs(t *testing.T) {
	t.Helper()
	oldStat := osStat
	oldReadDir := osReadDir
	oldRemoveAll := osRemoveAll
	oldGetwd := osGetwd
	oldAbs := filepathAbs
	t.Cleanup(func() {
		osStat = oldStat
		osReadDir = oldReadDir
		osRemoveAll = oldRemoveAll
		osGetwd = oldGetwd
		filepathAbs = oldAbs
	})
}

func TestNewCreatorWithGitRunner_NilUsesDefault(t *testing.T) {
	creator := NewCreatorWithGitRunner(fstest.MapFS{}, &bytes.Buffer{}, nil)
	if creator.runGit == nil {
		t.Fatal("runGit should default to git.Run when nil")
	}
}

func TestCreate_UnsupportedLanguageReturnsError(t *testing.T) {
	creator := NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	err := creator.Create(Options{Lang: "missing", ProjectName: "demo"})
	if err == nil {
		t.Fatal("Create() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported language") {
		t.Fatalf("Create() error = %v, want unsupported language error", err)
	}
}

func TestCreate_DryRunRejectsNonEmptyDirectory(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	withTempWorkingDir(t)
	if err := os.Mkdir("demo", 0755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join("demo", "old.txt"), []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, nil)
	err := creator.Create(Options{Lang: "go", ProjectName: "demo", TargetDir: "demo", AllowExistingEmptyDir: true, DryRun: true})
	if err == nil {
		t.Fatal("Create() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("Create() error = %v, want non-empty directory error", err)
	}
}

func TestCreate_DryRunForceAllowsPreviewWithoutDeletion(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	var out bytes.Buffer
	withTempWorkingDir(t)
	if err := os.Mkdir("demo", 0755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join("demo", "old.txt"), []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	creator := NewCreatorWithGitRunner(fsys, &out, nil)
	if err := creator.Create(Options{Lang: "go", ProjectName: "demo", TargetDir: "demo", Force: true, DryRun: true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join("demo", "old.txt")); err != nil {
		t.Fatalf("old.txt should remain after dry-run, stat error = %v", err)
	}
	if !strings.Contains(out.String(), "Dry-run mode") {
		t.Fatalf("output = %q, want dry-run message", out.String())
	}
}

func TestCreate_GitInitOnlyErrorIsReturned(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	withTempWorkingDir(t)

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		return errors.New("git init failed")
	})
	err := creator.Create(Options{Lang: "go", ProjectName: "demo", GitMode: GitModeInitOnly})
	if err == nil {
		t.Fatal("Create() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "git init failed") {
		t.Fatalf("Create() error = %v, want init failure", err)
	}
}

func TestMaybeInitGitRepo_InvalidModeReturnsError(t *testing.T) {
	creator := NewCreatorWithGitRunner(fstest.MapFS{}, &bytes.Buffer{}, nil)
	err := creator.maybeInitGitRepo(Options{ProjectName: "demo", GitMode: GitMode("invalid")})
	if err == nil {
		t.Fatal("maybeInitGitRepo() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid git mode") {
		t.Fatalf("maybeInitGitRepo() error = %v, want invalid git mode error", err)
	}
}

func TestListLangs_ReadDirError(t *testing.T) {
	creator := NewCreator(failingReadDirFS{err: errors.New("read dir failed")}, &bytes.Buffer{})
	_, err := creator.ListLangs()
	if err == nil {
		t.Fatal("ListLangs() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read templates") {
		t.Fatalf("ListLangs() error = %v, want wrapped read error", err)
	}
}

func TestWalkTemplateFilesErrors(t *testing.T) {
	creator := NewCreator(fstest.MapFS{"go/main.go.tmpl": {Data: []byte("package main")}}, &bytes.Buffer{})

	err := creator.walkTemplateFiles("missing", func(srcPath string, isTemplate bool) error {
		return nil
	})
	if err == nil {
		t.Fatal("walkTemplateFiles() expected error for missing directory, got nil")
	}

	err = creator.walkTemplateFiles("go", func(srcPath string, isTemplate bool) error {
		return errors.New("visit failed")
	})
	if err == nil {
		t.Fatal("walkTemplateFiles() expected visitor error, got nil")
	}
	if !strings.Contains(err.Error(), "visit failed") {
		t.Fatalf("walkTemplateFiles() error = %v, want visitor error", err)
	}
}

func TestIsEmptyDirErrorOnFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := isEmptyDir(file); err == nil {
		t.Fatal("isEmptyDir() expected error for file path, got nil")
	}
}

func TestIsCurrentDirFailsWhenWorkingDirMissing(t *testing.T) {
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

	_, err = isCurrentDir(".")
	if err == nil {
		t.Fatal("isCurrentDir() expected error, got nil")
	}
}

func TestIsCurrentDirFailsWhenGetwdFails(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	tmp := t.TempDir()
	absTarget := filepath.Join(tmp, "target")
	if err := os.MkdirAll(absTarget, 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", absTarget, err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir(%q) error = %v", tmp, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
	if err := os.RemoveAll(tmp); err != nil {
		t.Fatalf("RemoveAll(%q) error = %v", tmp, err)
	}

	_, err = isCurrentDir(absTarget)
	if err == nil {
		t.Fatal("isCurrentDir() expected error, got nil")
	}
}

func TestInspectDestDir_CurrentDirCheckError(t *testing.T) {
	c := NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	parent := t.TempDir()
	brokenCWD := filepath.Join(parent, "cwd")
	target := filepath.Join(parent, "target")
	if err := os.MkdirAll(brokenCWD, 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", brokenCWD, err)
	}
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", target, err)
	}
	if err := os.Chdir(brokenCWD); err != nil {
		t.Fatalf("Chdir(%q) error = %v", brokenCWD, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
	if err := os.RemoveAll(brokenCWD); err != nil {
		t.Fatalf("RemoveAll(%q) error = %v", brokenCWD, err)
	}

	err = c.inspectDestDir(Options{ProjectName: "demo", TargetDir: target, Force: true}, false)
	if err == nil {
		t.Fatal("inspectDestDir() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to inspect destination") {
		t.Fatalf("inspectDestDir() error = %v, want wrapped inspection error", err)
	}
}

func TestInspectDestDir_StatError(t *testing.T) {
	stubScaffoldOSFuncs(t)
	c := NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	osStat = func(name string) (os.FileInfo, error) {
		return nil, errors.New("stat failed")
	}

	err := c.inspectDestDir(Options{ProjectName: "demo", TargetDir: "demo"}, false)
	if err == nil {
		t.Fatal("inspectDestDir() expected stat error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to inspect destination") {
		t.Fatalf("inspectDestDir() error = %v, want wrapped stat error", err)
	}
}

func TestInspectDestDir_ReadDirError(t *testing.T) {
	stubScaffoldOSFuncs(t)
	c := NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	osStat = func(name string) (os.FileInfo, error) {
		return stubFileInfo{name: "demo", mode: os.ModeDir | 0755}, nil
	}
	osReadDir = func(name string) ([]os.DirEntry, error) {
		return nil, errors.New("read dir failed")
	}

	err := c.inspectDestDir(Options{ProjectName: "demo", TargetDir: "demo"}, false)
	if err == nil {
		t.Fatal("inspectDestDir() expected read dir error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to inspect destination") {
		t.Fatalf("inspectDestDir() error = %v, want wrapped read dir error", err)
	}
}

func TestInspectDestDir_RemoveAllError(t *testing.T) {
	stubScaffoldOSFuncs(t)
	c := NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	osStat = func(name string) (os.FileInfo, error) {
		return stubFileInfo{name: "demo", mode: os.ModeDir | 0755}, nil
	}
	osRemoveAll = func(path string) error {
		return errors.New("remove failed")
	}

	err := c.inspectDestDir(Options{ProjectName: "demo", TargetDir: "demo", Force: true}, false)
	if err == nil {
		t.Fatal("inspectDestDir() expected remove error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to remove existing directory") {
		t.Fatalf("inspectDestDir() error = %v, want wrapped remove error", err)
	}
}
