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

func TestNewCmd_RequiresLang(t *testing.T) {
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
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
		{name: "too many args", args: []string{"--lang", "go", "one", "two"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
			cmd := requireSubcommand(t, creator, commandKeyNew)
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

func TestNewCmd_RejectsMissingArgWithoutReplayOrConfig(t *testing.T) {
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--lang", "go"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Fatalf("Execute() error = %v, want arg count error", err)
	}
}

func TestNewCmd_RejectsConfigAndReplayTogether(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte("version = 1\n[new]\nlang = \"go\"\nproject_name = \"from-config\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}
	replayPath := writeReplayTOMLForTest(t, protocoltoml.Replay{
		Version:  protocoltoml.ReplayVersion,
		Mode:     string(appcreate.CommandNew),
		Template: protocoltoml.ReplayTemplate{Lang: "go"},
		Project:  protocoltoml.ReplayProject{Name: "from-replay", TargetDir: "from-replay", ModulePath: "example.com/from-replay"},
		Git:      protocoltoml.ReplayGit{Mode: domain.GitModeNone},
	})

	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := newRootCmd(newTestDependencies(creator))
	cmd.SetArgs([]string{"--config", configPath, "new", "--replay", replayPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--config and --replay cannot be combined") {
		t.Fatalf("Execute() error = %v, want config/replay conflict", err)
	}
}

func TestNewCmd_UsesConfigProjectNameAndLangWhenArgIsOmitted(t *testing.T) {
	fsys := fstest.MapFS{
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	configPath := filepath.Join(t.TempDir(), "config.toml")
	config := []byte("version = 1\n[new]\nlang = \"go\"\nproject_name = \"from-config\"\ngit_mode = \"none\"\n")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := newRootCmd(newTestDependencies(creator))
	cmd.SetArgs([]string{"--config", configPath, "new"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	mainGo, err := os.ReadFile(filepath.Join(workDir, "from-config", "main.go"))
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	if !strings.Contains(string(mainGo), `const Name = "from-config"`) {
		t.Fatalf("main.go content = %q, want rendered config project name", string(mainGo))
	}
}

func TestNewCmd_ExplainConfigWritesOnlyToStderr(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = false\n")},
		"go/go.mod.tmpl":                     {Data: []byte("module {{.ModulePath}}\ngo {{.GoVersion}}\n")},
	}
	withTempWorkingDir(t, "workspace")
	configPath := filepath.Join(t.TempDir(), "config.toml")
	config := []byte("version = 1\n[new]\nlang = \"go\"\nproject_name = \"from-config\"\nmodule = \"example.com/from-config\"\ngit_mode = \"none\"\nsignoff = false\n[new.inputs]\ngo_version = \"1.25\"\n")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	creator := appcreate.NewCreator(fsys, &stdout)
	cmd := newRootCmd(newTestDependencies(creator))
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--config", configPath, "--explain-config", "new", "--dry-run"})

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
		"  command: new\n",
		"  active_config_source: explicit-config\n",
		"  active_config_path: " + configPath + "\n",
		"  resolved values:\n",
		"    lang: go (source: explicit-config)\n",
		"    project_name: from-config (source: explicit-config)\n",
		"    target_dir: from-config (source: explicit-config)\n",
		"    module: example.com/from-config (source: explicit-config)\n",
		"    git_mode: none (source: explicit-config)\n",
		"    signoff: false (source: explicit-config)\n",
		"  template inputs:\n",
		"    module_path: example.com/from-config (source: explicit-config)\n",
		"    go_version: 1.25 (source: explicit-config)\n",
	})
}

func TestNewCmd_DryRunUsesResolvedConfigValues(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = false\n")},
		"go/go.mod.tmpl":                     {Data: []byte("module {{.ModulePath}}\ngo {{.GoVersion}}\n")},
	}
	withTempWorkingDir(t, "workspace")
	configPath := filepath.Join(t.TempDir(), "config.toml")
	config := []byte("version = 1\n[new]\nlang = \"go\"\nproject_name = \"from-config\"\nmodule = \"example.com/from-config\"\ngit_mode = \"none\"\n[new.inputs]\ngo_version = \"1.25\"\n")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	var out bytes.Buffer
	creator := appcreate.NewCreator(fsys, &out)
	cmd := newRootCmd(newTestDependencies(creator))
	cmd.SetArgs([]string{"--config", configPath, "new", "--dry-run"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	requireOrderedSubstrings(t, out.String(), []string{
		"Creating project with language: go, project name: from-config\n",
		"Dry-run mode: no files will be created\n",
		"template: go\n",
		"description: Go starter\n",
		"target_dir: from-config\n",
		"resolved inputs:\n",
		"  project_name: from-config\n",
		"  module_path: example.com/from-config\n",
		"  go_version: 1.25\n",
		"  git_mode: none\n",
		"explicit overrides:\n",
		"  (none)\n",
	})
}

func TestNewCmd_CreatesProjectFromArgument(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":  {Data: []byte("module {{.ModulePath}}\n")},
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
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

func TestNewCmd_CreatesRustProjectFromArgument(t *testing.T) {
	fsys := fstest.MapFS{
		"rust/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"rust\"\ndescription = \"Cargo-based Rust starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n\n[[inputs]]\nkey = \"year\"\ntemplate_var = \"Year\"\nrequired = false\n")},
		"rust/.github":                         {Mode: os.ModeDir},
		"rust/.github/dependabot.yml":          {Data: []byte("version: 2\n")},
		"rust/.github/workflows":               {Mode: os.ModeDir},
		"rust/.github/workflows/ci.yml":        {Data: []byte("name: ci\n")},
		"rust/.gitignore":                      {Data: []byte("/target/\n")},
		"rust/CONTRIBUTING.md.tmpl":            {Data: []byte("# Contributing to {{.ProjectName}}\n")},
		"rust/Cargo.toml.tmpl":                 {Data: []byte("[package]\nname = \"{{.ProjectName}}\"\nauthors = [\"{{.Author}}\"]\n")},
		"rust/README.md.tmpl":                  {Data: []byte("# {{.ProjectName}}\nGenerated in {{.Year}}\n")},
		"rust/SECURITY.md.tmpl":                {Data: []byte("# Security Policy\n{{.ProjectName}}\n")},
		"rust/dprint.json":                     {Data: []byte("{\n  \"includes\": [\"**/*.toml\"]\n}\n")},
		"rust/justfile.tmpl":                   {Data: []byte("test:\n    cargo test --all-features\n")},
		"rust/src/lib.rs.tmpl":                 {Data: []byte("pub fn greet(name: &str) -> String { format!(\"Hello, {name}!\") }\n")},
		"rust/src/main.rs.tmpl":                {Data: []byte("fn app_name() -> &'static str { \"{{.ProjectName}}\" }\nfn main() { println!(\"Hello, {}!\", app_name()); }\n")},
		"rust/typos.toml":                      {Data: []byte("[files]\nextend-exclude = [\"target\"]\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--lang", "rust", "--git", "none", "demo-rust"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	cargoToml, err := os.ReadFile(filepath.Join(workDir, "demo-rust", "Cargo.toml"))
	if err != nil {
		t.Fatalf("ReadFile(Cargo.toml) error = %v", err)
	}
	if !strings.Contains(string(cargoToml), "name = \"demo-rust\"") {
		t.Fatalf("Cargo.toml = %q, want rendered package name", string(cargoToml))
	}

	mainRS, err := os.ReadFile(filepath.Join(workDir, "demo-rust", "src", "main.rs"))
	if err != nil {
		t.Fatalf("ReadFile(src/main.rs) error = %v", err)
	}
	if !strings.Contains(string(mainRS), `"demo-rust"`) {
		t.Fatalf("src/main.rs content = %q, want rendered project name", string(mainRS))
	}

	readme, err := os.ReadFile(filepath.Join(workDir, "demo-rust", "README.md"))
	if err != nil {
		t.Fatalf("ReadFile(README.md) error = %v", err)
	}
	if !strings.Contains(string(readme), "# demo-rust") {
		t.Fatalf("README.md content = %q, want rendered title", string(readme))
	}

	if _, err := os.Stat(filepath.Join(workDir, "demo-rust", "dprint.json")); err != nil {
		t.Fatalf("Stat(dprint.json) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo-rust", "typos.toml")); err != nil {
		t.Fatalf("Stat(typos.toml) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo-rust", ".github", "workflows", "ci.yml")); err != nil {
		t.Fatalf("Stat(.github/workflows/ci.yml) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo-rust", "justfile")); err != nil {
		t.Fatalf("Stat(justfile) error = %v", err)
	}
}

func TestNewCmd_CreatesGoProjectWithRepoLevelFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n")},
		"go/.editorconfig":                   {Data: []byte("root = true\n")},
		"go/.github":                         {Mode: os.ModeDir},
		"go/.github/dependabot.yml":          {Data: []byte("version: 2\n")},
		"go/.github/workflows":               {Mode: os.ModeDir},
		"go/.github/workflows/ci.yml":        {Data: []byte("name: CI\n")},
		"go/.gitignore":                      {Data: []byte("bin/\n")},
		"go/go.mod.tmpl":                     {Data: []byte("module {{.ModulePath}}\n")},
		"go/justfile.tmpl":                   {Data: []byte("pre-commit:\n    go test ./...\n")},
		"go/main.go.tmpl":                    {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
		"go/typos.toml":                      {Data: []byte("[files]\nextend-exclude = [\"bin\"]\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "--module", "example.com/demo-go", "demo-go"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(workDir, "demo-go", ".editorconfig")); err != nil {
		t.Fatalf("Stat(.editorconfig) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo-go", ".github", "dependabot.yml")); err != nil {
		t.Fatalf("Stat(.github/dependabot.yml) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo-go", ".github", "workflows", "ci.yml")); err != nil {
		t.Fatalf("Stat(.github/workflows/ci.yml) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo-go", "typos.toml")); err != nil {
		t.Fatalf("Stat(typos.toml) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo-go", "justfile")); err != nil {
		t.Fatalf("Stat(justfile) error = %v", err)
	}
}

func TestNewCmd_CreatesCppProjectWithRepoLevelFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"cpp/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"cpp\"\ndescription = \"C++ starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n")},
		"cpp/.editorconfig":                   {Data: []byte("root = true\n")},
		"cpp/.github":                         {Mode: os.ModeDir},
		"cpp/.github/dependabot.yml":          {Data: []byte("version: 2\n")},
		"cpp/.github/workflows":               {Mode: os.ModeDir},
		"cpp/.github/workflows/ci.yml":        {Data: []byte("name: ci\n")},
		"cpp/.gitignore":                      {Data: []byte("/build/\n")},
		"cpp/CONTRIBUTING.md.tmpl":            {Data: []byte("# Contributing to {{.ProjectName}}\n")},
		"cpp/CMakeLists.txt.tmpl":             {Data: []byte("cmake_minimum_required(VERSION 3.20)\nproject({{.ProjectName}})\nadd_executable({{.ProjectName}} src/main.cc)\n")},
		"cpp/README.md.tmpl":                  {Data: []byte("# {{.ProjectName}}\n")},
		"cpp/justfile.tmpl":                   {Data: []byte("build:\n    cmake -S . -B build\n")},
		"cpp/src":                             {Mode: os.ModeDir},
		"cpp/src/main.cc.tmpl":                {Data: []byte("int main() { return 0; }\n")},
		"cpp/typos.toml":                      {Data: []byte("[files]\nextend-exclude = [\"build\"]\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--lang", "cpp", "--git", "none", "demo-cpp"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(workDir, "demo-cpp", ".editorconfig")); err != nil {
		t.Fatalf("Stat(.editorconfig) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo-cpp", ".github", "dependabot.yml")); err != nil {
		t.Fatalf("Stat(.github/dependabot.yml) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo-cpp", ".github", "workflows", "ci.yml")); err != nil {
		t.Fatalf("Stat(.github/workflows/ci.yml) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo-cpp", "CONTRIBUTING.md")); err != nil {
		t.Fatalf("Stat(CONTRIBUTING.md) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "demo-cpp", "typos.toml")); err != nil {
		t.Fatalf("Stat(typos.toml) error = %v", err)
	}
}

func TestNewCmd_GoModuleArgumentDerivesProjectNameAndModulePath(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":  {Data: []byte("module {{.ModulePath}}\n")},
		"go/main.go.tmpl": {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
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
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--lang", "go", "--dry-run", "--git", "none", "--module", "example.com/demo", "demo"})

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

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
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
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
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

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
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
	replayPath := writeReplayTOMLForTest(t, protocoltoml.Replay{
		Version:  protocoltoml.ReplayVersion,
		Mode:     string(appcreate.CommandInit),
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
	cmd := requireSubcommand(t, creator, commandKeyNew)
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

func TestNewCmd_ReplayRejectsLegacyJSONContent(t *testing.T) {
	workDir := withTempWorkingDir(t, "workspace")
	replayPath := filepath.Join(t.TempDir(), "replay.toml")
	if err := os.WriteFile(replayPath, []byte(`{"schema_version":1,"command":"new"}`), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", replayPath, err)
	}

	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
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

func TestNewCmd_ReplayCLIFlagsAndPositionalArgTakePrecedence(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":     {Data: []byte("module {{.ModulePath}}\n")},
		"go/main.go.tmpl":    {Data: []byte("package main\n\nconst Name = \"{{.ProjectName}}\"\n")},
		"cpp/main.cc.tmpl":   {Data: []byte("int main() { return 0; }\n")},
		"cpp/README.md.tmpl": {Data: []byte("# {{.ProjectName}}\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	replayPath := writeReplayTOMLForTest(t, protocoltoml.Replay{
		Version:  protocoltoml.ReplayVersion,
		Mode:     string(appcreate.CommandNew),
		Template: protocoltoml.ReplayTemplate{Lang: "cpp"},
		Project: protocoltoml.ReplayProject{
			Name:       "replay-name",
			TargetDir:  "replay-dir",
			ModulePath: "example.com/from-replay",
		},
		Git: protocoltoml.ReplayGit{
			Mode:    domain.GitModeInitOnly,
			Signoff: true,
		},
		Options: protocoltoml.ReplayOptions{Force: true},
		Inputs:  map[string]string{"module_path": "example.com/from-replay"},
	})

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
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
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--lang", "go", "--dry-run", "--write-replay", filepath.Join(t.TempDir(), "replay.toml"), "demo"})

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
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = false\n")},
		"go/go.mod.tmpl":                     {Data: []byte("module {{.ModulePath}}\ngo {{.GoVersion}}\n")},
	}
	withTempWorkingDir(t, "workspace")
	replayPath := filepath.Join(t.TempDir(), "replay.toml")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
	cmd.SetArgs([]string{"--lang", "go", "--git", "none", "--module", "example.com/demo", "--set", "go_version=1.25", "--write-replay", replayPath, "demo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	replay, err := protocoltoml.ReadReplay(replayPath)
	if err != nil {
		t.Fatalf("ReadReplay(%q) error = %v", replayPath, err)
	}

	if replay.Mode != string(appcreate.CommandNew) {
		t.Fatalf("replay.Mode = %q, want %q", replay.Mode, appcreate.CommandNew)
	}
	if replay.Template.Lang != "go" {
		t.Fatalf("replay.Template.Lang = %q, want %q", replay.Template.Lang, "go")
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
	if replay.Git.Signoff {
		t.Fatal("replay.Git.Signoff = true, want false")
	}
	if replay.Options.Force {
		t.Fatal("replay.Options.Force = true, want false")
	}
	if got := replay.Inputs["module_path"]; got != "example.com/demo" {
		t.Fatalf("replay.Inputs[module_path] = %q, want %q", got, "example.com/demo")
	}
	if got := replay.Inputs["go_version"]; got != "1.25" {
		t.Fatalf("replay.Inputs[go_version] = %q, want %q", got, "1.25")
	}
	if len(replay.Inputs) != 2 {
		t.Fatalf("len(replay.Inputs) = %d, want %d", len(replay.Inputs), 2)
	}
}

func TestNewCmd_WriteReplayIgnoresConfigMetadata(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = false\n")},
		"go/go.mod.tmpl":                     {Data: []byte("module {{.ModulePath}}\ngo {{.GoVersion}}\n")},
	}
	withTempWorkingDir(t, "workspace")
	configPath := filepath.Join(t.TempDir(), "config.toml")
	config := []byte("version = 1\n[new]\nlang = \"go\"\nproject_name = \"from-config\"\nmodule = \"example.com/from-config\"\ngit_mode = \"none\"\nsignoff = false\n[new.inputs]\ngo_version = \"1.25\"\n")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}
	replayPath := filepath.Join(t.TempDir(), "replay.toml")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := newRootCmd(newTestDependencies(creator))
	cmd.SetArgs([]string{"--config", configPath, "new", "--write-replay", replayPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	replay, err := protocoltoml.ReadReplay(replayPath)
	if err != nil {
		t.Fatalf("ReadReplay(%q) error = %v", replayPath, err)
	}
	if replay.Mode != string(appcreate.CommandNew) {
		t.Fatalf("replay.Mode = %q, want %q", replay.Mode, appcreate.CommandNew)
	}
	if replay.Template.Lang != "go" {
		t.Fatalf("replay.Template.Lang = %q, want %q", replay.Template.Lang, "go")
	}
	if replay.Project.Name != "from-config" {
		t.Fatalf("replay.Project.Name = %q, want %q", replay.Project.Name, "from-config")
	}
	if replay.Project.TargetDir != "from-config" {
		t.Fatalf("replay.Project.TargetDir = %q, want %q", replay.Project.TargetDir, "from-config")
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
	if got := replay.Inputs["go_version"]; got != "1.25" {
		t.Fatalf("replay.Inputs[go_version] = %q, want %q", got, "1.25")
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

func TestNewCmd_WriteReplayFailureReturnsWrappedError(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n")},
		"go/main.go.tmpl":                    {Data: []byte("package main\n")},
	}
	workDir := withTempWorkingDir(t, "workspace")
	replayPath := filepath.Join(t.TempDir(), "missing", "replay.toml")

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})
	cmd := requireSubcommand(t, creator, commandKeyNew)
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
			projectName, targetDir, modulePath, err := appcreate.ResolveNewProjectArgs("go", "", tt.arg)
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
			name:            "rust also uses literal arg",
			lang:            "rust",
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
			projectName, targetDir, modulePath, err := appcreate.ResolveNewProjectArgs(tt.lang, tt.module, tt.arg)
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
			if got := appcreate.ProjectNameFromGoModulePath(tt.modulePath); got != tt.want {
				t.Fatalf("projectNameFromGoModulePath(%q) = %q, want %q", tt.modulePath, got, tt.want)
			}
		})
	}
}
