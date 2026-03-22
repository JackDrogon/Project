package main

import (
	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

// newNewCmd creates the "new" subcommand that scaffolds a project from templates.
func newNewCmd(creator *appcreate.Creator) *cobra.Command {
	var shared scaffoldCommandFlags
	var force bool

	cmd := &cobra.Command{
		Use:   "new [project_name]",
		Short: "Create new project",
		Args: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("replay") {
				return cobra.RangeArgs(0, 1)(cmd, args)
			}

			return cobra.ExactArgs(1)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := newCreateService()
			opts, err := svc.BuildNewOptions(shared.newRequest(cmd, force, args))
			if err != nil {
				return err
			}

			return svc.ScaffoldAndMaybeWriteReplay(creator, shared.toAppFlags(), appcreate.CommandNew, opts)
		},
	}

	bindScaffoldCommandFlags(cmd, &shared)
	cmd.Flags().BoolVar(&force, "force", false, "Remove existing project directory before scaffolding")

	return cmd
}

func resolveNewProjectArgs(lang, module, arg string) (projectName string, targetDir string, modulePath string, err error) {
	return appcreate.ResolveNewProjectArgs(lang, module, arg)
}

func projectNameFromGoModulePath(modulePath string) string {
	return appcreate.ProjectNameFromGoModulePath(modulePath)
}
