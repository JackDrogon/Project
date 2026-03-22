package protocoltoml

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestManifestV2_LoadAndValidate(t *testing.T) {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n")},
	}

	manifest, found, err := LoadManifest(fsys, "go")
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	if !found {
		t.Fatal("LoadManifest() found = false, want true")
	}
	if manifest.Version != 2 || manifest.Name != "go" || manifest.Description != "Go starter" {
		t.Fatalf("manifest = %#v, want TOML manifest metadata", manifest)
	}
	if len(manifest.Inputs) != 1 || manifest.Inputs[0].Key != "module_path" {
		t.Fatalf("manifest.Inputs = %#v, want module_path input", manifest.Inputs)
	}
}

func TestManifestV2_RejectsLegacyJSON(t *testing.T) {
	_, err := DecodeManifest([]byte(`{"schema_version":1,"name":"go"}`), "go/.project-template-manifest.toml", "go")
	if err == nil {
		t.Fatal("DecodeManifest() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "legacy JSON") {
		t.Fatalf("DecodeManifest() error = %v, want legacy JSON rejection", err)
	}
}

func TestManifestV2_RejectsUnknownFieldsAndInvalidInputs(t *testing.T) {
	t.Run("unknown field", func(t *testing.T) {
		_, err := DecodeManifest([]byte("version = 2\nname = \"go\"\nunknown = true\n"), "go/.project-template-manifest.toml", "go")
		if err == nil {
			t.Fatal("DecodeManifest() expected error, got nil")
		}
	})

	t.Run("duplicate key", func(t *testing.T) {
		content := []byte("version = 2\nname = \"go\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"Author\"\n")
		_, err := DecodeManifest(content, "go/.project-template-manifest.toml", "go")
		if err == nil {
			t.Fatal("DecodeManifest() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "duplicate input key") {
			t.Fatalf("DecodeManifest() error = %v, want duplicate input error", err)
		}
	})
}
