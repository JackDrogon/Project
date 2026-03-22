package create

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/JackDrogon/project/internal/adapters/gitexec"
	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	"github.com/JackDrogon/project/internal/adapters/templatefs"
	domain "github.com/JackDrogon/project/internal/domain/scaffold"
)

var (
	osStat      = os.Stat
	osReadDir   = os.ReadDir
	osRemoveAll = os.RemoveAll
	osGetwd     = os.Getwd
	filepathAbs = filepath.Abs
)

const initialCommitMessage = "Initial commit"

type Options = domain.CreateRequest
type GitMode = domain.GitMode

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
	if err := c.preflightDestDir(opts); err != nil {
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

func (c *Creator) validateModulePath(opts Options) error {
	if opts.Lang != "go" {
		return nil
	}

	return domain.ValidateModulePath(defaultModulePath(opts))
}

func (c *Creator) validateTemplateInputs(opts Options) error {
	_, _, err := c.templateManifestAndVars(opts)
	return err
}

func (c *Creator) validateGitOptions(opts Options) error {
	mode, err := domain.ResolveGitMode(opts)
	if err != nil {
		return err
	}

	if opts.Signoff && mode != domain.GitModeInitCommit {
		return fmt.Errorf("--signoff requires --git=%s", domain.GitModeInitCommit)
	}

	return nil
}

func (c *Creator) checkDestDir(opts Options) error {
	return c.inspectDestDir(opts, false)
}

func (c *Creator) preflightDestDir(opts Options) error {
	return c.inspectDestDir(opts, true)
}

func (c *Creator) inspectDestDir(opts Options, previewOnly bool) error {
	targetDir := opts.DestinationDir()
	info, err := osStat(targetDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to inspect destination %q: %w", targetDir, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("destination %q already exists and is not a directory", targetDir)
	}

	if opts.Force {
		return c.handleForcedDestination(targetDir, previewOnly)
	}

	return validateReusableDestination(targetDir, opts.AllowExistingEmptyDir)
}

func (c *Creator) handleForcedDestination(targetDir string, previewOnly bool) error {
	currentDir, err := isCurrentDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to inspect destination %q: %w", targetDir, err)
	}
	if currentDir {
		return fmt.Errorf("refusing to remove current directory %q with --force", targetDir)
	}
	if previewOnly {
		return nil
	}

	_, _ = fmt.Fprintf(c.w, "Warning: directory %q already exists, removing due to --force\n", targetDir)
	if err := osRemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to remove existing directory %q: %w", targetDir, err)
	}

	return nil
}

func validateReusableDestination(targetDir string, allowExistingEmptyDir bool) error {
	empty, err := isEmptyDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to inspect destination %q: %w", targetDir, err)
	}

	if empty && allowExistingEmptyDir {
		return nil
	}

	if empty {
		return fmt.Errorf("directory %q already exists; use --force to overwrite", targetDir)
	}

	return fmt.Errorf("directory %q already exists and is not empty", targetDir)
}

func (c *Creator) copyTemplates(opts Options) error {
	_, vars, err := c.templateManifestAndVars(opts)
	if err != nil {
		return err
	}

	return templatefs.Materialize(c.w, c.fsys, opts.Lang, opts.DestinationDir(), vars, c.resolveMode)
}

func (c *Creator) initGitRepo(opts Options) error {
	commitArgs := []string{"commit", "-m", initialCommitMessage}
	if opts.Signoff {
		commitArgs = []string{"commit", "-s", "-m", initialCommitMessage}
	}

	for _, args := range [][]string{{"init"}, {"add", "."}, commitArgs} {
		if err := c.runGit(opts.DestinationDir(), args...); err != nil {
			return err
		}
	}

	return nil
}

func (c *Creator) maybeInitGitRepo(opts Options) error {
	mode, err := domain.ResolveGitMode(opts)
	if err != nil {
		return err
	}

	if mode == domain.GitModeNone {
		_, _ = fmt.Fprintln(c.w, "Skipping git initialization (--git none)")
		return nil
	}

	if mode == domain.GitModeInitOnly {
		_, _ = fmt.Fprintln(c.w, "Initializing git repository (--git init-only)")
		return c.runGit(opts.DestinationDir(), "init")
	}

	return c.initGitRepo(opts)
}

func (c *Creator) templateManifestAndVars(opts Options) (protocoltoml.Manifest, domain.TemplateVars, error) {
	manifest, found, err := protocoltoml.LoadManifest(c.fsys, opts.Lang)
	if err != nil {
		return protocoltoml.Manifest{}, domain.TemplateVars{}, err
	}
	if !found {
		manifest = protocoltoml.Manifest{Name: opts.Lang}
	}

	base := newDefaultTemplateVars(opts.ProjectName, opts.ModulePath)
	vars, err := domain.ResolveTemplateVars(manifestInputsToDomain(manifest.Inputs), opts, base)
	if err != nil {
		return protocoltoml.Manifest{}, domain.TemplateVars{}, err
	}

	return manifest, vars, nil
}

func manifestInputsToDomain(inputs []protocoltoml.ManifestInput) []domain.ManifestInput {
	result := make([]domain.ManifestInput, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, domain.ManifestInput{Name: input.Key, TemplateVar: input.TemplateVar})
	}
	return result
}

func isEmptyDir(dir string) (bool, error) {
	entries, err := osReadDir(dir)
	if err != nil {
		return false, err
	}

	return len(entries) == 0, nil
}

func isCurrentDir(targetDir string) (bool, error) {
	absTarget, err := filepathAbs(targetDir)
	if err != nil {
		return false, err
	}

	cwd, err := osGetwd()
	if err != nil {
		return false, err
	}

	return absTarget == cwd, nil
}

func defaultModulePath(opts Options) string {
	return domain.DefaultModulePath(opts)
}
