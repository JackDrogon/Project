package catalog

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"

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
		return errors.New("inspection query requires a language")
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
	minRank := governanceRank(q.MinGovernance)
	filtered := make([]Summary, 0, len(summaries))
	for _, summary := range summaries {
		if governanceRank(summary.GovernanceTier) < minRank {
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
	// Governance and repo-file counts sort descending, names ascending. cmp.Or
	// returns the first non-zero comparison, so each case reads as its
	// tie-breaking chain.
	byName := func(a, b Summary) int { return strings.Compare(a.Name, b.Name) }
	byGovernance := func(a, b Summary) int {
		return cmp.Compare(governanceRank(b.GovernanceTier), governanceRank(a.GovernanceTier))
	}
	byRepoFiles := func(a, b Summary) int { return cmp.Compare(b.RepoFileCount, a.RepoFileCount) }

	switch q.normalizedSort() {
	case SummarySortName:
		slices.SortFunc(summaries, byName)
	case SummarySortGovernance:
		slices.SortFunc(summaries, func(a, b Summary) int {
			return cmp.Or(byGovernance(a, b), byRepoFiles(a, b), byName(a, b))
		})
	case SummarySortRepoFiles:
		slices.SortFunc(summaries, func(a, b Summary) int {
			return cmp.Or(byRepoFiles(a, b), byGovernance(a, b), byName(a, b))
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
	if governanceRank(q.MinGovernance) == 0 {
		return fmt.Errorf("invalid --min-governance %q: must be one of minimal, basic, standard, rich", q.MinGovernance)
	}
	return nil
}

func (q SummaryQuery) validateAssets() error {
	for _, asset := range q.RequiredAssets {
		if !isKnownRepoAsset(asset) {
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
