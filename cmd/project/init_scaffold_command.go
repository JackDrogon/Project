package main

import (
	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

type initCommand struct {
	scaffoldCommandBase
}

func newInitCommand(creator *appcreate.Creator) *initCommand {
	return &initCommand{scaffoldCommandBase: newScaffoldCommandBase(creator)}
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
	return c.execute(cmd, args, appcreate.CommandInit, c.buildOptions)
}

func (c *initCommand) buildOptions(service *appcreate.Service, cmd *cobra.Command, args []string) (appcreate.Options, error) {
	return service.BuildInitOptions(c.flags.initRequest(cmd, args))
}

func init() {
	registerOrderedCommand(commandKeyInit, commandOrderInit, func(deps commandDependencies) *cobra.Command {
		return newInitCommand(deps.creator).buildCommand()
	})
}
