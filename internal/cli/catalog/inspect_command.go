package catalog

import (
	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	appconfig "github.com/JackDrogon/project/internal/app/config"
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
	builder := inspectCommandSpecBuilder{asTOML: c.asTOML, compact: c.compact, mode: c.filter}
	if len(args) > 0 {
		builder.lang = args[0]
	}
	applyInspectConfigDefaults(cmd, &builder, args)

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

func applyInspectConfigDefaults(cmd *cobra.Command, builder *inspectCommandSpecBuilder, args []string) {
	active, ok := appconfig.ActiveConfigFromContext(cmd.Context())
	if !ok || active.Config == nil || active.Config.Inspect == nil {
		return
	}

	configInspect := active.Config.Inspect
	flags := cmd.Flags()

	if len(args) == 0 && configInspect.Lang != nil {
		builder.lang = *configInspect.Lang
	}
	if !flags.Changed("toml") {
		applyInspectFormatDefault(configInspect, builder)
	}
	if !flags.Changed("compact") && configInspect.Compact != nil {
		builder.compact = *configInspect.Compact
	}
	if !flags.Changed("mode") && configInspect.Mode != nil {
		builder.mode = *configInspect.Mode
	}
}

func applyInspectFormatDefault(configInspect *protocoltoml.ConfigInspect, builder *inspectCommandSpecBuilder) {
	if configInspect.Format == nil {
		return
	}

	builder.asTOML = *configInspect.Format == outputFormatTOML
}
