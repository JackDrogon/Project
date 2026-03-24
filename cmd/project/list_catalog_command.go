package main

import (
	"fmt"
	"slices"
	"sort"

	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	"github.com/JackDrogon/project/internal/presenters"
	"github.com/spf13/cobra"
)

const (
	listSortName       = "name"
	listSortGovernance = "governance"
	listSortRepoFiles  = "repo-files"
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
	return &listCommand{sortBy: listSortName}
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
	cmd.Flags().StringVar(&c.sortBy, "sort", listSortName, "Sort detail rows by: name, governance, repo-files")
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
		summaries, err := service.ListSummaries()
		if err != nil {
			return err
		}
		summaries, err = filterSummaries(summaries, c.minGovernance, c.requiredAssets)
		if err != nil {
			return err
		}
		if err := sortSummaries(summaries, c.sortBy); err != nil {
			return err
		}
		return presenter.WriteSummaries(cmd.OutOrStdout(), summaries)
	}
	if c.table {
		return fmt.Errorf("--table requires --detail; plain list output only supports language names")
	}
	if c.sortBy != listSortName {
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

func filterSummaries(summaries []appcatalog.Summary, minGovernance string, requiredAssets []string) ([]appcatalog.Summary, error) {
	minRank, err := parseGovernanceFilter(minGovernance)
	if err != nil {
		return nil, err
	}
	for _, asset := range requiredAssets {
		if !isValidRepoAsset(asset) {
			return nil, fmt.Errorf("invalid --has-repo-asset %q: must match a known repo asset", asset)
		}
	}
	filtered := make([]appcatalog.Summary, 0, len(summaries))
	for _, summary := range summaries {
		if governanceSortRank(summary.GovernanceTier) < minRank {
			continue
		}
		if !summaryHasAllRepoAssets(summary, requiredAssets) {
			continue
		}
		filtered = append(filtered, summary)
	}
	return filtered, nil
}

func parseGovernanceFilter(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	rank := governanceSortRank(value)
	if rank == 0 {
		return 0, fmt.Errorf("invalid --min-governance %q: must be one of minimal, basic, standard, rich", value)
	}
	return rank, nil
}

func summaryHasAllRepoAssets(summary appcatalog.Summary, assets []string) bool {
	for _, asset := range assets {
		if !slices.Contains(summary.RepoAssets, asset) {
			return false
		}
	}
	return true
}

func isValidRepoAsset(asset string) bool {
	return appcatalog.IsKnownRepoAsset(asset)
}

func sortSummaries(summaries []appcatalog.Summary, mode string) error {
	switch mode {
	case "", listSortName:
		sort.Slice(summaries, func(i, j int) bool {
			return summaries[i].Name < summaries[j].Name
		})
	case listSortGovernance:
		sort.Slice(summaries, func(i, j int) bool {
			left := governanceSortRank(summaries[i].GovernanceTier)
			right := governanceSortRank(summaries[j].GovernanceTier)
			if left != right {
				return left > right
			}
			if summaries[i].RepoFileCount != summaries[j].RepoFileCount {
				return summaries[i].RepoFileCount > summaries[j].RepoFileCount
			}
			return summaries[i].Name < summaries[j].Name
		})
	case listSortRepoFiles:
		sort.Slice(summaries, func(i, j int) bool {
			if summaries[i].RepoFileCount != summaries[j].RepoFileCount {
				return summaries[i].RepoFileCount > summaries[j].RepoFileCount
			}
			left := governanceSortRank(summaries[i].GovernanceTier)
			right := governanceSortRank(summaries[j].GovernanceTier)
			if left != right {
				return left > right
			}
			return summaries[i].Name < summaries[j].Name
		})
	default:
		return fmt.Errorf("invalid --sort %q: must be one of %s, %s, %s", mode, listSortName, listSortGovernance, listSortRepoFiles)
	}
	return nil
}

func governanceSortRank(tier string) int {
	return appcatalog.GovernanceRank(tier)
}

func init() {
	registerOrderedCommand(commandKeyList, commandOrderList, func(commandDependencies) *cobra.Command {
		return newListCommand().buildCommand()
	})
}
