package presenters

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JackDrogon/project/internal/app/catalog"
	"github.com/JackDrogon/project/internal/scaffold"
)

func TestPresenter_WriteLangs(t *testing.T) {
	tests := []struct {
		name   string
		format string
		langs  []string
		want   []string
	}{
		{name: "text format", format: "text", langs: []string{"cpp", "go", "rust"}, want: []string{"cpp\ngo\nrust\n"}},
		{name: "toml format", format: "toml", langs: []string{"cpp", "go", "rust"}, want: []string{"languages = ['cpp', 'go', 'rust']", "languages = [\"cpp\", \"go\", \"rust\"]"}},
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
		{name: "text format", format: "text", want: []string{"go", `desc="Go starter"`, "manifest=v2", "inputs=[module_path]", "files=3", "templates=2", "vars=[ModulePath]", "repo=[ci, typos]", "repo_files=2", "governance=basic"}},
		{name: "toml format", format: "toml", want: []string{"[[templates]]", "name = 'go'", "manifest_version = 2", "input_names = ['module_path']", "repo_assets = ['ci', 'typos']", "repo_file_count = 2", "governance_tier = 'basic'"}},
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
		{name: "text format", format: "text", want: []string{"name: go", "description: Go starter", "manifest_version: 2", "inputs: module_path->ModulePath", "repo_assets: ci, typos", "shown: 1", "repo_files:", "- (none)", "language_files:", "- go.mod.tmpl -> go.mod (render)"}},
		{name: "toml format", format: "toml", want: []string{"name = 'go'", "repo_assets = ['ci', 'typos']", "shown_count = 1", "[[language_files]]", "source = 'go.mod.tmpl'"}},
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

	presenter, err := NewPresenter(OutputSpec{Format: "text", Summary: SummaryViewSpec{TextLayout: TextLayoutCompact}, Inspection: InspectionViewSpec{TextLayout: TextLayoutCompact}})
	if err != nil {
		t.Fatalf("NewPresenter() error = %v", err)
	}
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
	presenter, err := NewPresenter(OutputSpec{Format: "text", Summary: SummaryViewSpec{TextLayout: TextLayoutTable}, Inspection: DefaultInspectionViewSpec()})
	if err != nil {
		t.Fatalf("NewPresenter() error = %v", err)
	}
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
