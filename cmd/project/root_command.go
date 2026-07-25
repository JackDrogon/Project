package main

import (
	"context"

	appconfig "github.com/JackDrogon/project/internal/app/config"
	"github.com/spf13/cobra"
)

type rootCommandFlags struct {
	configPath    string
	explainConfig bool
}

// newRootCmd builds the command tree with all subcommands registered explicitly.
//
// Errors and usage are silenced on purpose: cobra prints the error to stderr
// but the usage block to the *out* stream, so main renders both on stderr
// exactly once instead.
func newRootCmd(deps dependencies) *cobra.Command {
	flags := rootCommandFlags{}
	rootCmd := &cobra.Command{
		Use:           "project",
		Short:         "project is a tool to create new project",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			loadCtx := appconfig.Context{ExplicitPath: flags.configPath}
			service := deps.newConfigService()
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

	rootCmd.AddCommand(subcommands(deps)...)

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
