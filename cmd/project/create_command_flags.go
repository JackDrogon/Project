package main

import (
	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

type scaffoldCommandFlags struct {
	lang            string
	module          string
	signoff         bool
	dryRun          bool
	noGit           bool
	gitMode         string
	replayPath      string
	writeReplayPath string
	setValues       []string
}

var newCreateService = func() *appcreate.Service {
	return appcreate.NewService()
}

func bindScaffoldCommandFlags(cmd *cobra.Command, flags *scaffoldCommandFlags) {
	cmd.Flags().StringVarP(&flags.lang, "lang", "l", "", "Programming language for the project")
	cmd.Flags().StringVarP(&flags.module, "module", "m", "", "Module path (e.g. github.com/user/project)")
	cmd.Flags().BoolVar(&flags.signoff, "signoff", false, "Add Signed-off-by trailer to the initial commit")
	cmd.Flags().BoolVarP(&flags.dryRun, "dry-run", "n", false, "Preview files without creating them")
	cmd.Flags().BoolVar(&flags.noGit, "no-git", false, "Skip git init/add/commit after scaffolding")
	cmd.Flags().StringVar(&flags.gitMode, "git", "", "Git workflow: none, init-only, init+commit (default: init+commit)")
	cmd.Flags().StringVar(&flags.replayPath, "replay", "", "Load project configuration from a TOML replay file")
	cmd.Flags().StringVar(&flags.writeReplayPath, "write-replay", "", "Write resolved project configuration to a TOML file after success")
	cmd.Flags().StringArrayVar(&flags.setValues, "set", nil, "Set a template-specific input value (key=value)")
	_ = cmd.Flags().MarkDeprecated("no-git", "use --git none instead")
}

func (flags scaffoldCommandFlags) toAppFlags() appcreate.Flags {
	return appcreate.Flags{
		Lang:            flags.lang,
		Module:          flags.module,
		Signoff:         flags.signoff,
		DryRun:          flags.dryRun,
		NoGit:           flags.noGit,
		GitMode:         flags.gitMode,
		ReplayPath:      flags.replayPath,
		WriteReplayPath: flags.writeReplayPath,
		SetValues:       flags.setValues,
	}
}

func (flags scaffoldCommandFlags) changed(cmd *cobra.Command, force bool) appcreate.Changed {
	return appcreate.Changed{
		Lang:    cmd.Flags().Changed("lang"),
		Module:  cmd.Flags().Changed("module"),
		Signoff: cmd.Flags().Changed("signoff"),
		NoGit:   cmd.Flags().Changed("no-git"),
		Git:     cmd.Flags().Changed("git"),
		Force:   force,
	}
}

func (flags scaffoldCommandFlags) newRequest(cmd *cobra.Command, force bool, args []string) appcreate.NewRequest {
	arg := ""
	hasArg := len(args) == 1
	if hasArg {
		arg = args[0]
	}

	return appcreate.NewRequest{
		Flags:   flags.toAppFlags(),
		Changed: flags.changed(cmd, cmd.Flags().Changed("force")),
		Force:   force,
		Arg:     arg,
		HasArg:  hasArg,
	}
}

func (flags scaffoldCommandFlags) initRequest(cmd *cobra.Command, args []string) appcreate.InitRequest {
	arg := ""
	hasArg := len(args) == 1
	if hasArg {
		arg = args[0]
	}

	return appcreate.InitRequest{
		Flags:   flags.toAppFlags(),
		Changed: flags.changed(cmd, false),
		Arg:     arg,
		HasArg:  hasArg,
	}
}
