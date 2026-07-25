package templatefs

import (
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

func TestRenderTemplate(t *testing.T) {
	vars := domain.NewTemplateVars("Demo", "example.com/demo", "1.25", "alice", 2030)

	t.Run("renders known fields", func(t *testing.T) {
		got, err := renderTemplate([]byte("module {{.ModulePath}} by {{.Author}}"), vars)
		if err != nil {
			t.Fatalf("renderTemplate() error = %v", err)
		}
		if string(got) != "module example.com/demo by alice" {
			t.Fatalf("renderTemplate() = %q, want %q", string(got), "module example.com/demo by alice")
		}
	})

	t.Run("rejects invalid template", func(t *testing.T) {
		if _, err := renderTemplate([]byte("{{.ProjectName"), vars); err == nil {
			t.Fatal("renderTemplate() expected error, got nil")
		}
	})
}

func TestWalkEntries_SkipsReservedManifestAndRendersPaths(t *testing.T) {
	fsys := fstest.MapFS{
		"lang/.project-template-manifest.toml":        {Data: []byte("version = 2\n")},
		"lang/README.md":                              {Data: []byte("# README\n")},
		"lang/cmd":                                    {Mode: fs.ModeDir | 0o755},
		"lang/cmd/{{.ProjectNameLower}}":              {Mode: fs.ModeDir | 0o755},
		"lang/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n")},
	}
	vars := domain.NewTemplateVars("Demo", "example.com/demo", "1.25", "alice", 2030)

	var got []string
	err := WalkEntries(fsys, "lang", "demo", vars, func(entry Entry) error {
		kind := "file"
		if entry.IsDir {
			kind = "dir"
		}
		got = append(got, kind+":"+entry.Destination)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkEntries() error = %v", err)
	}

	want := []string{
		"file:" + filepath.Join("demo", "README.md"),
		"dir:" + filepath.Join("demo", "cmd"),
		"dir:" + filepath.Join("demo", "cmd", "demo"),
		"file:" + filepath.Join("demo", "cmd", "demo", "main.go"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("WalkEntries() = %v, want %v", got, want)
	}
}

func TestWalkEntries_InvalidRenderedPathFails(t *testing.T) {
	fsys := fstest.MapFS{
		"lang/{{.ModulePath}}.tmpl": {Data: []byte("ignored")},
	}
	vars := domain.NewTemplateVars("Demo", "acme/demo", "1.25", "alice", 2030)

	err := WalkEntries(fsys, "lang", "demo", vars, func(Entry) error {
		return nil
	})
	if err == nil {
		t.Fatal("WalkEntries() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to render template path") {
		t.Fatalf("WalkEntries() error = %v, want wrapped path rendering error", err)
	}
}

func TestReadEntryAndRenderEntry(t *testing.T) {
	fsys := fstest.MapFS{
		"lang/hello.txt.tmpl": {Data: []byte("Hello, {{.ProjectName}}!")},
		"lang/raw.txt":        {Data: []byte("{{.ProjectName")},
	}
	vars := domain.NewTemplateVars("Demo", "example.com/demo", "1.25", "alice", 2030)

	entries := map[string]Entry{}
	err := WalkEntries(fsys, "lang", "demo", vars, func(entry Entry) error {
		entries[entry.Name] = entry
		return nil
	})
	if err != nil {
		t.Fatalf("WalkEntries() error = %v", err)
	}

	loaded, err := ReadEntry(fsys, entries["hello.txt.tmpl"])
	if err != nil {
		t.Fatalf("ReadEntry(template) error = %v", err)
	}
	rendered, err := RenderEntry(loaded, vars)
	if err != nil {
		t.Fatalf("RenderEntry(template) error = %v", err)
	}
	if string(rendered) != "Hello, Demo!" {
		t.Fatalf("RenderEntry(template) = %q, want %q", string(rendered), "Hello, Demo!")
	}

	loaded, err = ReadEntry(fsys, entries["raw.txt"])
	if err != nil {
		t.Fatalf("ReadEntry(raw) error = %v", err)
	}
	rendered, err = RenderEntry(loaded, vars)
	if err != nil {
		t.Fatalf("RenderEntry(raw) error = %v", err)
	}
	if string(rendered) != "{{.ProjectName" {
		t.Fatalf("RenderEntry(raw) = %q, want %q", string(rendered), "{{.ProjectName")
	}
}
