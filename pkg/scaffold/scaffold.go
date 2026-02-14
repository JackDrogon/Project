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
	fsys   fs.FS
	w      io.Writer
	runGit func(dir string, args ...string) error
}

// NewCreator returns a Creator that reads templates from fsys and writes
// progress output to w.
func NewCreator(fsys fs.FS, w io.Writer) *Creator {
	return NewCreatorWithGitRunner(fsys, w, git.Run)
}

// NewCreatorWithGitRunner returns a Creator with an injected git runner.
func NewCreatorWithGitRunner(fsys fs.FS, w io.Writer, runGit func(dir string, args ...string) error) *Creator {
	if runGit == nil {
		runGit = git.Run
	}

	return &Creator{fsys: fsys, w: w, runGit: runGit}
}

// Options holds all parameters for project creation.
type Options struct {
	Lang        string
	ProjectName string
	ModulePath  string
	Force       bool
	Signoff     bool
	DryRun      bool
	NoGit       bool
	GitMode     GitMode
}

// GitMode controls how git is initialized for new projects.
type GitMode string

const (
	GitModeNone       GitMode = "none"
	GitModeInitOnly   GitMode = "init-only"
	GitModeInitCommit GitMode = "init+commit"
)

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
	p := newPipeline(opts).step(c.validate).step(c.checkLang).step(c.validateGitOptions)
	if p.Err() != nil {
		return p.Err()
	}

	_, _ = fmt.Fprintf(c.w, "Creating project with language: %s, project name: %s\n", opts.Lang, opts.ProjectName)

	if opts.DryRun {
		_, _ = fmt.Fprintln(c.w, "Dry-run mode: no files will be created")
		return PreviewEmbedDir(c.w, c.fsys, opts.Lang, opts.ProjectName)
	}

	if err := p.step(c.checkDestDir).step(c.copyTemplates).step(c.maybeInitGitRepo).Err(); err != nil {
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

func (c *Creator) validateGitOptions(opts Options) error {
	mode, err := resolveGitMode(opts)
	if err != nil {
		return err
	}

	if opts.Signoff && mode != GitModeInitCommit {
		return fmt.Errorf("--signoff requires --git=%s", GitModeInitCommit)
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
		if err := c.runGit(opts.ProjectName, args...); err != nil {
			return err
		}
	}

	return nil
}

func (c *Creator) maybeInitGitRepo(opts Options) error {
	mode, err := resolveGitMode(opts)
	if err != nil {
		return err
	}

	switch mode {
	case GitModeNone:
		_, _ = fmt.Fprintln(c.w, "Skipping git initialization (--git none)")
		return nil
	case GitModeInitOnly:
		_, _ = fmt.Fprintln(c.w, "Initializing git repository (--git init-only)")
		return c.runGit(opts.ProjectName, "init")
	case GitModeInitCommit:
		return c.initGitRepo(opts)
	default:
		return fmt.Errorf("unsupported git mode: %q", mode)
	}
}

func resolveGitMode(opts Options) (GitMode, error) {
	if opts.NoGit {
		if opts.GitMode != "" && opts.GitMode != GitModeNone {
			return "", fmt.Errorf("conflicting git options: --no-git cannot be combined with --git=%s", opts.GitMode)
		}
		return GitModeNone, nil
	}

	if opts.GitMode == "" {
		return GitModeInitCommit, nil
	}

	switch opts.GitMode {
	case GitModeNone, GitModeInitOnly, GitModeInitCommit:
		return opts.GitMode, nil
	default:
		return "", fmt.Errorf("invalid git mode %q: must be one of %s, %s, %s", opts.GitMode, GitModeNone, GitModeInitOnly, GitModeInitCommit)
	}
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
