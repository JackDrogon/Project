package scaffold

import (
	"bytes"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestListTemplateSummaries(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":       {Data: []byte("module {{.ModulePath}}")},
		"go/main.go.tmpl":      {Data: []byte("package main\n")},
		"go/.gitignore":        {Data: []byte("bin/")},
		"cpp/src/main.cc.tmpl": {Data: []byte("// {{.ProjectName}}")},
		"cpp/README.md.tmpl":   {Data: []byte("By {{.Author}}")},
	}

	c := NewCreator(fsys, &bytes.Buffer{})
	summaries, err := c.ListTemplateSummaries()
	if err != nil {
		t.Fatalf("ListTemplateSummaries() error = %v", err)
	}

	if len(summaries) != 2 {
		t.Fatalf("len(summaries) = %d, want 2", len(summaries))
	}

	want := map[string]TemplateSummary{
		"cpp": {
			Name:          "cpp",
			FileCount:     2,
			TemplateCount: 2,
			Variables:     []string{"Author", "ProjectName"},
		},
		"go": {
			Name:          "go",
			FileCount:     3,
			TemplateCount: 2,
			Variables:     []string{"ModulePath"},
		},
	}

	for _, got := range summaries {
		expected, ok := want[got.Name]
		if !ok {
			t.Fatalf("unexpected summary: %q", got.Name)
		}
		if got.FileCount != expected.FileCount {
			t.Fatalf("%s FileCount = %d, want %d", got.Name, got.FileCount, expected.FileCount)
		}
		if got.TemplateCount != expected.TemplateCount {
			t.Fatalf("%s TemplateCount = %d, want %d", got.Name, got.TemplateCount, expected.TemplateCount)
		}
		if !reflect.DeepEqual(got.Variables, expected.Variables) {
			t.Fatalf("%s Variables = %v, want %v", got.Name, got.Variables, expected.Variables)
		}
	}
}

func TestInspectLang(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":  {Data: []byte("module {{.ModulePath}}")},
		"go/main.go.tmpl": {Data: []byte("// {{.ProjectName}}")},
		"go/.gitignore":   {Data: []byte("bin/")},
	}

	c := NewCreator(fsys, &bytes.Buffer{})
	details, err := c.InspectLang("go")
	if err != nil {
		t.Fatalf("InspectLang() error = %v", err)
	}

	if details.Name != "go" {
		t.Fatalf("details.Name = %q, want %q", details.Name, "go")
	}
	if details.FileCount != 3 {
		t.Fatalf("details.FileCount = %d, want %d", details.FileCount, 3)
	}
	if details.TemplateCount != 2 {
		t.Fatalf("details.TemplateCount = %d, want %d", details.TemplateCount, 2)
	}
	if !reflect.DeepEqual(details.Variables, []string{"ModulePath", "ProjectName"}) {
		t.Fatalf("details.Variables = %v, want %v", details.Variables, []string{"ModulePath", "ProjectName"})
	}

	if len(details.Files) != 3 {
		t.Fatalf("len(details.Files) = %d, want %d", len(details.Files), 3)
	}

	wantFiles := []TemplateFile{
		{Source: ".gitignore", Output: ".gitignore", IsTemplate: false},
		{Source: "go.mod.tmpl", Output: "go.mod", IsTemplate: true},
		{Source: "main.go.tmpl", Output: "main.go", IsTemplate: true},
	}

	if !reflect.DeepEqual(details.Files, wantFiles) {
		t.Fatalf("details.Files = %#v, want %#v", details.Files, wantFiles)
	}
}

func TestInspectLang_InvalidTemplateReturnsError(t *testing.T) {
	fsys := fstest.MapFS{
		"go/bad.txt.tmpl": {Data: []byte("{{.ProjectName")},
	}

	c := NewCreator(fsys, &bytes.Buffer{})
	if _, err := c.InspectLang("go"); err == nil {
		t.Fatal("InspectLang() expected error for invalid template, got nil")
	}
}
