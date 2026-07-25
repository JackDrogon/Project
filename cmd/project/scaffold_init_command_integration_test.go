package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	appcreate "github.com/JackDrogon/project/internal/app/create"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

func TestInitCmd_DefaultsToCurrentDirectory(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "demo")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
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

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
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
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", targetDir, err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("WriteFile(old.txt) error = %v", err)
	}

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
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
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
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
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
	cmd.SetArgs([]string{"--lang", "go", "one", "two"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "accepts between 0 and 1 arg") {
		t.Fatalf("Execute() error = %v, want arg count error", err)
	}
}

func TestInitCmd_RejectsConfigAndReplayTogether(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("version = 1\n[init]\nlang = \"go\"\ntarget_dir = \"demo\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}
	replayPath := writeReplayTOMLForTest(t, protocoltoml.Replay{
		Version:  protocoltoml.ReplayVersion,
		Mode:     string(appcreate.CommandInit),
		Template: protocoltoml.ReplayTemplate{Lang: "go"},
		Project:  protocoltoml.ReplayProject{Name: "from-replay", TargetDir: "from-replay", ModulePath: "example.com/from-replay"},
		Git:      protocoltoml.ReplayGit{Mode: domain.GitModeNone},
	})

	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newRootCmd(newTestDependencies(creator))
	cmd.SetArgs([]string{"--config", configPath, "init", "--replay", replayPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--config and --replay cannot be combined") {
		t.Fatalf("Execute() error = %v, want config/replay conflict", err)
	}
}

func TestInitCmd_UsesConfigLangAndTargetDirWhenOmitted(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	targetDir := filepath.Join("nested", "from-config")
	configPath := filepath.Join(t.TempDir(), "config.toml")
	config := []byte("version = 1\n[init]\nlang = \"go\"\ntarget_dir = \"nested/from-config\"\ngit_mode = \"none\"\n")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := newRootCmd(newTestDependencies(creator))
	cmd.SetArgs([]string{"--config", configPath, "init"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	mainGo, err := os.ReadFile(filepath.Join(workDir, targetDir, "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(mainGo), `const Name = "from-config"`) {
		t.Fatalf("main.go content = %q, want rendered config target project name", string(mainGo))
	}
	if _, err := os.Stat(filepath.Join(workDir, "main.go")); !os.IsNotExist(err) {
		t.Fatalf("config target_dir should override implicit current directory, stat err = %v", err)
	}
}

func TestInitCmd_ExplainConfigWritesOnlyToStderr(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n")},
		"go/go.mod.tmpl":                     {Data: []byte("module {{.ModulePath}}\n")},
		"go/README.md.tmpl":                  {Data: []byte("By {{.Author}}\n")},
	}
	withTempWorkingDir(t, "workspace")
	configPath := filepath.Join(t.TempDir(), "config.toml")
	config := []byte("version = 1\n[init]\nlang = \"go\"\ntarget_dir = \"nested/from-config\"\nmodule = \"example.com/from-config\"\ngit_mode = \"none\"\nsignoff = false\n[init.inputs]\nauthor = \"config-author\"\n")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	creator := appcreate.NewCreator(fsys, &stdout)
	cmd := newRootCmd(newTestDependencies(creator))
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--config", configPath, "--explain-config", "init", "--dry-run"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if strings.Contains(stdout.String(), "config source report:") {
		t.Fatalf("stdout = %q, want no explain report", stdout.String())
	}
	if strings.Contains(stderr.String(), "Creating project with language") {
		t.Fatalf("stderr = %q, want no dry-run stdout text", stderr.String())
	}

	requireOrderedSubstrings(t, stderr.String(), []string{
		"config source report:\n",
		"  command: init\n",
		"  active_config_source: explicit-config\n",
		"  active_config_path: " + configPath + "\n",
		"  resolved values:\n",
		"    lang: go (source: explicit-config)\n",
		"    project_name: from-config (source: explicit-config)\n",
		"    target_dir: nested/from-config (source: explicit-config)\n",
		"    module: example.com/from-config (source: explicit-config)\n",
		"    git_mode: none (source: explicit-config)\n",
		"    signoff: false (source: explicit-config)\n",
		"  template inputs:\n",
		"    module_path: example.com/from-config (source: explicit-config)\n",
		"    author: config-author (source: explicit-config)\n",
	})
}

func TestInitCmd_PropagatesGitOptionConflict(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n")},
	}
	withTempWorkingDir(t, "workspace")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
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
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", targetDir, err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("WriteFile(old.txt) error = %v", err)
	}

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
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
		"go/.project-template-manifest.toml":        {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = false\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n\n[[inputs]]\nkey = \"year\"\ntemplate_var = \"Year\"\nrequired = false\n")},
		"go/README.md":                              {Data: []byte("# README\n")},
		"go/cmd":                                    {Mode: os.ModeDir},
		"go/cmd/{{.ProjectNameLower}}":              {Mode: os.ModeDir},
		"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	var out bytes.Buffer

	creator := appcreate.NewCreator(fsys, &out)
	cmd := requireSubcommand(t, creator, commandKeyInit)
	targetDir := filepath.Join("nested", "demo")
	cmd.SetArgs([]string{"--lang", "go", "--dry-run", "--git", "none", "--module", "example.com/demo", targetDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := appcreate.NewTemplateVars("demo", "example.com/demo")
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

func TestInitCmd_DryRunUsesResolvedConfigValues(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n")},
		"go/go.mod.tmpl":                     {Data: []byte("module {{.ModulePath}}\n")},
		"go/README.md.tmpl":                  {Data: []byte("By {{.Author}}\n")},
	}
	withTempWorkingDir(t, "workspace")
	configPath := filepath.Join(t.TempDir(), "config.toml")
	config := []byte("version = 1\n[init]\nlang = \"go\"\ntarget_dir = \"nested/from-config\"\nmodule = \"example.com/from-config\"\ngit_mode = \"none\"\n[init.inputs]\nauthor = \"config-author\"\n")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	var out bytes.Buffer
	creator := appcreate.NewCreator(fsys, &out)
	cmd := newRootCmd(newTestDependencies(creator))
	cmd.SetArgs([]string{"--config", configPath, "init", "--dry-run"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	requireOrderedSubstrings(t, out.String(), []string{
		"Creating project with language: go, project name: from-config\n",
		"Dry-run mode: no files will be created\n",
		"template: go\n",
		"description: Go starter\n",
		"target_dir: nested/from-config\n",
		"resolved inputs:\n",
		"  project_name: from-config\n",
		"  module_path: example.com/from-config\n",
		"  author: config-author\n",
		"  git_mode: none\n",
		"explicit overrides:\n",
		"  (none)\n",
	})
}

func TestInitCmd_RejectsWriteReplayWithDryRun(t *testing.T) {
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
	cmd.SetArgs([]string{"--lang", "go", "--dry-run", "--write-replay", filepath.Join(t.TempDir(), "replay.toml")})

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
	replayPath := writeReplayTOMLForTest(t, protocoltoml.Replay{
		Version:  protocoltoml.ReplayVersion,
		Mode:     string(appcreate.CommandNew),
		Template: protocoltoml.ReplayTemplate{Lang: "go"},
		Project: protocoltoml.ReplayProject{
			Name:       "replayed-demo",
			TargetDir:  "replayed-demo",
			ModulePath: "example.com/replayed-demo",
		},
		Git:     protocoltoml.ReplayGit{Mode: domain.GitModeNone},
		Options: protocoltoml.ReplayOptions{},
		Inputs:  map[string]string{"module_path": "example.com/replayed-demo"},
	})

	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
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

func TestInitCmd_ReplayRejectsLegacyJSONContent(t *testing.T) {
	workDir := withTempWorkingDir(t, "workspace")
	replayPath := filepath.Join(t.TempDir(), "replay.toml")
	if err := os.WriteFile(replayPath, []byte(`{"schema_version":1,"command":"init"}`), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", replayPath, err)
	}

	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
	cmd.SetArgs([]string{"--replay", replayPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "legacy JSON") {
		t.Fatalf("Execute() error = %v, want legacy JSON rejection", err)
	}
	if _, statErr := os.Stat(filepath.Join(workDir, "replayed-demo")); !os.IsNotExist(statErr) {
		t.Fatalf("legacy JSON input should fail before project creation, stat err = %v", statErr)
	}
}

func TestInitCmd_WriteReplayRecordsResolvedInputs(t *testing.T) {
	fsys := fstest.MapFS{
		"cpp/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"cpp\"\ndescription = \"C++ starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n")},
		"cpp/README.md.tmpl":                  {Data: []byte("By {{.Author}}\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	replayPath := filepath.Join(t.TempDir(), "replay.toml")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
	cmd.SetArgs([]string{"--lang", "cpp", "--git", "none", "--set", "author=alice", "--write-replay", replayPath, "demo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	replay, err := protocoltoml.ReadReplay(replayPath)
	if err != nil {
		t.Fatalf("ReadReplay(%q) error = %v", replayPath, err)
	}

	if replay.Mode != string(appcreate.CommandInit) {
		t.Fatalf("replay.Mode = %q, want %q", replay.Mode, appcreate.CommandInit)
	}
	if replay.Template.Lang != "cpp" {
		t.Fatalf("replay.Template.Lang = %q, want %q", replay.Template.Lang, "cpp")
	}
	if replay.Project.Name != "demo" {
		t.Fatalf("replay.Project.Name = %q, want %q", replay.Project.Name, "demo")
	}
	if replay.Project.TargetDir != "demo" {
		t.Fatalf("replay.Project.TargetDir = %q, want %q", replay.Project.TargetDir, "demo")
	}
	if replay.Git.Mode != domain.GitModeNone {
		t.Fatalf("replay.Git.Mode = %q, want %q", replay.Git.Mode, domain.GitModeNone)
	}
	if got := replay.Inputs["author"]; got != "alice" {
		t.Fatalf("replay.Inputs[author] = %q, want %q", got, "alice")
	}
	if len(replay.Inputs) != 1 {
		t.Fatalf("len(replay.Inputs) = %d, want %d", len(replay.Inputs), 1)
	}

	readme, err := os.ReadFile(filepath.Join(workDir, "demo", "README.md"))
	if err != nil {
		t.Fatalf("ReadFile(README.md) error = %v", err)
	}
	if string(readme) != "By alice\n" {
		t.Fatalf("README.md = %q, want %q", string(readme), "By alice\n")
	}
}

func TestInitCmd_WriteReplayIgnoresConfigMetadata(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n")},
		"go/go.mod.tmpl":                     {Data: []byte("module {{.ModulePath}}\n")},
		"go/README.md.tmpl":                  {Data: []byte("By {{.Author}}\n")},
	}
	withTempWorkingDir(t, "workspace")
	configPath := filepath.Join(t.TempDir(), "config.toml")
	config := []byte("version = 1\n[init]\nlang = \"go\"\ntarget_dir = \"nested/from-config\"\nmodule = \"example.com/from-config\"\ngit_mode = \"none\"\nsignoff = false\n[init.inputs]\nauthor = \"config-author\"\n")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}
	replayPath := filepath.Join(t.TempDir(), "replay.toml")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := newRootCmd(newTestDependencies(creator))
	cmd.SetArgs([]string{"--config", configPath, "init", "--write-replay", replayPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	replay, err := protocoltoml.ReadReplay(replayPath)
	if err != nil {
		t.Fatalf("ReadReplay(%q) error = %v", replayPath, err)
	}
	if replay.Mode != string(appcreate.CommandInit) {
		t.Fatalf("replay.Mode = %q, want %q", replay.Mode, appcreate.CommandInit)
	}
	if replay.Template.Lang != "go" {
		t.Fatalf("replay.Template.Lang = %q, want %q", replay.Template.Lang, "go")
	}
	if replay.Project.Name != "from-config" {
		t.Fatalf("replay.Project.Name = %q, want %q", replay.Project.Name, "from-config")
	}
	if replay.Project.TargetDir != filepath.Join("nested", "from-config") {
		t.Fatalf("replay.Project.TargetDir = %q, want %q", replay.Project.TargetDir, filepath.Join("nested", "from-config"))
	}
	if replay.Project.ModulePath != "example.com/from-config" {
		t.Fatalf("replay.Project.ModulePath = %q, want %q", replay.Project.ModulePath, "example.com/from-config")
	}
	if replay.Git.Mode != domain.GitModeNone {
		t.Fatalf("replay.Git.Mode = %q, want %q", replay.Git.Mode, domain.GitModeNone)
	}
	if replay.Git.Signoff {
		t.Fatal("replay.Git.Signoff = true, want false")
	}
	if got := replay.Inputs["module_path"]; got != "example.com/from-config" {
		t.Fatalf("replay.Inputs[module_path] = %q, want %q", got, "example.com/from-config")
	}
	if got := replay.Inputs["author"]; got != "config-author" {
		t.Fatalf("replay.Inputs[author] = %q, want %q", got, "config-author")
	}

	raw, err := os.ReadFile(replayPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", replayPath, err)
	}
	for _, forbidden := range []string{configPath, "active_config_source", "active_config_path", "explicit-config", "user-config"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("replay file = %q, want no config metadata %q", string(raw), forbidden)
		}
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

	_, err = appcreate.ProjectNameFromTargetDir(".")
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

	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyInit)
	cmd.SetArgs([]string{"--lang", "go"})

	err = cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to resolve target directory") {
		t.Fatalf("Execute() error = %v, want resolution error", err)
	}
}
