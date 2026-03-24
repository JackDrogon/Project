package create

import (
	"fmt"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

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
