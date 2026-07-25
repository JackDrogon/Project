package catalog

import (
	"fmt"
	"io/fs"
	"slices"

	"golang.org/x/sync/errgroup"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	"github.com/JackDrogon/project/internal/adapters/templatefs"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

// ManifestLoader loads the template manifest for one language. A nil loader
// passed to NewService falls back to protocoltoml.LoadManifest.
type ManifestLoader func(fsys fs.FS, lang string) (protocoltoml.Manifest, bool, error)

type Service struct {
	fsys         fs.FS
	loadManifest ManifestLoader
}

func NewService(fsys fs.FS, loader ManifestLoader) *Service {
	if loader == nil {
		loader = protocoltoml.LoadManifest
	}
	return &Service{fsys: fsys, loadManifest: loader}
}

func (s *Service) ListLangs() ([]string, error) {
	entries, err := fs.ReadDir(s.fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates: %w", err)
	}

	langs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			langs = append(langs, entry.Name())
		}
	}
	slices.Sort(langs)
	return langs, nil
}

func (s *Service) listSummaries() ([]Summary, error) {
	return s.QuerySummaries(DefaultSummaryQuery())
}

func (s *Service) QuerySummaries(query SummaryQuery) ([]Summary, error) {
	langs, err := s.ListLangs()
	if err != nil {
		return nil, err
	}

	summaries := make([]Summary, len(langs))
	var g errgroup.Group
	for i, lang := range langs {
		g.Go(func() error {
			analysis, err := s.analyze(lang)
			if err != nil {
				return err
			}
			summaries[i] = analysis.summary()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return query.Apply(summaries)
}

func (s *Service) Inspect(lang string, mode InspectMode) (Inspection, error) {
	return s.QueryInspection(InspectionQuery{Lang: lang, Mode: mode})
}

func (s *Service) QueryInspection(query InspectionQuery) (Inspection, error) {
	if err := query.Validate(); err != nil {
		return Inspection{}, err
	}

	analysis, err := s.analyze(query.Lang)
	if err != nil {
		return Inspection{}, err
	}
	return analysis.inspection(query.Mode)
}

// analysis is the per-language raw material shared by the summary and
// inspection views.
type analysis struct {
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

func (s *Service) analyze(lang string) (analysis, error) {
	manifest, found, err := s.loadManifest(s.fsys, lang)
	if err != nil {
		return analysis{}, err
	}
	if !found {
		return analysis{}, fmt.Errorf("unsupported language: %s", lang)
	}

	details, err := templatefs.CollectDetails(s.fsys, lang, templatefs.Manifest{
		SchemaVersion: manifest.Version,
		Name:          manifest.Name,
		Description:   manifest.Description,
		Inputs:        manifest.DomainInputs(),
	})
	if err != nil {
		return analysis{}, err
	}

	files := make([]InspectionFile, 0, len(details.Files))
	for _, file := range details.Files {
		files = append(files, inspectionFileFromDetail(file))
	}

	return analysis{
		name:            details.Name,
		description:     details.Description,
		manifestVersion: details.ManifestVersion,
		inputs:          append([]domain.ManifestInput(nil), details.Inputs...),
		fileCount:       details.FileCount,
		templateCount:   details.TemplateCount,
		variables:       append([]string(nil), details.Variables...),
		repoAssets:      activeRepoAssets.AssetsForFiles(files),
		files:           files,
	}, nil
}

func (a analysis) summary() Summary {
	inspection, _ := a.inspection(InspectModeAll)
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
		GovernanceTier:  governanceTier(inspection),
	}
}

func (a analysis) inspection(mode InspectMode) (Inspection, error) {
	normalized, err := ParseInspectMode(string(mode))
	if err != nil {
		return Inspection{}, err
	}

	files := make([]InspectionFile, 0, len(a.files))
	for _, file := range a.files {
		if inspectModeMatches(file, normalized) {
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

func inspectionFileFromDetail(file templatefs.FileDetail) InspectionFile {
	action := FileActionCopy
	if file.IsTemplate {
		action = FileActionRender
	}
	return InspectionFile{Source: file.Source, Output: file.Output, Action: action, Group: activeRepoAssets.GroupForSource(file.Source)}
}

func inputNames(inputs []domain.ManifestInput) []string {
	result := make([]string, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, input.Name)
	}
	return result
}
