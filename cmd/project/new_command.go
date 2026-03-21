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
			opts, err := shared.resolveNewOptions(cmd, force, args)
			if err != nil {
				return err
			}

			return shared.scaffoldAndMaybeWriteReplay(creator, scaffold.ReplayCommandNew, opts)
		},
	}

	bindScaffoldCommandFlags(cmd, &shared)
	cmd.Flags().BoolVar(&force, "force", false, "Remove existing project directory before scaffolding")

	return cmd
}

func (flags scaffoldCommandFlags) resolveNewOptions(cmd *cobra.Command, force bool, args []string) (scaffold.Options, error) {
	runtime, err := flags.runtimeState(scaffold.ReplayCommandNew)
	if err != nil {
		return scaffold.Options{}, err
	}

	lang, err := flags.resolveLang(cmd, runtime)
	if err != nil {
		return scaffold.Options{}, err
	}

	signoff := flags.resolveSignoff(cmd, runtime)
	noGit := flags.resolveNoGit(cmd)
	gitMode := flags.resolveGitMode(cmd, runtime)
	replayModulePath := flags.resolveModulePath(cmd, runtime)

	resolvedForce := force
	if !cmd.Flags().Changed("force") && runtime.hasReplay {
		resolvedForce = runtime.replay.Create.Force
	}

	if len(args) == 0 {
		if !runtime.hasReplay {
			return scaffold.Options{}, cobra.ExactArgs(1)(cmd, args)
		}

		opts := flags.options(
			lang,
			runtime.replay.Create.ProjectName,
			runtime.replay.Create.TargetDir,
			replayModulePath,
			signoff,
			noGit,
			gitMode,
		)
		opts.TemplateInputValues = flags.resolveTemplateInputValues(runtime)
		opts.Force = resolvedForce
		return opts, nil
	}

	explicitModulePath := ""
	if cmd.Flags().Changed("module") {
		explicitModulePath = flags.module
	}

	projectName, targetDir, modulePath, err := resolveNewProjectArgs(lang, explicitModulePath, args[0])
	if err != nil {
		return scaffold.Options{}, err
	}
	if explicitModulePath == "" && modulePath == "" {
		modulePath = replayModulePath
	}

	opts := flags.options(lang, projectName, targetDir, modulePath, signoff, noGit, gitMode)
	opts.TemplateInputValues = flags.resolveTemplateInputValues(runtime)
	opts.Force = resolvedForce
	return opts, nil
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
