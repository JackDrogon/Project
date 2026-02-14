package scaffold

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/JackDrogon/project/pkg/git"
)

// Creator scaffolds new projects from embedded templates.
type Creator struct {
	fsys fs.FS
	w    io.Writer
}

// NewCreator returns a Creator that reads templates from fsys and writes
// progress output to w.
func NewCreator(fsys fs.FS, w io.Writer) *Creator {
	return &Creator{fsys: fsys, w: w}
}

// Options holds all parameters for project creation.
type Options struct {
	Lang        string
	ProjectName string
	ModulePath  string
	Force       bool
	Signoff     bool
	DryRun      bool
}

// pipeline chains fallible steps, short-circuiting on the first error.
type pipeline struct {
	err  error
	opts Options
}

func newPipeline(opts Options) *pipeline {
	return &pipeline{opts: opts}
}

func (p *pipeline) step(fn func(Options) error) *pipeline {
	if p.err == nil {
		p.err = fn(p.opts)
	}
	return p
}

func (p *pipeline) Err() error { return p.err }

// Create scaffolds a new project based on the given options.
func (c *Creator) Create(opts Options) error {
	p := newPipeline(opts).step(c.validate).step(c.checkLang)
	if p.Err() != nil {
		return p.Err()
	}

	_, _ = fmt.Fprintf(c.w, "Creating project with language: %s, project name: %s\n", opts.Lang, opts.ProjectName)

	if opts.DryRun {
		_, _ = fmt.Fprintln(c.w, "Dry-run mode: no files will be created")
		return PreviewEmbedDir(c.w, c.fsys, opts.Lang, opts.ProjectName)
	}

	if err := p.step(c.checkDestDir).step(c.copyTemplates).step(c.initGitRepo).Err(); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(c.w, "Project created successfully")
	return nil
}

func (c *Creator) validate(opts Options) error {
	return ValidateProjectName(opts.ProjectName)
}

func (c *Creator) checkLang(opts Options) error {
	if _, err := fs.ReadDir(c.fsys, opts.Lang); err != nil {
		return fmt.Errorf("unsupported language: %s", opts.Lang)
	}
	return nil
}

func (c *Creator) checkDestDir(opts Options) error {
	info, err := os.Stat(opts.ProjectName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to inspect destination %q: %w", opts.ProjectName, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("destination %q already exists and is not a directory", opts.ProjectName)
	}

	if !opts.Force {
		return fmt.Errorf("directory %q already exists; use --force to overwrite", opts.ProjectName)
	}

	_, _ = fmt.Fprintf(c.w, "Warning: directory %q already exists, removing due to --force\n", opts.ProjectName)
	if err := os.RemoveAll(opts.ProjectName); err != nil {
		return fmt.Errorf("failed to remove existing directory %q: %w", opts.ProjectName, err)
	}

	return nil
}

func (c *Creator) copyTemplates(opts Options) error {
	vars := NewTemplateVars(opts.ProjectName, opts.ModulePath)
	return CopyEmbedDir(c.w, c.fsys, opts.Lang, opts.ProjectName, vars)
}

func (c *Creator) initGitRepo(opts Options) error {
	commitArgs := []string{"commit", "-m", "Initial commit"}
	if opts.Signoff {
		commitArgs = []string{"commit", "-s", "-m", "Initial commit"}
	}

	for _, args := range [][]string{{"init"}, {"add", "."}, commitArgs} {
		if err := git.Run(opts.ProjectName, args...); err != nil {
			return err
		}
	}

	return nil
}

// ListLangs returns the available template language names.
func (c *Creator) ListLangs() ([]string, error) {
	entries, err := fs.ReadDir(c.fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates: %w", err)
	}

	var langs []string
	for _, entry := range entries {
		if entry.IsDir() {
			langs = append(langs, entry.Name())
		}
	}
	return langs, nil
}
