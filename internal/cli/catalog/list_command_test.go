package catalog

import (
	"bytes"
	"strings"
	"testing"
)

func TestSelectedOutputFormat(t *testing.T) {
	t.Parallel()

	if got := selectedOutputFormat(false); got != outputFormatText {
		t.Fatalf("selectedOutputFormat(false) = %q, want %q", got, outputFormatText)
	}
	if got := selectedOutputFormat(true); got != outputFormatTOML {
		t.Fatalf("selectedOutputFormat(true) = %q, want %q", got, outputFormatTOML)
	}
}

func TestListCmdOutputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		wantContain []string
		wantEither  [][2]string
	}{
		{name: "text", wantContain: []string{"cpp", "go", "rust"}},
		{name: "toml", args: []string{"--toml"}, wantContain: []string{"languages = ", "cpp", "go", "rust"}},
		{name: "detail text", args: []string{"--detail"}, wantContain: []string{`desc="C++ starter"`, `manifest=v2`, `inputs=[author]`, `files=8`, `vars=[Author, ProjectName]`, `repo=[ci, contributing, dependabot, editorconfig, gitignore, typos]`, `repo_files=6`, `governance=standard`}},
		{name: "detail compact text", args: []string{"--detail", "--compact"}, wantContain: []string{"cpp [standard]", "counts: files=8 templates=3 repo_files=6 manifest=v2", "repo: ci, contributing, dependabot, editorconfig, gitignore, typos"}},
		{name: "detail table text", args: []string{"--detail", "--table"}, wantContain: []string{"NAME", "GOVERNANCE", "REPO_FILES", "cpp", "go", "rust", "ci,dependabot,editorconfig,gitignore,typos"}},
		{name: "detail min governance filter", args: []string{"--detail", "--min-governance", "rich"}, wantContain: []string{"rust\tdesc=", "governance=rich"}},
		{name: "detail repo asset filter", args: []string{"--detail", "--has-repo-asset", "security"}, wantContain: []string{"rust\tdesc=", "security"}},
		{name: "detail governance sort", args: []string{"--detail", "--sort", "governance"}, wantContain: []string{"rust\tdesc=", "governance=rich", "cpp\tdesc=", "go\tdesc="}},
		{name: "detail repo-files sort", args: []string{"--detail", "--sort", "repo-files"}, wantContain: []string{"rust\tdesc=", "repo_files=8", "cpp\tdesc=", "repo_files=6", "go\tdesc=", "repo_files=5"}},
		{name: "detail toml", args: []string{"--detail", "--toml"}, wantContain: []string{"[[templates]]", "manifest_version = 2", "file_count = 8", "template_count = 3", "Cargo-based Rust starter", "file_count = 17", "template_count = 7", "Go starter", "file_count = 7", "template_count = 2", "repo_assets", "repo_file_count", "governance_tier"}, wantEither: [][2]string{{"name = 'cpp'", `name = "cpp"`}, {"input_names = ['author']", `input_names = ["author"]`}, {"name = 'rust'", `name = "rust"`}, {"input_names = ['author', 'year']", `input_names = ["author", "year"]`}, {"name = 'go'", `name = "go"`}, {"input_names = ['module_path', 'go_version']", `input_names = ["module_path", "go_version"]`}, {"repo_assets = ['ci', 'dependabot', 'editorconfig', 'gitignore', 'typos']", `repo_assets = ["ci", "dependabot", "editorconfig", "gitignore", "typos"]`}, {"repo_file_count = 5", `repo_file_count = 5`}, {"governance_tier = 'rich'", `governance_tier = "rich"`}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := NewListCommand(newTestDependencies(newCommandTestCatalogService))
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			got := buf.String()
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Fatalf("output = %q, want contains %q", got, want)
				}
			}
			for _, want := range tt.wantEither {
				if !strings.Contains(got, want[0]) && !strings.Contains(got, want[1]) {
					t.Fatalf("output = %q, want contains either %q or %q", got, want[0], want[1])
				}
			}
		})
	}
}

func TestListCmd_ConfigDefaultsApplyToDetailAndFormat(t *testing.T) {
	t.Parallel()

	config := `version = 1

[list]
format = "toml"
detail = true
sort = "governance"
min_governance = "minimal"
required_assets = ["ci"]
`

	var buf bytes.Buffer
	cmd := NewListCommand(newTestDependenciesWithTOML(t, newCommandTestCatalogService, config))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "[[templates]]") {
		t.Fatalf("output = %q, want TOML detail output", got)
	}
	if !strings.Contains(got, "governance_tier") {
		t.Fatalf("output = %q, want detailed governance fields", got)
	}

	rustIndex := strings.Index(got, "name = \"rust\"")
	if rustIndex == -1 {
		rustIndex = strings.Index(got, "name = 'rust'")
	}
	cppIndex := strings.Index(got, "name = \"cpp\"")
	if cppIndex == -1 {
		cppIndex = strings.Index(got, "name = 'cpp'")
	}
	goIndex := strings.Index(got, "name = \"go\"")
	if goIndex == -1 {
		goIndex = strings.Index(got, "name = 'go'")
	}
	if rustIndex == -1 || cppIndex == -1 || goIndex == -1 {
		t.Fatalf("output = %q, want rust/cpp/go templates", got)
	}
	if rustIndex >= cppIndex || cppIndex >= goIndex {
		t.Fatalf("output = %q, want governance sort order rust -> cpp -> go", got)
	}
}

func TestListCmd_FlagsOverrideConfigDefaults(t *testing.T) {
	t.Parallel()

	config := `version = 1

[list]
format = "toml"
compact = true
detail = false
table = true
sort = "repo-files"
min_governance = "rich"
required_assets = ["security"]
`

	var buf bytes.Buffer
	cmd := NewListCommand(newTestDependenciesWithTOML(t, newCommandTestCatalogService, config))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--detail", "--toml=false", "--compact=false", "--table=false", "--sort", "name", "--min-governance", "minimal", "--has-repo-asset", "ci"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := buf.String()
	if strings.Contains(got, "[[templates]]") {
		t.Fatalf("output = %q, want text format after --toml=false override", got)
	}
	if strings.Contains(got, "NAME") || strings.Contains(got, "GOVERNANCE") {
		t.Fatalf("output = %q, want non-table detail output after --table=false override", got)
	}
	if !strings.Contains(got, "cpp\tdesc=") || !strings.Contains(got, "go\tdesc=") || !strings.Contains(got, "rust\tdesc=") {
		t.Fatalf("output = %q, want all templates after flag overrides", got)
	}
}

func TestListCmd_InvalidConfigCombinationStillFailsValidation(t *testing.T) {
	t.Parallel()

	config := `version = 1

[list]
detail = false
table = true
`

	cmd := NewListCommand(newTestDependenciesWithTOML(t, newCommandTestCatalogService, config))

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--table requires --detail") {
		t.Fatalf("Execute() error = %v, want table/detail validation error", err)
	}
}
