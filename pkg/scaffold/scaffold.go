package scaffold

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/JackDrogon/project/pkg/git"
)

var (
	osStat      = os.Stat
	osReadDir   = os.ReadDir
	osRemoveAll = os.RemoveAll
	osGetwd     = os.Getwd
	filepathAbs = filepath.Abs
)

const initialCommitMessage = "Initial commit"

// This file owns the stateful scaffolding pipeline.
//
// The workflow is intentionally staged in a fixed order:
//  1. Validate options before touching the filesystem.
//  2. Announce the requested operation once inputs are known-good.
//  3. For dry-run, inspect the destination and validate the rendered plan without writing it.
//  4. For real creation, validate the destination, materialize files, then run git.
//
// Keeping git as the final phase preserves the existing resumability properties:
// reruns only need to reason about generated files and destination checks, while git
// side effects happen after templates have been written successfully.

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
	Lang                  string
	ProjectName           string
	TargetDir             string
	ModulePath            string
	TemplateInputValues   map[string]string
	Force                 bool
	AllowExistingEmptyDir bool
	Signoff               bool
	DryRun                bool
	NoGit                 bool
	GitMode               GitMode
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

func (opts Options) destinationDir() string {
	if opts.TargetDir != "" {
		return opts.TargetDir
	}

	return opts.ProjectName
}

// Create scaffolds a new project based on the given options.
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
	return newPipeline(opts).
		step(c.validate).
		step(c.checkLang).
		step(c.validateModulePath).
		step(c.validateTemplateInputs).
		step(c.validateGitOptions).
		Err()
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
	return newPipeline(opts).
		step(c.checkDestDir).
		step(c.copyTemplates).
		step(c.maybeInitGitRepo).
		Err()
}

func (c *Creator) templateVars(opts Options) (TemplateVars, error) {
	_, vars, err := c.templateManifestAndVars(opts)
	if err != nil {
		return TemplateVars{}, err
	}

	return vars, nil
}

func (c *Creator) validate(opts Options) error {
	return ValidateProjectName(opts.ProjectName)
}

func (c *Creator) validateModulePath(opts Options) error {
	if opts.Lang != "go" {
		return nil
	}

	return ValidateModulePath(c.defaultModulePath(opts))
}

func (c *Creator) defaultModulePath(opts Options) string {
	if opts.ModulePath != "" {
		return opts.ModulePath
	}

	return opts.ProjectName
}

func (c *Creator) templateManifestAndVars(opts Options) (TemplateManifest, TemplateVars, error) {
	manifest, _, err := loadTemplateManifest(c.fsys, opts.Lang)
	if err != nil {
		return TemplateManifest{}, TemplateVars{}, err
	}

	vars, err := resolveTemplateVars(manifest, opts)
	if err != nil {
		return TemplateManifest{}, TemplateVars{}, err
	}

	return manifest, vars, nil
}

func (c *Creator) checkLang(opts Options) error {
	if _, err := fs.ReadDir(c.fsys, opts.Lang); err != nil {
		return fmt.Errorf("unsupported language: %s", opts.Lang)
	}
	return nil
}

func (c *Creator) validateTemplateInputs(opts Options) error {
	_, _, err := c.templateManifestAndVars(opts)
	return err
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
	return c.inspectDestDir(opts, false)
}

func (c *Creator) preflightDestDir(opts Options) error {
	return c.inspectDestDir(opts, true)
}

func (c *Creator) inspectDestDir(opts Options, previewOnly bool) error {
	targetDir := opts.destinationDir()
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
	vars, err := c.templateVars(opts)
	if err != nil {
		return err
	}

	return CopyEmbedDir(c.w, c.fsys, opts.Lang, opts.destinationDir(), vars)
}

func (c *Creator) initGitRepo(opts Options) error {
	commitArgs := []string{"commit", "-m", initialCommitMessage}
	if opts.Signoff {
		commitArgs = []string{"commit", "-s", "-m", initialCommitMessage}
	}

	for _, args := range [][]string{{"init"}, {"add", "."}, commitArgs} {
		if err := c.runGit(opts.destinationDir(), args...); err != nil {
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

	if mode == GitModeNone {
		_, _ = fmt.Fprintln(c.w, "Skipping git initialization (--git none)")
		return nil
	}

	if mode == GitModeInitOnly {
		_, _ = fmt.Fprintln(c.w, "Initializing git repository (--git init-only)")
		return c.runGit(opts.destinationDir(), "init")
	}

	return c.initGitRepo(opts)
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

// ReplayFile returns the resolved replay payload for a completed create request.
func (c *Creator) ReplayFile(command ReplayCommand, opts Options) (ReplayFile, error) {
	manifest, _, err := c.templateManifestAndVars(opts)
	if err != nil {
		return ReplayFile{}, err
	}

	mode, err := resolveGitMode(opts)
	if err != nil {
		return ReplayFile{}, err
	}

	replay := ReplayFile{
		Command: command,
		Lang:    opts.Lang,
		Create: ReplayFileCreate{
			ProjectName: opts.ProjectName,
			TargetDir:   opts.destinationDir(),
			GitMode:     mode,
			Signoff:     opts.Signoff,
			Force:       opts.Force,
		},
		TemplateInputs: map[string]string{},
	}

	for _, input := range manifest.Inputs {
		switch input.TemplateVar {
		case "ModulePath":
			replay.TemplateInputs[input.Name] = c.defaultModulePath(opts)
		default:
			value, ok := opts.TemplateInputValues[input.Name]
			if ok {
				replay.TemplateInputs[input.Name] = value
			}
		}
	}

	return replay, nil
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
