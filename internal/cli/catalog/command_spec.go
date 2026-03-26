package catalog

import (
	"fmt"

	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	"github.com/JackDrogon/project/internal/presenters"
)

type listCommandSpec struct {
	query      appcatalog.SummaryQuery
	outputSpec presenters.OutputSpec
	detail     bool
}

type inspectCommandSpec struct {
	query      appcatalog.InspectionQuery
	outputSpec presenters.OutputSpec
}

type listCommandSpecBuilder struct {
	asTOML         bool
	compact        bool
	detail         bool
	table          bool
	sortBy         string
	minGovernance  string
	requiredAssets []string
}

const defaultListSortBy = string(appcatalog.SummarySortName)

type inspectCommandSpecBuilder struct {
	asTOML  bool
	compact bool
	lang    string
	mode    string
}

func (b listCommandSpecBuilder) Build() (listCommandSpec, error) {
	if !b.detail {
		if b.table {
			return listCommandSpec{}, fmt.Errorf("--table requires --detail; plain list output only supports language names")
		}
		if b.sortBy != defaultListSortBy {
			return listCommandSpec{}, fmt.Errorf("--sort=%s requires --detail; plain list output only supports name order", b.sortBy)
		}
		if b.minGovernance != "" || len(b.requiredAssets) > 0 {
			return listCommandSpec{}, fmt.Errorf("repo-level filters require --detail; plain list output only supports name order")
		}
		return listCommandSpec{detail: false, outputSpec: defaultCatalogOutputSpec(b.asTOML, b.compact)}, nil
	}

	outputSpec := defaultCatalogOutputSpec(b.asTOML, b.compact)
	if b.table {
		var err error
		outputSpec, err = listTableOutputSpec(b.asTOML, b.compact)
		if err != nil {
			return listCommandSpec{}, err
		}
	}

	query := appcatalog.SummaryQuery{
		SortBy:         appcatalog.SummarySort(b.sortBy),
		MinGovernance:  b.minGovernance,
		RequiredAssets: append([]string(nil), b.requiredAssets...),
	}
	if err := query.Validate(); err != nil {
		return listCommandSpec{}, err
	}

	return listCommandSpec{detail: true, query: query, outputSpec: outputSpec}, nil
}

func (b inspectCommandSpecBuilder) Build() (inspectCommandSpec, error) {
	mode, err := appcatalog.ParseInspectMode(b.mode)
	if err != nil {
		return inspectCommandSpec{}, err
	}
	query := appcatalog.DefaultInspectionQuery(b.lang)
	query.Mode = mode
	if err := query.Validate(); err != nil {
		return inspectCommandSpec{}, err
	}
	return inspectCommandSpec{query: query, outputSpec: inspectOutputSpec(b.asTOML, b.compact)}, nil
}
