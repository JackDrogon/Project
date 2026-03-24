package main

import (
	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

type newCommand struct {
	scaffoldCommandBase
	force bool
}

func newNewCommand(creator *appcreate.Creator) *newCommand {
	return &newCommand{scaffoldCommandBase: newScaffoldCommandBase(creator)}
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
	if cmd.Flags().Changed("replay") {
		return cobra.RangeArgs(0, 1)(cmd, args)
	}

	return cobra.ExactArgs(1)(cmd, args)
}

func (c *newCommand) run(cmd *cobra.Command, args []string) error {
	return c.execute(cmd, args, newScaffoldCommandSpecBuilder{flags: &c.flags, force: &c.force})
}

func init() {
	registerOrderedCommand(commandKeyNew, commandOrderNew, func(deps commandDependencies) *cobra.Command {
		return newNewCommand(deps.creator).buildCommand()
	})
}
