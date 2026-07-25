package version

import (
	"fmt"

	appversion "github.com/JackDrogon/project/internal/app/version"
	"github.com/spf13/cobra"
)

func NewCommand(deps Dependencies) *cobra.Command {
	var verbose bool
	service := deps.newService()

	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if appversion.ResolveVerbose(verbose, cmd.Flags().Changed("verbose"), deps.activeConfig()) {
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
