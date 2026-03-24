package catalog

import (
	"fmt"
	"io/fs"
	"slices"
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

var repoAssetLabelsByPath = map[string]string{
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
}

var repoAssetPathSet = func() map[string]struct{} {
	paths := make(map[string]struct{}, len(repoAssetLabelsByPath))
	for _, path := range repoAssetLabelsByPath {
		paths[path] = struct{}{}
	}
	return paths
}()

func repoAssetsFromInspectionFiles(files []InspectionFile) []string {
	seen := make(map[string]struct{})
	assets := make([]string, 0, len(repoAssetLabelsByPath))
	for _, file := range files {
		for label, path := range repoAssetLabelsByPath {
			if file.Source != path {
				continue
			}
			if _, exists := seen[label]; exists {
				break
			}
			seen[label] = struct{}{}
			assets = append(assets, label)
			break
		}
	}
	sort.Strings(assets)
	return assets
}

func matchesInspectMode(file InspectionFile, mode InspectMode) bool {
	switch mode {
	case InspectModeAll:
		return true
	case InspectModeRender:
		return file.Action == FileActionRender
	case InspectModeCopy:
		return file.Action == FileActionCopy
	default:
		return false
	}
}

func governanceTierForInspection(inspection Inspection) string {
	repoFileCount := len(inspection.RepoFiles())
	hasCI := slices.Contains(inspection.RepoAssets, "ci")
	hasTypos := slices.Contains(inspection.RepoAssets, "typos")
	hasDependabot := slices.Contains(inspection.RepoAssets, "dependabot")
	hasContributing := slices.Contains(inspection.RepoAssets, "contributing")
	hasSecurity := slices.Contains(inspection.RepoAssets, "security")

	switch {
	case repoFileCount >= 8 || (hasCI && hasTypos && hasDependabot && hasContributing && hasSecurity):
		return "rich"
	case repoFileCount >= 5 && hasCI && hasTypos:
		return "standard"
	case repoFileCount > 0:
		return "basic"
	default:
		return "minimal"
	}
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
	group := FileGroupLanguage
	if _, ok := repoAssetPathSet[file.Source]; ok {
		group = FileGroupRepo
	}
	action := FileActionCopy
	if file.IsTemplate {
		action = FileActionRender
	}
	return InspectionFile{Source: file.Source, Output: file.Output, Action: action, Group: group}
}
