package main

import (
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	"github.com/spf13/cobra"
)

type inspectCommand struct {
	catalogCommandBase
	filter string
}

func newInspectCommand() *inspectCommand {
	return &inspectCommand{filter: string(appcatalog.InspectModeAll)}
}

func (c *inspectCommand) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect [lang]",
		Short: "Inspect one template language",
		Args:  cobra.ExactArgs(1),
		RunE:  c.run,
	}

	c.bindSharedFlags(cmd)
	cmd.Flags().StringVar(&c.filter, "mode", string(appcatalog.InspectModeAll), "File mode: all, render, copy")
	return cmd
}

func (c *inspectCommand) run(cmd *cobra.Command, args []string) error {
	presenter, err := c.newPresenter()
	if err != nil {
		return err
	}

	mode, err := appcatalog.ParseInspectMode(c.filter)
	if err != nil {
		return err
	}

	query := appcatalog.InspectionQuery{Lang: args[0], Mode: mode}
	inspection, err := c.newService().QueryInspection(query)
	if err != nil {
		return err
	}

	return presenter.WriteInspection(cmd.OutOrStdout(), inspection)
}

func init() {
	registerOrderedCommand(commandKeyInspect, commandOrderInspect, func(commandDependencies) *cobra.Command {
		return newInspectCommand().buildCommand()
	})
}
