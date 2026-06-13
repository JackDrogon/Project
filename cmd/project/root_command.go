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
			loadCtx := appconfig.Context{ExplicitPath: flags.configPath}
			service := newConfigService()
			active, err := service.LoadActiveConfig(loadCtx)
			if err != nil {
				if !isConfigCommand(cmd) {
					return err
				}
				active = appconfig.ActiveConfig{Source: appconfig.SourceNone}
			}

			sharedContext := appconfig.WithActiveConfig(cmd.Context(), active)
			sharedContext = appconfig.WithLoadContext(sharedContext, loadCtx)
			sharedContext = appconfig.WithLoadError(sharedContext, err)
			applyContextToCommandTree(cmd.Root(), sharedContext)
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&flags.configPath, "config", "", "Path to CLI config file")
	rootCmd.PersistentFlags().BoolVar(&flags.explainConfig, "explain-config", false, "Explain resolved config sources on stderr")

	addRegisteredCommands(rootCmd, deps)

	return rootCmd
}

func isConfigCommand(cmd *cobra.Command) bool {
	for current := cmd; current != nil; current = current.Parent() {
		if current.Name() == string(commandKeyConfig) {
			return true
		}
	}
	return false
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
