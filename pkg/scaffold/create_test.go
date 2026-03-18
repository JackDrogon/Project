package scaffold

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func withTempWorkingDir(t *testing.T) string {
	t.Helper()

	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir(%q) error = %v", tmp, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	return tmp
}

func TestCreate_NoGitSkipsGit(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	var out bytes.Buffer
	gitCalled := false

	creator := NewCreatorWithGitRunner(fsys, &out, func(dir string, args ...string) error {
		gitCalled = true
		return nil
	})

	tmp := withTempWorkingDir(t)
	if err := creator.Create(Options{Lang: "go", ProjectName: "demo", NoGit: true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if gitCalled {
		t.Fatal("git runner should not be called when NoGit is true")
	}

	got, err := os.ReadFile(filepath.Join(tmp, "demo", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(got), `const Name = "demo"`) {
		t.Fatalf("main.go content = %q, want rendered project name", string(got))
	}
}

func TestCreate_GitCommandsInOrder(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	tmp := withTempWorkingDir(t)

	type gitCall struct {
		dir  string
		args []string
	}
	var calls []gitCall

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		calls = append(calls, gitCall{dir: dir, args: append([]string(nil), args...)})
		return nil
	})

	if err := creator.Create(Options{Lang: "go", ProjectName: "demo"}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	want := [][]string{{"init"}, {"add", "."}, {"commit", "-m", "Initial commit"}}
	if len(calls) != len(want) {
		t.Fatalf("git call count = %d, want %d", len(calls), len(want))
	}

	for i := range calls {
		if calls[i].dir != "demo" {
			t.Fatalf("git call[%d] dir = %q, want %q", i, calls[i].dir, "demo")
		}
		if !reflect.DeepEqual(calls[i].args, want[i]) {
			t.Fatalf("git call[%d] args = %v, want %v", i, calls[i].args, want[i])
		}
	}

	if _, err := os.Stat(filepath.Join(tmp, "demo", "main.go")); err != nil {
		t.Fatalf("generated file missing: %v", err)
	}
}

func TestCreate_SignoffUsesSignedCommit(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	withTempWorkingDir(t)

	var commitArgs []string
	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		if len(args) > 0 && args[0] == "commit" {
			commitArgs = append([]string(nil), args...)
		}
		return nil
	})

	if err := creator.Create(Options{Lang: "go", ProjectName: "demo", Signoff: true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	want := []string{"commit", "-s", "-m", "Initial commit"}
	if !reflect.DeepEqual(commitArgs, want) {
		t.Fatalf("commit args = %v, want %v", commitArgs, want)
	}
}

func TestCreate_TargetDirControlsOutputAndTemplateVars(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	var out bytes.Buffer

	creator := NewCreatorWithGitRunner(fsys, &out, func(dir string, args ...string) error {
		return nil
	})

	tmp := withTempWorkingDir(t)
	targetDir := filepath.Join("workspace", "output")
	if err := creator.Create(Options{Lang: "go", ProjectName: "logical-name", TargetDir: targetDir, NoGit: true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(tmp, targetDir, "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(got), `const Name = "logical-name"`) {
		t.Fatalf("main.go content = %q, want rendered logical project name", string(got))
	}
}

func TestCreate_TargetDirIsUsedForGitOperations(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	withTempWorkingDir(t)

	var dirs []string
	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		dirs = append(dirs, dir)
		return nil
	})

	targetDir := filepath.Join("workspace", "output")
	if err := creator.Create(Options{Lang: "go", ProjectName: "logical-name", TargetDir: targetDir}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if len(dirs) != 3 {
		t.Fatalf("git dir count = %d, want %d", len(dirs), 3)
	}
	for i, dir := range dirs {
		if dir != targetDir {
			t.Fatalf("git dir[%d] = %q, want %q", i, dir, targetDir)
		}
	}
}

func TestCreate_DryRunSkipsWritesAndGit(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	var out bytes.Buffer
	gitCalled := false

	creator := NewCreatorWithGitRunner(fsys, &out, func(dir string, args ...string) error {
		gitCalled = true
		return nil
	})

	tmp := withTempWorkingDir(t)
	if err := creator.Create(Options{Lang: "go", ProjectName: "demo", DryRun: true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if gitCalled {
		t.Fatal("git runner should not be called during dry-run")
	}

	if _, err := os.Stat(filepath.Join(tmp, "demo")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create destination directory, stat err = %v", err)
	}

	if !strings.Contains(out.String(), "Dry-run mode") {
		t.Fatalf("output = %q, want dry-run message", out.String())
	}
}

func TestCreate_GitErrorIsReturned(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	withTempWorkingDir(t)

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		if len(args) > 0 && args[0] == "commit" {
			return errors.New("git commit failed")
		}
		return nil
	})

	err := creator.Create(Options{Lang: "go", ProjectName: "demo"})
	if err == nil {
		t.Fatal("Create() expected git error, got nil")
	}
	if !strings.Contains(err.Error(), "git commit failed") {
		t.Fatalf("Create() error = %v, want commit failure", err)
	}
}

func TestCreate_GitModeInitOnlyRunsInitOnly(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	withTempWorkingDir(t)

	var calls [][]string
	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		calls = append(calls, append([]string(nil), args...))
		return nil
	})

	err := creator.Create(Options{Lang: "go", ProjectName: "demo", GitMode: GitModeInitOnly})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("git call count = %d, want %d", len(calls), 1)
	}
	if !reflect.DeepEqual(calls[0], []string{"init"}) {
		t.Fatalf("git call = %v, want %v", calls[0], []string{"init"})
	}
}

func TestCreate_GitModeNoneSkipsGit(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	withTempWorkingDir(t)

	called := false
	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		called = true
		return nil
	})

	err := creator.Create(Options{Lang: "go", ProjectName: "demo", GitMode: GitModeNone})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if called {
		t.Fatal("git runner should not be called for GitModeNone")
	}
}

func TestCreate_InvalidModulePathReturnsError(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	withTempWorkingDir(t)

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		return nil
	})

	err := creator.Create(Options{Lang: "go", ProjectName: "demo", ModulePath: "https://example.com/demo"})
	if err == nil {
		t.Fatal("Create() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "module path") {
		t.Fatalf("Create() error = %v, want module path validation error", err)
	}
	if _, statErr := os.Stat("demo"); !os.IsNotExist(statErr) {
		t.Fatalf("invalid module path should not create project directory, stat err = %v", statErr)
	}
}

func TestCreate_DefaultGoModulePathIsValidated(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	withTempWorkingDir(t)

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		return nil
	})

	err := creator.Create(Options{Lang: "go", ProjectName: "demo.", NoGit: true})
	if err == nil {
		t.Fatal("Create() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "module path") {
		t.Fatalf("Create() error = %v, want module path validation error", err)
	}
	if _, statErr := os.Stat("demo."); !os.IsNotExist(statErr) {
		t.Fatalf("invalid default module path should not create project directory, stat err = %v", statErr)
	}
}

func TestValidateModulePath_SkipsNonGoTemplates(t *testing.T) {
	creator := NewCreatorWithGitRunner(fstest.MapFS{}, &bytes.Buffer{}, nil)
	err := creator.validateModulePath(Options{Lang: "cpp", ModulePath: "https://example.com/demo"})
	if err != nil {
		t.Fatalf("validateModulePath() error = %v, want nil for non-go template", err)
	}
}

func TestCreate_InvalidGitModeReturnsError(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	tmp := withTempWorkingDir(t)

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		return nil
	})

	err := creator.Create(Options{Lang: "go", ProjectName: "demo", GitMode: GitMode("invalid")})
	if err == nil {
		t.Fatal("Create() expected invalid git mode error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid git mode") {
		t.Fatalf("Create() error = %v, want invalid git mode message", err)
	}

	if _, statErr := os.Stat(filepath.Join(tmp, "demo")); !os.IsNotExist(statErr) {
		t.Fatalf("demo directory should not be created for invalid options, stat err = %v", statErr)
	}
}

func TestCreate_ConflictingNoGitAndGitModeReturnsError(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	tmp := withTempWorkingDir(t)

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		return nil
	})

	err := creator.Create(Options{Lang: "go", ProjectName: "demo", NoGit: true, GitMode: GitModeInitOnly})
	if err == nil {
		t.Fatal("Create() expected conflict error, got nil")
	}
	if !strings.Contains(err.Error(), "conflicting git options") {
		t.Fatalf("Create() error = %v, want conflict message", err)
	}

	if _, statErr := os.Stat(filepath.Join(tmp, "demo")); !os.IsNotExist(statErr) {
		t.Fatalf("demo directory should not be created for invalid options, stat err = %v", statErr)
	}
}

func TestCreate_SignoffRequiresCommitMode(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	tmp := withTempWorkingDir(t)

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		return nil
	})

	err := creator.Create(Options{Lang: "go", ProjectName: "demo", GitMode: GitModeInitOnly, Signoff: true})
	if err == nil {
		t.Fatal("Create() expected signoff mode error, got nil")
	}
	if !strings.Contains(err.Error(), "--signoff") {
		t.Fatalf("Create() error = %v, want --signoff message", err)
	}

	if _, statErr := os.Stat(filepath.Join(tmp, "demo")); !os.IsNotExist(statErr) {
		t.Fatalf("demo directory should not be created for invalid options, stat err = %v", statErr)
	}
}
