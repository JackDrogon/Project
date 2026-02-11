package scaffold

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

const tmplSuffix = ".tmpl"

// RenderTemplate applies TemplateVars to content using text/template.
// If parsing fails (e.g. the file is not a Go template), the original
// content is returned unchanged.
func RenderTemplate(content []byte, vars TemplateVars) ([]byte, error) {
	tmpl, err := template.New("").Option("missingkey=zero").Parse(string(content))
	if err != nil {
		// Not a valid template — return as-is
		return content, nil
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return content, nil
	}
	return buf.Bytes(), nil
}

// CopyEmbedDir recursively copies a directory from an embedded filesystem
// to the local filesystem, rendering template variables in file contents.
func CopyEmbedDir(w io.Writer, fsys fs.FS, srcDir, destDir string, vars TemplateVars) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	entries, err := fs.ReadDir(fsys, srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// embed.FS always uses forward slashes
		srcPath := path.Join(srcDir, entry.Name())
		// Strip .tmpl suffix so "go.mod.tmpl" becomes "go.mod"
		destName := strings.TrimSuffix(entry.Name(), tmplSuffix)
		destPath := filepath.Join(destDir, destName)
		_, _ = fmt.Fprintf(w, "  create %s\n", destPath)

		if entry.IsDir() {
			if err := CopyEmbedDir(w, fsys, srcPath, destPath, vars); err != nil {
				return err
			}
			continue
		}

		content, err := fs.ReadFile(fsys, srcPath)
		if err != nil {
			return err
		}

		rendered, err := RenderTemplate(content, vars)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", srcPath, err)
		}

		if err := os.WriteFile(destPath, rendered, 0644); err != nil {
			return err
		}
	}
	return nil
}

// PreviewEmbedDir prints what files would be created without writing anything.
func PreviewEmbedDir(w io.Writer, fsys fs.FS, srcDir, destDir string) error {
	entries, err := fs.ReadDir(fsys, srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := path.Join(srcDir, entry.Name())
		destName := strings.TrimSuffix(entry.Name(), tmplSuffix)
		destPath := filepath.Join(destDir, destName)

		if entry.IsDir() {
			_, _ = fmt.Fprintf(w, "  create %s/\n", destPath)
			if err := PreviewEmbedDir(w, fsys, srcPath, destPath); err != nil {
				return err
			}
			continue
		}

		_, _ = fmt.Fprintf(w, "  create %s\n", destPath)
	}
	return nil
}
