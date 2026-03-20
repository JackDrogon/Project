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
			targetDir := "."
			if len(args) == 1 {
				targetDir = args[0]
			}

			projectName, err := projectNameFromTargetDir(targetDir)
			if err != nil {
				return err
			}

			opts := shared.options(projectName, targetDir, shared.module)
			opts.AllowExistingEmptyDir = true
			return creator.Create(opts)
		},
	}

	bindCreateCommandFlags(cmd, &shared)

	return cmd
}

func projectNameFromTargetDir(targetDir string) (string, error) {
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target directory %q: %w", targetDir, err)
	}

	return filepath.Base(absTarget), nil
}
