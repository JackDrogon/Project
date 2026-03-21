package scaffold

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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

func TestCreate_SkipsReservedManifestFile(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.json": {Data: []byte(`{"schema_version":1,"name":"go","description":"Production-ready Go CLI starter","inputs":[{"name":"module_path","template_var":"ModulePath"}]}`)},
		"go/main.go.tmpl":                    {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, nil)
	tmp := withTempWorkingDir(t)

	if err := creator.Create(Options{Lang: "go", ProjectName: "demo", NoGit: true}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "demo", ".project-template-manifest.json")); !os.IsNotExist(err) {
		t.Fatalf("reserved manifest should be skipped, stat err = %v", err)
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

func TestCreate_DryRunOutputsResolvedExecutionPlan(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.json":        {Data: []byte(`{"schema_version":1,"name":"go","description":"Go starter","inputs":[{"name":"module_path","template_var":"ModulePath"},{"name":"go_version","template_var":"GoVersion"},{"name":"author","template_var":"Author"},{"name":"year","template_var":"Year"}]}`)},
		"go/README.md":                              {Data: []byte("# README\n")},
		"go/cmd":                                    {Mode: os.ModeDir},
		"go/cmd/{{.ProjectNameLower}}":              {Mode: os.ModeDir},
		"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
	}
	var out bytes.Buffer

	creator := NewCreatorWithGitRunner(fsys, &out, func(dir string, args ...string) error {
		t.Fatal("git runner should not be called during dry-run")
		return nil
	})

	withTempWorkingDir(t)
	opts := Options{Lang: "go", ProjectName: "Demo", ModulePath: "example.com/demo", DryRun: true, GitMode: GitModeNone}
	vars, err := creator.templateVars(opts)
	if err != nil {
		t.Fatalf("templateVars() error = %v", err)
	}

	if err := creator.Create(opts); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got := out.String()
	requireOrderedSubstrings(t, got, []string{
		"Creating project with language: go, project name: Demo\n",
		"Dry-run mode: no files will be created\n",
		"template: go\n",
		"description: Go starter\n",
		"target_dir: Demo\n",
		"resolved inputs:\n",
		"  project_name: Demo\n",
		"  module_path: example.com/demo\n",
		"  go_version: " + vars.GoVersion + "\n",
		"  author: " + vars.Author + "\n",
		"  year: " + strconv.Itoa(vars.Year) + "\n",
		"  git_mode: none\n",
		"explicit overrides:\n",
		"  module_path: example.com/demo\n",
		"  git_mode: none\n",
		"actions:\n",
		"  copy go/README.md -> " + filepath.Join("Demo", "README.md") + "\n",
		"  create " + filepath.Join("Demo", "cmd") + string(filepath.Separator) + "\n",
		"  create " + filepath.Join("Demo", "cmd", "demo") + string(filepath.Separator) + "\n",
		"  render go/cmd/{{.ProjectNameLower}}/main.go.tmpl -> " + filepath.Join("Demo", "cmd", "demo", "main.go") + "\n",
		"  render go/go.mod.tmpl -> " + filepath.Join("Demo", "go.mod") + "\n",
	})

	if strings.Contains(got, templateManifestFilename) {
		t.Fatalf("output = %q, want reserved manifest omitted from dry-run actions", got)
	}
	if _, err := os.Stat(filepath.Join("Demo", "README.md")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create files, stat err = %v", err)
	}
}

func TestCreate_DryRunPlanStillFailsAfterPrintingHeader(t *testing.T) {
	fsys := fstest.MapFS{
		"go/bad.txt.tmpl": {Data: []byte("{{.ProjectName")},
	}
	var out bytes.Buffer

	creator := NewCreatorWithGitRunner(fsys, &out, func(dir string, args ...string) error {
		t.Fatal("git runner should not be called during dry-run")
		return nil
	})

	withTempWorkingDir(t)
	err := creator.Create(Options{Lang: "go", ProjectName: "demo", DryRun: true})
	if err == nil {
		t.Fatal("Create() expected dry-run plan error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to render template") {
		t.Fatalf("Create() error = %v, want rendered template failure", err)
	}

	got := out.String()
	requireOrderedSubstrings(t, got, []string{
		"Creating project with language: go, project name: demo\n",
		"Dry-run mode: no files will be created\n",
	})
	if strings.Contains(got, "actions:\n") {
		t.Fatalf("output = %q, want failure before action plan is rendered", got)
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

func TestCreate_AppliesTemplateInputOverrides(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.json": {Data: []byte(`{"schema_version":1,"name":"go","description":"Go starter","inputs":[{"name":"module_path","template_var":"ModulePath"},{"name":"go_version","template_var":"GoVersion"},{"name":"author","template_var":"Author"},{"name":"year","template_var":"Year"}]}`)},
		"go/README.md.tmpl":                  {Data: []byte("module={{.ModulePath}}\ngo={{.GoVersion}}\nauthor={{.Author}}\nyear={{.Year}}\n")},
	}
	workDir := withTempWorkingDir(t)

	creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
		return nil
	})

	err := creator.Create(Options{
		Lang:        "go",
		ProjectName: "demo",
		ModulePath:  "example.com/demo",
		TemplateInputValues: map[string]string{
			"go_version": "1.25",
			"author":     "alice",
			"year":       "2030",
		},
		NoGit: true,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(workDir, "demo", "README.md"))
	if err != nil {
		t.Fatalf("ReadFile(README.md) error = %v", err)
	}

	want := "module=example.com/demo\ngo=1.25\nauthor=alice\nyear=2030\n"
	if string(got) != want {
		t.Fatalf("README.md = %q, want %q", string(got), want)
	}
}

func TestCreate_RejectsUnknownOrInvalidTemplateInputs(t *testing.T) {
	tests := []struct {
		name                string
		projectName         string
		templateInputValues map[string]string
		wantErr             string
	}{
		{
			name:                "unknown input",
			projectName:         "unknown-demo",
			templateInputValues: map[string]string{"author": "alice"},
			wantErr:             `template input "author" is not declared by template "go"`,
		},
		{
			name:                "invalid year",
			projectName:         "invalid-year-demo",
			templateInputValues: map[string]string{"year": "twenty"},
			wantErr:             `template input "year" must be a valid year`,
		},
		{
			name:                "module path must stay first class",
			projectName:         "module-path-demo",
			templateInputValues: map[string]string{"module_path": "example.com/demo"},
			wantErr:             `template input "module_path" must be provided via module path options`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := withTempWorkingDir(t)
			fsys := fstest.MapFS{
				"go/.project-template-manifest.json": {Data: []byte(`{"schema_version":1,"name":"go","description":"Go starter","inputs":[{"name":"module_path","template_var":"ModulePath"},{"name":"go_version","template_var":"GoVersion"},{"name":"year","template_var":"Year"}]}`)},
				"go/README.md.tmpl":                  {Data: []byte("# {{.ProjectName}}\n")},
			}

			creator := NewCreatorWithGitRunner(fsys, &bytes.Buffer{}, func(dir string, args ...string) error {
				return nil
			})

			err := creator.Create(Options{
				Lang:                "go",
				ProjectName:         tt.projectName,
				TemplateInputValues: tt.templateInputValues,
				NoGit:               true,
			})
			if err == nil {
				t.Fatal("Create() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Create() error = %v, want contains %q", err, tt.wantErr)
			}

			if _, statErr := os.Stat(filepath.Join(workDir, tt.projectName)); !os.IsNotExist(statErr) {
				t.Fatalf("project directory should not be created for invalid template inputs, stat err = %v", statErr)
			}
		})
	}
}
