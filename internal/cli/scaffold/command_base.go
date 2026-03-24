package scaffold

import (
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
	service := c.deps.newService()
	spec, err := builder.Build(service, cmd, args)
	if err != nil {
		return err
	}

	return service.ExecuteScaffoldSpec(c.deps.creator(), spec)
}
