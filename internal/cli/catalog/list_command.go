package catalog

import (
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
		sortBy:      "name",
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
	cmd.Flags().StringVar(&c.sortBy, "sort", "name", "Sort detail rows by: name, governance, repo-files")
	cmd.Flags().StringVar(&c.minGovernance, "min-governance", "", "Filter detail rows by minimum governance tier: minimal, basic, standard, rich")
	cmd.Flags().StringArrayVar(&c.requiredAssets, "has-repo-asset", nil, "Filter detail rows to templates containing the given repo asset (repeatable)")
	return cmd
}

func (c *listCommand) run(cmd *cobra.Command, args []string) error {
	spec, err := listCommandSpecBuilder{
		asTOML:         c.asTOML,
		compact:        c.compact,
		detail:         c.detail,
		table:          c.table,
		sortBy:         c.sortBy,
		minGovernance:  c.minGovernance,
		requiredAssets: c.requiredAssets,
	}.Build()
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
