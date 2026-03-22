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

func ReadEntry(fsys fs.FS, entry Entry) (Entry, error) {
	content, err := fs.ReadFile(fsys, entry.SourcePath)
	if err != nil {
		return Entry{}, err
	}

	entry.RawContents = content
	return entry, nil
}

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
