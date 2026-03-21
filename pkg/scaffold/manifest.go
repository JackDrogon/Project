package scaffold

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"slices"
)

const (
	templateManifestFilename      = ".project-template-manifest.json"
	templateManifestSchemaVersion = 1
)

var allowedManifestTemplateVars = []string{"ModulePath", "GoVersion", "Author", "Year"}

type TemplateManifest struct {
	SchemaVersion int                     `json:"schema_version"`
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	Inputs        []TemplateManifestInput `json:"inputs"`
}

type TemplateManifestInput struct {
	Name        string `json:"name"`
	TemplateVar string `json:"template_var"`
}

func isReservedTemplateFile(name string) bool {
	return path.Base(name) == templateManifestFilename
}

func loadTemplateManifest(fsys fs.FS, lang string) (TemplateManifest, bool, error) {
	manifestPath := path.Join(lang, templateManifestFilename)
	content, err := fs.ReadFile(fsys, manifestPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return TemplateManifest{}, false, nil
		}
		return TemplateManifest{}, false, fmt.Errorf("failed to read template manifest %s: %w", manifestPath, err)
	}

	manifest, err := decodeTemplateManifest(content, manifestPath, lang)
	if err != nil {
		return TemplateManifest{}, false, err
	}

	return manifest, true, nil
}

func decodeTemplateManifest(content []byte, manifestPath, lang string) (TemplateManifest, error) {
	if !isReservedTemplateFile(manifestPath) {
		return TemplateManifest{}, fmt.Errorf("template manifest %s must use reserved filename %s", manifestPath, templateManifestFilename)
	}

	var manifest TemplateManifest
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return TemplateManifest{}, fmt.Errorf("failed to decode template manifest %s: %w", manifestPath, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return TemplateManifest{}, fmt.Errorf("failed to decode template manifest %s: expected a single JSON object", manifestPath)
	}

	if err := validateTemplateManifest(manifest, manifestPath, lang); err != nil {
		return TemplateManifest{}, err
	}

	return manifest, nil
}

func validateTemplateManifest(manifest TemplateManifest, manifestPath, lang string) error {
	if manifest.SchemaVersion != templateManifestSchemaVersion {
		return fmt.Errorf("template manifest %s has unsupported schema_version %d", manifestPath, manifest.SchemaVersion)
	}

	if manifest.Name != lang {
		return fmt.Errorf("template manifest %s name %q must match template directory %q", manifestPath, manifest.Name, lang)
	}

	seenInputs := make(map[string]struct{}, len(manifest.Inputs))
	for _, input := range manifest.Inputs {
		if _, ok := seenInputs[input.Name]; ok {
			return fmt.Errorf("template manifest %s has duplicate input name %q", manifestPath, input.Name)
		}
		seenInputs[input.Name] = struct{}{}

		if !slices.Contains(allowedManifestTemplateVars, input.TemplateVar) {
			return fmt.Errorf("template manifest %s input %q has unsupported template_var %q", manifestPath, input.Name, input.TemplateVar)
		}
	}

	return nil
}
