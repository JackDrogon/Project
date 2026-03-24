package create

import (
	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	"github.com/JackDrogon/project/internal/adapters/templatefs"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

func (c *Creator) validateModulePath(opts Options) error {
	if opts.Lang != "go" {
		return nil
	}

	return domain.ValidateModulePath(defaultModulePath(opts))
}

func (c *Creator) validateTemplateInputs(opts Options) error {
	_, _, err := c.templateManifestAndVars(opts)
	return err
}

func (c *Creator) copyTemplates(opts Options) error {
	_, vars, err := c.templateManifestAndVars(opts)
	if err != nil {
		return err
	}

	return templatefs.Materialize(c.w, c.fsys, opts.Lang, opts.DestinationDir(), vars, c.resolveMode)
}

func (c *Creator) templateManifestAndVars(opts Options) (protocoltoml.Manifest, domain.TemplateVars, error) {
	manifest, found, err := protocoltoml.LoadManifest(c.fsys, opts.Lang)
	if err != nil {
		return protocoltoml.Manifest{}, domain.TemplateVars{}, err
	}
	if !found {
		manifest = protocoltoml.Manifest{Name: opts.Lang}
	}

	base := newDefaultTemplateVars(opts.ProjectName, opts.ModulePath)
	vars, err := domain.ResolveTemplateVars(manifestInputsToDomain(manifest.Inputs), opts, base)
	if err != nil {
		return protocoltoml.Manifest{}, domain.TemplateVars{}, err
	}

	return manifest, vars, nil
}

func manifestInputsToDomain(inputs []protocoltoml.ManifestInput) []domain.ManifestInput {
	result := make([]domain.ManifestInput, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, domain.ManifestInput{Name: input.Key, TemplateVar: input.TemplateVar})
	}
	return result
}

func defaultModulePath(opts Options) string {
	return domain.DefaultModulePath(opts)
}
