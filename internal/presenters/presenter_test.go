package presenters

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/JackDrogon/project/internal/app/catalog"
	"github.com/JackDrogon/project/internal/scaffold"
)

type stubSummaryWriter struct{ text string }

func (w stubSummaryWriter) WriteSummaries(out io.Writer, summaries []catalog.Summary) error {
	_, err := io.WriteString(out, w.text)
	return err
}

type stubInspectionWriter struct{ text string }

func (w stubInspectionWriter) WriteInspection(out io.Writer, inspection catalog.Inspection) error {
	_, err := io.WriteString(out, w.text)
	return err
}

type stubTextFormatterRegistry struct {
	summary    summaryWriter
	inspection inspectionWriter
}

func (r stubTextFormatterRegistry) SummaryWriter(SummaryViewSpec) (summaryWriter, error) {
	return r.summary, nil
}

func (r stubTextFormatterRegistry) InspectionWriter(InspectionViewSpec) (inspectionWriter, error) {
	return r.inspection, nil
}

type stubFormatterFactory struct{ formatter Formatter }

func (f stubFormatterFactory) Build(OutputSpec) (Formatter, error) {
	if f.formatter == nil {
		return nil, fmt.Errorf("missing formatter")
	}
	return f.formatter, nil
}

func TestNewPresenter(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		summary SummaryViewSpec
		inspect InspectionViewSpec
		wantErr bool
	}{
		{"text format", "text", DefaultSummaryViewSpec(), DefaultInspectionViewSpec(), false},
		{"compact text format", "text", SummaryViewSpec{TextLayout: TextLayoutCompact}, InspectionViewSpec{TextLayout: TextLayoutCompact}, false},
		{"toml format", "toml", DefaultSummaryViewSpec(), DefaultInspectionViewSpec(), false},
		{"compact toml rejected", "toml", SummaryViewSpec{TextLayout: TextLayoutCompact}, DefaultInspectionViewSpec(), true},
		{"invalid format", "json", DefaultSummaryViewSpec(), DefaultInspectionViewSpec(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter, err := NewPresenter(OutputSpec{Format: tt.format, Summary: tt.summary, Inspection: tt.inspect})
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

func TestNewCompactTextPresenter(t *testing.T) {
	presenter := NewCompactTextPresenter()
	if presenter == nil {
		t.Fatal("NewCompactTextPresenter() = nil")
	}
}

func TestNewTableTextPresenter(t *testing.T) {
	presenter := NewTableTextPresenter()
	if presenter == nil {
		t.Fatal("NewTableTextPresenter() = nil")
	}
}

func TestNewPresenter_RejectsInspectionTableLayout(t *testing.T) {
	_, err := NewPresenter(OutputSpec{
		Format:     "text",
		Summary:    DefaultSummaryViewSpec(),
		Inspection: InspectionViewSpec{TextLayout: TextLayoutTable},
	})
	if err == nil {
		t.Fatal("NewPresenter() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "table output is only supported for summary text views") {
		t.Fatalf("NewPresenter() error = %v, want inspection table error", err)
	}
}

func TestNewPresenterWithFactory_UsesInjectedFactory(t *testing.T) {
	presenter, err := NewPresenterWithFactory(OutputSpec{Format: "text"}, stubFormatterFactory{formatter: &tomlFormatter{}})
	if err != nil {
		t.Fatalf("NewPresenterWithFactory() error = %v", err)
	}
	if presenter == nil {
		t.Fatal("NewPresenterWithFactory() = nil")
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
			presenter, err := NewPresenter(OutputSpec{Format: tt.format, Summary: DefaultSummaryViewSpec(), Inspection: DefaultInspectionViewSpec()})
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
		RepoAssets:      []string{"ci", "typos"},
		RepoFileCount:   2,
		GovernanceTier:  "basic",
	}}

	tests := []struct {
		name   string
		format string
		want   []string
	}{
		{
			name:   "text format",
			format: "text",
			want:   []string{"go", `desc="Go starter"`, "manifest=v2", "inputs=[module_path]", "files=3", "templates=2", "vars=[ModulePath]", "repo=[ci, typos]", "repo_files=2", "governance=basic"},
		},
		{
			name:   "toml format",
			format: "toml",
			want:   []string{"[[templates]]", "name = 'go'", "manifest_version = 2", "input_names = ['module_path']", "repo_assets = ['ci', 'typos']", "repo_file_count = 2", "governance_tier = 'basic'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter, err := NewPresenter(OutputSpec{Format: tt.format, Summary: DefaultSummaryViewSpec(), Inspection: DefaultInspectionViewSpec()})
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
		RepoAssets:      []string{"ci", "typos"},
		Mode:            catalog.InspectModeRender,
		Files:           []catalog.InspectionFile{{Source: "go.mod.tmpl", Output: "go.mod", Action: catalog.FileActionRender, Group: catalog.FileGroupLanguage}},
	}

	tests := []struct {
		name   string
		format string
		want   []string
	}{
		{
			name:   "text format",
			format: "text",
			want:   []string{"name: go", "description: Go starter", "manifest_version: 2", "inputs: module_path->ModulePath", "repo_assets: ci, typos", "shown: 1", "repo_files:", "- (none)", "language_files:", "- go.mod.tmpl -> go.mod (render)"},
		},
		{
			name:   "toml format",
			format: "toml",
			want:   []string{"name = 'go'", "repo_assets = ['ci', 'typos']", "shown_count = 1", "[[language_files]]", "source = 'go.mod.tmpl'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter, err := NewPresenter(OutputSpec{Format: tt.format, Summary: DefaultSummaryViewSpec(), Inspection: DefaultInspectionViewSpec()})
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

func TestPresenter_WriteCompactTextOutputs(t *testing.T) {
	summaries := []catalog.Summary{{
		Name:            "go",
		Description:     "Go starter",
		ManifestVersion: 2,
		InputNames:      []string{"module_path"},
		FileCount:       3,
		TemplateCount:   2,
		Variables:       []string{"ModulePath"},
		RepoAssets:      []string{"ci", "typos"},
		RepoFileCount:   2,
		GovernanceTier:  "basic",
	}}
	inspection := catalog.Inspection{
		Name:            "go",
		Description:     "Go starter",
		ManifestVersion: 2,
		Inputs:          []scaffold.ManifestInput{{Name: "module_path", TemplateVar: "ModulePath"}},
		FileCount:       3,
		TemplateCount:   2,
		Variables:       []string{"ModulePath"},
		RepoAssets:      []string{"ci", "typos"},
		Mode:            catalog.InspectModeRender,
		Files:           []catalog.InspectionFile{{Source: "go.mod.tmpl", Output: "go.mod", Action: catalog.FileActionRender, Group: catalog.FileGroupLanguage}},
	}

	presenter := NewCompactTextPresenter()
	var summaryBuf bytes.Buffer
	if err := presenter.WriteSummaries(&summaryBuf, summaries); err != nil {
		t.Fatalf("WriteSummaries() error = %v", err)
	}
	summaryOut := summaryBuf.String()
	for _, want := range []string{"go [basic]", "counts: files=3 templates=2 repo_files=2 manifest=v2", "repo: ci, typos"} {
		if !strings.Contains(summaryOut, want) {
			t.Fatalf("compact summary output = %q, want contains %q", summaryOut, want)
		}
	}

	var inspectBuf bytes.Buffer
	if err := presenter.WriteInspection(&inspectBuf, inspection); err != nil {
		t.Fatalf("WriteInspection() error = %v", err)
	}
	inspectOut := inspectBuf.String()
	for _, want := range []string{"go — Go starter", "manifest: v2 | mode: render | shown: 1 | files: 3 | templates: 2", "repo_files: (none)", "language_files:", "go.mod.tmpl -> go.mod (render)"} {
		if !strings.Contains(inspectOut, want) {
			t.Fatalf("compact inspection output = %q, want contains %q", inspectOut, want)
		}
	}
}

func TestPresenter_WriteTableTextSummaries(t *testing.T) {
	summaries := []catalog.Summary{
		{Name: "go", RepoFileCount: 5, FileCount: 7, TemplateCount: 2, GovernanceTier: "standard", InputNames: []string{"module_path", "go_version"}, RepoAssets: []string{"ci", "typos"}},
		{Name: "rust", RepoFileCount: 8, FileCount: 17, TemplateCount: 7, GovernanceTier: "rich", InputNames: []string{"author", "year"}, RepoAssets: []string{"ci", "security", "typos"}},
	}
	presenter := NewTableTextPresenter()
	var buf bytes.Buffer
	if err := presenter.WriteSummaries(&buf, summaries); err != nil {
		t.Fatalf("WriteSummaries() error = %v", err)
	}
	got := buf.String()
	for _, want := range []string{"NAME", "GOVERNANCE", "REPO_FILES", "go", "rust", "standard", "rich", "module_path,go_version", "ci,security,typos"} {
		if !strings.Contains(got, want) {
			t.Fatalf("table summary output = %q, want contains %q", got, want)
		}
	}
}

func TestDefaultFormatterFactory_UsesInjectedTextRegistry(t *testing.T) {
	factory := defaultFormatterFactory{
		textRegistry: stubTextFormatterRegistry{
			summary:    stubSummaryWriter{text: "summary-from-registry"},
			inspection: stubInspectionWriter{text: "inspection-from-registry"},
		},
	}

	formatter, err := factory.Build(OutputSpec{Format: "text", Summary: DefaultSummaryViewSpec(), Inspection: DefaultInspectionViewSpec()})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	var summaryBuf bytes.Buffer
	if err := formatter.WriteSummaries(&summaryBuf, nil); err != nil {
		t.Fatalf("WriteSummaries() error = %v", err)
	}
	if got := summaryBuf.String(); got != "summary-from-registry" {
		t.Fatalf("WriteSummaries() = %q, want summary-from-registry", got)
	}

	var inspectBuf bytes.Buffer
	if err := formatter.WriteInspection(&inspectBuf, catalog.Inspection{}); err != nil {
		t.Fatalf("WriteInspection() error = %v", err)
	}
	if got := inspectBuf.String(); got != "inspection-from-registry" {
		t.Fatalf("WriteInspection() = %q, want inspection-from-registry", got)
	}
}
