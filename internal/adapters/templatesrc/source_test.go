package templatesrc

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	source := New()
	if source == nil {
		t.Fatal("New() = nil")
	}
	if source.FS() == nil {
		t.Fatal("Source.FS() = nil")
	}
}

func TestModeForPath_CoversAllTemplatePaths(t *testing.T) {
	t.Parallel()

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
		slices.Sort(missing)
		t.Fatalf("mode metadata missing paths: %v", missing)
	}
}

func TestModeForPath_KnownModes(t *testing.T) {
	t.Parallel()

	// Every expectation below is derivable from git's mode for the path:
	// directories and executables are 0o755, everything else 0o644. Before
	// canonicalization these were umask fossils (0o664, 0o775, and one 0o764).
	tests := map[string]fs.FileMode{
		"cpp/.github/workflows/ci.yml":             0o644,
		"cpp/CONTRIBUTING.md.tmpl":                 0o644,
		"cpp/typos.toml":                           0o644,
		"cpp/dev-tools/apply-format":               0o755,
		"cpp/dev-tools/git-pre-commit-format":      0o755,
		"go/.github/dependabot.yml":                0o644,
		"go/.goreleaser.yml.tmpl":                  0o644,
		"go/.project-template-manifest.toml":       0o644,
		"go/typos.toml":                            0o644,
		"go/cmd/{{.ProjectNameLower}}":             0o755,
		"go/internal/version/version_test.go.tmpl": 0o644,
		"rust/.project-template-manifest.toml":     0o644,
		"rust/dprint.json":                         0o644,
		"rust/Cargo.toml.tmpl":                     0o644,
		"rust/justfile.tmpl":                       0o755,
		"rust/typos.toml":                          0o644,
		"rust/src":                                 0o755,
	}

	for path, want := range tests {
		if got, ok := ModeForPath(path); !ok {
			t.Errorf("ModeForPath(%q) missing metadata", path)
		} else if got != want {
			t.Errorf("ModeForPath(%q) = %o, want %o", path, got, want)
		}
	}
}

func TestTemplateModeMetadata_OnlyReferencesEmbeddedPaths(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	walkedPaths := make(map[string]struct{})
	for _, dir := range templateDirs {
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() && isEmptyDir(path) {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			normalizedPath := filepath.ToSlash(path)
			walkedPaths[normalizedPath] = struct{}{}
			generatedMode, ok := templateModeMetadata[normalizedPath]
			if !ok {
				mismatches = append(mismatches, normalizedPath+" (missing in metadata)")
				return nil
			}
			if canonicalMode(info) != generatedMode {
				mismatches = append(mismatches, normalizedPath)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("WalkDir(%q) error = %v", dir, err)
		}
	}

	for sourcePath := range templateModeMetadata {
		normalizedPath := filepath.ToSlash(sourcePath)
		if _, ok := walkedPaths[normalizedPath]; !ok {
			mismatches = append(mismatches, normalizedPath+" (extra in metadata)")
		}
	}

	if len(mismatches) > 0 {
		slices.Sort(mismatches)
		t.Fatalf("permissions_generated.go is stale, run: go generate ./internal/adapters/templatesrc/\nmismatched paths: %v", mismatches)
	}
}

// canonicalMode mirrors the rule in gen/gen_permissions.go. The two copies are
// deliberate: importing this package from the generator would mean `go generate`
// could not run whenever permissions_generated.go is missing or malformed, which
// is exactly when it needs to run. TestGeneratedModesAreCanonical guards the drift.
func canonicalMode(info fs.FileInfo) fs.FileMode {
	const ownerExecutable = 0o100

	if info.IsDir() || info.Mode().Perm()&ownerExecutable != 0 {
		return 0o755
	}

	return 0o644
}

// TestGeneratedModesAreCanonical pins the property that made the table
// reproducible: every recorded mode is one of the two values git can represent.
// Any 0o664/0o775/0o764 entry means a umask leaked into `go generate` again.
func TestGeneratedModesAreCanonical(t *testing.T) {
	t.Parallel()

	for sourcePath, mode := range templateModeMetadata {
		if mode != 0o644 && mode != 0o755 {
			t.Errorf("templateModeMetadata[%q] = 0o%03o, want 0o644 or 0o755", sourcePath, mode)
		}
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
