package presenters

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JackDrogon/project/internal/app/catalog"
	"github.com/JackDrogon/project/internal/scaffold"
)

func TestNewPresenter(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"text format", "text", false},
		{"toml format", "toml", false},
		{"invalid format", "json", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter, err := NewPresenter(tt.format)
			if tt.wantErr {
				if err == nil {
					t.Fatal("NewPresenter() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewPresenter() error = %v", err)
			}
			if presenter == nil {
				t.Fatal("NewPresenter() = nil")
			}
		})
	}
}

func TestNewTextPresenter(t *testing.T) {
	presenter := NewTextPresenter()
	if presenter == nil {
		t.Fatal("NewTextPresenter() = nil")
	}
}

func TestNewTOMLPresenter(t *testing.T) {
	presenter := NewTOMLPresenter()
	if presenter == nil {
		t.Fatal("NewTOMLPresenter() = nil")
	}
}

func TestPresenter_WriteLangs(t *testing.T) {
	tests := []struct {
		name   string
		format string
		langs  []string
		want   []string
	}{
		{
			name:   "text format",
			format: "text",
			langs:  []string{"cpp", "go", "rust"},
			want:   []string{"cpp\ngo\nrust\n"},
		},
		{
			name:   "toml format",
			format: "toml",
			langs:  []string{"cpp", "go", "rust"},
			want:   []string{"languages = ['cpp', 'go', 'rust']", "languages = [\"cpp\", \"go\", \"rust\"]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter, err := NewPresenter(tt.format)
			if err != nil {
				t.Fatalf("NewPresenter() error = %v", err)
			}

			var buf bytes.Buffer
			if err := presenter.WriteLangs(&buf, tt.langs); err != nil {
				t.Fatalf("WriteLangs() error = %v", err)
			}

			got := buf.String()
			matched := false
			for _, want := range tt.want {
				if strings.Contains(got, want) || got == want {
					matched = true
					break
				}
			}
			if !matched {
				t.Fatalf("WriteLangs() = %q, want one of %v", got, tt.want)
			}
		})
	}
}

func TestPresenter_WriteSummaries(t *testing.T) {
	summaries := []catalog.Summary{{
		Name:            "go",
		Description:     "Go starter",
		ManifestVersion: 2,
		InputNames:      []string{"module_path"},
		FileCount:       3,
		TemplateCount:   2,
		Variables:       []string{"ModulePath"},
	}}

	tests := []struct {
		name   string
		format string
		want   []string
	}{
		{
			name:   "text format",
			format: "text",
			want:   []string{"go", `desc="Go starter"`, "manifest=v2", "inputs=[module_path]", "files=3", "templates=2", "vars=[ModulePath]"},
		},
		{
			name:   "toml format",
			format: "toml",
			want:   []string{"[[templates]]", "name = 'go'", "manifest_version = 2", "input_names = ['module_path']"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter, err := NewPresenter(tt.format)
			if err != nil {
				t.Fatalf("NewPresenter() error = %v", err)
			}

			var buf bytes.Buffer
			if err := presenter.WriteSummaries(&buf, summaries); err != nil {
				t.Fatalf("WriteSummaries() error = %v", err)
			}

			got := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(got, want) && !strings.Contains(got, strings.ReplaceAll(want, "'", `"`)) {
					t.Fatalf("WriteSummaries() output = %q, want contains %q", got, want)
				}
			}
		})
	}
}

func TestPresenter_WriteInspection(t *testing.T) {
	inspection := catalog.Inspection{
		Name:            "go",
		Description:     "Go starter",
		ManifestVersion: 2,
		Inputs:          []scaffold.ManifestInput{{Name: "module_path", TemplateVar: "ModulePath"}},
		FileCount:       3,
		TemplateCount:   2,
		Variables:       []string{"ModulePath"},
		Mode:            catalog.InspectModeRender,
		ShownCount:      1,
		Files:           []catalog.FileDetail{{Source: "go.mod.tmpl", Output: "go.mod", IsTemplate: true}},
	}

	tests := []struct {
		name   string
		format string
		want   []string
	}{
		{
			name:   "text format",
			format: "text",
			want:   []string{"name: go", "description: Go starter", "manifest_version: 2", "inputs: module_path->ModulePath", "shown: 1", "- go.mod.tmpl -> go.mod (render)"},
		},
		{
			name:   "toml format",
			format: "toml",
			want:   []string{"name = 'go'", "shown_count = 1", "source = 'go.mod.tmpl'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter, err := NewPresenter(tt.format)
			if err != nil {
				t.Fatalf("NewPresenter() error = %v", err)
			}

			var buf bytes.Buffer
			if err := presenter.WriteInspection(&buf, inspection); err != nil {
				t.Fatalf("WriteInspection() error = %v", err)
			}

			got := buf.String()
			for _, want := range tt.want {
				if !strings.Contains(got, want) && !strings.Contains(got, strings.ReplaceAll(want, "'", `"`)) {
					t.Fatalf("WriteInspection() output = %q, want contains %q", got, want)
				}
			}
		})
	}
}
