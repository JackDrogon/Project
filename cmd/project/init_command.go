package main

import (
	"fmt"
	"path/filepath"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

func newInitCmd(creator *scaffold.Creator) *cobra.Command {
	var lang string
	var module string
	var signoff bool
	var dryRun bool
	var noGit bool
	var gitMode string

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

			return creator.Create(scaffold.Options{
				Lang:                  lang,
				ProjectName:           projectName,
				TargetDir:             targetDir,
				ModulePath:            module,
				AllowExistingEmptyDir: true,
				Signoff:               signoff,
				DryRun:                dryRun,
				NoGit:                 noGit,
				GitMode:               scaffold.GitMode(gitMode),
			})
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "", "Programming language for the project")
	_ = cmd.MarkFlagRequired("lang")
	cmd.Flags().StringVarP(&module, "module", "m", "", "Module path (e.g. github.com/user/project)")
	cmd.Flags().BoolVar(&signoff, "signoff", false, "Add Signed-off-by trailer to the initial commit")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview files without creating them")
	cmd.Flags().BoolVar(&noGit, "no-git", false, "Skip git init/add/commit after scaffolding")
	cmd.Flags().StringVar(&gitMode, "git", "", "Git workflow: none, init-only, init+commit (default: init+commit)")
	_ = cmd.Flags().MarkDeprecated("no-git", "use --git none instead")

	return cmd
}

func projectNameFromTargetDir(targetDir string) (string, error) {
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target directory %q: %w", targetDir, err)
	}

	return filepath.Base(absTarget), nil
}
