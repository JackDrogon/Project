package catalog

import (
	"slices"
)

// governanceTierFunc scores an inspection into a governance tier. Service holds
// one so tests can substitute a scoring rule without swapping a global.
type governanceTierFunc func(Inspection) string

type repoAssetRegistry struct {
	pathsByAsset map[string]string
	assetByPath  map[string]string
}

func newRepoAssetRegistry(pathsByAsset map[string]string) repoAssetRegistry {
	assetByPath := make(map[string]string, len(pathsByAsset))
	for asset, path := range pathsByAsset {
		assetByPath[path] = asset
	}
	return repoAssetRegistry{pathsByAsset: pathsByAsset, assetByPath: assetByPath}
}

func defaultRepoAssets() repoAssetRegistry {
	return newRepoAssetRegistry(map[string]string{
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
}

func (r repoAssetRegistry) KnownAssets() []string {
	assets := make([]string, 0, len(r.pathsByAsset))
	for asset := range r.pathsByAsset {
		assets = append(assets, asset)
	}
	slices.Sort(assets)
	return assets
}

func (r repoAssetRegistry) HasAsset(asset string) bool {
	_, ok := r.pathsByAsset[asset]
	return ok
}

func (r repoAssetRegistry) GroupForSource(source string) FileGroup {
	if _, ok := r.assetByPath[source]; ok {
		return FileGroupRepo
	}
	return FileGroupLanguage
}

func (r repoAssetRegistry) AssetsForFiles(files []InspectionFile) []string {
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
	slices.Sort(assets)
	return assets
}

func inspectModeMatches(file InspectionFile, mode InspectMode) bool {
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

func governanceRank(tier string) int {
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

func defaultGovernanceTier(inspection Inspection) string {
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
