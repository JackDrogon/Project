package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JackDrogon/project/internal/app/catalog"
	"github.com/JackDrogon/project/internal/domain/scaffold"
)

func TestNew(t *testing.T) {
	presenter := New()
	if presenter == nil {
		t.Fatal("New() = nil")
	}
}

func TestWriteTextLangs(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTextLangs(&buf, []string{"cpp", "go"}); err != nil {
		t.Fatalf("WriteTextLangs() error = %v", err)
	}
	if got := buf.String(); got != "cpp\ngo\n" {
		t.Fatalf("WriteTextLangs() = %q, want %q", got, "cpp\ngo\n")
	}
}

func TestWriteTextSummaries(t *testing.T) {
	var buf bytes.Buffer
	summaries := []catalog.Summary{{Name: "go", Description: "Go starter", ManifestVersion: 2, InputNames: []string{"module_path"}, FileCount: 3, TemplateCount: 2, Variables: []string{"ModulePath"}}}
	if err := WriteTextSummaries(&buf, summaries); err != nil {
		t.Fatalf("WriteTextSummaries() error = %v", err)
	}
	got := buf.String()
	for _, want := range []string{"go", `desc="Go starter"`, "manifest=v2", "inputs=[module_path]", "files=3", "templates=2", "vars=[ModulePath]"} {
		if !strings.Contains(got, want) {
			t.Fatalf("WriteTextSummaries() output = %q, want contains %q", got, want)
		}
	}
}

func TestWriteTextInspection(t *testing.T) {
	var buf bytes.Buffer
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
	if err := WriteTextInspection(&buf, inspection); err != nil {
		t.Fatalf("WriteTextInspection() error = %v", err)
	}
	got := buf.String()
	for _, want := range []string{"name: go", "description: Go starter", "manifest_version: 2", "inputs: module_path->ModulePath", "shown: 1", "- go.mod.tmpl -> go.mod (render)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("WriteTextInspection() output = %q, want contains %q", got, want)
		}
	}
}

func TestWriteTOMLPresentations(t *testing.T) {
	t.Run("langs", func(t *testing.T) {
		var buf bytes.Buffer
		if err := WriteTOMLLangs(&buf, []string{"cpp", "go"}); err != nil {
			t.Fatalf("WriteTOMLLangs() error = %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, "languages = ['cpp', 'go']") && !strings.Contains(got, "languages = [\"cpp\", \"go\"]") {
			t.Fatalf("WriteTOMLLangs() output = %q", got)
		}
	})

	t.Run("summaries", func(t *testing.T) {
		var buf bytes.Buffer
		summaries := []catalog.Summary{{Name: "cpp", Description: "C++ starter", ManifestVersion: 2, InputNames: []string{"author"}, FileCount: 2, TemplateCount: 2, Variables: []string{"Author", "ProjectName"}}}
		if err := WriteTOMLSummaries(&buf, summaries); err != nil {
			t.Fatalf("WriteTOMLSummaries() error = %v", err)
		}
		got := buf.String()
		for _, want := range []string{"[[templates]]", "name = 'cpp'", "manifest_version = 2", "input_names = ['author']"} {
			if !strings.Contains(got, want) && !strings.Contains(got, strings.ReplaceAll(want, "'", `"`)) {
				t.Fatalf("WriteTOMLSummaries() output = %q, want contains %q", got, want)
			}
		}
	})

	t.Run("inspection", func(t *testing.T) {
		var buf bytes.Buffer
		inspection := catalog.Inspection{
			Name:            "go",
			Description:     "Go starter",
			ManifestVersion: 2,
			Inputs:          []scaffold.ManifestInput{{Name: "module_path", TemplateVar: "ModulePath"}},
			FileCount:       3,
			TemplateCount:   2,
			Variables:       []string{"ModulePath"},
			Mode:            catalog.InspectModeCopy,
			ShownCount:      1,
			Files:           []catalog.FileDetail{{Source: ".gitignore", Output: ".gitignore", IsTemplate: false}},
		}
		if err := WriteTOMLInspection(&buf, inspection); err != nil {
			t.Fatalf("WriteTOMLInspection() error = %v", err)
		}
		got := buf.String()
		for _, want := range []string{"name = 'go'", "manifest_version = 2", "mode = 'copy'", "shown_count = 1", "[[files]]", "source = '.gitignore'"} {
			if !strings.Contains(got, want) && !strings.Contains(got, strings.ReplaceAll(want, "'", `"`)) {
				t.Fatalf("WriteTOMLInspection() output = %q, want contains %q", got, want)
			}
		}
	})
}
