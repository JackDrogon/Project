package scaffold

import (
	"bytes"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadTemplateManifest_ValidatesReservedFilenameAndSchemaVersion(t *testing.T) {
	t.Run("loads valid manifest from reserved filename", func(t *testing.T) {
		fsys := fstest.MapFS{
			"go/.project-template.json": {Data: []byte(`{"schema_version":1,"name":"go","inputs":[{"name":"module","template_var":"ModulePath"}]}`)},
		}

		manifest, found, err := loadTemplateManifest(fsys, "go")
		if err != nil {
			t.Fatalf("loadTemplateManifest() error = %v", err)
		}
		if !found {
			t.Fatal("loadTemplateManifest() found = false, want true")
		}

		want := TemplateManifest{
			SchemaVersion: 1,
			Name:          "go",
			Inputs:        []TemplateManifestInput{{Name: "module", TemplateVar: "ModulePath"}},
		}
		if !reflect.DeepEqual(manifest, want) {
			t.Fatalf("loadTemplateManifest() = %#v, want %#v", manifest, want)
		}
	})

	t.Run("rejects non reserved filename", func(t *testing.T) {
		_, err := decodeTemplateManifest([]byte(`{"schema_version":1,"name":"go"}`), "go/manifest.json", "go")
		if err == nil {
			t.Fatal("decodeTemplateManifest() expected reserved filename error, got nil")
		}
		if !strings.Contains(err.Error(), "must use reserved filename") {
			t.Fatalf("decodeTemplateManifest() error = %v, want reserved filename error", err)
		}
	})

	t.Run("rejects unsupported schema version", func(t *testing.T) {
		fsys := fstest.MapFS{
			"go/.project-template.json": {Data: []byte(`{"schema_version":2,"name":"go"}`)},
		}

		_, _, err := loadTemplateManifest(fsys, "go")
		if err == nil {
			t.Fatal("loadTemplateManifest() expected schema version error, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported schema_version 2") {
			t.Fatalf("loadTemplateManifest() error = %v, want schema version error", err)
		}
	})

	t.Run("rejects mismatched template name", func(t *testing.T) {
		fsys := fstest.MapFS{
			"go/.project-template.json": {Data: []byte(`{"schema_version":1,"name":"cpp"}`)},
		}

		_, _, err := loadTemplateManifest(fsys, "go")
		if err == nil {
			t.Fatal("loadTemplateManifest() expected name mismatch error, got nil")
		}
		if !strings.Contains(err.Error(), `must match template directory "go"`) {
			t.Fatalf("loadTemplateManifest() error = %v, want name mismatch error", err)
		}
	})
}

func TestLoadTemplateManifest_RejectsUnknownFieldsAndUnsupportedTemplateVars(t *testing.T) {
	t.Run("rejects unknown fields", func(t *testing.T) {
		fsys := fstest.MapFS{
			"go/.project-template.json": {Data: []byte(`{"schema_version":1,"name":"go","unknown":true}`)},
		}

		_, _, err := loadTemplateManifest(fsys, "go")
		if err == nil {
			t.Fatal("loadTemplateManifest() expected unknown field error, got nil")
		}
		if !strings.Contains(err.Error(), "unknown field \"unknown\"") {
			t.Fatalf("loadTemplateManifest() error = %v, want unknown field error", err)
		}
	})

	t.Run("rejects unsupported template vars", func(t *testing.T) {
		fsys := fstest.MapFS{
			"go/.project-template.json": {Data: []byte(`{"schema_version":1,"name":"go","inputs":[{"name":"project","template_var":"ProjectName"}]}`)},
		}

		_, _, err := loadTemplateManifest(fsys, "go")
		if err == nil {
			t.Fatal("loadTemplateManifest() expected template_var error, got nil")
		}
		if !strings.Contains(err.Error(), `unsupported template_var "ProjectName"`) {
			t.Fatalf("loadTemplateManifest() error = %v, want template_var error", err)
		}
	})

	t.Run("rejects duplicate input names", func(t *testing.T) {
		fsys := fstest.MapFS{
			"go/.project-template.json": {Data: []byte(`{"schema_version":1,"name":"go","inputs":[{"name":"module","template_var":"ModulePath"},{"name":"module","template_var":"Author"}]}`)},
		}

		_, _, err := loadTemplateManifest(fsys, "go")
		if err == nil {
			t.Fatal("loadTemplateManifest() expected duplicate input error, got nil")
		}
		if !strings.Contains(err.Error(), `duplicate input name "module"`) {
			t.Fatalf("loadTemplateManifest() error = %v, want duplicate input error", err)
		}
	})
}

func TestTemplateWalkers_SkipReservedManifestFile(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template.json": {Data: []byte(`{"schema_version":1,"name":"go"}`)},
		"go/README.md.tmpl":         {Data: []byte("# {{.ProjectName}}\n")},
		"go/sub/nested.txt":         {Data: []byte("plain")},
	}

	vars := TemplateVars{ProjectName: "Demo", ProjectNameLower: "demo"}

	var entryPaths []string
	err := walkTemplateEntries(fsys, "go", "demo", vars, func(entry templateEntry) error {
		entryPaths = append(entryPaths, filepath.ToSlash(entry.srcPath))
		return nil
	})
	if err != nil {
		t.Fatalf("walkTemplateEntries() error = %v", err)
	}
	if slices.Contains(entryPaths, "go/.project-template.json") {
		t.Fatalf("walkTemplateEntries() visited reserved manifest: %v", entryPaths)
	}

	creator := NewCreator(fsys, &bytes.Buffer{})
	var filePaths []string
	err = creator.walkTemplateFiles("go", func(srcPath string, isTemplate bool) error {
		filePaths = append(filePaths, filepath.ToSlash(srcPath))
		return nil
	})
	if err != nil {
		t.Fatalf("walkTemplateFiles() error = %v", err)
	}
	if slices.Contains(filePaths, "go/.project-template.json") {
		t.Fatalf("walkTemplateFiles() visited reserved manifest: %v", filePaths)
	}

	if len(filePaths) != 2 {
		t.Fatalf("len(filePaths) = %d, want 2", len(filePaths))
	}

	var preview bytes.Buffer
	if err := PreviewEmbedDir(&preview, fsys, "go", "demo", vars); err != nil {
		t.Fatalf("PreviewEmbedDir() error = %v", err)
	}
	if strings.Contains(preview.String(), templateManifestFilename) {
		t.Fatalf("PreviewEmbedDir() output = %q, want reserved manifest skipped", preview.String())
	}

	details, err := creator.InspectLang("go")
	if err != nil {
		t.Fatalf("InspectLang() error = %v", err)
	}
	if details.FileCount != 2 {
		t.Fatalf("InspectLang() FileCount = %d, want 2", details.FileCount)
	}
	for _, file := range details.Files {
		if file.Source == templateManifestFilename {
			t.Fatalf("InspectLang() included reserved manifest in files: %#v", details.Files)
		}
	}
}
