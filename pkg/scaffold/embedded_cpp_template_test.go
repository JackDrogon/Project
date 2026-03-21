package scaffold

import (
	"reflect"
	"testing"

	"github.com/JackDrogon/project/pkg/templates"
)

func TestEmbeddedCPPTemplateManifestMetadata(t *testing.T) {
	manifest, found, err := loadTemplateManifest(templates.FS, "cpp")
	if err != nil {
		t.Fatalf("loadTemplateManifest(cpp) error = %v", err)
	}
	if !found {
		t.Fatal("loadTemplateManifest(cpp) found = false, want true")
	}

	if manifest.SchemaVersion != 1 {
		t.Fatalf("manifest.SchemaVersion = %d, want %d", manifest.SchemaVersion, 1)
	}
	if manifest.Name != "cpp" {
		t.Fatalf("manifest.Name = %q, want %q", manifest.Name, "cpp")
	}
	if manifest.Description != "CMake-based C++ starter" {
		t.Fatalf("manifest.Description = %q, want %q", manifest.Description, "CMake-based C++ starter")
	}

	wantInputs := []TemplateManifestInput{{Name: "author", TemplateVar: "Author"}}
	if !reflect.DeepEqual(manifest.Inputs, wantInputs) {
		t.Fatalf("manifest.Inputs = %#v, want %#v", manifest.Inputs, wantInputs)
	}
}
