package scaffold

import (
	"github.com/spf13/cobra"
)

type initCommand struct {
	scaffoldCommandBase
}

func NewInitCommand(deps Dependencies) *cobra.Command {
	return (&initCommand{scaffoldCommandBase: newScaffoldCommandBase(deps)}).buildCommand()
}

func (c *initCommand) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [target_dir]",
		Short: "Initialize project in current or target directory",
		Args:  cobra.RangeArgs(0, 1),
		RunE:  c.run,
	}

	c.bindSharedFlags(cmd)
	return cmd
}

func (c *initCommand) run(cmd *cobra.Command, args []string) error {
	return c.execute(cmd, args, initScaffoldCommandSpecBuilder{flags: &c.flags})
}
