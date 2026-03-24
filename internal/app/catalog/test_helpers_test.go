package catalog

import (
	"io/fs"
	"slices"
	"strings"
)

type failingFS struct{ err error }

func (f failingFS) Open(string) (fs.File, error)          { return nil, f.err }
func (f failingFS) ReadDir(string) ([]fs.DirEntry, error) { return nil, f.err }

type stubRepoAssetRegistry struct {
	known  []string
	groups map[string]FileGroup
	assets []string
}

func (r stubRepoAssetRegistry) KnownAssets() []string      { return append([]string(nil), r.known...) }
func (r stubRepoAssetRegistry) HasAsset(asset string) bool { return slices.Contains(r.known, asset) }
func (r stubRepoAssetRegistry) AssetsForFiles([]InspectionFile) []string {
	return append([]string(nil), r.assets...)
}

func (r stubRepoAssetRegistry) GroupForSource(source string) FileGroup {
	if group, ok := r.groups[source]; ok {
		return group
	}
	return FileGroupLanguage
}

type stubGovernancePolicy struct {
	tier string
	rank map[string]int
}

func (p stubGovernancePolicy) Tier(Inspection) string { return p.tier }
func (p stubGovernancePolicy) Rank(tier string) int {
	if rank, ok := p.rank[tier]; ok {
		return rank
	}
	return 0
}

func filepathToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
