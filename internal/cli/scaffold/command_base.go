package scaffold

import (
	"fmt"

	"github.com/spf13/cobra"
)

type scaffoldCommandBase struct {
	deps  Dependencies
	flags scaffoldCommandFlags
}

func newScaffoldCommandBase(deps Dependencies) scaffoldCommandBase {
	return scaffoldCommandBase{deps: deps}
}

func (c *scaffoldCommandBase) bindSharedFlags(cmd *cobra.Command) {
	bindScaffoldCommandFlags(cmd, &c.flags)
}

func (c *scaffoldCommandBase) execute(
	cmd *cobra.Command,
	args []string,
	builder scaffoldCommandSpecBuilder,
) error {
	if err := validateConfigReplayConflict(cmd); err != nil {
		return err
	}

	service := c.deps.newService()
	spec, err := builder.Build(service, cmd, args)
	if err != nil {
		return err
	}

	return service.ExecuteScaffoldSpec(cmd.Context(), c.deps.newCreator(cmd.OutOrStdout()), spec)
}

func validateConfigReplayConflict(cmd *cobra.Command) error {
	if !cmd.Flags().Changed("config") || !cmd.Flags().Changed("replay") {
		return nil
	}

	return fmt.Errorf("--config and --replay cannot be combined")
}
