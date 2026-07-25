package catalog

import (
	"fmt"
	"slices"
	"sort"

	appconfig "github.com/JackDrogon/project/internal/app/config"
)

// configFormatTOML is the [list]/[inspect] config section format value that
// selects TOML output.
const configFormatTOML = "toml"

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

// ListSettings are the effective `list` command settings after config
// defaults have been applied.
type ListSettings struct {
	TOML           bool
	Compact        bool
	Detail         bool
	Table          bool
	SortBy         string
	MinGovernance  string
	RequiredAssets []string
}

// ListSettingsChanged records which `list` flags were set explicitly.
type ListSettingsChanged struct {
	TOML           bool
	Compact        bool
	Detail         bool
	Table          bool
	Sort           bool
	MinGovernance  bool
	RequiredAssets bool
}

// ResolveListSettings fills settings from the [list] config section for
// every flag that was not set explicitly.
func ResolveListSettings(settings ListSettings, changed ListSettingsChanged, active appconfig.ActiveConfig) ListSettings {
	if active.Config == nil || active.Config.List == nil {
		return settings
	}

	configList := active.Config.List
	if !changed.TOML && configList.Format != nil {
		settings.TOML = *configList.Format == configFormatTOML
	}
	if !changed.Compact && configList.Compact != nil {
		settings.Compact = *configList.Compact
	}
	if !changed.Detail && configList.Detail != nil {
		settings.Detail = *configList.Detail
	}
	if !changed.Table && configList.Table != nil {
		settings.Table = *configList.Table
	}
	if !changed.Sort && configList.Sort != nil {
		settings.SortBy = *configList.Sort
	}
	if !changed.MinGovernance && configList.MinGovernance != nil {
		settings.MinGovernance = *configList.MinGovernance
	}
	if !changed.RequiredAssets && configList.RequiredAsset != nil {
		settings.RequiredAssets = append([]string(nil), configList.RequiredAsset...)
	}

	return settings
}

// InspectSettings are the effective `inspect` command settings after config
// defaults have been applied.
type InspectSettings struct {
	Lang    string
	TOML    bool
	Compact bool
	Mode    string
}

// InspectSettingsChanged records which `inspect` inputs were set explicitly;
// Lang marks a positional language argument.
type InspectSettingsChanged struct {
	Lang    bool
	TOML    bool
	Compact bool
	Mode    bool
}

// ResolveInspectSettings fills settings from the [inspect] config section
// for every input that was not set explicitly.
func ResolveInspectSettings(settings InspectSettings, changed InspectSettingsChanged, active appconfig.ActiveConfig) InspectSettings {
	if active.Config == nil || active.Config.Inspect == nil {
		return settings
	}

	configInspect := active.Config.Inspect
	if !changed.Lang && configInspect.Lang != nil {
		settings.Lang = *configInspect.Lang
	}
	if !changed.TOML && configInspect.Format != nil {
		settings.TOML = *configInspect.Format == configFormatTOML
	}
	if !changed.Compact && configInspect.Compact != nil {
		settings.Compact = *configInspect.Compact
	}
	if !changed.Mode && configInspect.Mode != nil {
		settings.Mode = *configInspect.Mode
	}

	return settings
}
