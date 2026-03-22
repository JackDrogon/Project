package templatesrc

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestNew(t *testing.T) {
	source := New()
	if source == nil {
		t.Fatal("New() = nil")
	}
	if source.FS() == nil {
		t.Fatal("Source.FS() = nil")
	}
}

func TestModeForPath_CoversAllTemplatePaths(t *testing.T) {
	source := New()
	var missing []string
	err := fs.WalkDir(source.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		if _, ok := ModeForPath(filepath.ToSlash(path)); !ok {
			missing = append(missing, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir() error = %v", err)
	}
	if len(missing) != 0 {
		sort.Strings(missing)
		t.Fatalf("mode metadata missing paths: %v", missing)
	}
}

func TestModeForPath_KnownModes(t *testing.T) {
	tests := map[string]fs.FileMode{
		"cpp/dev-tools/apply-format":               0o755,
		"cpp/dev-tools/git-pre-commit-format":      0o755,
		"go/.goreleaser.yml.tmpl":                  0o664,
		"go/.project-template-manifest.toml":       0o664,
		"go/cmd/{{.ProjectNameLower}}":             0o775,
		"go/internal/version/version_test.go.tmpl": 0o664,
	}

	for path, want := range tests {
		if got, ok := ModeForPath(path); !ok {
			t.Fatalf("ModeForPath(%q) missing metadata", path)
		} else if got != want {
			t.Fatalf("ModeForPath(%q) = %o, want %o", path, got, want)
		}
	}
}

func TestTemplateModeMetadata_OnlyReferencesEmbeddedPaths(t *testing.T) {
	source := New()
	for sourcePath := range templateModeMetadata {
		if _, err := fs.Stat(source.FS(), sourcePath); err != nil {
			if os.IsNotExist(err) {
				t.Fatalf("templateModeMetadata contains missing path %q", sourcePath)
			}
			t.Fatalf("Stat(%q) error = %v", sourcePath, err)
		}
	}
	if len(templateModeMetadata) == 0 {
		t.Fatal("templateModeMetadata is empty")
	}
}

func TestLookupMode_UsesExplicitMetadataTableOnly(t *testing.T) {
	metadata := map[string]fs.FileMode{
		"lang/script.sh": 0o755,
		"lang/README.md": 0o644,
	}

	if got, ok := lookupMode(metadata, `./lang\\script.sh`); !ok {
		t.Fatal("lookupMode() did not find explicit metadata entry")
	} else if got != 0o755 {
		t.Fatalf("lookupMode() = %o, want %o", got, fs.FileMode(0o755))
	}

	if _, ok := lookupMode(metadata, "lang/missing.sh"); ok {
		t.Fatal("lookupMode() unexpectedly inferred metadata for missing path")
	}
}
