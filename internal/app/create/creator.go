package create

import (
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

type Creator struct {
	fsys        fs.FS
	w           io.Writer
	runGit      func(dir string, args ...string) error
	resolveMode templatefs.ModeResolver
}

func NewCreator(fsys fs.FS, w io.Writer) *Creator {
	return NewCreatorWithDeps(fsys, w, gitexec.New().Run, nil)
}

func NewCreatorWithDeps(fsys fs.FS, w io.Writer, runGit func(dir string, args ...string) error, resolveMode templatefs.ModeResolver) *Creator {
	if runGit == nil {
		runGit = gitexec.New().Run
	}

	return &Creator{fsys: fsys, w: w, runGit: runGit, resolveMode: resolveMode}
}

func (c *Creator) Create(opts Options) error {
	if err := c.validateCreateOptions(opts); err != nil {
		return err
	}

	c.writeCreateStart(opts)

	if opts.DryRun {
		return c.previewCreate(opts)
	}

	if err := c.materializeCreate(opts); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(c.w, "Project created successfully")
	return nil
}

func (c *Creator) validateCreateOptions(opts Options) error {
	if err := domain.ValidateProjectName(opts.ProjectName); err != nil {
		return err
	}
	if err := c.checkLang(opts); err != nil {
		return err
	}
	if err := c.validateModulePath(opts); err != nil {
		return err
	}
	if err := c.validateTemplateInputs(opts); err != nil {
		return err
	}
	return c.validateGitOptions(opts)
}

func (c *Creator) writeCreateStart(opts Options) {
	_, _ = fmt.Fprintf(c.w, "Creating project with language: %s, project name: %s\n", opts.Lang, opts.ProjectName)
}

func (c *Creator) previewCreate(opts Options) error {
	if err := c.checkDestDir(opts); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(c.w, "Dry-run mode: no files will be created")

	plan, err := c.BuildDryRunPlan(opts)
	if err != nil {
		return err
	}

	return writeDryRunPlan(c.w, plan, opts)
}

func (c *Creator) materializeCreate(opts Options) error {
	if err := c.checkDestDir(opts); err != nil {
		return err
	}
	if err := c.copyTemplates(opts); err != nil {
		return err
	}
	return c.maybeInitGitRepo(opts)
}

func (c *Creator) checkLang(opts Options) error {
	if _, err := fs.ReadDir(c.fsys, opts.Lang); err != nil {
		return fmt.Errorf("unsupported language: %s", opts.Lang)
	}
	return nil
}
