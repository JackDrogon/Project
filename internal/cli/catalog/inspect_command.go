package catalog

import (
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
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
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.run,
	}

	c.bindSharedFlags(cmd)
	cmd.Flags().StringVar(&c.filter, "mode", "all", "File mode: all, render, copy")
	return cmd
}

func (c *inspectCommand) run(cmd *cobra.Command, args []string) error {
	settings := c.resolveSettings(cmd, args)
	builder := inspectCommandSpecBuilder{
		asTOML:  settings.TOML,
		compact: settings.Compact,
		lang:    settings.Lang,
		mode:    settings.Mode,
	}

	spec, err := builder.Build()
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

func (c *inspectCommand) resolveSettings(cmd *cobra.Command, args []string) appcatalog.InspectSettings {
	lang := ""
	if len(args) > 0 {
		lang = args[0]
	}

	active := c.deps.activeConfig()
	flags := cmd.Flags()

	return appcatalog.ResolveInspectSettings(
		appcatalog.InspectSettings{
			Lang:    lang,
			TOML:    c.asTOML,
			Compact: c.compact,
			Mode:    c.filter,
		},
		appcatalog.InspectSettingsChanged{
			Lang:    len(args) > 0,
			TOML:    flags.Changed("toml"),
			Compact: flags.Changed("compact"),
			Mode:    flags.Changed("mode"),
		},
		active,
	)
}
