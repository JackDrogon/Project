package catalog

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	"github.com/JackDrogon/project/internal/adapters/templatefs"
	domain "github.com/JackDrogon/project/internal/domain/scaffold"
)

const (
	InspectModeAll    = "all"
	InspectModeRender = "render"
	InspectModeCopy   = "copy"
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
}

type Summary struct {
	Name            string
	Description     string
	ManifestVersion int
	InputNames      []string
	FileCount       int
	TemplateCount   int
	Variables       []string
}

type FileDetail struct {
	Source     string
	Output     string
	IsTemplate bool
}

type Inspection struct {
	Name            string
	Description     string
	ManifestVersion int
	Inputs          []domain.ManifestInput
	FileCount       int
	TemplateCount   int
	Variables       []string
	Mode            string
	ShownCount      int
	Files           []FileDetail
}

func NewService(fsys fs.FS, loader ManifestLoader) *Service {
	if loader == nil {
		loader = manifestLoaderFunc(protocoltoml.LoadManifest)
	}
	return &Service{fsys: fsys, manifestLoader: loader}
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
	langs, err := s.ListLangs()
	if err != nil {
		return nil, err
	}

	results := make([]Summary, 0, len(langs))
	for _, lang := range langs {
		inspection, err := s.Inspect(lang, InspectModeAll)
		if err != nil {
			return nil, err
		}
		results = append(results, Summary{
			Name:            inspection.Name,
			Description:     inspection.Description,
			ManifestVersion: inspection.ManifestVersion,
			InputNames:      inputNames(inspection.Inputs),
			FileCount:       inspection.FileCount,
			TemplateCount:   inspection.TemplateCount,
			Variables:       append([]string(nil), inspection.Variables...),
		})
	}
	return results, nil
}

func (s *Service) Inspect(lang, mode string) (Inspection, error) {
	manifest, found, err := s.manifestLoader.Load(s.fsys, lang)
	if err != nil {
		return Inspection{}, err
	}
	if !found {
		return Inspection{}, fmt.Errorf("unsupported language: %s", lang)
	}

	details, err := templatefs.CollectDetails(s.fsys, lang, templatefs.Manifest{
		SchemaVersion: manifest.Version,
		Name:          manifest.Name,
		Description:   manifest.Description,
		Inputs:        manifestInputsToDomain(manifest.Inputs),
	})
	if err != nil {
		return Inspection{}, err
	}

	normalized, err := normalizeInspectMode(mode)
	if err != nil {
		return Inspection{}, err
	}

	files := make([]FileDetail, 0, len(details.Files))
	for _, file := range details.Files {
		switch normalized {
		case InspectModeAll:
			files = append(files, fileDetailFromTemplatefs(file))
		case InspectModeRender:
			if file.IsTemplate {
				files = append(files, fileDetailFromTemplatefs(file))
			}
		case InspectModeCopy:
			if !file.IsTemplate {
				files = append(files, fileDetailFromTemplatefs(file))
			}
		}
	}

	return Inspection{
		Name:            details.Name,
		Description:     details.Description,
		ManifestVersion: details.ManifestVersion,
		Inputs:          append([]domain.ManifestInput(nil), details.Inputs...),
		FileCount:       details.FileCount,
		TemplateCount:   details.TemplateCount,
		Variables:       append([]string(nil), details.Variables...),
		Mode:            normalized,
		ShownCount:      len(files),
		Files:           files,
	}, nil
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

func fileDetailFromTemplatefs(file templatefs.FileDetail) FileDetail {
	return FileDetail{Source: file.Source, Output: file.Output, IsTemplate: file.IsTemplate}
}

func normalizeInspectMode(mode string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		return InspectModeAll, nil
	}

	switch normalized {
	case InspectModeAll, InspectModeRender, InspectModeCopy:
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid --mode %q: must be one of %s, %s, %s", mode, InspectModeAll, InspectModeRender, InspectModeCopy)
	}
}
