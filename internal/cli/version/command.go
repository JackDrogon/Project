package version

import (
	"fmt"

	appconfig "github.com/JackDrogon/project/internal/app/config"
	"github.com/spf13/cobra"
)

func NewCommand(deps Dependencies) *cobra.Command {
	var verbose bool
	service := deps.newService()

	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version",
		RunE: func(cmd *cobra.Command, args []string) error {
			applyVersionConfigDefaults(cmd, &verbose)

			if verbose {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), service.Verbose())
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), service.Info())
			return err
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed version info")
	return cmd
}

func applyVersionConfigDefaults(cmd *cobra.Command, verbose *bool) {
	if cmd.Flags().Changed("verbose") {
		return
	}

	active, ok := appconfig.ActiveConfigFromContext(cmd.Context())
	if !ok || active.Config == nil || active.Config.VersionCmd == nil || active.Config.VersionCmd.Verbose == nil {
		return
	}

	*verbose = *active.Config.VersionCmd.Verbose
}
