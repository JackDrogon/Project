package catalog

import (
	"fmt"
	"strings"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

type InspectMode string

const (
	InspectModeAll    InspectMode = "all"
	InspectModeRender InspectMode = "render"
	InspectModeCopy   InspectMode = "copy"
)

type FileAction string

const (
	FileActionRender FileAction = "render"
	FileActionCopy   FileAction = "copy"
)

type FileGroup string

const (
	FileGroupRepo     FileGroup = "repo"
	FileGroupLanguage FileGroup = "language"
)

type Summary struct {
	Name            string
	Description     string
	ManifestVersion int
	InputNames      []string
	FileCount       int
	TemplateCount   int
	Variables       []string
	RepoAssets      []string
	RepoFileCount   int
	GovernanceTier  string
}

type InspectionFile struct {
	Source string
	Output string
	Action FileAction
	Group  FileGroup
}

func (f InspectionFile) IsTemplate() bool {
	return f.Action == FileActionRender
}

type Inspection struct {
	Name            string
	Description     string
	ManifestVersion int
	Inputs          []domain.ManifestInput
	FileCount       int
	TemplateCount   int
	Variables       []string
	RepoAssets      []string
	Mode            InspectMode
	Files           []InspectionFile
}

func (i Inspection) ShownCount() int {
	return len(i.Files)
}

func (i Inspection) RepoFiles() []InspectionFile {
	return filesByGroup(i.Files, FileGroupRepo)
}

func (i Inspection) LanguageFiles() []InspectionFile {
	return filesByGroup(i.Files, FileGroupLanguage)
}

func ParseInspectMode(mode string) (InspectMode, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		return InspectModeAll, nil
	}

	switch InspectMode(normalized) {
	case InspectModeAll, InspectModeRender, InspectModeCopy:
		return InspectMode(normalized), nil
	default:
		return "", fmt.Errorf("invalid --mode %q: must be one of %s, %s, %s", mode, InspectModeAll, InspectModeRender, InspectModeCopy)
	}
}

func KnownRepoAssets() []string {
	return DefaultDependencies().RepoAssets.KnownAssets()
}

func IsKnownRepoAsset(asset string) bool {
	return DefaultDependencies().RepoAssets.HasAsset(asset)
}

func GovernanceRank(tier string) int {
	return DefaultDependencies().Governance.Rank(tier)
}

func filesByGroup(files []InspectionFile, group FileGroup) []InspectionFile {
	filtered := make([]InspectionFile, 0, len(files))
	for _, file := range files {
		if file.Group == group {
			filtered = append(filtered, file)
		}
	}
	return filtered
}
