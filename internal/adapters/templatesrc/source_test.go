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
		"cpp/.github/workflows/ci.yml":             0o664,
		"cpp/CONTRIBUTING.md.tmpl":                 0o664,
		"cpp/typos.toml":                           0o664,
		"cpp/dev-tools/apply-format":               0o755,
		"cpp/dev-tools/git-pre-commit-format":      0o755,
		"go/.github/dependabot.yml":                0o664,
		"go/.goreleaser.yml.tmpl":                  0o664,
		"go/.project-template-manifest.toml":       0o664,
		"go/typos.toml":                            0o664,
		"go/cmd/{{.ProjectNameLower}}":             0o775,
		"go/internal/version/version_test.go.tmpl": 0o664,
		"rust/.project-template-manifest.toml":     0o664,
		"rust/dprint.json":                         0o664,
		"rust/Cargo.toml.tmpl":                     0o664,
		"rust/justfile.tmpl":                       0o764,
		"rust/typos.toml":                          0o664,
		"rust/src":                                 0o775,
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

func TestPermissionsGeneratedUpToDate(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	var templateDirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "gen" || name[0] == '.' {
			continue
		}
		templateDirs = append(templateDirs, name)
	}

	var mismatches []string
	for _, dir := range templateDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && isEmptyDir(path) {
				return nil
			}
			normalizedPath := filepath.ToSlash(path)
			actualMode := info.Mode().Perm()
			generatedMode, ok := templateModeMetadata[normalizedPath]
			if !ok {
				mismatches = append(mismatches, normalizedPath+" (missing in metadata)")
				return nil
			}
			if actualMode != generatedMode {
				mismatches = append(mismatches, normalizedPath)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Walk(%q) error = %v", dir, err)
		}
	}

	if len(mismatches) > 0 {
		sort.Strings(mismatches)
		t.Fatalf("permissions_generated.go is stale, run: go generate ./internal/adapters/templatesrc/\nmismatched paths: %v", mismatches)
	}
}

func isEmptyDir(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			return false
		}
		subPath := filepath.Join(path, entry.Name())
		if !isEmptyDir(subPath) {
			return false
		}
	}
	return true
}
