package catalog

import (
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
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
		Args:  cobra.NoArgs,
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
	settings := c.resolveSettings(cmd)
	builder := listCommandSpecBuilder{
		asTOML:         settings.TOML,
		compact:        settings.Compact,
		detail:         settings.Detail,
		table:          settings.Table,
		sortBy:         settings.SortBy,
		minGovernance:  settings.MinGovernance,
		requiredAssets: settings.RequiredAssets,
	}

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

func (c *listCommand) resolveSettings(cmd *cobra.Command) appcatalog.ListSettings {
	active := c.deps.activeConfig()
	flags := cmd.Flags()

	return appcatalog.ResolveListSettings(
		appcatalog.ListSettings{
			TOML:           c.asTOML,
			Compact:        c.compact,
			Detail:         c.detail,
			Table:          c.table,
			SortBy:         c.sortBy,
			MinGovernance:  c.minGovernance,
			RequiredAssets: c.requiredAssets,
		},
		appcatalog.ListSettingsChanged{
			TOML:           flags.Changed("toml"),
			Compact:        flags.Changed("compact"),
			Detail:         flags.Changed("detail"),
			Table:          flags.Changed("table"),
			Sort:           flags.Changed("sort"),
			MinGovernance:  flags.Changed("min-governance"),
			RequiredAssets: flags.Changed("has-repo-asset"),
		},
		active,
	)
}
