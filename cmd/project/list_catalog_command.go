package main

import "github.com/spf13/cobra"

type listCommand struct {
	catalogCommandBase
	detail bool
}

func newListCommand() *listCommand {
	return &listCommand{}
}

func (c *listCommand) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all supported languages",
		RunE:  c.run,
	}

	c.bindSharedFlags(cmd)
	cmd.Flags().BoolVar(&c.detail, "detail", false, "Show file/template counts and variables")
	return cmd
}

func (c *listCommand) run(cmd *cobra.Command, args []string) error {
	presenter, err := c.newPresenter()
	if err != nil {
		return err
	}

	service := c.newService()
	if c.detail {
		summaries, err := service.ListSummaries()
		if err != nil {
			return err
		}
		return presenter.WriteSummaries(cmd.OutOrStdout(), summaries)
	}

	langs, err := service.ListLangs()
	if err != nil {
		return err
	}
	return presenter.WriteLangs(cmd.OutOrStdout(), langs)
}

func init() {
	registerOrderedCommand(commandKeyList, commandOrderList, func(commandDependencies) *cobra.Command {
		return newListCommand().buildCommand()
	})
}
