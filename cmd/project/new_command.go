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
	var lang string
	var module string
	var force bool
	var signoff bool
	var dryRun bool
	var noGit bool
	var gitMode string

	cmd := &cobra.Command{
		Use:   "new [project_name]",
		Short: "Create new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName, targetDir, modulePath, err := resolveNewProjectArgs(lang, module, args[0])
			if err != nil {
				return err
			}

			return creator.Create(scaffold.Options{
				Lang:        lang,
				ProjectName: projectName,
				TargetDir:   targetDir,
				ModulePath:  modulePath,
				Force:       force,
				Signoff:     signoff,
				DryRun:      dryRun,
				NoGit:       noGit,
				GitMode:     scaffold.GitMode(gitMode),
			})
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "", "Programming language for the project")
	_ = cmd.MarkFlagRequired("lang")
	cmd.Flags().StringVarP(&module, "module", "m", "", "Module path (e.g. github.com/user/project)")
	cmd.Flags().BoolVar(&force, "force", false, "Remove existing project directory before scaffolding")
	cmd.Flags().BoolVar(&signoff, "signoff", false, "Add Signed-off-by trailer to the initial commit")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview files without creating them")
	cmd.Flags().BoolVar(&noGit, "no-git", false, "Skip git init/add/commit after scaffolding")
	cmd.Flags().StringVar(&gitMode, "git", "", "Git workflow: none, init-only, init+commit (default: init+commit)")
	_ = cmd.Flags().MarkDeprecated("no-git", "use --git none instead")

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
