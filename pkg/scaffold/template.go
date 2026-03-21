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
		if isReservedTemplateFile(entry.Name()) {
			continue
		}

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

func writeDryRunPlan(w io.Writer, plan DryRunPlan, opts Options) error {
	mode, err := resolveGitMode(opts)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "template: %s\n", plan.Template)
	_, _ = fmt.Fprintf(w, "description: %s\n", plan.Description)
	_, _ = fmt.Fprintf(w, "target_dir: %s\n", plan.TargetDir)
	_, _ = fmt.Fprintln(w, "resolved inputs:")
	_, _ = fmt.Fprintf(w, "  project_name: %s\n", opts.ProjectName)

	modulePath, includeModulePath := dryRunPlanModulePath(plan, opts)
	if includeModulePath {
		_, _ = fmt.Fprintf(w, "  module_path: %s\n", modulePath)
	}

	for _, input := range plan.ResolvedInputs {
		if includeModulePath && input.TemplateVar == "ModulePath" {
			continue
		}
		_, _ = fmt.Fprintf(w, "  %s: %s\n", input.Name, input.Value)
	}
	_, _ = fmt.Fprintf(w, "  git_mode: %s\n", mode)

	_, _ = fmt.Fprintln(w, "explicit overrides:")
	overrides := dryRunPlanOverrides(plan, opts, mode, modulePath, includeModulePath)
	if len(overrides) == 0 {
		_, _ = fmt.Fprintln(w, "  (none)")
	} else {
		for _, override := range overrides {
			_, _ = fmt.Fprintf(w, "  %s: %s\n", override.name, override.value)
		}
	}

	_, _ = fmt.Fprintln(w, "actions:")
	for _, action := range plan.Actions {
		_, _ = fmt.Fprintf(w, "  %s\n", formatDryRunAction(action))
	}

	return nil
}

type dryRunPlanOverride struct {
	name  string
	value string
}

func dryRunPlanModulePath(plan DryRunPlan, opts Options) (string, bool) {
	if value, ok := dryRunPlanResolvedInputValue(plan, "ModulePath"); ok {
		return value, true
	}
	if opts.Lang == "go" {
		return effectiveModulePath(opts), true
	}

	return "", false
}

func dryRunPlanResolvedInputValue(plan DryRunPlan, templateVar string) (string, bool) {
	for _, input := range plan.ResolvedInputs {
		if input.TemplateVar == templateVar {
			return input.Value, true
		}
	}

	return "", false
}

func dryRunPlanOverrides(plan DryRunPlan, opts Options, mode GitMode, modulePath string, includeModulePath bool) []dryRunPlanOverride {
	var overrides []dryRunPlanOverride
	if includeModulePath && modulePath != opts.ProjectName {
		overrides = append(overrides, dryRunPlanOverride{name: "module_path", value: modulePath})
	}
	for _, input := range plan.ResolvedInputs {
		if input.TemplateVar == "ModulePath" {
			continue
		}

		if value, ok := opts.TemplateInputValues[input.Name]; ok {
			overrides = append(overrides, dryRunPlanOverride{name: input.Name, value: value})
		}
	}
	if opts.NoGit || opts.GitMode != "" {
		overrides = append(overrides, dryRunPlanOverride{name: "git_mode", value: string(mode)})
	}

	return overrides
}

func effectiveModulePath(opts Options) string {
	if opts.ModulePath != "" {
		return opts.ModulePath
	}

	return opts.ProjectName
}

func formatDryRunAction(action DryRunAction) string {
	switch action.Kind {
	case DryRunActionCreateDir:
		return fmt.Sprintf("create %s%s", action.Target, string(filepath.Separator))
	case DryRunActionRenderFile:
		return fmt.Sprintf("render %s -> %s", action.Source, action.Target)
	case DryRunActionCopyFile:
		return fmt.Sprintf("copy %s -> %s", action.Source, action.Target)
	default:
		return fmt.Sprintf("create %s", action.Target)
	}
}
