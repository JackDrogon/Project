package main

import (
	"bytes"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/JackDrogon/project/pkg/scaffold"
)

type failingCommandFS struct {
	err error
}

func (f failingCommandFS) Open(name string) (fs.File, error) {
	return nil, f.err
}

func (f failingCommandFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, f.err
}

func newCommandTestCreator() *scaffold.Creator {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.json":  {Data: []byte(`{"schema_version":1,"name":"go","description":"Go starter","inputs":[{"name":"module_path","template_var":"ModulePath"},{"name":"go_version","template_var":"GoVersion"}]}`)},
		"cpp/.project-template-manifest.json": {Data: []byte(`{"schema_version":1,"name":"cpp","description":"C++ starter","inputs":[{"name":"author","template_var":"Author"}]}`)},
		"go/go.mod.tmpl":                      {Data: []byte("module {{.ModulePath}}")},
		"go/main.go.tmpl":                     {Data: []byte("package main\n")},
		"go/.gitignore":                       {Data: []byte("bin/")},
		"cpp/src/main.cc.tmpl":                {Data: []byte("// {{.ProjectName}}")},
		"cpp/README.md.tmpl":                  {Data: []byte("By {{.Author}}")},
	}

	return scaffold.NewCreator(fsys, &bytes.Buffer{})
}

func TestSelectedOutputFormat(t *testing.T) {
	tests := []struct {
		name    string
		asJSON  bool
		asYAML  bool
		want    string
		wantErr bool
	}{
		{name: "text default", want: outputFormatText},
		{name: "json", asJSON: true, want: outputFormatJSON},
		{name: "yaml", asYAML: true, want: outputFormatYAML},
		{name: "conflict", asJSON: true, asYAML: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectedOutputFormat(tt.asJSON, tt.asYAML)
			if tt.wantErr {
				if err == nil {
					t.Fatal("selectedOutputFormat() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("selectedOutputFormat() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("selectedOutputFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildInspectOutput(t *testing.T) {
	details := scaffold.TemplateDetails{
		TemplateSummary: scaffold.TemplateSummary{
			Name:            "go",
			Description:     "Go starter",
			ManifestVersion: 1,
			InputNames:      []string{"module_path", "go_version"},
			FileCount:       3,
			TemplateCount:   2,
			Variables:       []string{"ModulePath", "ProjectName"},
		},
		Inputs: []scaffold.TemplateManifestInput{{Name: "module_path", TemplateVar: "ModulePath"}, {Name: "go_version", TemplateVar: "GoVersion"}},
		Files: []scaffold.TemplateFile{
			{Source: ".gitignore", Output: ".gitignore", IsTemplate: false},
			{Source: "go.mod.tmpl", Output: "go.mod", IsTemplate: true},
			{Source: "main.go.tmpl", Output: "main.go", IsTemplate: true},
		},
	}

	tests := []struct {
		name      string
		mode      string
		shown     int
		firstFile string
		wantErr   bool
	}{
		{name: "all", mode: inspectModeAll, shown: 3, firstFile: ".gitignore"},
		{name: "render", mode: inspectModeRender, shown: 2, firstFile: "go.mod.tmpl"},
		{name: "copy", mode: inspectModeCopy, shown: 1, firstFile: ".gitignore"},
		{name: "blank falls back to all", mode: "  ", shown: 3, firstFile: ".gitignore"},
		{name: "invalid", mode: "bogus", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildInspectOutput(details, tt.mode)
			if tt.wantErr {
				if err == nil {
					t.Fatal("buildInspectOutput() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("buildInspectOutput() error = %v", err)
			}
			if got.Description != details.Description {
				t.Fatalf("Description = %q, want %q", got.Description, details.Description)
			}
			if got.ManifestVersion != details.ManifestVersion {
				t.Fatalf("ManifestVersion = %d, want %d", got.ManifestVersion, details.ManifestVersion)
			}
			if len(got.Inputs) != len(details.Inputs) {
				t.Fatalf("len(Inputs) = %d, want %d", len(got.Inputs), len(details.Inputs))
			}
			if got.ShownCount != tt.shown {
				t.Fatalf("ShownCount = %d, want %d", got.ShownCount, tt.shown)
			}
			if len(got.Files) == 0 || got.Files[0].Source != tt.firstFile {
				t.Fatalf("first file = %#v, want %q", got.Files, tt.firstFile)
			}
		})
	}
}

func TestYAMLWriters(t *testing.T) {
	t.Run("langs", func(t *testing.T) {
		var buf bytes.Buffer
		if err := writeYAMLLangs(&buf, []string{"go", "cpp"}); err != nil {
			t.Fatalf("writeYAMLLangs() error = %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, `- "go"`) || !strings.Contains(got, `- "cpp"`) {
			t.Fatalf("writeYAMLLangs() output = %q", got)
		}
	})

	t.Run("summaries", func(t *testing.T) {
		var buf bytes.Buffer
		summaries := []scaffold.TemplateSummary{
			{Name: "go", Description: "Go starter", ManifestVersion: 1, InputNames: []string{"module_path"}, FileCount: 3, TemplateCount: 2, Variables: []string{"ModulePath"}},
			{Name: "txt", Description: "", ManifestVersion: 1, InputNames: nil, FileCount: 1, TemplateCount: 0, Variables: nil},
		}
		if err := writeYAMLSummaries(&buf, summaries); err != nil {
			t.Fatalf("writeYAMLSummaries() error = %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, `name: "go"`) || !strings.Contains(got, `input_names:`) || !strings.Contains(got, `variables: []`) {
			t.Fatalf("writeYAMLSummaries() output = %q", got)
		}
	})

	t.Run("inspect output", func(t *testing.T) {
		var buf bytes.Buffer
		output := inspectOutput{
			Name:            "go",
			Description:     "Go starter",
			ManifestVersion: 1,
			Inputs:          []scaffold.TemplateManifestInput{{Name: "module_path", TemplateVar: "ModulePath"}},
			FileCount:       3,
			TemplateCount:   2,
			Variables:       []string{"ModulePath"},
			Mode:            inspectModeRender,
			ShownCount:      1,
			Files:           []scaffold.TemplateFile{{Source: "go.mod.tmpl", Output: "go.mod", IsTemplate: true}},
		}
		if err := writeYAMLInspectOutput(&buf, output); err != nil {
			t.Fatalf("writeYAMLInspectOutput() error = %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, `shown_count: 1`) || !strings.Contains(got, `template_var: "ModulePath"`) || !strings.Contains(got, `source: "go.mod.tmpl"`) {
			t.Fatalf("writeYAMLInspectOutput() output = %q", got)
		}
	})

	t.Run("empty inspect files", func(t *testing.T) {
		var buf bytes.Buffer
		output := inspectOutput{Name: "go", Mode: inspectModeCopy}
		if err := writeYAMLInspectOutput(&buf, output); err != nil {
			t.Fatalf("writeYAMLInspectOutput() error = %v", err)
		}
		if !strings.Contains(buf.String(), `files: []`) {
			t.Fatalf("writeYAMLInspectOutput() output = %q", buf.String())
		}
	})

	if got := yamlQuote("go"); got != `"go"` {
		t.Fatalf("yamlQuote() = %q, want %q", got, `"go"`)
	}
}

func TestListCmdOutputs(t *testing.T) {
	creator := newCommandTestCreator()
	tests := []struct {
		name        string
		args        []string
		wantContain []string
		wantErr     string
	}{
		{name: "text", wantContain: []string{"cpp", "go"}},
		{name: "json", args: []string{"--json"}, wantContain: []string{`"cpp"`, `"go"`}},
		{name: "yaml", args: []string{"--yaml"}, wantContain: []string{`- "cpp"`, `- "go"`}},
		{name: "detail text", args: []string{"--detail"}, wantContain: []string{"files=", "templates=", "vars=["}},
		{name: "detail json", args: []string{"--detail", "--json"}, wantContain: []string{`"name": "cpp"`, `"description": "C++ starter"`, `"manifest_version": 1`, `"input_names": [`}},
		{name: "detail yaml", args: []string{"--detail", "--yaml"}, wantContain: []string{"description:", "manifest_version:", "input_names:", "variables:"}},
		{name: "conflict", args: []string{"--json", "--yaml"}, wantErr: "cannot be used together"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := newListCmd(creator)
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
		})
	}
}

func TestListCmd_PropagatesCreatorErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "langs", args: nil},
		{name: "summaries", args: []string{"--detail"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := scaffold.NewCreator(failingCommandFS{err: errors.New("boom")}, &bytes.Buffer{})
			cmd := newListCmd(creator)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("Execute() expected error, got nil")
			}
			if !strings.Contains(err.Error(), "failed to read templates") {
				t.Fatalf("Execute() error = %v, want template read error", err)
			}
		})
	}
}

func TestInspectCmdOutputs(t *testing.T) {
	creator := newCommandTestCreator()
	tests := []struct {
		name        string
		args        []string
		wantContain []string
		wantErr     string
	}{
		{name: "text", args: []string{"go"}, wantContain: []string{"name: go", "mode: all", "shown: 3"}},
		{name: "json render", args: []string{"go", "--json", "--mode", "render"}, wantContain: []string{`"description": "Go starter"`, `"inputs": [`, `"mode": "render"`, `"shown_count": 2`, `"source": "go.mod.tmpl"`}},
		{name: "yaml copy", args: []string{"go", "--yaml", "--mode", "copy"}, wantContain: []string{`manifest_version: 1`, `template_var: "ModulePath"`, `mode: "copy"`, `shown_count: 1`, `source: ".gitignore"`}},
		{name: "conflicting formats", args: []string{"go", "--json", "--yaml"}, wantErr: "cannot be used together"},
		{name: "invalid mode", args: []string{"go", "--mode", "bogus"}, wantErr: "invalid --mode"},
		{name: "unsupported language", args: []string{"missing"}, wantErr: "unsupported language"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := newInspectCmd(creator)
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
		})
	}
}

func TestListCmd_DetailOutputsManifestFields(t *testing.T) {
	creator := newCommandTestCreator()
	tests := []struct {
		name        string
		args        []string
		wantContain []string
	}{
		{name: "text", args: []string{"--detail"}, wantContain: []string{`desc="C++ starter"`, `manifest=v1`, `inputs=[author]`, `files=2`, `vars=[Author, ProjectName]`}},
		{name: "json", args: []string{"--detail", "--json"}, wantContain: []string{`"description": "C++ starter"`, `"manifest_version": 1`, `"input_names": [`, `"file_count": 2`, `"variables": [`}},
		{name: "yaml", args: []string{"--detail", "--yaml"}, wantContain: []string{`description: "C++ starter"`, `manifest_version: 1`, `input_names:`, `file_count: 2`, `variables:`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := newListCmd(creator)
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
		})
	}
}

func TestInspectCmd_OutputsManifestInputs(t *testing.T) {
	creator := newCommandTestCreator()
	tests := []struct {
		name        string
		args        []string
		wantContain []string
	}{
		{name: "text", args: []string{"go"}, wantContain: []string{`description: Go starter`, `manifest_version: 1`, `inputs: module_path->ModulePath, go_version->GoVersion`, `variables: ModulePath`}},
		{name: "json", args: []string{"go", "--json"}, wantContain: []string{`"description": "Go starter"`, `"manifest_version": 1`, `"inputs": [`, `"name": "module_path"`, `"template_var": "ModulePath"`, `"files": [`}},
		{name: "yaml", args: []string{"go", "--yaml"}, wantContain: []string{`description: "Go starter"`, `manifest_version: 1`, `inputs:`, `name: "module_path"`, `template_var: "ModulePath"`, `files:`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := newInspectCmd(creator)
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
		})
	}
}
