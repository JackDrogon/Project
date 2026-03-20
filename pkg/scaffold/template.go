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

var (
	osMkdirAll  = os.MkdirAll
	osWriteFile = os.WriteFile
)

const tmplSuffix = ".tmpl"

// This file implements the template materialization pipeline used by both real
// creation and dry-run preview.
//
// The key invariant is that path rendering happens before any file writes, using
// the same traversal order for preview and copy. That keeps dry-run output aligned
// with real execution and avoids subtle drift between the two code paths.

type templateEntry struct {
	srcPath     string
	destPath    string
	name        string
	entry       fs.DirEntry
	isDir       bool
	isTemplate  bool
	rawContents []byte
}

// RenderTemplate applies TemplateVars to content using text/template.
// It returns an error when template syntax is invalid or references unknown keys.
func RenderTemplate(content []byte, vars TemplateVars) ([]byte, error) {
	rendered, err := renderTemplateString(string(content), vars)
	if err != nil {
		return nil, err
	}
	return []byte(rendered), nil
}

func renderTemplateString(content string, vars TemplateVars) (string, error) {
	tmpl, err := template.New("").Option("missingkey=error").Parse(string(content))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderTemplatePathSegment(name string, vars TemplateVars) (string, error) {
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

func walkTemplateEntries(fsys fs.FS, srcDir, destDir string, vars TemplateVars, visit func(templateEntry) error) error {
	entries, err := fs.ReadDir(fsys, srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := path.Join(srcDir, entry.Name())
		destName, err := renderTemplatePathSegment(strings.TrimSuffix(entry.Name(), tmplSuffix), vars)
		if err != nil {
			return fmt.Errorf("failed to render template path %s: %w", srcPath, err)
		}

		current := templateEntry{
			srcPath:    srcPath,
			destPath:   filepath.Join(destDir, destName),
			name:       entry.Name(),
			entry:      entry,
			isDir:      entry.IsDir(),
			isTemplate: strings.HasSuffix(entry.Name(), tmplSuffix),
		}

		if err := visit(current); err != nil {
			return err
		}

		if current.isDir {
			if err := walkTemplateEntries(fsys, srcPath, current.destPath, vars, visit); err != nil {
				return err
			}
		}
	}

	return nil
}

func readTemplateEntry(fsys fs.FS, entry templateEntry) (templateEntry, error) {
	content, err := fs.ReadFile(fsys, entry.srcPath)
	if err != nil {
		return templateEntry{}, err
	}

	entry.rawContents = content
	return entry, nil
}

func renderTemplateEntry(entry templateEntry, vars TemplateVars) ([]byte, error) {
	if !entry.isTemplate {
		return entry.rawContents, nil
	}

	rendered, err := RenderTemplate(entry.rawContents, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to render template %s: %w", entry.srcPath, err)
	}

	return rendered, nil
}

func templateEntryMode(entry templateEntry) fs.FileMode {
	mode := fs.FileMode(0644)
	if info, err := entry.entry.Info(); err == nil {
		if perm := info.Mode().Perm(); perm != 0 {
			mode = perm
		}
	}

	return mode
}

// CopyEmbedDir recursively copies a directory from an embedded filesystem
// to the local filesystem, rendering template variables in file contents.
func CopyEmbedDir(w io.Writer, fsys fs.FS, srcDir, destDir string, vars TemplateVars) error {
	if err := osMkdirAll(destDir, 0755); err != nil {
		return err
	}

	return walkTemplateEntries(fsys, srcDir, destDir, vars, func(entry templateEntry) error {
		_, _ = fmt.Fprintf(w, "  create %s\n", entry.destPath)

		if entry.isDir {
			return osMkdirAll(entry.destPath, 0755)
		}

		loaded, err := readTemplateEntry(fsys, entry)
		if err != nil {
			return err
		}

		rendered, err := renderTemplateEntry(loaded, vars)
		if err != nil {
			return err
		}

		return osWriteFile(entry.destPath, rendered, templateEntryMode(entry))
	})
}

// PreviewEmbedDir prints what files would be created without writing anything.
func PreviewEmbedDir(w io.Writer, fsys fs.FS, srcDir, destDir string, vars TemplateVars) error {
	return walkTemplateEntries(fsys, srcDir, destDir, vars, func(entry templateEntry) error {
		if entry.isDir {
			_, _ = fmt.Fprintf(w, "  create %s/\n", entry.destPath)
			return nil
		}

		_, _ = fmt.Fprintf(w, "  create %s\n", entry.destPath)
		loaded, err := readTemplateEntry(fsys, entry)
		if err != nil {
			return err
		}

		_, err = renderTemplateEntry(loaded, vars)
		if err != nil {
			return err
		}

		return nil
	})
}
