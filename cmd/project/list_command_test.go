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
		"go/go.mod.tmpl":       {Data: []byte("module {{.ModulePath}}")},
		"go/main.go.tmpl":      {Data: []byte("package main\n")},
		"go/.gitignore":        {Data: []byte("bin/")},
		"cpp/src/main.cc.tmpl": {Data: []byte("// {{.ProjectName}}")},
		"cpp/README.md.tmpl":   {Data: []byte("By {{.Author}}")},
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
			Name:          "go",
			FileCount:     3,
			TemplateCount: 2,
			Variables:     []string{"ModulePath", "ProjectName"},
		},
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
			{Name: "go", FileCount: 3, TemplateCount: 2, Variables: []string{"ModulePath"}},
			{Name: "txt", FileCount: 1, TemplateCount: 0, Variables: nil},
		}
		if err := writeYAMLSummaries(&buf, summaries); err != nil {
			t.Fatalf("writeYAMLSummaries() error = %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, `name: "go"`) || !strings.Contains(got, `variables: []`) {
			t.Fatalf("writeYAMLSummaries() output = %q", got)
		}
	})

	t.Run("inspect output", func(t *testing.T) {
		var buf bytes.Buffer
		output := inspectOutput{
			Name:          "go",
			FileCount:     3,
			TemplateCount: 2,
			Variables:     []string{"ModulePath"},
			Mode:          inspectModeRender,
			ShownCount:    1,
			Files:         []scaffold.TemplateFile{{Source: "go.mod.tmpl", Output: "go.mod", IsTemplate: true}},
		}
		if err := writeYAMLInspectOutput(&buf, output); err != nil {
			t.Fatalf("writeYAMLInspectOutput() error = %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, `shown_count: 1`) || !strings.Contains(got, `source: "go.mod.tmpl"`) {
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
		{name: "detail json", args: []string{"--detail", "--json"}, wantContain: []string{`"name": "cpp"`, `"file_count": 2`}},
		{name: "detail yaml", args: []string{"--detail", "--yaml"}, wantContain: []string{"file_count:", "template_count:", "variables:"}},
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
		{name: "json render", args: []string{"go", "--json", "--mode", "render"}, wantContain: []string{`"mode": "render"`, `"shown_count": 2`, `"source": "go.mod.tmpl"`}},
		{name: "yaml copy", args: []string{"go", "--yaml", "--mode", "copy"}, wantContain: []string{`mode: "copy"`, `shown_count: 1`, `source: ".gitignore"`}},
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
