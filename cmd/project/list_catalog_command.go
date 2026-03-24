package main

import (
	"fmt"

	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	"github.com/JackDrogon/project/internal/presenters"
	"github.com/spf13/cobra"
)

type listCommand struct {
	catalogCommandBase
	detail         bool
	table          bool
	sortBy         string
	minGovernance  string
	requiredAssets []string
}

func newListCommand() *listCommand {
	return &listCommand{sortBy: string(appcatalog.SummarySortName)}
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
	cmd.Flags().StringVar(&c.sortBy, "sort", string(appcatalog.SummarySortName), "Sort detail rows by: name, governance, repo-files")
	cmd.Flags().StringVar(&c.minGovernance, "min-governance", "", "Filter detail rows by minimum governance tier: minimal, basic, standard, rich")
	cmd.Flags().StringArrayVar(&c.requiredAssets, "has-repo-asset", nil, "Filter detail rows to templates containing the given repo asset (repeatable)")
	return cmd
}

func (c *listCommand) run(cmd *cobra.Command, args []string) error {
	service := c.newService()
	if c.detail {
		presenter, err := c.newListPresenter()
		if err != nil {
			return err
		}
		query := c.summaryQuery()
		summaries, err := service.QuerySummaries(query)
		if err != nil {
			return err
		}
		return presenter.WriteSummaries(cmd.OutOrStdout(), summaries)
	}
	if c.table {
		return fmt.Errorf("--table requires --detail; plain list output only supports language names")
	}
	if c.sortBy != string(appcatalog.SummarySortName) {
		return fmt.Errorf("--sort=%s requires --detail; plain list output only supports name order", c.sortBy)
	}
	if c.minGovernance != "" || len(c.requiredAssets) > 0 {
		return fmt.Errorf("repo-level filters require --detail; plain list output only supports name order")
	}

	langs, err := service.ListLangs()
	if err != nil {
		return err
	}
	presenter, err := c.newPresenter()
	if err != nil {
		return err
	}
	return presenter.WriteLangs(cmd.OutOrStdout(), langs)
}

func (c *listCommand) newListPresenter() (*presenters.Presenter, error) {
	if c.table {
		spec, err := listTableOutputSpec(c.asTOML, c.compact)
		if err != nil {
			return nil, err
		}
		return presenters.NewPresenter(spec)
	}
	return c.newPresenter()
}

func (c *listCommand) summaryQuery() appcatalog.SummaryQuery {
	return appcatalog.SummaryQuery{
		SortBy:         appcatalog.SummarySort(c.sortBy),
		MinGovernance:  c.minGovernance,
		RequiredAssets: append([]string(nil), c.requiredAssets...),
	}
}

func init() {
	registerOrderedCommand(commandKeyList, commandOrderList, func(commandDependencies) *cobra.Command {
		return newListCommand().buildCommand()
	})
}
