package create

import (
	"fmt"
	"maps"
	"strings"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
)

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
	lang, err := resolveLang(flags, changed, runtime)
	if err != nil {
		return resolvedScaffoldSettings{}, err
	}

	return resolvedScaffoldSettings{
		Lang:                lang,
		ModulePath:          resolveModulePath(flags, changed, runtime),
		Signoff:             resolveSignoff(flags, changed, runtime),
		NoGit:               flags.NoGit,
		GitMode:             resolveGitMode(flags, changed, runtime),
		TemplateInputValues: resolveTemplateInputValues(runtime),
	}, nil
}

func (defaultNewTargetResolver) Resolve(flags Flags, runtime Runtime, changed Changed, force bool, arg string, hasArg bool, settings resolvedScaffoldSettings) (targetResolution, error) {
	resolvedForce := force
	if !changed.Force && runtime.HasReplay {
		resolvedForce = runtime.Replay.Options.Force
	}

	if !hasArg {
		if !runtime.HasReplay {
			config := activeConfigNewSection(runtime)
			if config == nil || config.ProjectName == nil || strings.TrimSpace(*config.ProjectName) == "" {
				return targetResolution{}, fmt.Errorf("accepts 1 arg(s), received 0")
			}
			arg = *config.ProjectName
			hasArg = true
		} else {
			return targetResolution{
				ProjectName: runtime.Replay.Project.Name,
				TargetDir:   runtime.Replay.Project.TargetDir,
				ModulePath:  settings.ModulePath,
				Force:       resolvedForce,
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
	if explicitModulePath == "" && modulePath == "" {
		modulePath = settings.ModulePath
	}

	return targetResolution{
		ProjectName: projectName,
		TargetDir:   targetDir,
		ModulePath:  modulePath,
		Force:       resolvedForce,
	}, nil
}

func (defaultInitTargetResolver) Resolve(runtime Runtime, arg string, hasArg bool, settings resolvedScaffoldSettings) (targetResolution, error) {
	targetDir := "."
	projectName := ""
	var err error

	if hasArg {
		targetDir = arg
		projectName, err = projectNameFromTargetDir(arg)
		if err != nil {
			return targetResolution{}, err
		}
	} else if runtime.HasReplay {
		projectName = runtime.Replay.Project.Name
		targetDir = runtime.Replay.Project.TargetDir
	} else {
		config := activeConfigInitSection(runtime)
		if config != nil && config.TargetDir != nil && strings.TrimSpace(*config.TargetDir) != "" {
			targetDir = *config.TargetDir
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

	return fmt.Errorf("accepts 1 arg(s), received 0")
}

func resolveLang(flags Flags, changed Changed, runtime Runtime) (string, error) {
	if changed.Lang {
		return flags.Lang, nil
	}
	if runtime.HasReplay {
		return runtime.Replay.Template.Lang, nil
	}
	if value := activeConfigLang(runtime); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("required flag(s) \"lang\" not set")
}

func resolveSignoff(flags Flags, changed Changed, runtime Runtime) bool {
	if changed.Signoff {
		return flags.Signoff
	}
	if runtime.HasReplay {
		return runtime.Replay.Git.Signoff
	}
	if value, ok := activeConfigSignoff(runtime); ok {
		return value
	}
	return flags.Signoff
}

func resolveGitMode(flags Flags, changed Changed, runtime Runtime) string {
	if changed.Git {
		return flags.GitMode
	}
	if changed.NoGit {
		return ""
	}
	if runtime.HasReplay {
		return string(runtime.Replay.Git.Mode)
	}
	if value := activeConfigGitMode(runtime); value != "" {
		return value
	}
	return flags.GitMode
}

func resolveModulePath(flags Flags, changed Changed, runtime Runtime) string {
	if changed.Module {
		return flags.Module
	}
	if runtime.HasReplay {
		if runtime.Replay.Project.ModulePath != "" {
			return runtime.Replay.Project.ModulePath
		}
		return runtime.Replay.Inputs["module_path"]
	}
	if value := activeConfigModule(runtime); value != "" {
		return value
	}
	return flags.Module
}

func resolveTemplateInputValues(runtime Runtime) map[string]string {
	if len(runtime.TemplateInputValues) == 0 {
		return nil
	}
	return maps.Clone(runtime.TemplateInputValues)
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
