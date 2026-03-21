package main

import (
	"fmt"
	"path/filepath"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

func newInitCmd(creator *scaffold.Creator) *cobra.Command {
	var shared createCommandFlags

	cmd := &cobra.Command{
		Use:   "init [target_dir]",
		Short: "Initialize project in current or target directory",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := shared.resolveInitOptions(cmd, args)
			if err != nil {
				return err
			}

			return shared.createWithOptionalAnswers(creator, scaffold.AnswersCommandInit, opts)
		},
	}

	bindCreateCommandFlags(cmd, &shared)

	return cmd
}

func (flags createCommandFlags) resolveInitOptions(cmd *cobra.Command, args []string) (scaffold.Options, error) {
	runtime, err := flags.runtimeState(scaffold.AnswersCommandInit)
	if err != nil {
		return scaffold.Options{}, err
	}

	lang, err := flags.resolveLang(cmd, runtime)
	if err != nil {
		return scaffold.Options{}, err
	}

	targetDir := "."
	projectName := ""
	if len(args) == 1 {
		targetDir = args[0]
		projectName, err = projectNameFromTargetDir(targetDir)
		if err != nil {
			return scaffold.Options{}, err
		}
	} else if runtime.hasReplay {
		projectName = runtime.replay.Create.ProjectName
		targetDir = runtime.replay.Create.TargetDir
	} else {
		projectName, err = projectNameFromTargetDir(targetDir)
		if err != nil {
			return scaffold.Options{}, err
		}
	}

	opts := flags.options(
		lang,
		projectName,
		targetDir,
		flags.resolveModulePath(cmd, runtime),
		flags.resolveSignoff(cmd, runtime),
		flags.resolveNoGit(cmd),
		flags.resolveGitMode(cmd, runtime),
	)
	opts.TemplateInputValues = flags.resolveTemplateInputValues(runtime)
	opts.AllowExistingEmptyDir = true
	return opts, nil
}

func projectNameFromTargetDir(targetDir string) (string, error) {
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target directory %q: %w", targetDir, err)
	}

	return filepath.Base(absTarget), nil
}
