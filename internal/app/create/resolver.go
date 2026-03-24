package create

import (
	"fmt"
	"maps"
	"path/filepath"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	domain "github.com/JackDrogon/project/internal/scaffold"
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

func mergeReplayInputs(replay protocoltoml.Replay, templateInputValues map[string]string) map[string]string {
	mergedInputs := make(map[string]string, len(replay.Inputs)+len(templateInputValues))
	for key, value := range replay.Inputs {
		if key == "module_path" {
			continue
		}
		mergedInputs[key] = value
	}
	maps.Copy(mergedInputs, templateInputValues)
	return mergedInputs
}

func buildReplay(command Command, opts Options) (protocoltoml.Replay, error) {
	resolvedGitMode, err := domain.ResolveGitMode(domain.CreateRequest{NoGit: opts.NoGit, GitMode: domain.GitMode(opts.GitMode)})
	if err != nil {
		return protocoltoml.Replay{}, fmt.Errorf("failed to resolve replay after project creation: %w", err)
	}

	inputs := maps.Clone(opts.TemplateInputValues)
	if inputs == nil {
		inputs = map[string]string{}
	}
	if opts.ModulePath != "" {
		inputs["module_path"] = opts.ModulePath
	}

	return protocoltoml.Replay{
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
	}, nil
}

func projectNameFromTargetDir(targetDir string) (string, error) {
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target directory %q: %w", targetDir, err)
	}
	return filepath.Base(absTarget), nil
}
