package create

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/JackDrogon/project/internal/adapters/gitexec"
	"github.com/JackDrogon/project/internal/adapters/templatefs"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

type (
	Options = domain.CreateRequest
	GitMode = domain.GitMode
)

// RunGit runs one git invocation in dir. It takes a context because git is an
// external process: cancelling the context must kill it.
type RunGit func(ctx context.Context, dir string, args ...string) error

type Creator struct {
	fsys        fs.FS
	w           io.Writer
	runGit      RunGit
	resolveMode templatefs.ModeResolver
}

func NewCreator(fsys fs.FS, w io.Writer) *Creator {
	return NewCreatorWithDeps(fsys, w, gitexec.New().Run, nil)
}

func NewCreatorWithDeps(fsys fs.FS, w io.Writer, runGit RunGit, resolveMode templatefs.ModeResolver) *Creator {
	if runGit == nil {
		runGit = gitexec.New().Run
	}

	return &Creator{fsys: fsys, w: w, runGit: runGit, resolveMode: resolveMode}
}

// Create scaffolds the project. ctx only reaches the external commands this
// package shells out to (git, and `go env` for version detection); the
// in-memory steps stay context-free.
func (c *Creator) Create(ctx context.Context, opts Options) error {
	if err := c.validateCreateOptions(ctx, opts); err != nil {
		return err
	}

	c.writeCreateStart(opts)

	if opts.DryRun {
		return c.previewCreate(ctx, opts)
	}

	if err := c.materializeCreate(ctx, opts); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(c.w, "Project created successfully")
	return nil
}

func (c *Creator) validateCreateOptions(ctx context.Context, opts Options) error {
	if err := domain.ValidateProjectName(opts.ProjectName); err != nil {
		return err
	}
	if err := c.checkLang(opts); err != nil {
		return err
	}
	if err := c.validateModulePath(opts); err != nil {
		return err
	}
	if err := c.validateTemplateInputs(ctx, opts); err != nil {
		return err
	}
	return c.validateGitOptions(opts)
}

func (c *Creator) writeCreateStart(opts Options) {
	_, _ = fmt.Fprintf(c.w, "Creating project with language: %s, project name: %s\n", opts.Lang, opts.ProjectName)
}

func (c *Creator) previewCreate(ctx context.Context, opts Options) error {
	if err := c.checkDestDir(opts); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(c.w, "Dry-run mode: no files will be created")

	plan, err := c.BuildDryRunPlan(ctx, opts)
	if err != nil {
		return err
	}

	return writeDryRunPlan(c.w, plan, opts)
}

func (c *Creator) materializeCreate(ctx context.Context, opts Options) error {
	if err := c.checkDestDir(opts); err != nil {
		return err
	}
	if err := c.copyTemplates(ctx, opts); err != nil {
		return err
	}
	return c.maybeInitGitRepo(ctx, opts)
}

// checkLang distinguishes "this template does not exist" from "the template
// tree could not be read". Only the former is a user error; collapsing both
// into "unsupported language" hid real I/O failures behind a wrong message.
func (c *Creator) checkLang(opts Options) error {
	if !fs.ValidPath(opts.Lang) {
		return fmt.Errorf("unsupported language: %s", opts.Lang)
	}

	if _, err := fs.ReadDir(c.fsys, opts.Lang); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("unsupported language: %s", opts.Lang)
		}
		return fmt.Errorf("failed to read template %q: %w", opts.Lang, err)
	}
	return nil
}
