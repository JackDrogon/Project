package main

import (
	"fmt"
	"strings"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

type createCommandFlags struct {
	lang             string
	module           string
	signoff          bool
	dryRun           bool
	noGit            bool
	gitMode          string
	replayPath       string
	writeAnswersPath string
	setValues        []string
}

type createCommandRuntime struct {
	replay              scaffold.AnswersFile
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

func bindCreateCommandFlags(cmd *cobra.Command, flags *createCommandFlags) {
	cmd.Flags().StringVarP(&flags.lang, "lang", "l", "", "Programming language for the project")
	cmd.Flags().StringVarP(&flags.module, "module", "m", "", "Module path (e.g. github.com/user/project)")
	cmd.Flags().BoolVar(&flags.signoff, "signoff", false, "Add Signed-off-by trailer to the initial commit")
	cmd.Flags().BoolVarP(&flags.dryRun, "dry-run", "n", false, "Preview files without creating them")
	cmd.Flags().BoolVar(&flags.noGit, "no-git", false, "Skip git init/add/commit after scaffolding")
	cmd.Flags().StringVar(&flags.gitMode, "git", "", "Git workflow: none, init-only, init+commit (default: init+commit)")
	cmd.Flags().StringVar(&flags.replayPath, "replay", "", "Load project configuration from a JSON replay file")
	cmd.Flags().StringVar(&flags.writeAnswersPath, "write-answers", "", "Write resolved project configuration to a JSON file after success")
	cmd.Flags().StringArrayVar(&flags.setValues, "set", nil, "Set a template-specific input value (key=value)")
	_ = cmd.Flags().MarkDeprecated("no-git", "use --git none instead")
}

func (flags createCommandFlags) runtimeState(expectedCommand scaffold.AnswersCommand) (createCommandRuntime, error) {
	templateInputValues, err := flags.parseSetValues()
	if err != nil {
		return createCommandRuntime{}, err
	}

	if flags.writeAnswersPath != "" && flags.dryRun {
		return createCommandRuntime{}, fmt.Errorf("--write-answers cannot be combined with --dry-run")
	}

	if flags.replayPath == "" {
		return createCommandRuntime{templateInputValues: templateInputValues}, nil
	}

	replay, err := scaffold.ReadAnswersFile(flags.replayPath)
	if err != nil {
		return createCommandRuntime{}, err
	}
	if replay.Command != expectedCommand {
		return createCommandRuntime{}, fmt.Errorf(
			"invalid --replay %q: answers command %q does not match %q",
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

	return createCommandRuntime{replay: replay, hasReplay: true, templateInputValues: mergedInputs}, nil
}

func (flags createCommandFlags) parseSetValues() (map[string]string, error) {
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

func (flags createCommandFlags) resolveLang(cmd *cobra.Command, runtime createCommandRuntime) (string, error) {
	if cmd.Flags().Changed("lang") {
		return flags.lang, nil
	}
	if runtime.hasReplay {
		return runtime.replay.Lang, nil
	}

	return "", fmt.Errorf("required flag(s) \"lang\" not set")
}

func (flags createCommandFlags) resolveSignoff(cmd *cobra.Command, runtime createCommandRuntime) bool {
	if cmd.Flags().Changed("signoff") {
		return flags.signoff
	}
	if runtime.hasReplay {
		return runtime.replay.Create.Signoff
	}

	return flags.signoff
}

func (flags createCommandFlags) resolveNoGit(cmd *cobra.Command) bool {
	if cmd.Flags().Changed("no-git") {
		return flags.noGit
	}

	return flags.noGit
}

func (flags createCommandFlags) resolveGitMode(cmd *cobra.Command, runtime createCommandRuntime) string {
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

func (flags createCommandFlags) resolveModulePath(cmd *cobra.Command, runtime createCommandRuntime) string {
	if cmd.Flags().Changed("module") {
		return flags.module
	}
	if runtime.hasReplay {
		return runtime.replay.TemplateInputs["module_path"]
	}

	return flags.module
}

func (flags createCommandFlags) resolveTemplateInputValues(runtime createCommandRuntime) map[string]string {
	if len(runtime.templateInputValues) == 0 {
		return nil
	}

	resolved := make(map[string]string, len(runtime.templateInputValues))
	for key, value := range runtime.templateInputValues {
		resolved[key] = value
	}

	return resolved
}

func (flags createCommandFlags) options(lang, projectName, targetDir, modulePath string, signoff, noGit bool, gitMode string) scaffold.Options {
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

func (flags createCommandFlags) createWithOptionalAnswers(creator *scaffold.Creator, command scaffold.AnswersCommand, opts scaffold.Options) error {
	if err := creator.Create(opts); err != nil {
		return err
	}
	if flags.writeAnswersPath == "" {
		return nil
	}

	answers, err := creator.AnswersFile(command, opts)
	if err != nil {
		return fmt.Errorf("failed to resolve answers after project creation: %w", err)
	}
	if err := scaffold.WriteAnswersFile(flags.writeAnswersPath, answers); err != nil {
		return fmt.Errorf("failed to write resolved answers after project creation: %w", err)
	}

	return nil
}
