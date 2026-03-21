package main

import (
	"fmt"
	"strings"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

type scaffoldCommandFlags struct {
	lang            string
	module          string
	signoff         bool
	dryRun          bool
	noGit           bool
	gitMode         string
	replayPath      string
	writeReplayPath string
	setValues       []string
}

type scaffoldCommandRuntime struct {
	replay              scaffold.ReplayFile
	hasReplay           bool
	templateInputValues map[string]string
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

func bindScaffoldCommandFlags(cmd *cobra.Command, flags *scaffoldCommandFlags) {
	cmd.Flags().StringVarP(&flags.lang, "lang", "l", "", "Programming language for the project")
	cmd.Flags().StringVarP(&flags.module, "module", "m", "", "Module path (e.g. github.com/user/project)")
	cmd.Flags().BoolVar(&flags.signoff, "signoff", false, "Add Signed-off-by trailer to the initial commit")
	cmd.Flags().BoolVarP(&flags.dryRun, "dry-run", "n", false, "Preview files without creating them")
	cmd.Flags().BoolVar(&flags.noGit, "no-git", false, "Skip git init/add/commit after scaffolding")
	cmd.Flags().StringVar(&flags.gitMode, "git", "", "Git workflow: none, init-only, init+commit (default: init+commit)")
	cmd.Flags().StringVar(&flags.replayPath, "replay", "", "Load project configuration from a JSON replay file")
	cmd.Flags().StringVar(&flags.writeReplayPath, "write-replay", "", "Write resolved project configuration to a JSON file after success")
	cmd.Flags().StringArrayVar(&flags.setValues, "set", nil, "Set a template-specific input value (key=value)")
	_ = cmd.Flags().MarkDeprecated("no-git", "use --git none instead")
}

func (flags scaffoldCommandFlags) runtimeState(expectedCommand scaffold.ReplayCommand) (scaffoldCommandRuntime, error) {
	templateInputValues, err := flags.parseSetValues()
	if err != nil {
		return scaffoldCommandRuntime{}, err
	}

	if flags.writeReplayPath != "" && flags.dryRun {
		return scaffoldCommandRuntime{}, fmt.Errorf("--write-replay cannot be combined with --dry-run")
	}

	if flags.replayPath == "" {
		return scaffoldCommandRuntime{templateInputValues: templateInputValues}, nil
	}

	replay, err := scaffold.ReadReplayFile(flags.replayPath)
	if err != nil {
		return scaffoldCommandRuntime{}, err
	}
	if replay.Command != expectedCommand {
		return scaffoldCommandRuntime{}, fmt.Errorf(
			"invalid --replay %q: replay command %q does not match %q",
			flags.replayPath,
			replay.Command,
			expectedCommand,
		)
	}

	mergedInputs := make(map[string]string, len(replay.TemplateInputs)+len(templateInputValues))
	for key, value := range replay.TemplateInputs {
		if key == "module_path" {
			continue
		}
		mergedInputs[key] = value
	}
	for key, value := range templateInputValues {
		mergedInputs[key] = value
	}

	return scaffoldCommandRuntime{replay: replay, hasReplay: true, templateInputValues: mergedInputs}, nil
}

func (flags scaffoldCommandFlags) parseSetValues() (map[string]string, error) {
	values := make(map[string]string, len(flags.setValues))
	for _, raw := range flags.setValues {
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

func (flags scaffoldCommandFlags) resolveLang(cmd *cobra.Command, runtime scaffoldCommandRuntime) (string, error) {
	if cmd.Flags().Changed("lang") {
		return flags.lang, nil
	}
	if runtime.hasReplay {
		return runtime.replay.Lang, nil
	}

	return "", fmt.Errorf("required flag(s) \"lang\" not set")
}

func (flags scaffoldCommandFlags) resolveSignoff(cmd *cobra.Command, runtime scaffoldCommandRuntime) bool {
	if cmd.Flags().Changed("signoff") {
		return flags.signoff
	}
	if runtime.hasReplay {
		return runtime.replay.Create.Signoff
	}

	return flags.signoff
}

func (flags scaffoldCommandFlags) resolveNoGit(cmd *cobra.Command) bool {
	if cmd.Flags().Changed("no-git") {
		return flags.noGit
	}

	return flags.noGit
}

func (flags scaffoldCommandFlags) resolveGitMode(cmd *cobra.Command, runtime scaffoldCommandRuntime) string {
	if cmd.Flags().Changed("git") {
		return flags.gitMode
	}
	if cmd.Flags().Changed("no-git") {
		return ""
	}
	if runtime.hasReplay {
		return string(runtime.replay.Create.GitMode)
	}

	return flags.gitMode
}

func (flags scaffoldCommandFlags) resolveModulePath(cmd *cobra.Command, runtime scaffoldCommandRuntime) string {
	if cmd.Flags().Changed("module") {
		return flags.module
	}
	if runtime.hasReplay {
		return runtime.replay.TemplateInputs["module_path"]
	}

	return flags.module
}

func (flags scaffoldCommandFlags) resolveTemplateInputValues(runtime scaffoldCommandRuntime) map[string]string {
	if len(runtime.templateInputValues) == 0 {
		return nil
	}

	resolved := make(map[string]string, len(runtime.templateInputValues))
	for key, value := range runtime.templateInputValues {
		resolved[key] = value
	}

	return resolved
}

func (flags scaffoldCommandFlags) options(lang, projectName, targetDir, modulePath string, signoff, noGit bool, gitMode string) scaffold.Options {
	return scaffold.Options{
		Lang:        lang,
		ProjectName: projectName,
		TargetDir:   targetDir,
		ModulePath:  modulePath,
		Signoff:     signoff,
		DryRun:      flags.dryRun,
		NoGit:       noGit,
		GitMode:     scaffold.GitMode(gitMode),
	}
}

func (flags scaffoldCommandFlags) scaffoldAndMaybeWriteReplay(creator *scaffold.Creator, command scaffold.ReplayCommand, opts scaffold.Options) error {
	if err := creator.Create(opts); err != nil {
		return err
	}
	if flags.writeReplayPath == "" {
		return nil
	}

	replay, err := creator.ReplayFile(command, opts)
	if err != nil {
		return fmt.Errorf("failed to resolve replay after project creation: %w", err)
	}
	if err := scaffold.WriteReplayFile(flags.writeReplayPath, replay); err != nil {
		return fmt.Errorf("failed to write resolved replay after project creation: %w", err)
	}

	return nil
}
