//go:generate go run gen/gen_permissions.go

package templatesrc

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

// lookupMode takes the metadata table as a parameter so tests can drive it with
// a table of their own; ModeForPath binds it to the generated one.
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
