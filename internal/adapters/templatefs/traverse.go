package templatefs

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

const TmplSuffix = ".tmpl"

var reservedTemplateFilenames = map[string]struct{}{
	".project-template-manifest.toml": {},
}

type Entry struct {
	SourcePath  string
	Destination string
	Name        string
	DirEntry    fs.DirEntry
	IsDir       bool
	IsTemplate  bool
	RawContents []byte
}

// RenderTemplate applies template variable substitution to the given content.
// It uses Go's text/template engine with strict error handling for missing keys.
// Returns the rendered content or an error if template execution fails.
func RenderTemplate(content []byte, vars domain.TemplateVars) ([]byte, error) {
	rendered, err := renderTemplateString(string(content), vars)
	if err != nil {
		return nil, err
	}
	return []byte(rendered), nil
}

func renderTemplateString(content string, vars domain.TemplateVars) (string, error) {
	tmpl, err := template.New("").Option("missingkey=error").Parse(content)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderPathSegment renders a single path component (file or directory name) with template variables.
// It validates that the result is a valid path segment without separators or special names.
// Returns an error if rendering produces an invalid path component.
func RenderPathSegment(name string, vars domain.TemplateVars) (string, error) {
	rendered, err := renderTemplateString(name, vars)
	if err != nil {
		return "", err
	}
	if rendered == "" || rendered == "." || rendered == ".." {
		return "", fmt.Errorf("invalid rendered path segment %q", rendered)
	}
	if strings.Contains(rendered, "/") || strings.Contains(rendered, `\`) {
		return "", fmt.Errorf("rendered path segment %q must not contain path separators", rendered)
	}

	return rendered, nil
}

// WalkEntries recursively traverses a template directory tree and invokes the visit function for each entry.
// It renders path segments with template variables and skips reserved template files.
// Returns an error if directory traversal or path rendering fails.
func WalkEntries(fsys fs.FS, srcDir, destDir string, vars domain.TemplateVars, visit func(Entry) error) error {
	entries, err := fs.ReadDir(fsys, srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if isReservedTemplateFile(entry.Name()) {
			continue
		}

		srcPath := path.Join(srcDir, entry.Name())
		destName, err := RenderPathSegment(strings.TrimSuffix(entry.Name(), TmplSuffix), vars)
		if err != nil {
			return fmt.Errorf("failed to render template path %s: %w", srcPath, err)
		}

		current := Entry{
			SourcePath:  srcPath,
			Destination: filepath.Join(destDir, destName),
			Name:        entry.Name(),
			DirEntry:    entry,
			IsDir:       entry.IsDir(),
			IsTemplate:  strings.HasSuffix(entry.Name(), TmplSuffix),
		}

		if err := visit(current); err != nil {
			return err
		}

		if current.IsDir {
			if err := WalkEntries(fsys, srcPath, current.Destination, vars, visit); err != nil {
				return err
			}
		}
	}

	return nil
}

// ReadEntry loads the file content for a given entry from the filesystem.
// It populates the RawContents field of the entry. Returns an error if the file cannot be read.
func ReadEntry(fsys fs.FS, entry Entry) (Entry, error) {
	content, err := fs.ReadFile(fsys, entry.SourcePath)
	if err != nil {
		return Entry{}, err
	}

	entry.RawContents = content
	return entry, nil
}

// RenderEntry processes an entry's content, applying template rendering if it's a .tmpl file.
// Non-template files are returned as-is. Returns an error if template rendering fails.
func RenderEntry(entry Entry, vars domain.TemplateVars) ([]byte, error) {
	if !entry.IsTemplate {
		return entry.RawContents, nil
	}

	rendered, err := RenderTemplate(entry.RawContents, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render template %s: %w", entry.SourcePath, err)
	}

	return rendered, nil
}

func isReservedTemplateFile(name string) bool {
	_, ok := reservedTemplateFilenames[path.Base(name)]
	return ok
}
