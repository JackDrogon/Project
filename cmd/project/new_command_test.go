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

func writeReplayFileForTest(t *testing.T, replay scaffold.ReplayFile) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "replay.json")
	if err := scaffold.WriteReplayFile(path, replay); err != nil {
		t.Fatalf("WriteReplayFile(%q) error = %v", path, err)
	}

	return path
}

func TestNewCmd_RequiresLang(t *testing.T) {
	creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"demo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "required flag") {
		t.Fatalf("Execute() error = %v, want required flag error", err)
	}
}

func TestNewCmd_RejectsInvalidArgCount(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing project name", args: []string{"--lang", "go"}},
		{name: "too many args", args: []string{"--lang", "go", "one", "two"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
			cmd := newNewCmd(creator)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("Execute() expected error, got nil")
			}
			if !strings.Contains(err.Error(), "accepts 1 arg") {
				t.Fatalf("Execute() error = %v, want arg count error", err)
			}
		})
	}
}

func TestNewCmd_CreatesProjectFromArgument(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":  {Data: []byte("module {{.ModulePath}}\n")},
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "--module", "example.com/demo", "demo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(workDir, "demo", "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if string(goMod) != "module example.com/demo\n" {
		t.Fatalf("go.mod = %q, want %q", string(goMod), "module example.com/demo\n")
	}

	mainGo, err := os.ReadFile(filepath.Join(workDir, "demo", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(mainGo), `const Name = "demo"`) {
		t.Fatalf("main.go content = %q, want rendered project name", string(mainGo))
	}
}

func TestNewCmd_GoModuleArgumentDerivesProjectNameAndModulePath(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":  {Data: []byte("module {{.ModulePath}}\n")},
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "github.com/JackDrogon/agent-village"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(workDir, "agent-village", "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if string(goMod) != "module github.com/JackDrogon/agent-village\n" {
		t.Fatalf("go.mod = %q, want %q", string(goMod), "module github.com/JackDrogon/agent-village\n")
	}

	mainGo, err := os.ReadFile(filepath.Join(workDir, "agent-village", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(mainGo), `const Name = "agent-village"`) {
		t.Fatalf("main.go content = %q, want rendered project name", string(mainGo))
	}
}

func TestNewCmd_DryRunUsesEnhancedPlanOutput(t *testing.T) {
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
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--dry-run", "--git", "none", "--module", "example.com/demo", "demo"})

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
		"target_dir: demo\n",
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
		"  copy go/README.md -> " + filepath.Join("demo", "README.md") + "\n",
		"  create " + filepath.Join("demo", "cmd") + string(filepath.Separator) + "\n",
		"  create " + filepath.Join("demo", "cmd", "demo") + string(filepath.Separator) + "\n",
		"  render go/cmd/{{.ProjectNameLower}}/main.go.tmpl -> " + filepath.Join("demo", "cmd", "demo", "main.go") + "\n",
		"  render go/go.mod.tmpl -> " + filepath.Join("demo", "go.mod") + "\n",
	})

	if _, err := os.Stat(filepath.Join(workDir, "demo")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create destination directory, stat err = %v", err)
	}
}

func TestNewCmd_GoModuleArgumentWithMajorVersionSuffixUsesRepositoryName(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":  {Data: []byte("module {{.ModulePath}}\n")},
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "github.com/acme/agent-village/v2"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(workDir, "agent-village", "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if string(goMod) != "module github.com/acme/agent-village/v2\n" {
		t.Fatalf("go.mod = %q, want %q", string(goMod), "module github.com/acme/agent-village/v2\n")
	}

	mainGo, err := os.ReadFile(filepath.Join(workDir, "agent-village", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(mainGo), `const Name = "agent-village"`) {
		t.Fatalf("main.go content = %q, want rendered repository name", string(mainGo))
	}

	if _, err := os.Stat(filepath.Join(workDir, "v2")); !os.IsNotExist(err) {
		t.Fatalf("version suffix directory should not be created, stat err = %v", err)
	}
}

func TestNewCmd_InvalidDerivedProjectNameReturnsResolveError(t *testing.T) {
	creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "github.com/acme/9agent"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "project name") {
		t.Fatalf("Execute() error = %v, want project name validation error", err)
	}
}

func TestNewCmd_ReplayAllowsOmittedLangAndProjectArg(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":  {Data: []byte("module {{.ModulePath}}\n")},
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
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

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--replay", replayPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(workDir, "replayed-demo", "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if string(goMod) != "module example.com/replayed-demo\n" {
		t.Fatalf("go.mod = %q, want %q", string(goMod), "module example.com/replayed-demo\n")
	}

	mainGo, err := os.ReadFile(filepath.Join(workDir, "replayed-demo", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(mainGo), `const Name = "replayed-demo"`) {
		t.Fatalf("main.go content = %q, want rendered replay project name", string(mainGo))
	}
}

func TestNewCmd_ReplayRejectsMismatchedCommand(t *testing.T) {
	workDir := withTempWorkingDir(t, "workspace")
	replayPath := writeReplayFileForTest(t, scaffold.ReplayFile{
		Command: scaffold.ReplayCommandInit,
		Lang:    "go",
		Create: scaffold.ReplayFileCreate{
			ProjectName: "replayed-demo",
			TargetDir:   "replayed-demo",
			GitMode:     scaffold.GitModeNone,
		},
		TemplateInputs: map[string]string{"module_path": "example.com/replayed-demo"},
	})

	creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--replay", replayPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected replay command mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --replay") {
		t.Fatalf("Execute() error = %v, want replay validation prefix", err)
	}
	if !strings.Contains(err.Error(), `command "init"`) || !strings.Contains(err.Error(), `match "new"`) {
		t.Fatalf("Execute() error = %v, want mismatch details", err)
	}
	if _, statErr := os.Stat(filepath.Join(workDir, "replayed-demo")); !os.IsNotExist(statErr) {
		t.Fatalf("replay mismatch should fail before project creation, stat err = %v", statErr)
	}
}

func TestNewCmd_ReplayCLIFlagsAndPositionalArgTakePrecedence(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":     {Data: []byte("module {{.ModulePath}}\n")},
		"go/main.go.tmpl":    {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
		"cpp/main.cc.tmpl":   {Data: []byte("int main() { return 0; }\n")},
		"cpp/README.md.tmpl": {Data: []byte("# {{.ProjectName}}\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	replayPath := writeReplayFileForTest(t, scaffold.ReplayFile{
		Command: scaffold.ReplayCommandNew,
		Lang:    "cpp",
		Create: scaffold.ReplayFileCreate{
			ProjectName: "replay-name",
			TargetDir:   "replay-dir",
			GitMode:     scaffold.GitModeInitOnly,
			Signoff:     true,
			Force:       true,
		},
		TemplateInputs: map[string]string{"module_path": "example.com/from-replay"},
	})

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--replay", replayPath, "--lang", "go", "--module", "example.com/from-cli", "--git", "none", "--signoff=false", "cli-demo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(workDir, "cli-demo", "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile(go.mod) error = %v", err)
	}
	if string(goMod) != "module example.com/from-cli\n" {
		t.Fatalf("go.mod = %q, want %q", string(goMod), "module example.com/from-cli\n")
	}

	mainGo, err := os.ReadFile(filepath.Join(workDir, "cli-demo", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(mainGo), `const Name = "cli-demo"`) {
		t.Fatalf("main.go content = %q, want CLI project name", string(mainGo))
	}

	if _, err := os.Stat(filepath.Join(workDir, "replay-dir")); !os.IsNotExist(err) {
		t.Fatalf("replay target directory should not be created, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "cli-demo", "main.cc")); !os.IsNotExist(err) {
		t.Fatalf("cpp template should not be used, stat err = %v", err)
	}
}

func TestNewCmd_RejectsWriteReplayWithDryRun(t *testing.T) {
	creator := scaffold.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--dry-run", "--write-replay", filepath.Join(t.TempDir(), "replay.json"), "demo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--write-replay cannot be combined with --dry-run") {
		t.Fatalf("Execute() error = %v, want dry-run conflict error", err)
	}
}

func TestNewCmd_WriteReplayRecordsResolvedInputs(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.json": {Data: []byte(`{"schema_version":1,"name":"go","description":"Go starter","inputs":[{"name":"module_path","template_var":"ModulePath"},{"name":"go_version","template_var":"GoVersion"}]}`)},
		"go/go.mod.tmpl":                     {Data: []byte("module {{.ModulePath}}\ngo {{.GoVersion}}\n")},
	}
	withTempWorkingDir(t, "workspace")
	replayPath := filepath.Join(t.TempDir(), "replay.json")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "--module", "example.com/demo", "--set", "go_version=1.25", "--write-replay", replayPath, "demo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	replay, err := scaffold.ReadReplayFile(replayPath)
	if err != nil {
		t.Fatalf("ReadReplayFile(%q) error = %v", replayPath, err)
	}

	if replay.Command != scaffold.ReplayCommandNew {
		t.Fatalf("replay.Command = %q, want %q", replay.Command, scaffold.ReplayCommandNew)
	}
	if replay.Lang != "go" {
		t.Fatalf("replay.Lang = %q, want %q", replay.Lang, "go")
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
	if replay.Create.Signoff {
		t.Fatal("replay.Create.Signoff = true, want false")
	}
	if replay.Create.Force {
		t.Fatal("replay.Create.Force = true, want false")
	}
	if got := replay.TemplateInputs["module_path"]; got != "example.com/demo" {
		t.Fatalf("replay.TemplateInputs[module_path] = %q, want %q", got, "example.com/demo")
	}
	if got := replay.TemplateInputs["go_version"]; got != "1.25" {
		t.Fatalf("replay.TemplateInputs[go_version] = %q, want %q", got, "1.25")
	}
	if len(replay.TemplateInputs) != 2 {
		t.Fatalf("len(replay.TemplateInputs) = %d, want %d", len(replay.TemplateInputs), 2)
	}
}

func TestNewCmd_WriteReplayFailureReturnsWrappedError(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.json": {Data: []byte(`{"schema_version":1,"name":"go","description":"Go starter","inputs":[{"name":"module_path","template_var":"ModulePath"}]}`)},
		"go/main.go.tmpl":                    {Data: []byte("package main\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	replayPath := filepath.Join(t.TempDir(), "missing", "replay.json")

	creator := scaffold.NewCreator(fsys, &bytes.Buffer{})
	cmd := newNewCmd(creator)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "--write-replay", replayPath, "demo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to write resolved replay after project creation") {
		t.Fatalf("Execute() error = %v, want wrapped replay write error", err)
	}
	if !strings.Contains(err.Error(), "failed to write replay file") {
		t.Fatalf("Execute() error = %v, want underlying replay file error", err)
	}

	if _, statErr := os.Stat(filepath.Join(workDir, "demo", "main.go")); statErr != nil {
		t.Fatalf("generated project should remain on disk after replay write failure, stat err = %v", statErr)
	}
	if _, statErr := os.Stat(replayPath); !os.IsNotExist(statErr) {
		t.Fatalf("replay file should not be created on failure, stat err = %v", statErr)
	}
}

func TestResolveNewProjectArgs_GoModuleVersionHeuristic(t *testing.T) {
	tests := []struct {
		name            string
		arg             string
		wantProjectName string
		wantTargetDir   string
		wantModulePath  string
	}{
		{
			name:            "plain module path uses final segment",
			arg:             "github.com/acme/agent-village",
			wantProjectName: "agent-village",
			wantTargetDir:   "agent-village",
			wantModulePath:  "github.com/acme/agent-village",
		},
		{
			name:            "major version suffix uses repository segment",
			arg:             "github.com/acme/agent-village/v2",
			wantProjectName: "agent-village",
			wantTargetDir:   "agent-village",
			wantModulePath:  "github.com/acme/agent-village/v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectName, targetDir, modulePath, err := resolveNewProjectArgs("go", "", tt.arg)
			if err != nil {
				t.Fatalf("resolveNewProjectArgs() error = %v", err)
			}
			if projectName != tt.wantProjectName {
				t.Fatalf("projectName = %q, want %q", projectName, tt.wantProjectName)
			}
			if targetDir != tt.wantTargetDir {
				t.Fatalf("targetDir = %q, want %q", targetDir, tt.wantTargetDir)
			}
			if modulePath != tt.wantModulePath {
				t.Fatalf("modulePath = %q, want %q", modulePath, tt.wantModulePath)
			}
		})
	}
}

func TestResolveNewProjectArgs_FallbacksAndErrors(t *testing.T) {
	tests := []struct {
		name            string
		lang            string
		module          string
		arg             string
		wantProjectName string
		wantTargetDir   string
		wantModulePath  string
		wantErr         string
	}{
		{
			name:            "non-go language uses literal arg",
			lang:            "cpp",
			arg:             "github.com/acme/agent-village",
			wantProjectName: "github.com/acme/agent-village",
			wantTargetDir:   "github.com/acme/agent-village",
			wantModulePath:  "",
		},
		{
			name:            "explicit module bypasses go module heuristic",
			lang:            "go",
			module:          "example.com/custom/module",
			arg:             "github.com/acme/agent-village",
			wantProjectName: "github.com/acme/agent-village",
			wantTargetDir:   "github.com/acme/agent-village",
			wantModulePath:  "example.com/custom/module",
		},
		{
			name:    "derived repository name must also be a valid project name",
			lang:    "go",
			arg:     "github.com/acme/9agent",
			wantErr: "project name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectName, targetDir, modulePath, err := resolveNewProjectArgs(tt.lang, tt.module, tt.arg)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("resolveNewProjectArgs() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("resolveNewProjectArgs() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("resolveNewProjectArgs() error = %v", err)
			}
			if projectName != tt.wantProjectName {
				t.Fatalf("projectName = %q, want %q", projectName, tt.wantProjectName)
			}
			if targetDir != tt.wantTargetDir {
				t.Fatalf("targetDir = %q, want %q", targetDir, tt.wantTargetDir)
			}
			if modulePath != tt.wantModulePath {
				t.Fatalf("modulePath = %q, want %q", modulePath, tt.wantModulePath)
			}
		})
	}
}

func TestProjectNameFromGoModulePath(t *testing.T) {
	tests := []struct {
		name       string
		modulePath string
		want       string
	}{
		{name: "plain final segment", modulePath: "github.com/acme/agent-village", want: "agent-village"},
		{name: "major version uses parent segment", modulePath: "github.com/acme/agent-village/v2", want: "agent-village"},
		{name: "bare major version keeps suffix", modulePath: "v2", want: "v2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := projectNameFromGoModulePath(tt.modulePath); got != tt.want {
				t.Fatalf("projectNameFromGoModulePath(%q) = %q, want %q", tt.modulePath, got, tt.want)
			}
		})
	}
}
