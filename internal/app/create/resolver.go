package create

import (
	"fmt"
	"maps"
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
		NoGit:               resolveNoGit(flags, changed),
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
			return targetResolution{}, fmt.Errorf("accepts 1 arg(s), received 0")
		}
		return targetResolution{
			ProjectName: runtime.Replay.Project.Name,
			TargetDir:   runtime.Replay.Project.TargetDir,
			ModulePath:  settings.ModulePath,
			Force:       resolvedForce,
		}, nil
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

func resolveLang(flags Flags, changed Changed, runtime Runtime) (string, error) {
	if changed.Lang {
		return flags.Lang, nil
	}
	if runtime.HasReplay {
		return runtime.Replay.Template.Lang, nil
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
	return flags.Signoff
}

func resolveNoGit(flags Flags, changed Changed) bool {
	if changed.NoGit {
		return flags.NoGit
	}
	return flags.NoGit
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
	return flags.Module
}

func resolveTemplateInputValues(runtime Runtime) map[string]string {
	if len(runtime.TemplateInputValues) == 0 {
		return nil
	}
	return maps.Clone(runtime.TemplateInputValues)
}
