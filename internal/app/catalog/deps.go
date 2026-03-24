package catalog

type Dependencies struct {
	RepoAssets   RepoAssetRegistry
	InspectModes InspectModePolicy
	Governance   GovernancePolicy
}

func DefaultDependencies() Dependencies {
	return Dependencies{
		RepoAssets: newRepoAssetRegistry(map[string]string{
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
		}),
		InspectModes: newInspectModePolicy(),
		Governance:   newGovernancePolicy(),
	}
}

func (d Dependencies) withDefaults() Dependencies {
	defaults := DefaultDependencies()
	if d.RepoAssets == nil {
		d.RepoAssets = defaults.RepoAssets
	}
	if d.InspectModes == nil {
		d.InspectModes = defaults.InspectModes
	}
	if d.Governance == nil {
		d.Governance = defaults.Governance
	}
	return d
}
