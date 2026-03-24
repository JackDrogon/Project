package create

import (
	"fmt"
	"strings"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	domain "github.com/JackDrogon/project/internal/scaffold"
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

type ScaffoldSpec struct {
	Command Command
	Flags   Flags
	Options Options
}

type resolvedScaffoldSettings struct {
	Lang                string
	ModulePath          string
	Signoff             bool
	NoGit               bool
	GitMode             string
	TemplateInputValues map[string]string
}

type targetResolution struct {
	ProjectName           string
	TargetDir             string
	ModulePath            string
	Force                 bool
	AllowExistingEmptyDir bool
}

type Service struct {
	settingsResolver   ScaffoldSettingsResolver
	newTargetResolver  NewTargetResolver
	initTargetResolver InitTargetResolver
}

func NewService() *Service {
	return NewServiceWithDeps(DefaultDependencies())
}

func NewServiceWithDeps(deps Dependencies) *Service {
	deps = deps.withDefaults()
	return &Service{
		settingsResolver:   deps.SettingsResolver,
		newTargetResolver:  deps.NewTargetResolver,
		initTargetResolver: deps.InitTargetResolver,
	}
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

	mergedInputs := mergeReplayInputs(replay, templateInputValues)

	return Runtime{Replay: replay, HasReplay: true, TemplateInputValues: mergedInputs}, nil
}

func (s *Service) BuildNewOptions(req NewRequest) (Options, error) {
	runtime, err := s.RuntimeState(req.Flags, CommandNew)
	if err != nil {
		return Options{}, err
	}

	return s.NewOptions(req.Flags, runtime, req.Changed, req.Force, req.Arg, req.HasArg)
}

func (s *Service) BuildNewSpec(req NewRequest) (ScaffoldSpec, error) {
	options, err := s.BuildNewOptions(req)
	if err != nil {
		return ScaffoldSpec{}, err
	}
	return ScaffoldSpec{Command: CommandNew, Flags: req.Flags, Options: options}, nil
}

func (s *Service) BuildInitOptions(req InitRequest) (Options, error) {
	runtime, err := s.RuntimeState(req.Flags, CommandInit)
	if err != nil {
		return Options{}, err
	}

	return s.InitOptions(req.Flags, runtime, req.Changed, req.Arg, req.HasArg)
}

func (s *Service) BuildInitSpec(req InitRequest) (ScaffoldSpec, error) {
	options, err := s.BuildInitOptions(req)
	if err != nil {
		return ScaffoldSpec{}, err
	}
	return ScaffoldSpec{Command: CommandInit, Flags: req.Flags, Options: options}, nil
}

func (s *Service) ResolveScaffoldSettings(flags Flags, changed Changed, runtime Runtime) (resolvedScaffoldSettings, error) {
	return s.settingsResolver.Resolve(flags, changed, runtime)
}

func (s resolvedScaffoldSettings) Options(flags Flags, target targetResolution) Options {
	options := Options{
		Lang:                  s.Lang,
		ProjectName:           target.ProjectName,
		TargetDir:             target.TargetDir,
		ModulePath:            target.ModulePath,
		Signoff:               s.Signoff,
		DryRun:                flags.DryRun,
		NoGit:                 s.NoGit,
		GitMode:               domain.GitMode(s.GitMode),
		TemplateInputValues:   s.TemplateInputValues,
		Force:                 target.Force,
		AllowExistingEmptyDir: target.AllowExistingEmptyDir,
	}
	return options
}

func (s *Service) ResolveNewTarget(flags Flags, runtime Runtime, changed Changed, force bool, arg string, hasArg bool, settings resolvedScaffoldSettings) (targetResolution, error) {
	return s.newTargetResolver.Resolve(flags, runtime, changed, force, arg, hasArg, settings)
}

func (s *Service) ResolveInitTarget(runtime Runtime, arg string, hasArg bool, settings resolvedScaffoldSettings) (targetResolution, error) {
	return s.initTargetResolver.Resolve(runtime, arg, hasArg, settings)
}

func (s *Service) NewOptions(flags Flags, runtime Runtime, changed Changed, force bool, arg string, hasArg bool) (Options, error) {
	settings, err := s.ResolveScaffoldSettings(flags, changed, runtime)
	if err != nil {
		return Options{}, err
	}
	target, err := s.ResolveNewTarget(flags, runtime, changed, force, arg, hasArg, settings)
	if err != nil {
		return Options{}, err
	}
	return settings.Options(flags, target), nil
}

func (s *Service) InitOptions(flags Flags, runtime Runtime, changed Changed, arg string, hasArg bool) (Options, error) {
	settings, err := s.ResolveScaffoldSettings(flags, changed, runtime)
	if err != nil {
		return Options{}, err
	}
	target, err := s.ResolveInitTarget(runtime, arg, hasArg, settings)
	if err != nil {
		return Options{}, err
	}
	return settings.Options(flags, target), nil
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

	replay, err := buildReplay(command, opts)
	if err != nil {
		return err
	}

	if err := protocoltoml.WriteReplay(flags.WriteReplayPath, replay); err != nil {
		return fmt.Errorf("failed to write resolved replay after project creation: %w", err)
	}

	return nil
}

func (s *Service) ExecuteScaffoldSpec(creator *Creator, spec ScaffoldSpec) error {
	return s.ScaffoldAndMaybeWriteReplay(creator, spec.Flags, spec.Command, spec.Options)
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

// ResolveNewProjectArgs resolves the project name, target directory, and module path
// from the provided language, module flag, and positional argument.
// For Go projects, it handles both simple project names and full module paths.
// Returns an error if the arguments are invalid or cannot be resolved.
func ResolveNewProjectArgs(lang, module, arg string) (projectName string, targetDir string, modulePath string, err error) {
	return resolveNewProjectArgs(lang, module, arg)
}

// ProjectNameFromGoModulePath extracts the project name from a Go module path.
// For example, "github.com/user/myapp" returns "myapp", and "github.com/user/myapp/v2" also returns "myapp".
func ProjectNameFromGoModulePath(modulePath string) string {
	return domain.ProjectNameFromGoModulePath(modulePath)
}

// ProjectNameFromTargetDir derives the project name from the target directory path.
// It returns the base name of the absolute path. Returns an error if the path cannot be resolved.
func ProjectNameFromTargetDir(targetDir string) (string, error) {
	return projectNameFromTargetDir(targetDir)
}
