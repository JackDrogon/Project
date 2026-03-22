package create

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	domain "github.com/JackDrogon/project/internal/domain/scaffold"
)

type Command string

const (
	CommandNew  Command = "new"
	CommandInit Command = "init"
)

type Flags struct {
	Lang            string
	Module          string
	Signoff         bool
	DryRun          bool
	NoGit           bool
	GitMode         string
	ReplayPath      string
	WriteReplayPath string
	SetValues       []string
}

type Changed struct {
	Lang    bool
	Module  bool
	Signoff bool
	NoGit   bool
	Git     bool
	Force   bool
}

type Runtime struct {
	Replay              protocoltoml.Replay
	HasReplay           bool
	TemplateInputValues map[string]string
}

type NewRequest struct {
	Flags   Flags
	Changed Changed
	Force   bool
	Arg     string
	HasArg  bool
}

type InitRequest struct {
	Flags   Flags
	Changed Changed
	Arg     string
	HasArg  bool
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

var reservedSetKeys = map[string]struct{}{
	"lang":         {},
	"project_name": {},
	"target_dir":   {},
	"module_path":  {},
	"git_mode":     {},
	"signoff":      {},
	"force":        {},
	"dry_run":      {},
}

func (s *Service) ParseSetValues(flags Flags) (map[string]string, error) {
	values := make(map[string]string, len(flags.SetValues))
	for _, raw := range flags.SetValues {
		key, value, ok := strings.Cut(raw, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --set value %q: must be key=value", raw)
		}
		if key == "" {
			return nil, fmt.Errorf("invalid --set value %q: key must not be empty", raw)
		}
		if _, reserved := reservedSetKeys[key]; reserved {
			return nil, fmt.Errorf("invalid --set key %q: reserved for command options", key)
		}
		if _, exists := values[key]; exists {
			return nil, fmt.Errorf("invalid --set key %q: specified more than once", key)
		}
		values[key] = value
	}

	return values, nil
}

func (s *Service) RuntimeState(flags Flags, expected Command) (Runtime, error) {
	templateInputValues, err := s.ParseSetValues(flags)
	if err != nil {
		return Runtime{}, err
	}

	if flags.WriteReplayPath != "" && flags.DryRun {
		return Runtime{}, fmt.Errorf("--write-replay cannot be combined with --dry-run")
	}

	if flags.ReplayPath == "" {
		return Runtime{TemplateInputValues: templateInputValues}, nil
	}

	replay, err := protocoltoml.ReadReplay(flags.ReplayPath)
	if err != nil {
		return Runtime{}, err
	}
	if replay.Mode != string(expected) {
		return Runtime{}, fmt.Errorf(
			"invalid --replay %q: replay command %q does not match %q",
			flags.ReplayPath,
			replay.Mode,
			expected,
		)
	}

	mergedInputs := make(map[string]string, len(replay.Inputs)+len(templateInputValues))
	for key, value := range replay.Inputs {
		if key == "module_path" {
			continue
		}
		mergedInputs[key] = value
	}
	for key, value := range templateInputValues {
		mergedInputs[key] = value
	}

	return Runtime{Replay: replay, HasReplay: true, TemplateInputValues: mergedInputs}, nil
}

func (s *Service) BuildNewOptions(req NewRequest) (Options, error) {
	runtime, err := s.RuntimeState(req.Flags, CommandNew)
	if err != nil {
		return Options{}, err
	}

	return s.NewOptions(req.Flags, runtime, req.Changed, req.Force, req.Arg, req.HasArg)
}

func (s *Service) BuildInitOptions(req InitRequest) (Options, error) {
	runtime, err := s.RuntimeState(req.Flags, CommandInit)
	if err != nil {
		return Options{}, err
	}

	return s.InitOptions(req.Flags, runtime, req.Changed, req.Arg, req.HasArg)
}

func (s *Service) ResolveLang(flags Flags, changed Changed, runtime Runtime) (string, error) {
	if changed.Lang {
		return flags.Lang, nil
	}
	if runtime.HasReplay {
		return runtime.Replay.Template.Lang, nil
	}

	return "", fmt.Errorf("required flag(s) \"lang\" not set")
}

func (s *Service) ResolveSignoff(flags Flags, changed Changed, runtime Runtime) bool {
	if changed.Signoff {
		return flags.Signoff
	}
	if runtime.HasReplay {
		return runtime.Replay.Git.Signoff
	}

	return flags.Signoff
}

func (s *Service) ResolveNoGit(flags Flags, changed Changed) bool {
	if changed.NoGit {
		return flags.NoGit
	}

	return flags.NoGit
}

func (s *Service) ResolveGitMode(flags Flags, changed Changed, runtime Runtime) string {
	if changed.Git {
		return flags.GitMode
	}
	if changed.NoGit {
		return ""
	}
	if runtime.HasReplay {
		return string(runtime.Replay.Git.Mode)
	}

	return flags.GitMode
}

func (s *Service) ResolveModulePath(flags Flags, changed Changed, runtime Runtime) string {
	if changed.Module {
		return flags.Module
	}
	if runtime.HasReplay {
		if runtime.Replay.Project.ModulePath != "" {
			return runtime.Replay.Project.ModulePath
		}
		return runtime.Replay.Inputs["module_path"]
	}

	return flags.Module
}

func (s *Service) ResolveTemplateInputValues(runtime Runtime) map[string]string {
	if len(runtime.TemplateInputValues) == 0 {
		return nil
	}

	resolved := make(map[string]string, len(runtime.TemplateInputValues))
	for key, value := range runtime.TemplateInputValues {
		resolved[key] = value
	}

	return resolved
}

func (s *Service) NewOptions(flags Flags, runtime Runtime, changed Changed, force bool, arg string, hasArg bool) (Options, error) {
	lang, err := s.ResolveLang(flags, changed, runtime)
	if err != nil {
		return Options{}, err
	}

	signoff := s.ResolveSignoff(flags, changed, runtime)
	noGit := s.ResolveNoGit(flags, changed)
	gitMode := s.ResolveGitMode(flags, changed, runtime)
	replayModulePath := s.ResolveModulePath(flags, changed, runtime)

	resolvedForce := force
	if !changed.Force && runtime.HasReplay {
		resolvedForce = runtime.Replay.Options.Force
	}

	if !hasArg {
		if !runtime.HasReplay {
			return Options{}, fmt.Errorf("accepts 1 arg(s), received 0")
		}

		opts := s.Options(flags, lang, runtime.Replay.Project.Name, runtime.Replay.Project.TargetDir, replayModulePath, signoff, noGit, gitMode)
		opts.TemplateInputValues = s.ResolveTemplateInputValues(runtime)
		opts.Force = resolvedForce
		return opts, nil
	}

	explicitModulePath := ""
	if changed.Module {
		explicitModulePath = flags.Module
	}

	projectName, targetDir, modulePath, err := resolveNewProjectArgs(lang, explicitModulePath, arg)
	if err != nil {
		return Options{}, err
	}
	if explicitModulePath == "" && modulePath == "" {
		modulePath = replayModulePath
	}

	opts := s.Options(flags, lang, projectName, targetDir, modulePath, signoff, noGit, gitMode)
	opts.TemplateInputValues = s.ResolveTemplateInputValues(runtime)
	opts.Force = resolvedForce
	return opts, nil
}

func (s *Service) InitOptions(flags Flags, runtime Runtime, changed Changed, arg string, hasArg bool) (Options, error) {
	lang, err := s.ResolveLang(flags, changed, runtime)
	if err != nil {
		return Options{}, err
	}

	targetDir := "."
	projectName := ""
	if hasArg {
		targetDir = arg
		projectName, err = projectNameFromTargetDir(targetDir)
		if err != nil {
			return Options{}, err
		}
	} else if runtime.HasReplay {
		projectName = runtime.Replay.Project.Name
		targetDir = runtime.Replay.Project.TargetDir
	} else {
		projectName, err = projectNameFromTargetDir(targetDir)
		if err != nil {
			return Options{}, err
		}
	}

	opts := s.Options(
		flags,
		lang,
		projectName,
		targetDir,
		s.ResolveModulePath(flags, changed, runtime),
		s.ResolveSignoff(flags, changed, runtime),
		s.ResolveNoGit(flags, changed),
		s.ResolveGitMode(flags, changed, runtime),
	)
	opts.TemplateInputValues = s.ResolveTemplateInputValues(runtime)
	opts.AllowExistingEmptyDir = true
	return opts, nil
}

func (s *Service) Options(flags Flags, lang, projectName, targetDir, modulePath string, signoff, noGit bool, gitMode string) Options {
	return Options{
		Lang:        lang,
		ProjectName: projectName,
		TargetDir:   targetDir,
		ModulePath:  modulePath,
		Signoff:     signoff,
		DryRun:      flags.DryRun,
		NoGit:       noGit,
		GitMode:     domain.GitMode(gitMode),
	}
}

func (s *Service) ScaffoldAndMaybeWriteReplay(creator *Creator, flags Flags, command Command, opts Options) error {
	if err := creator.Create(opts); err != nil {
		return err
	}
	if flags.WriteReplayPath == "" {
		return nil
	}

	resolvedGitMode, err := domain.ResolveGitMode(domain.CreateRequest{NoGit: opts.NoGit, GitMode: domain.GitMode(opts.GitMode)})
	if err != nil {
		return fmt.Errorf("failed to resolve replay after project creation: %w", err)
	}

	inputs := map[string]string{}
	for key, value := range opts.TemplateInputValues {
		inputs[key] = value
	}
	if opts.ModulePath != "" {
		inputs["module_path"] = opts.ModulePath
	}

	replay := protocoltoml.Replay{
		Version:  protocoltoml.ReplayVersion,
		Mode:     string(command),
		Template: protocoltoml.ReplayTemplate{Lang: opts.Lang},
		Project: protocoltoml.ReplayProject{
			Name:       opts.ProjectName,
			TargetDir:  opts.DestinationDir(),
			ModulePath: opts.ModulePath,
		},
		Git:     protocoltoml.ReplayGit{Mode: domain.GitMode(resolvedGitMode), Signoff: opts.Signoff},
		Options: protocoltoml.ReplayOptions{Force: opts.Force},
		Inputs:  inputs,
	}

	if err := protocoltoml.WriteReplay(flags.WriteReplayPath, replay); err != nil {
		return fmt.Errorf("failed to write resolved replay after project creation: %w", err)
	}

	return nil
}

func resolveNewProjectArgs(lang, module, arg string) (projectName string, targetDir string, modulePath string, err error) {
	if lang == "go" && module == "" {
		if projectErr := domain.ValidateProjectName(arg); projectErr != nil {
			if moduleErr := domain.ValidateModulePath(arg); moduleErr == nil {
				name := domain.ProjectNameFromGoModulePath(arg)
				if err := domain.ValidateProjectName(name); err != nil {
					return "", "", "", err
				}
				return name, name, arg, nil
			}
		}
	}

	return arg, arg, module, nil
}

func ResolveNewProjectArgs(lang, module, arg string) (projectName string, targetDir string, modulePath string, err error) {
	return resolveNewProjectArgs(lang, module, arg)
}

func ProjectNameFromGoModulePath(modulePath string) string {
	return domain.ProjectNameFromGoModulePath(modulePath)
}

func projectNameFromTargetDir(targetDir string) (string, error) {
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target directory %q: %w", targetDir, err)
	}

	return filepath.Base(absTarget), nil
}

func ProjectNameFromTargetDir(targetDir string) (string, error) {
	return projectNameFromTargetDir(targetDir)
}
