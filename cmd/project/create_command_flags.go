package main

import (
	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

type createCommandFlags struct {
	lang    string
	module  string
	signoff bool
	dryRun  bool
	noGit   bool
	gitMode string
}

func bindCreateCommandFlags(cmd *cobra.Command, flags *createCommandFlags) {
	cmd.Flags().StringVarP(&flags.lang, "lang", "l", "", "Programming language for the project")
	_ = cmd.MarkFlagRequired("lang")
	cmd.Flags().StringVarP(&flags.module, "module", "m", "", "Module path (e.g. github.com/user/project)")
	cmd.Flags().BoolVar(&flags.signoff, "signoff", false, "Add Signed-off-by trailer to the initial commit")
	cmd.Flags().BoolVarP(&flags.dryRun, "dry-run", "n", false, "Preview files without creating them")
	cmd.Flags().BoolVar(&flags.noGit, "no-git", false, "Skip git init/add/commit after scaffolding")
	cmd.Flags().StringVar(&flags.gitMode, "git", "", "Git workflow: none, init-only, init+commit (default: init+commit)")
	_ = cmd.Flags().MarkDeprecated("no-git", "use --git none instead")
}

func (flags createCommandFlags) options(projectName, targetDir, modulePath string) scaffold.Options {
	return scaffold.Options{
		Lang:        flags.lang,
		ProjectName: projectName,
		TargetDir:   targetDir,
		ModulePath:  modulePath,
		Signoff:     flags.signoff,
		DryRun:      flags.dryRun,
		NoGit:       flags.noGit,
		GitMode:     scaffold.GitMode(flags.gitMode),
	}
}
