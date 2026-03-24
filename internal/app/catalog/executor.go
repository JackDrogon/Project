package catalog

type QueryExecutor interface {
	QuerySummaries(SummaryQuery) ([]Summary, error)
	QueryInspection(InspectionQuery) (Inspection, error)
}

type queryExecutor struct {
	langSource LanguageSource
	analyzer   Analyzer
}

type LanguageSource interface {
	ListLangs() ([]string, error)
}

func newQueryExecutor(langSource LanguageSource, analyzer Analyzer) QueryExecutor {
	return &queryExecutor{langSource: langSource, analyzer: analyzer}
}

func (e *queryExecutor) QuerySummaries(query SummaryQuery) ([]Summary, error) {
	langs, err := e.langSource.ListLangs()
	if err != nil {
		return nil, err
	}

	results, err := projectLangs(langs, e.analyzer, summaryProjector{})
	if err != nil {
		return nil, err
	}
	return query.Apply(results)
}

func (e *queryExecutor) QueryInspection(query InspectionQuery) (Inspection, error) {
	if err := query.Validate(); err != nil {
		return Inspection{}, err
	}

	analysis, err := e.analyzer.Analyze(query.Lang)
	if err != nil {
		return Inspection{}, err
	}

	return inspectionProjector{mode: query.Mode}.Project(analysis)
}
