package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
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

func TestInitCmd_DryRunUsesEnhancedPlanOutput(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.json":        {Data: []byte(`{"schema_version":1,"name":"go","description":"Go starter","inputs":[{"name":"module_path","template_var":"ModulePath"},{"name":"go_version","template_var":"GoVersion"},{"name":"author","template_var":"Author"},{"name":"year","template_var":"Year"}]}`)},
		"go/README.md":                              {Data: []byte("# README\n")},
		"go/cmd":                                    {Mode: os.ModeDir},
		"go/cmd/{{.ProjectNameLower}}":              {Mode: os.ModeDir},
		"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	var out bytes.Buffer

	creator := scaffold.NewCreator(fsys, &out)
	cmd := newInitCmd(creator)
	targetDir := filepath.Join("nested", "demo")
	cmd.SetArgs([]string{"--lang", "go", "--dry-run", "--git", "none", "--module", "example.com/demo", targetDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := scaffold.NewTemplateVars("demo", "example.com/demo")
	got := out.String()
	requireOrderedSubstrings(t, got, []string{
		"Creating project with language: go, project name: demo\n",
		"Dry-run mode: no files will be created\n",
		"template: go\n",
		"description: Go starter\n",
		"target_dir: " + targetDir + "\n",
		"resolved inputs:\n",
		"  project_name: demo\n",
		"  module_path: example.com/demo\n",
		"  go_version: " + vars.GoVersion + "\n",
		"  author: " + vars.Author + "\n",
		"  year: " + strconv.Itoa(vars.Year) + "\n",
		"  git_mode: none\n",
		"explicit overrides:\n",
		"  module_path: example.com/demo\n",
		"  git_mode: none\n",
		"actions:\n",
		"  copy go/README.md -> " + filepath.Join(targetDir, "README.md") + "\n",
		"  create " + filepath.Join(targetDir, "cmd") + string(filepath.Separator) + "\n",
		"  create " + filepath.Join(targetDir, "cmd", "demo") + string(filepath.Separator) + "\n",
		"  render go/cmd/{{.ProjectNameLower}}/main.go.tmpl -> " + filepath.Join(targetDir, "cmd", "demo", "main.go") + "\n",
		"  render go/go.mod.tmpl -> " + filepath.Join(targetDir, "go.mod") + "\n",
	})

	if _, err := os.Stat(filepath.Join(workDir, targetDir)); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create target directory, stat err = %v", err)
	}
}

func TestInitCmd_RejectsWriteReplayWithDryRun(t *testing.T) {
	creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newInitCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--dry-run", "--write-replay", filepath.Join(t.TempDir(), "replay.json")})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--write-replay cannot be combined with --dry-run") {
		t.Fatalf("Execute() error = %v, want dry-run conflict error", err)
	}
}

func TestInitCmd_ReplayRejectsMismatchedCommand(t *testing.T) {
	workDir := withTempWorkingDir(t, "workspace")
	replayPath := writeReplayFileForTest(t, scaffold.ReplayFile{
		Command: scaffold.ReplayCommandNew,
		Lang:    "go",
		Create: scaffold.ReplayFileCreate{
			ProjectName: "replayed-demo",
			TargetDir:   "replayed-demo",
			GitMode:     scaffold.GitModeNone,
		},
		TemplateInputs: map[string]string{"module_path": "example.com/replayed-demo"},
	})

	creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newInitCmd(creator)
	cmd.SetArgs([]string{"--replay", replayPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected replay command mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --replay") {
		t.Fatalf("Execute() error = %v, want replay validation prefix", err)
	}
	if !strings.Contains(err.Error(), `command "new"`) || !strings.Contains(err.Error(), `match "init"`) {
		t.Fatalf("Execute() error = %v, want mismatch details", err)
	}
	if _, statErr := os.Stat(filepath.Join(workDir, "replayed-demo")); !os.IsNotExist(statErr) {
		t.Fatalf("replay mismatch should fail before project creation, stat err = %v", statErr)
	}
}

func TestInitCmd_WriteReplayRecordsResolvedInputs(t *testing.T) {
	fsys := fstest.MapFS{
		"cpp/.project-template-manifest.json": {Data: []byte(`{"schema_version":1,"name":"cpp","description":"C++ starter","inputs":[{"name":"author","template_var":"Author"}]}`)},
		"cpp/README.md.tmpl":                  {Data: []byte("By {{.Author}}\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	replayPath := filepath.Join(t.TempDir(), "replay.json")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newInitCmd(creator)
	cmd.SetArgs([]string{"--lang", "cpp", "--git", "none", "--set", "author=alice", "--write-replay", replayPath, "demo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	replay, err := scaffold.ReadReplayFile(replayPath)
	if err != nil {
		t.Fatalf("ReadReplayFile(%q) error = %v", replayPath, err)
	}

	if replay.Command != scaffold.ReplayCommandInit {
		t.Fatalf("replay.Command = %q, want %q", replay.Command, scaffold.ReplayCommandInit)
	}
	if replay.Lang != "cpp" {
		t.Fatalf("replay.Lang = %q, want %q", replay.Lang, "cpp")
	}
	if replay.Create.ProjectName != "demo" {
		t.Fatalf("replay.Create.ProjectName = %q, want %q", replay.Create.ProjectName, "demo")
	}
	if replay.Create.TargetDir != "demo" {
		t.Fatalf("replay.Create.TargetDir = %q, want %q", replay.Create.TargetDir, "demo")
	}
	if replay.Create.GitMode != scaffold.GitModeNone {
		t.Fatalf("replay.Create.GitMode = %q, want %q", replay.Create.GitMode, scaffold.GitModeNone)
	}
	if got := replay.TemplateInputs["author"]; got != "alice" {
		t.Fatalf("replay.TemplateInputs[author] = %q, want %q", got, "alice")
	}
	if len(replay.TemplateInputs) != 1 {
		t.Fatalf("len(replay.TemplateInputs) = %d, want %d", len(replay.TemplateInputs), 1)
	}

	readme, err := os.ReadFile(filepath.Join(workDir, "demo", "README.md"))
	if err != nil {
		t.Fatalf("ReadFile(README.md) error = %v", err)
	}
	if string(readme) != "By alice\n" {
		t.Fatalf("README.md = %q, want %q", string(readme), "By alice\n")
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
