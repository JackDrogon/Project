package templatefs

import (
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

func TestCollectDetails(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml":        {Data: []byte("{}")},
		"go/.gitignore":                             {Data: []byte("bin/\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
		"go/cmd/{{.ProjectNameLower}}":              {Mode: fs.ModeDir | 0o755},
		"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n")},
	}
	manifest := Manifest{
		SchemaVersion: 1,
		Name:          "go",
		Description:   "Go starter",
		Inputs: []domain.ManifestInput{
			{Name: "module_path", TemplateVar: "ModulePath"},
			{Name: "go_version", TemplateVar: "GoVersion"},
		},
	}

	details, err := CollectDetails(fsys, "go", manifest)
	if err != nil {
		t.Fatalf("CollectDetails() error = %v", err)
	}

	if details.Name != "go" || details.Description != "Go starter" || details.ManifestVersion != 1 {
		t.Fatalf("CollectDetails() summary = %#v, want manifest metadata preserved", details.Summary)
	}
	if details.FileCount != 3 || details.TemplateCount != 2 {
		t.Fatalf("CollectDetails() counts = (%d, %d), want (3, 2)", details.FileCount, details.TemplateCount)
	}
	if !reflect.DeepEqual(details.InputNames, []string{"module_path", "go_version"}) {
		t.Fatalf("InputNames = %v, want module_path/go_version", details.InputNames)
	}
	if !reflect.DeepEqual(details.Variables, []string{"ModulePath", "ProjectNameLower"}) {
		t.Fatalf("Variables = %v, want ModulePath and ProjectNameLower", details.Variables)
	}

	wantFiles := []FileDetail{
		{Source: ".gitignore", Output: ".gitignore", IsTemplate: false},
		{Source: filepath.ToSlash(filepath.Join("cmd", "{{.ProjectNameLower}}", "main.go.tmpl")), Output: filepath.ToSlash(filepath.Join("cmd", "{{.ProjectNameLower}}", "main.go")), IsTemplate: true},
		{Source: "go.mod.tmpl", Output: "go.mod", IsTemplate: true},
	}
	if !reflect.DeepEqual(details.Files, wantFiles) {
		t.Fatalf("Files = %#v, want %#v", details.Files, wantFiles)
	}
}

func TestCollectDetails_RejectsInvalidTemplate(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"go/bad.txt.tmpl": {Data: []byte("{{.ProjectName")},
	}
	_, err := CollectDetails(fsys, "go", Manifest{Name: "go"})
	if err == nil {
		t.Fatal("CollectDetails() expected error, got nil")
	}
}
