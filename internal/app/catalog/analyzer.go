package catalog

import (
	"fmt"
	"io/fs"

	"golang.org/x/sync/errgroup"

	"github.com/JackDrogon/project/internal/adapters/templatefs"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

type Analyzer interface {
	Analyze(lang string) (Analysis, error)
}

type analysisProjector[T any] interface {
	Project(Analysis) (T, error)
}

type templateAnalyzer struct {
	fsys           fs.FS
	manifestLoader ManifestLoader
	deps           Dependencies
}

type Analysis struct {
	deps            Dependencies
	name            string
	description     string
	manifestVersion int
	inputs          []domain.ManifestInput
	fileCount       int
	templateCount   int
	variables       []string
	repoAssets      []string
	files           []InspectionFile
}

type AnalysisViewBuilder interface {
	Summary() Summary
	Inspection(mode InspectMode) (Inspection, error)
}

type analysisViewBuilder struct {
	analysis Analysis
}

type SummaryBuilder interface {
	Build() Summary
}

type InspectionBuilder interface {
	Build(mode InspectMode) (Inspection, error)
}

type summaryBuilder struct {
	analysis Analysis
}

type inspectionBuilder struct {
	analysis Analysis
}

func newTemplateAnalyzerWithDeps(fsys fs.FS, loader ManifestLoader, deps Dependencies) Analyzer {
	return &templateAnalyzer{fsys: fsys, manifestLoader: loader, deps: deps.withDefaults()}
}

func (a *templateAnalyzer) Analyze(lang string) (Analysis, error) {
	manifest, found, err := a.manifestLoader.Load(a.fsys, lang)
	if err != nil {
		return Analysis{}, err
	}
	if !found {
		return Analysis{}, unsupportedLanguageError(lang)
	}

	details, err := templatefs.CollectDetails(a.fsys, lang, templatefs.Manifest{
		SchemaVersion: manifest.Version,
		Name:          manifest.Name,
		Description:   manifest.Description,
		Inputs:        manifestInputsToDomain(manifest.Inputs),
	})
	if err != nil {
		return Analysis{}, err
	}

	files := make([]InspectionFile, 0, len(details.Files))
	for _, file := range details.Files {
		files = append(files, inspectionFileFromTemplateDetailWithRegistry(file, a.deps.RepoAssets))
	}

	return Analysis{
		deps:            a.deps,
		name:            details.Name,
		description:     details.Description,
		manifestVersion: details.ManifestVersion,
		inputs:          append([]domain.ManifestInput(nil), details.Inputs...),
		fileCount:       details.FileCount,
		templateCount:   details.TemplateCount,
		variables:       append([]string(nil), details.Variables...),
		repoAssets:      a.deps.RepoAssets.AssetsForFiles(files),
		files:           files,
	}, nil
}

func (a Analysis) View() AnalysisViewBuilder {
	return analysisViewBuilder{analysis: a}
}

func (b analysisViewBuilder) Summary() Summary {
	return summaryBuilder(b).Build()
}

func (b analysisViewBuilder) Inspection(mode InspectMode) (Inspection, error) {
	return inspectionBuilder(b).Build(mode)
}

func (b inspectionBuilder) Build(mode InspectMode) (Inspection, error) {
	a := b.analysis
	normalized, err := ParseInspectMode(string(mode))
	if err != nil {
		return Inspection{}, err
	}

	files := make([]InspectionFile, 0, len(a.files))
	for _, file := range a.files {
		if a.deps.InspectModes.Matches(file, normalized) {
			files = append(files, file)
		}
	}

	return Inspection{
		Name:            a.name,
		Description:     a.description,
		ManifestVersion: a.manifestVersion,
		Inputs:          append([]domain.ManifestInput(nil), a.inputs...),
		FileCount:       a.fileCount,
		TemplateCount:   a.templateCount,
		Variables:       append([]string(nil), a.variables...),
		RepoAssets:      append([]string(nil), a.repoAssets...),
		Mode:            normalized,
		Files:           files,
	}, nil
}

func (b summaryBuilder) Build() Summary {
	a := b.analysis
	inspection, _ := inspectionBuilder{analysis: a}.Build(InspectModeAll)
	return Summary{
		Name:            a.name,
		Description:     a.description,
		ManifestVersion: a.manifestVersion,
		InputNames:      inputNames(a.inputs),
		FileCount:       a.fileCount,
		TemplateCount:   a.templateCount,
		Variables:       append([]string(nil), a.variables...),
		RepoAssets:      append([]string(nil), a.repoAssets...),
		RepoFileCount:   len(inspection.RepoFiles()),
		GovernanceTier:  a.deps.Governance.Tier(inspection),
	}
}

type summaryProjector struct{}

func (summaryProjector) Project(analysis Analysis) (Summary, error) {
	return analysis.View().Summary(), nil
}

type inspectionProjector struct {
	mode InspectMode
}

func (p inspectionProjector) Project(analysis Analysis) (Inspection, error) {
	return analysis.View().Inspection(p.mode)
}

func projectLangs[T any](langs []string, analyzer Analyzer, projector analysisProjector[T]) ([]T, error) {
	results := make([]T, len(langs))
	var g errgroup.Group
	for i, lang := range langs {
		g.Go(func() error {
			analysis, err := analyzer.Analyze(lang)
			if err != nil {
				return err
			}
			projected, err := projector.Project(analysis)
			if err != nil {
				return err
			}
			results[i] = projected
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		var zero []T
		return zero, err
	}
	return results, nil
}

func unsupportedLanguageError(lang string) error {
	return fmt.Errorf("unsupported language: %s", lang)
}
