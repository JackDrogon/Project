package catalog

import (
	"github.com/JackDrogon/project/internal/presenters"
	"github.com/spf13/cobra"
)

type inspectCommand struct {
	commandBase
	filter string
}

func NewInspectCommand(deps Dependencies) *cobra.Command {
	command := &inspectCommand{
		commandBase: commandBase{deps: deps},
		filter:      "all",
	}
	return command.buildCommand()
}

func (c *inspectCommand) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect [lang]",
		Short: "Inspect one template language",
		Args:  cobra.ExactArgs(1),
		RunE:  c.run,
	}

	c.bindSharedFlags(cmd)
	cmd.Flags().StringVar(&c.filter, "mode", "all", "File mode: all, render, copy")
	return cmd
}

func (c *inspectCommand) run(cmd *cobra.Command, args []string) error {
	spec, err := inspectCommandSpecBuilder{asTOML: c.asTOML, compact: c.compact, lang: args[0], mode: c.filter}.Build()
	if err != nil {
		return err
	}
	presenter, err := presenters.NewPresenter(spec.outputSpec)
	if err != nil {
		return err
	}
	inspection, err := c.newService().QueryInspection(spec.query)
	if err != nil {
		return err
	}

	return presenter.WriteInspection(cmd.OutOrStdout(), inspection)
}
