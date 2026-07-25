package create

import (
	"context"
	"fmt"
	"os"

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
	options, _, err := s.buildNew(req)
	if err != nil {
		return Options{}, err
	}

	return options, nil
}

func (s *Service) BuildNewSpec(req NewRequest) (ScaffoldSpec, error) {
	options, origins, err := s.buildNew(req)
	if err != nil {
		return ScaffoldSpec{}, err
	}
	return ScaffoldSpec{Command: CommandNew, Flags: req.Flags, Options: options, Origins: origins}, nil
}

func (s *Service) BuildInitOptions(req InitRequest) (Options, error) {
	options, _, err := s.buildInit(req)
	if err != nil {
		return Options{}, err
	}

	return options, nil
}

func (s *Service) BuildInitSpec(req InitRequest) (ScaffoldSpec, error) {
	options, origins, err := s.buildInit(req)
	if err != nil {
		return ScaffoldSpec{}, err
	}
	return ScaffoldSpec{Command: CommandInit, Flags: req.Flags, Options: options, Origins: origins}, nil
}

func (s *Service) buildNew(req NewRequest) (Options, ResolutionOrigins, error) {
	runtime, err := s.RuntimeState(req.Flags, CommandNew)
	if err != nil {
		return Options{}, ResolutionOrigins{}, err
	}
	if err := validateNewArgFallback(runtime, req.HasArg); err != nil {
		return Options{}, ResolutionOrigins{}, err
	}

	settings, err := s.ResolveScaffoldSettings(req.Flags, req.Changed, runtime)
	if err != nil {
		return Options{}, ResolutionOrigins{}, err
	}
	target, err := s.ResolveNewTarget(req.Flags, runtime, req.Changed, req.Force, req.Arg, req.HasArg, settings)
	if err != nil {
		return Options{}, ResolutionOrigins{}, err
	}

	options := settings.Options(req.Flags, target)
	return options, scaffoldResolutionOrigins(settings, target), nil
}

func (s *Service) buildInit(req InitRequest) (Options, ResolutionOrigins, error) {
	runtime, err := s.RuntimeState(req.Flags, CommandInit)
	if err != nil {
		return Options{}, ResolutionOrigins{}, err
	}

	settings, err := s.ResolveScaffoldSettings(req.Flags, req.Changed, runtime)
	if err != nil {
		return Options{}, ResolutionOrigins{}, err
	}
	target, err := s.ResolveInitTarget(runtime, req.Arg, req.HasArg, settings)
	if err != nil {
		return Options{}, ResolutionOrigins{}, err
	}

	options := settings.Options(req.Flags, target)
	return options, scaffoldResolutionOrigins(settings, target), nil
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

func (s *Service) ScaffoldAndMaybeWriteReplay(ctx context.Context, creator *Creator, flags Flags, command Command, opts Options) error {
	if err := creator.Create(ctx, opts); err != nil {
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

func (s *Service) ExecuteScaffoldSpec(ctx context.Context, creator *Creator, spec ScaffoldSpec) error {
	if spec.Flags.ExplainConfig {
		report, err := buildScaffoldExplainReport(ctx, creator, spec)
		if err != nil {
			return err
		}

		stderr := spec.Flags.Stderr
		if stderr == nil {
			stderr = os.Stderr
		}
		_, _ = fmt.Fprint(stderr, report)
	}

	if spec.Options.DryRun {
		return s.executeDryRunSpec(ctx, creator, spec)
	}

	return s.ScaffoldAndMaybeWriteReplay(ctx, creator, spec.Flags, spec.Command, spec.Options)
}

func (s *Service) executeDryRunSpec(ctx context.Context, creator *Creator, spec ScaffoldSpec) error {
	if err := creator.validateCreateOptions(ctx, spec.Options); err != nil {
		return err
	}

	creator.writeCreateStart(spec.Options)

	if err := creator.checkDestDir(spec.Options); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(creator.w, "Dry-run mode: no files will be created")

	plan, err := creator.BuildDryRunPlan(ctx, spec.Options)
	if err != nil {
		return err
	}

	return writeDryRunPlanWithOrigins(creator.w, plan, spec.Options, spec.Origins)
}

// scaffoldResolutionOrigins collects the origins captured while resolving the
// values, so reported origins can never drift from the resolved values.
func scaffoldResolutionOrigins(settings resolvedScaffoldSettings, target targetResolution) ResolutionOrigins {
	return ResolutionOrigins{
		Lang:           settings.Origins.Lang,
		ProjectName:    target.Origins.ProjectName,
		TargetDir:      target.Origins.TargetDir,
		Module:         target.Origins.Module,
		GitMode:        settings.Origins.GitMode,
		Signoff:        settings.Origins.Signoff,
		TemplateInputs: settings.Origins.TemplateInputs,
	}
}
