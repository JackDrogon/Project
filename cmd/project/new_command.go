package main

import (
	"path"
	"regexp"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

var goMajorVersionSuffix = regexp.MustCompile(`^v[2-9][0-9]*$`)

// newNewCmd creates the "new" subcommand that scaffolds a project from templates.
func newNewCmd(creator *scaffold.Creator) *cobra.Command {
	var shared createCommandFlags
	var force bool

	cmd := &cobra.Command{
		Use:   "new [project_name]",
		Short: "Create new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName, targetDir, modulePath, err := resolveNewProjectArgs(shared.lang, shared.module, args[0])
			if err != nil {
				return err
			}

			opts := shared.options(projectName, targetDir, modulePath)
			opts.Force = force
			return creator.Create(opts)
		},
	}

	bindCreateCommandFlags(cmd, &shared)
	cmd.Flags().BoolVar(&force, "force", false, "Remove existing project directory before scaffolding")

	return cmd
}

func resolveNewProjectArgs(lang, module, arg string) (projectName string, targetDir string, modulePath string, err error) {
	if lang == "go" && module == "" {
		if projectErr := scaffold.ValidateProjectName(arg); projectErr != nil {
			if moduleErr := scaffold.ValidateModulePath(arg); moduleErr == nil {
				name := projectNameFromGoModulePath(arg)
				if err := scaffold.ValidateProjectName(name); err != nil {
					return "", "", "", err
				}

				return name, name, arg, nil
			}
		}
	}

	return arg, arg, module, nil
}

func projectNameFromGoModulePath(modulePath string) string {
	name := path.Base(modulePath)
	if !goMajorVersionSuffix.MatchString(name) {
		return name
	}

	parent := path.Base(path.Dir(modulePath))
	if parent == "." || parent == "/" || parent == "" {
		return name
	}

	return parent
}
