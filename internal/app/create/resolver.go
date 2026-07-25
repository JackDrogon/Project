package create

import (
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
)

// errMissingProjectNameArg mirrors cobra's ExactArgs(1) message so `new`
// reports the same error whether the arg check fails in cobra or here.
var errMissingProjectNameArg = errors.New("accepts 1 arg(s), received 0")

type ScaffoldSettingsResolver interface {
	Resolve(Flags, Changed, Runtime) (resolvedScaffoldSettings, error)
}

type NewTargetResolver interface {
	Resolve(Flags, Runtime, Changed, bool, string, bool, resolvedScaffoldSettings) (targetResolution, error)
}

type InitTargetResolver interface {
	Resolve(Runtime, string, bool, resolvedScaffoldSettings) (targetResolution, error)
}

type (
	defaultScaffoldSettingsResolver struct{}
	defaultNewTargetResolver        struct{}
	defaultInitTargetResolver       struct{}
)

func newScaffoldSettingsResolver() ScaffoldSettingsResolver { return defaultScaffoldSettingsResolver{} }
func newNewTargetResolver() NewTargetResolver               { return defaultNewTargetResolver{} }
func newInitTargetResolver() InitTargetResolver             { return defaultInitTargetResolver{} }

func (defaultScaffoldSettingsResolver) Resolve(flags Flags, changed Changed, runtime Runtime) (resolvedScaffoldSettings, error) {
	lang, langOrigin, err := resolveLang(flags, changed, runtime)
	if err != nil {
		return resolvedScaffoldSettings{}, err
	}

	modulePath, moduleOrigin, moduleDefaulted := resolveModulePath(flags, changed, runtime)
	signoff, signoffOrigin := resolveSignoff(flags, changed, runtime)
	gitMode, gitModeOrigin := resolveGitMode(flags, changed, runtime)

	return resolvedScaffoldSettings{
		Lang:                lang,
		ModulePath:          modulePath,
		Signoff:             signoff,
		NoGit:               flags.NoGit,
		GitMode:             gitMode,
		TemplateInputValues: resolveTemplateInputValues(runtime),
		Origins: settingsOrigins{
			Lang:            langOrigin,
			Module:          moduleOrigin,
			GitMode:         gitModeOrigin,
			Signoff:         signoffOrigin,
			ModuleDefaulted: moduleDefaulted,
			TemplateInputs:  resolveTemplateInputOrigins(runtime),
		},
	}, nil
}

func (defaultNewTargetResolver) Resolve(flags Flags, runtime Runtime, changed Changed, force bool, arg string, hasArg bool, settings resolvedScaffoldSettings) (targetResolution, error) {
	resolvedForce := force
	if !changed.Force && runtime.HasReplay {
		resolvedForce = runtime.Replay.Options.Force
	}

	// `new` derives the target dir FROM the project name, so both origins key
	// off the same source (arg, replay, or config project_name).
	nameOrigin := ValueOriginArg

	if !hasArg {
		if !runtime.HasReplay {
			config := activeConfigNewSection(runtime)
			if config == nil || !hasNonBlankString(config.ProjectName) {
				return targetResolution{}, errMissingProjectNameArg
			}
			arg = *config.ProjectName
			nameOrigin = activeConfigValueOrigin(runtime)
		} else {
			return targetResolution{
				ProjectName: runtime.Replay.Project.Name,
				TargetDir:   runtime.Replay.Project.TargetDir,
				ModulePath:  settings.ModulePath,
				Force:       resolvedForce,
				Origins: targetOrigins{
					ProjectName: ValueOriginReplay,
					TargetDir:   ValueOriginReplay,
					Module:      settings.Origins.Module,
				},
			}, nil
		}
	}

	explicitModulePath := ""
	if changed.Module {
		explicitModulePath = flags.Module
	}

	projectName, targetDir, modulePath, err := resolveNewProjectArgs(settings.Lang, explicitModulePath, arg)
	if err != nil {
		return targetResolution{}, err
	}

	moduleOrigin := settings.Origins.Module
	// Only a positional arg may re-attribute the module origin, and only when
	// flag, replay, and config all missed (see resolveModulePath).
	if hasArg && settings.Origins.ModuleDefaulted && modulePath != "" {
		moduleOrigin = ValueOriginArg
	}

	if explicitModulePath == "" && modulePath == "" {
		modulePath = settings.ModulePath
	}

	return targetResolution{
		ProjectName: projectName,
		TargetDir:   targetDir,
		ModulePath:  modulePath,
		Force:       resolvedForce,
		Origins: targetOrigins{
			ProjectName: nameOrigin,
			TargetDir:   nameOrigin,
			Module:      moduleOrigin,
		},
	}, nil
}

func (defaultInitTargetResolver) Resolve(runtime Runtime, arg string, hasArg bool, settings resolvedScaffoldSettings) (targetResolution, error) {
	targetDir := "."
	projectName := ""
	// `init` derives the project name FROM the target dir, so both origins key
	// off the same source (arg, replay, or config target_dir).
	nameOrigin := ValueOriginDefault
	var err error

	if hasArg {
		nameOrigin = ValueOriginArg
		targetDir = arg
		projectName, err = projectNameFromTargetDir(arg)
		if err != nil {
			return targetResolution{}, err
		}
	} else if runtime.HasReplay {
		nameOrigin = ValueOriginReplay
		projectName = runtime.Replay.Project.Name
		targetDir = runtime.Replay.Project.TargetDir
	} else {
		config := activeConfigInitSection(runtime)
		if config != nil && hasNonBlankString(config.TargetDir) {
			targetDir = *config.TargetDir
			nameOrigin = activeConfigValueOrigin(runtime)
		}
		projectName, err = projectNameFromTargetDir(targetDir)
		if err != nil {
			return targetResolution{}, err
		}
	}

	return targetResolution{
		ProjectName:           projectName,
		TargetDir:             targetDir,
		ModulePath:            settings.ModulePath,
		AllowExistingEmptyDir: true,
		Origins: targetOrigins{
			ProjectName: nameOrigin,
			TargetDir:   nameOrigin,
			Module:      settings.Origins.Module,
		},
	}, nil
}

func validateNewArgFallback(runtime Runtime, hasArg bool) error {
	if hasArg || runtime.HasReplay {
		return nil
	}

	config := activeConfigNewSection(runtime)
	if config != nil && config.ProjectName != nil && strings.TrimSpace(*config.ProjectName) != "" {
		return nil
	}

	return errMissingProjectNameArg
}

func resolveLang(flags Flags, changed Changed, runtime Runtime) (string, ValueOrigin, error) {
	if changed.Lang {
		return flags.Lang, ValueOriginFlag, nil
	}
	if runtime.HasReplay {
		return runtime.Replay.Template.Lang, ValueOriginReplay, nil
	}
	if value := activeConfigLang(runtime); value != "" {
		return value, activeConfigValueOrigin(runtime), nil
	}
	return "", ValueOriginDefault, fmt.Errorf("required flag(s) \"lang\" not set")
}

func resolveSignoff(flags Flags, changed Changed, runtime Runtime) (bool, ValueOrigin) {
	if changed.Signoff {
		return flags.Signoff, ValueOriginFlag
	}
	if runtime.HasReplay {
		return runtime.Replay.Git.Signoff, ValueOriginReplay
	}
	if value, ok := activeConfigSignoff(runtime); ok {
		return value, activeConfigValueOrigin(runtime)
	}
	return flags.Signoff, ValueOriginDefault
}

func resolveGitMode(flags Flags, changed Changed, runtime Runtime) (string, ValueOrigin) {
	if changed.Git {
		return flags.GitMode, ValueOriginFlag
	}
	if changed.NoGit {
		return "", ValueOriginFlag
	}
	if runtime.HasReplay {
		return string(runtime.Replay.Git.Mode), ValueOriginReplay
	}
	if value := activeConfigGitMode(runtime); value != "" {
		return value, activeConfigValueOrigin(runtime)
	}
	return flags.GitMode, ValueOriginDefault
}

// resolveModulePath reports defaulted=true only when flag, replay, and config
// all missed; `new` uses it to attribute an arg-derived module to the arg.
func resolveModulePath(flags Flags, changed Changed, runtime Runtime) (value string, origin ValueOrigin, defaulted bool) {
	if changed.Module {
		return flags.Module, ValueOriginFlag, false
	}
	if runtime.HasReplay {
		if runtime.Replay.Project.ModulePath != "" {
			return runtime.Replay.Project.ModulePath, ValueOriginReplay, false
		}
		return runtime.Replay.Inputs["module_path"], ValueOriginReplay, false
	}
	if value := activeConfigModule(runtime); value != "" {
		return value, activeConfigValueOrigin(runtime), false
	}
	return flags.Module, ValueOriginDefault, true
}

func resolveTemplateInputValues(runtime Runtime) map[string]string {
	if len(runtime.TemplateInputValues) == 0 {
		return nil
	}
	return maps.Clone(runtime.TemplateInputValues)
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

func activeConfigNewSection(runtime Runtime) *protocoltoml.ConfigNewSection {
	if runtime.HasReplay || runtime.ActiveConfig.Config == nil {
		return nil
	}
	return runtime.ActiveConfig.Config.New
}

func activeConfigInitSection(runtime Runtime) *protocoltoml.ConfigInitSection {
	if runtime.HasReplay || runtime.ActiveConfig.Config == nil {
		return nil
	}
	return runtime.ActiveConfig.Config.Init
}

func activeConfigLang(runtime Runtime) string {
	switch runtime.Command {
	case CommandNew:
		if section := activeConfigNewSection(runtime); section != nil && section.Lang != nil {
			return *section.Lang
		}
	case CommandInit:
		if section := activeConfigInitSection(runtime); section != nil && section.Lang != nil {
			return *section.Lang
		}
	}
	return ""
}

func activeConfigModule(runtime Runtime) string {
	switch runtime.Command {
	case CommandNew:
		if section := activeConfigNewSection(runtime); section != nil && section.Module != nil {
			return *section.Module
		}
	case CommandInit:
		if section := activeConfigInitSection(runtime); section != nil && section.Module != nil {
			return *section.Module
		}
	}
	return ""
}

func activeConfigGitMode(runtime Runtime) string {
	switch runtime.Command {
	case CommandNew:
		if section := activeConfigNewSection(runtime); section != nil && section.GitMode != nil {
			return *section.GitMode
		}
	case CommandInit:
		if section := activeConfigInitSection(runtime); section != nil && section.GitMode != nil {
			return *section.GitMode
		}
	}
	return ""
}

func activeConfigSignoff(runtime Runtime) (bool, bool) {
	switch runtime.Command {
	case CommandNew:
		if section := activeConfigNewSection(runtime); section != nil && section.Signoff != nil {
			return *section.Signoff, true
		}
	case CommandInit:
		if section := activeConfigInitSection(runtime); section != nil && section.Signoff != nil {
			return *section.Signoff, true
		}
	}
	return false, false
}
