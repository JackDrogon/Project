package main

import (
	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

// newNewCmd creates the "new" subcommand that scaffolds a project from templates.
func newNewCmd(creator *scaffold.Creator) *cobra.Command {
	var lang string
	var module string
	var force bool
	var signoff bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "new [project_name]",
		Short: "Create new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return creator.Create(scaffold.Options{
				Lang:        lang,
				ProjectName: args[0],
				ModulePath:  module,
				Force:       force,
				Signoff:     signoff,
				DryRun:      dryRun,
			})
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "", "Programming language for the project")
	_ = cmd.MarkFlagRequired("lang")
	cmd.Flags().StringVarP(&module, "module", "m", "", "Module path (e.g. github.com/user/project)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing project directory")
	cmd.Flags().BoolVar(&signoff, "signoff", false, "Add Signed-off-by trailer to the initial commit")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview files without creating them")

	return cmd
}
