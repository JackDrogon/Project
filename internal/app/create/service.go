package create

import (
	"fmt"
	"os"
	"strings"

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
	return options, buildNewResolutionOrigins(req, runtime, settings), nil
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
	return options, buildInitResolutionOrigins(req, runtime), nil
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
	if err := validateNewArgFallback(runtime, hasArg); err != nil {
		return Options{}, err
	}

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
	if spec.Flags.ExplainConfig {
		report, err := buildScaffoldExplainReport(creator, spec)
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
		return s.executeDryRunSpec(creator, spec)
	}

	return s.ScaffoldAndMaybeWriteReplay(creator, spec.Flags, spec.Command, spec.Options)
}

func (s *Service) executeDryRunSpec(creator *Creator, spec ScaffoldSpec) error {
	if err := creator.validateCreateOptions(spec.Options); err != nil {
		return err
	}

	creator.writeCreateStart(spec.Options)

	if err := creator.preflightDestDir(spec.Options); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(creator.w, "Dry-run mode: no files will be created")

	plan, err := creator.BuildDryRunPlan(spec.Options)
	if err != nil {
		return err
	}

	return writeDryRunPlanWithOrigins(creator.w, plan, spec.Options, spec.Origins)
}

func buildNewResolutionOrigins(req NewRequest, runtime Runtime, settings resolvedScaffoldSettings) ResolutionOrigins {
	return ResolutionOrigins{
		Lang:           resolveLangOrigin(req.Changed, runtime),
		ProjectName:    resolveNewProjectNameOrigin(req, runtime),
		TargetDir:      resolveNewTargetDirOrigin(req, runtime),
		Module:         resolveNewModuleOrigin(req, runtime, settings),
		GitMode:        resolveGitModeOrigin(req.Changed, runtime),
		Signoff:        resolveSignoffOrigin(req.Changed, runtime),
		TemplateInputs: resolveTemplateInputOrigins(runtime),
	}
}

func buildInitResolutionOrigins(req InitRequest, runtime Runtime) ResolutionOrigins {
	return ResolutionOrigins{
		Lang:           resolveLangOrigin(req.Changed, runtime),
		ProjectName:    resolveInitProjectNameOrigin(req, runtime),
		TargetDir:      resolveInitTargetDirOrigin(req, runtime),
		Module:         resolveInitModuleOrigin(req, runtime),
		GitMode:        resolveGitModeOrigin(req.Changed, runtime),
		Signoff:        resolveSignoffOrigin(req.Changed, runtime),
		TemplateInputs: resolveTemplateInputOrigins(runtime),
	}
}

func resolveLangOrigin(changed Changed, runtime Runtime) ValueOrigin {
	if changed.Lang {
		return ValueOriginFlag
	}
	if runtime.HasReplay {
		return ValueOriginReplay
	}
	if activeConfigLang(runtime) != "" {
		return activeConfigValueOrigin(runtime)
	}
	return ValueOriginDefault
}

func resolveNewProjectNameOrigin(req NewRequest, runtime Runtime) ValueOrigin {
	if req.HasArg {
		return ValueOriginArg
	}
	if runtime.HasReplay {
		return ValueOriginReplay
	}
	section := activeConfigNewSection(runtime)
	if section != nil && hasNonBlankString(section.ProjectName) {
		return activeConfigValueOrigin(runtime)
	}
	return ValueOriginDefault
}

func resolveNewTargetDirOrigin(req NewRequest, runtime Runtime) ValueOrigin {
	if req.HasArg {
		return ValueOriginArg
	}
	if runtime.HasReplay {
		return ValueOriginReplay
	}
	section := activeConfigNewSection(runtime)
	if section != nil && hasNonBlankString(section.ProjectName) {
		return activeConfigValueOrigin(runtime)
	}
	return ValueOriginDefault
}

func resolveInitProjectNameOrigin(req InitRequest, runtime Runtime) ValueOrigin {
	if req.HasArg {
		return ValueOriginArg
	}
	if runtime.HasReplay {
		return ValueOriginReplay
	}
	section := activeConfigInitSection(runtime)
	if section != nil && hasNonBlankString(section.TargetDir) {
		return activeConfigValueOrigin(runtime)
	}
	return ValueOriginDefault
}

func resolveInitTargetDirOrigin(req InitRequest, runtime Runtime) ValueOrigin {
	if req.HasArg {
		return ValueOriginArg
	}
	if runtime.HasReplay {
		return ValueOriginReplay
	}
	section := activeConfigInitSection(runtime)
	if section != nil && hasNonBlankString(section.TargetDir) {
		return activeConfigValueOrigin(runtime)
	}
	return ValueOriginDefault
}

func resolveNewModuleOrigin(req NewRequest, runtime Runtime, settings resolvedScaffoldSettings) ValueOrigin {
	if req.Changed.Module {
		return ValueOriginFlag
	}
	if runtime.HasReplay {
		return ValueOriginReplay
	}
	if activeConfigModule(runtime) != "" {
		return activeConfigValueOrigin(runtime)
	}
	if req.HasArg {
		_, _, derivedModulePath, err := resolveNewProjectArgs(settings.Lang, "", req.Arg)
		if err == nil && derivedModulePath != "" {
			return ValueOriginArg
		}
	}
	return ValueOriginDefault
}

func resolveInitModuleOrigin(req InitRequest, runtime Runtime) ValueOrigin {
	if req.Changed.Module {
		return ValueOriginFlag
	}
	if runtime.HasReplay {
		return ValueOriginReplay
	}
	if activeConfigModule(runtime) != "" {
		return activeConfigValueOrigin(runtime)
	}
	return ValueOriginDefault
}

func resolveGitModeOrigin(changed Changed, runtime Runtime) ValueOrigin {
	if changed.Git || changed.NoGit {
		return ValueOriginFlag
	}
	if runtime.HasReplay {
		return ValueOriginReplay
	}
	if activeConfigGitMode(runtime) != "" {
		return activeConfigValueOrigin(runtime)
	}
	return ValueOriginDefault
}

func resolveSignoffOrigin(changed Changed, runtime Runtime) ValueOrigin {
	if changed.Signoff {
		return ValueOriginFlag
	}
	if runtime.HasReplay {
		return ValueOriginReplay
	}
	if _, ok := activeConfigSignoff(runtime); ok {
		return activeConfigValueOrigin(runtime)
	}
	return ValueOriginDefault
}

func resolveTemplateInputOrigins(runtime Runtime) map[string]ValueOrigin {
	origins := map[string]ValueOrigin{}
	if runtime.HasReplay {
		for key := range runtime.Replay.Inputs {
			if key == "module_path" {
				continue
			}
			origins[key] = ValueOriginReplay
		}
	} else {
		configOrigin := activeConfigValueOrigin(runtime)
		for key := range activeConfigInputs(runtime.Command, runtime.ActiveConfig) {
			origins[key] = configOrigin
		}
	}
	for key := range runtime.ExplicitSetValues {
		origins[key] = ValueOriginSet
	}
	if len(origins) == 0 {
		return nil
	}
	return origins
}

func activeConfigValueOrigin(runtime Runtime) ValueOrigin {
	source := strings.TrimSpace(string(runtime.ActiveConfig.Source))
	if source == "" || source == "none" {
		return ValueOriginDefault
	}
	return ValueOrigin(source)
}

func hasNonBlankString(v *string) bool {
	return v != nil && strings.TrimSpace(*v) != ""
}

func buildScaffoldExplainReport(creator *Creator, spec ScaffoldSpec) (string, error) {
	manifest, vars, err := creator.templateManifestAndVars(spec.Options)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("config source report:\n")
	_, _ = fmt.Fprintf(&b, "  command: %s\n", spec.Command)
	_, _ = fmt.Fprintf(&b, "  active_config_source: %s\n", explainActiveConfigSource(spec))
	_, _ = fmt.Fprintf(&b, "  active_config_path: %s\n", explainActiveConfigPath(spec))
	b.WriteString("  resolved values:\n")
	writeExplainField(&b, "lang", spec.Options.Lang, spec.Origins.Lang)
	writeExplainField(&b, "project_name", spec.Options.ProjectName, spec.Origins.ProjectName)
	writeExplainField(&b, "target_dir", spec.Options.DestinationDir(), spec.Origins.TargetDir)
	writeExplainField(&b, "module", spec.Options.ModulePath, spec.Origins.Module)
	writeExplainField(&b, "git_mode", string(spec.Options.GitMode), spec.Origins.GitMode)
	writeExplainField(&b, "signoff", fmt.Sprintf("%t", spec.Options.Signoff), spec.Origins.Signoff)
	b.WriteString("  template inputs:\n")

	inputs := resolveDryRunInputs(manifestInputsToDomain(manifest.Inputs), vars)
	if len(inputs) == 0 {
		b.WriteString("    (none)\n")
		return b.String(), nil
	}
	for _, input := range inputs {
		writeExplainField(&b, input.Name, input.Value, explainInputOrigin(spec.Origins, input))
	}

	return b.String(), nil
}

func explainActiveConfigSource(spec ScaffoldSpec) string {
	source := strings.TrimSpace(string(spec.Flags.ActiveConfig.Source))
	if source == "" {
		return "none"
	}
	return source
}

func explainActiveConfigPath(spec ScaffoldSpec) string {
	path := strings.TrimSpace(spec.Flags.ActiveConfig.Path)
	if path == "" {
		return "(none)"
	}
	return path
}

func writeExplainField(b *strings.Builder, name, value string, origin ValueOrigin) {
	_, _ = fmt.Fprintf(b, "    %s: %s (source: %s)\n", name, value, normalizeOrigin(origin))
}

func explainInputOrigin(origins ResolutionOrigins, input domain.DryRunResolvedInput) ValueOrigin {
	if input.TemplateVar == "ModulePath" {
		return normalizeOrigin(origins.Module)
	}
	if origin, ok := origins.TemplateInputs[input.Name]; ok {
		return normalizeOrigin(origin)
	}
	return ValueOriginDefault
}

func normalizeOrigin(origin ValueOrigin) ValueOrigin {
	if origin == "" {
		return ValueOriginDefault
	}
	return origin
}
