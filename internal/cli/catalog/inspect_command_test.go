package catalog

import (
	"bytes"
	"strings"
	"testing"
)

func TestInspectCmdOutputs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantContain []string
		wantEither  [][2]string
		wantErr     string
	}{
		{name: "text", args: []string{"go"}, wantContain: []string{"name: go", "manifest_version: 2", "repo_assets: ci, dependabot, editorconfig, gitignore, typos", "mode: all", "shown: 7", "repo_files:", "language_files:"}},
		{name: "compact text", args: []string{"go", "--compact"}, wantContain: []string{"go — Go starter", "manifest: v2 | mode: all | shown: 7 | files: 7 | templates: 2", "repo_files:", "language_files:"}},
		{name: "toml render", args: []string{"go", "--toml", "--mode", "render"}, wantContain: []string{"manifest_version = 2", "mode = ", "shown_count = 2", "[[files]]", "[[language_files]]", "repo_files = []", "source = "}, wantEither: [][2]string{{"name = 'go'", `name = "go"`}, {"mode = 'render'", `mode = "render"`}, {"source = 'go.mod.tmpl'", `source = "go.mod.tmpl"`}}},
		{name: "compact toml rejected", args: []string{"go", "--compact", "--toml"}, wantErr: "compact output is only supported for text format"},
		{name: "invalid mode", args: []string{"go", "--mode", "bogus"}, wantErr: "invalid --mode"},
		{name: "unsupported language", args: []string{"missing"}, wantErr: "unsupported language"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := NewInspectCommand(newTestDependencies(newCommandTestCatalogService))
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("Execute() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Execute() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
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

func TestInspectCmd_ConfigDefaultsSupplyLangAndMode(t *testing.T) {
	config := `version = 1

[inspect]
lang = "go"
format = "toml"
mode = "render"
`

	var buf bytes.Buffer
	cmd := NewInspectCommand(newTestDependencies(newCommandTestCatalogService))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(withActiveConfigContext(t, config))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "shown_count = 2") {
		t.Fatalf("output = %q, want render-mode filtered output", got)
	}
	if !strings.Contains(got, "go.mod.tmpl") {
		t.Fatalf("output = %q, want go template output from config lang", got)
	}
	if !strings.Contains(got, "mode = \"render\"") && !strings.Contains(got, "mode = 'render'") {
		t.Fatalf("output = %q, want render mode from config", got)
	}
}

func TestInspectCmd_PositionalLangOverridesConfig(t *testing.T) {
	config := `version = 1

[inspect]
lang = "rust"
`

	var buf bytes.Buffer
	cmd := NewInspectCommand(newTestDependencies(newCommandTestCatalogService))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetContext(withActiveConfigContext(t, config))
	cmd.SetArgs([]string{"go"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "name: go") {
		t.Fatalf("output = %q, want positional language to win", got)
	}
}

func TestInspectCmd_MissingLangStillFailsWithoutArgOrConfig(t *testing.T) {
	cmd := NewInspectCommand(newTestDependencies(newCommandTestCatalogService))

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "inspection query requires a language") {
		t.Fatalf("Execute() error = %v, want missing language validation error", err)
	}
}
