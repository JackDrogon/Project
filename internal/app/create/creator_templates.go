package create

import (
	"context"

	"github.com/JackDrogon/project/internal/adapters/protocoltoml"
	"github.com/JackDrogon/project/internal/adapters/templatefs"
	domain "github.com/JackDrogon/project/internal/scaffold"
)

func (c *Creator) validateModulePath(opts Options) error {
	if opts.Lang != langGo {
		return nil
	}

	return domain.ValidateModulePath(defaultModulePath(opts))
}

func (c *Creator) validateTemplateInputs(ctx context.Context, opts Options) error {
	_, _, err := c.templateManifestAndVars(ctx, opts)
	return err
}

func (c *Creator) copyTemplates(ctx context.Context, opts Options) error {
	_, vars, err := c.templateManifestAndVars(ctx, opts)
	if err != nil {
		return err
	}

	return templatefs.Materialize(c.w, c.fsys, opts.Lang, opts.DestinationDir(), vars, c.resolveMode)
}

func (c *Creator) templateManifestAndVars(ctx context.Context, opts Options) (protocoltoml.Manifest, domain.TemplateVars, error) {
	manifest, found, err := protocoltoml.LoadManifest(c.fsys, opts.Lang)
	if err != nil {
		return protocoltoml.Manifest{}, domain.TemplateVars{}, err
	}
	if !found {
		manifest = protocoltoml.Manifest{Name: opts.Lang}
	}

	base := newDefaultTemplateVars(ctx, opts.ProjectName, opts.ModulePath)
	vars, err := domain.ResolveTemplateVars(manifest.DomainInputs(), opts, base)
	if err != nil {
		return protocoltoml.Manifest{}, domain.TemplateVars{}, err
	}

	return manifest, vars, nil
}

func defaultModulePath(opts Options) string {
	return domain.DefaultModulePath(opts)
}
