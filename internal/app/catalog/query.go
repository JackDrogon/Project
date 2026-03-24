package catalog

import (
	"fmt"
	"slices"
	"sort"
)

type SummarySort string

const (
	SummarySortName       SummarySort = "name"
	SummarySortGovernance SummarySort = "governance"
	SummarySortRepoFiles  SummarySort = "repo-files"
)

type SummaryQuery struct {
	SortBy         SummarySort
	MinGovernance  string
	RequiredAssets []string
}

type InspectionQuery struct {
	Lang string
	Mode InspectMode
}

func DefaultInspectionQuery(lang string) InspectionQuery {
	return InspectionQuery{Lang: lang, Mode: InspectModeAll}
}

func (q InspectionQuery) Validate() error {
	if q.Lang == "" {
		return fmt.Errorf("inspection query requires a language")
	}
	_, err := ParseInspectMode(string(q.Mode))
	return err
}

func DefaultSummaryQuery() SummaryQuery {
	return SummaryQuery{SortBy: SummarySortName}
}

func (q SummaryQuery) Validate() error {
	if q.SortBy == "" {
		q.SortBy = SummarySortName
	}
	if err := q.validateSort(); err != nil {
		return err
	}
	if err := q.validateGovernance(); err != nil {
		return err
	}
	if err := q.validateAssets(); err != nil {
		return err
	}
	return nil
}

func (q SummaryQuery) Apply(summaries []Summary) ([]Summary, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	filtered := q.filter(summaries)
	q.sort(filtered)
	return filtered, nil
}

func (q SummaryQuery) filter(summaries []Summary) []Summary {
	minRank := GovernanceRank(q.MinGovernance)
	filtered := make([]Summary, 0, len(summaries))
	for _, summary := range summaries {
		if GovernanceRank(summary.GovernanceTier) < minRank {
			continue
		}
		if !summaryHasAllRepoAssets(summary, q.RequiredAssets) {
			continue
		}
		filtered = append(filtered, summary)
	}
	return filtered
}

func (q SummaryQuery) sort(summaries []Summary) {
	switch q.normalizedSort() {
	case SummarySortName:
		sort.Slice(summaries, func(i, j int) bool {
			return summaries[i].Name < summaries[j].Name
		})
	case SummarySortGovernance:
		sort.Slice(summaries, func(i, j int) bool {
			left := GovernanceRank(summaries[i].GovernanceTier)
			right := GovernanceRank(summaries[j].GovernanceTier)
			if left != right {
				return left > right
			}
			if summaries[i].RepoFileCount != summaries[j].RepoFileCount {
				return summaries[i].RepoFileCount > summaries[j].RepoFileCount
			}
			return summaries[i].Name < summaries[j].Name
		})
	case SummarySortRepoFiles:
		sort.Slice(summaries, func(i, j int) bool {
			if summaries[i].RepoFileCount != summaries[j].RepoFileCount {
				return summaries[i].RepoFileCount > summaries[j].RepoFileCount
			}
			left := GovernanceRank(summaries[i].GovernanceTier)
			right := GovernanceRank(summaries[j].GovernanceTier)
			if left != right {
				return left > right
			}
			return summaries[i].Name < summaries[j].Name
		})
	}
}

func (q SummaryQuery) validateSort() error {
	switch q.normalizedSort() {
	case SummarySortName, SummarySortGovernance, SummarySortRepoFiles:
		return nil
	default:
		return fmt.Errorf("invalid --sort %q: must be one of %s, %s, %s", q.SortBy, SummarySortName, SummarySortGovernance, SummarySortRepoFiles)
	}
}

func (q SummaryQuery) validateGovernance() error {
	if q.MinGovernance == "" {
		return nil
	}
	if GovernanceRank(q.MinGovernance) == 0 {
		return fmt.Errorf("invalid --min-governance %q: must be one of minimal, basic, standard, rich", q.MinGovernance)
	}
	return nil
}

func (q SummaryQuery) validateAssets() error {
	for _, asset := range q.RequiredAssets {
		if !IsKnownRepoAsset(asset) {
			return fmt.Errorf("invalid --has-repo-asset %q: must match a known repo asset", asset)
		}
	}
	return nil
}

func (q SummaryQuery) normalizedSort() SummarySort {
	if q.SortBy == "" {
		return SummarySortName
	}
	return q.SortBy
}

func summaryHasAllRepoAssets(summary Summary, assets []string) bool {
	for _, asset := range assets {
		if !slices.Contains(summary.RepoAssets, asset) {
			return false
		}
	}
	return true
}
