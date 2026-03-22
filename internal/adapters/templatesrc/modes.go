package templatesrc

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

var templateModeMetadata = map[string]fs.FileMode{
	"cpp":                                       0o755,
	"cpp/.project-template-manifest.toml":       0o644,
	"cpp/CHANGELOG":                             0o644,
	"cpp/CMakeLists.txt.tmpl":                   0o644,
	"cpp/CPPLINT.cfg":                           0o644,
	"cpp/README.md.tmpl":                        0o644,
	"cpp/dev-tools":                             0o755,
	"cpp/dev-tools/.gitkeep":                    0o644,
	"cpp/dev-tools/apply-format":                0o755,
	"cpp/dev-tools/check-style.py":              0o755,
	"cpp/dev-tools/cpplint.py":                  0o755,
	"cpp/dev-tools/git-pre-commit-format":       0o755,
	"cpp/include":                               0o755,
	"cpp/include/.gitkeep":                      0o644,
	"cpp/justfile.tmpl":                         0o644,
	"cpp/src":                                   0o755,
	"cpp/src/main.cc.tmpl":                      0o644,
	"go":                                        0o755,
	"go/.github":                                0o755,
	"go/.github/workflows":                      0o755,
	"go/.github/workflows/ci.yml":               0o644,
	"go/.gitignore":                             0o644,
	"go/.golangci.yml":                          0o644,
	"go/.goreleaser.yml.tmpl":                   0o644,
	"go/.project-template-manifest.toml":        0o644,
	"go/CODE_OF_CONDUCT.md.tmpl":                0o644,
	"go/CONTRIBUTING.md.tmpl":                   0o644,
	"go/README.md.tmpl":                         0o644,
	"go/cmd":                                    0o755,
	"go/cmd/{{.ProjectNameLower}}":              0o755,
	"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": 0o644,
	"go/codecov.yml":                            0o644,
	"go/go.mod.tmpl":                            0o644,
	"go/internal":                               0o755,
	"go/internal/app":                           0o755,
	"go/internal/app/app.go.tmpl":               0o644,
	"go/internal/app/app_test.go.tmpl":          0o644,
	"go/internal/version":                       0o755,
	"go/internal/version/version.go.tmpl":       0o644,
	"go/internal/version/version_test.go.tmpl":  0o644,
	"go/justfile.tmpl":                          0o644,
}

func LookupMode(path string) (fs.FileMode, bool) {
	return lookupMode(templateModeMetadata, path)
}

func lookupMode(metadata map[string]fs.FileMode, sourcePath string) (fs.FileMode, bool) {
	normalized := normalizeSourcePath(sourcePath)
	mode, ok := metadata[normalized]
	return mode, ok
}

func normalizeSourcePath(sourcePath string) string {
	cleaned := strings.ReplaceAll(strings.TrimSpace(sourcePath), `\\`, "/")
	cleaned = filepath.ToSlash(cleaned)
	cleaned = path.Clean(cleaned)
	if cleaned == "." {
		return cleaned
	}

	return strings.TrimPrefix(cleaned, "./")
}
