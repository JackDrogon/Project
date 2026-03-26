package main

import (
	"context"
	"fmt"
	"io"
	"os"

	appconfig "github.com/JackDrogon/project/internal/app/config"
	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

var (
	exitFunc               = os.Exit
	stderrWriter io.Writer = os.Stderr
)

type rootCommandFlags struct {
	configPath    string
	explainConfig bool
}

// newRootCmd builds the command tree with all subcommands registered explicitly.
func newRootCmd(creator *appcreate.Creator) *cobra.Command {
	deps := commandDependencies{creator: creator}
	flags := rootCommandFlags{}
	rootCmd := &cobra.Command{
		Use:   "project",
		Short: "project is a tool to create new project",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			active, err := newConfigService().LoadActiveConfig(appconfig.Context{ExplicitPath: flags.configPath})
			if err != nil {
				return err
			}

			sharedContext := appconfig.WithActiveConfig(cmd.Context(), active)
			applyContextToCommandTree(cmd.Root(), sharedContext)
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&flags.configPath, "config", "", "Path to CLI config file")
	rootCmd.PersistentFlags().BoolVar(&flags.explainConfig, "explain-config", false, "Explain resolved config sources on stderr")

	addRegisteredCommands(rootCmd, deps)

	return rootCmd
}

func applyContextToCommandTree(cmd *cobra.Command, ctx context.Context) {
	cmd.SetContext(ctx)
	for _, subCmd := range cmd.Commands() {
		applyContextToCommandTree(subCmd, ctx)
	}
}

// Execute runs the root command.
// If an error occurs during execution, it prints the error to stderr
// and exits the program with status code 1.
func Execute(creator *appcreate.Creator) {
	if err := newRootCmd(creator).Execute(); err != nil {
		_, _ = fmt.Fprintln(stderrWriter, err)
		exitFunc(1)
	}
}
