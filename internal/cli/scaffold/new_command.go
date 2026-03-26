package scaffold

import (
	"github.com/spf13/cobra"
)

type newCommand struct {
	scaffoldCommandBase
	force bool
}

func NewNewCommand(deps Dependencies) *cobra.Command {
	return (&newCommand{scaffoldCommandBase: newScaffoldCommandBase(deps)}).buildCommand()
}

func (c *newCommand) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [project_name]",
		Short: "Create new project",
		Args:  c.validateArgs,
		RunE:  c.run,
	}

	c.bindSharedFlags(cmd)
	cmd.Flags().BoolVar(&c.force, "force", false, "Remove existing project directory before scaffolding")
	return cmd
}

func (c *newCommand) validateArgs(cmd *cobra.Command, args []string) error {
	if len(args) <= 1 {
		return nil
	}

	return cobra.ExactArgs(1)(cmd, args)
}

func (c *newCommand) run(cmd *cobra.Command, args []string) error {
	return c.execute(cmd, args, newScaffoldCommandSpecBuilder{flags: &c.flags, force: &c.force})
}
