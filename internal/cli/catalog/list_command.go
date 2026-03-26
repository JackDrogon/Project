package catalog

import (
	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	appconfig "github.com/JackDrogon/project/internal/app/config"
	"github.com/JackDrogon/project/internal/presenters"
	"github.com/spf13/cobra"
)

type listCommand struct {
	commandBase
	detail         bool
	table          bool
	sortBy         string
	minGovernance  string
	requiredAssets []string
}

func NewListCommand(deps Dependencies) *cobra.Command {
	command := &listCommand{
		commandBase: commandBase{deps: deps},
		sortBy:      defaultListSortBy,
	}
	return command.buildCommand()
}

func (c *listCommand) buildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all supported languages",
		RunE:  c.run,
	}

	c.bindSharedFlags(cmd)
	cmd.Flags().BoolVar(&c.detail, "detail", false, "Show file/template counts and variables")
	cmd.Flags().BoolVar(&c.table, "table", false, "Render detail rows as a comparison table (text only)")
	cmd.Flags().StringVar(&c.sortBy, "sort", defaultListSortBy, "Sort detail rows by: name, governance, repo-files")
	cmd.Flags().StringVar(&c.minGovernance, "min-governance", "", "Filter detail rows by minimum governance tier: minimal, basic, standard, rich")
	cmd.Flags().StringArrayVar(&c.requiredAssets, "has-repo-asset", nil, "Filter detail rows to templates containing the given repo asset (repeatable)")
	return cmd
}

func (c *listCommand) run(cmd *cobra.Command, args []string) error {
	builder := listCommandSpecBuilder{
		asTOML:         c.asTOML,
		compact:        c.compact,
		detail:         c.detail,
		table:          c.table,
		sortBy:         c.sortBy,
		minGovernance:  c.minGovernance,
		requiredAssets: c.requiredAssets,
	}
	applyListConfigDefaults(cmd, &builder)

	spec, err := builder.Build()
	if err != nil {
		return err
	}

	service := c.newService()
	if spec.detail {
		presenter, err := presenters.NewPresenter(spec.outputSpec)
		if err != nil {
			return err
		}
		summaries, err := service.QuerySummaries(spec.query)
		if err != nil {
			return err
		}
		return presenter.WriteSummaries(cmd.OutOrStdout(), summaries)
	}

	langs, err := service.ListLangs()
	if err != nil {
		return err
	}
	presenter, err := presenters.NewPresenter(spec.outputSpec)
	if err != nil {
		return err
	}
	return presenter.WriteLangs(cmd.OutOrStdout(), langs)
}

func applyListConfigDefaults(cmd *cobra.Command, builder *listCommandSpecBuilder) {
	active, ok := appconfig.ActiveConfigFromContext(cmd.Context())
	if !ok || active.Config == nil || active.Config.List == nil {
		return
	}

	configList := active.Config.List
	flags := cmd.Flags()

	if !flags.Changed("toml") {
		applyListFormatDefault(configList, builder)
	}
	if !flags.Changed("compact") && configList.Compact != nil {
		builder.compact = *configList.Compact
	}
	if !flags.Changed("detail") && configList.Detail != nil {
		builder.detail = *configList.Detail
	}
	if !flags.Changed("table") && configList.Table != nil {
		builder.table = *configList.Table
	}
	if !flags.Changed("sort") && configList.Sort != nil {
		builder.sortBy = *configList.Sort
	}
	if !flags.Changed("min-governance") && configList.MinGovernance != nil {
		builder.minGovernance = *configList.MinGovernance
	}
	if !flags.Changed("has-repo-asset") && configList.RequiredAsset != nil {
		builder.requiredAssets = append([]string(nil), configList.RequiredAsset...)
	}
}

func applyListFormatDefault(configList *protocoltoml.ConfigListSection, builder *listCommandSpecBuilder) {
	if configList.Format == nil {
		return
	}

	builder.asTOML = *configList.Format == outputFormatTOML
}
