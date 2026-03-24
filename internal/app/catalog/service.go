package catalog

import (
	"fmt"
	"io/fs"
	"sort"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	"github.com/JackDrogon/project/internal/adapters/templatefs"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

type ManifestLoader interface {
	Load(fsys fs.FS, lang string) (protocoltoml.Manifest, bool, error)
}

type manifestLoaderFunc func(fsys fs.FS, lang string) (protocoltoml.Manifest, bool, error)

func (fn manifestLoaderFunc) Load(fsys fs.FS, lang string) (protocoltoml.Manifest, bool, error) {
	return fn(fsys, lang)
}

type Service struct {
	fsys           fs.FS
	manifestLoader ManifestLoader
	analyzer       Analyzer
	executor       QueryExecutor
}

var defaultRepoAssetRegistry = newRepoAssetRegistry(map[string]string{
	"editorconfig": ".editorconfig",
	"dependabot":   ".github/dependabot.yml",
	"ci":           ".github/workflows/ci.yml",
	"gitignore":    ".gitignore",
	"golangci":     ".golangci.yml",
	"goreleaser":   ".goreleaser.yml.tmpl",
	"codecov":      "codecov.yml",
	"contributing": "CONTRIBUTING.md.tmpl",
	"security":     "SECURITY.md.tmpl",
	"justfile":     "justfile.tmpl",
	"typos":        "typos.toml",
})

var (
	defaultInspectModeMatcher  = newInspectModePolicy()
	defaultGovernanceEvaluator = newGovernancePolicy()
)

func NewService(fsys fs.FS, loader ManifestLoader) *Service {
	if loader == nil {
		loader = manifestLoaderFunc(protocoltoml.LoadManifest)
	}
	analyzer := newTemplateAnalyzer(fsys, loader)
	service := &Service{fsys: fsys, manifestLoader: loader, analyzer: analyzer}
	service.executor = newQueryExecutor(service, analyzer)
	return service
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
	sort.Strings(langs)
	return langs, nil
}

func (s *Service) ListSummaries() ([]Summary, error) {
	return s.QuerySummaries(DefaultSummaryQuery())
}

func (s *Service) QuerySummaries(query SummaryQuery) ([]Summary, error) {
	return s.executor.QuerySummaries(query)
}

func (s *Service) Inspect(lang string, mode InspectMode) (Inspection, error) {
	return s.QueryInspection(InspectionQuery{Lang: lang, Mode: mode})
}

func (s *Service) QueryInspection(query InspectionQuery) (Inspection, error) {
	return s.executor.QueryInspection(query)
}

func manifestInputsToDomain(inputs []protocoltoml.ManifestInput) []domain.ManifestInput {
	result := make([]domain.ManifestInput, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, domain.ManifestInput{Name: input.Key, TemplateVar: input.TemplateVar})
	}
	return result
}

func inputNames(inputs []domain.ManifestInput) []string {
	result := make([]string, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, input.Name)
	}
	return result
}

func inspectionFileFromTemplateDetail(file templatefs.FileDetail) InspectionFile {
	action := FileActionCopy
	if file.IsTemplate {
		action = FileActionRender
	}
	return InspectionFile{Source: file.Source, Output: file.Output, Action: action, Group: defaultRepoAssetRegistry.GroupForSource(file.Source)}
}
