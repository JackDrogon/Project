package main

import (
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
//
// The resolved config is shared with the subcommands through a pointer rather
// than through context values. Subcommands are constructed before --config has
// been parsed, so they capture the pointer here and read through it once
// PersistentPreRunE has filled it in. That leaves cmd.Context() meaning
// cancellation and nothing else.
func newRootCmd(deps dependencies) *cobra.Command {
	flags := rootCommandFlags{}
	resolved := &appconfig.Resolved{}

	rootCmd := &cobra.Command{
		Use:           "project",
		Short:         "project is a tool to create new project",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			options := appconfig.LoadOptions{ExplicitPath: flags.configPath}
			active, err := deps.newConfigService().LoadActiveConfig(options)
			if err != nil {
				// `config` has to keep working on a broken config file - that
				// is how the user inspects and fixes it - so it sees the load
				// error instead of being blocked by it.
				if !isConfigCommand(cmd) {
					return err
				}
				active = appconfig.ActiveConfig{Source: appconfig.SourceNone}
			}

			*resolved = appconfig.Resolved{Active: active, Options: options, LoadErr: err}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&flags.configPath, "config", "", "Path to CLI config file")
	rootCmd.PersistentFlags().BoolVar(&flags.explainConfig, "explain-config", false, "Explain resolved config sources on stderr")

	rootCmd.AddCommand(subcommands(deps, resolved)...)

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
