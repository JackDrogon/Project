package catalog

import (
	"slices"
	"sort"
)

type RepoAssetRegistry interface {
	KnownAssets() []string
	HasAsset(string) bool
	GroupForSource(string) FileGroup
	AssetsForFiles([]InspectionFile) []string
}

type InspectModePolicy interface {
	Matches(InspectionFile, InspectMode) bool
}

type GovernancePolicy interface {
	Rank(string) int
	Tier(Inspection) string
}

type staticRepoAssetRegistry struct {
	pathsByAsset map[string]string
	assetByPath  map[string]string
}

type (
	defaultInspectModePolicy struct{}
	defaultGovernancePolicy  struct{}
)

func newRepoAssetRegistry(pathsByAsset map[string]string) RepoAssetRegistry {
	assetByPath := make(map[string]string, len(pathsByAsset))
	for asset, path := range pathsByAsset {
		assetByPath[path] = asset
	}
	return &staticRepoAssetRegistry{pathsByAsset: pathsByAsset, assetByPath: assetByPath}
}

func newInspectModePolicy() InspectModePolicy {
	return defaultInspectModePolicy{}
}

func newGovernancePolicy() GovernancePolicy {
	return defaultGovernancePolicy{}
}

func (r *staticRepoAssetRegistry) KnownAssets() []string {
	assets := make([]string, 0, len(r.pathsByAsset))
	for asset := range r.pathsByAsset {
		assets = append(assets, asset)
	}
	sort.Strings(assets)
	return assets
}

func (r *staticRepoAssetRegistry) HasAsset(asset string) bool {
	_, ok := r.pathsByAsset[asset]
	return ok
}

func (r *staticRepoAssetRegistry) GroupForSource(source string) FileGroup {
	if _, ok := r.assetByPath[source]; ok {
		return FileGroupRepo
	}
	return FileGroupLanguage
}

func (r *staticRepoAssetRegistry) AssetsForFiles(files []InspectionFile) []string {
	seen := make(map[string]struct{})
	assets := make([]string, 0, len(r.pathsByAsset))
	for _, file := range files {
		asset, ok := r.assetByPath[file.Source]
		if !ok {
			continue
		}
		if _, exists := seen[asset]; exists {
			continue
		}
		seen[asset] = struct{}{}
		assets = append(assets, asset)
	}
	sort.Strings(assets)
	return assets
}

func (defaultInspectModePolicy) Matches(file InspectionFile, mode InspectMode) bool {
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

func (defaultGovernancePolicy) Rank(tier string) int {
	switch tier {
	case "rich":
		return 4
	case "standard":
		return 3
	case "basic":
		return 2
	case "minimal":
		return 1
	default:
		return 0
	}
}

func (p defaultGovernancePolicy) Tier(inspection Inspection) string {
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
