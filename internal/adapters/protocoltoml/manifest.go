package protocoltoml

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	domain "github.com/JackDrogon/project/internal/scaffold"
	toml "github.com/pelletier/go-toml/v2"
)

const (
	ManifestFilename = ".project-template-manifest.toml"
	ManifestVersion  = 2
)

var allowedManifestTemplateVars = []string{"ModulePath", "GoVersion", "Author", "Year"}

type Manifest struct {
	Version     int             `toml:"version"`
	Name        string          `toml:"name"`
	Description string          `toml:"description"`
	Inputs      []ManifestInput `toml:"inputs"`
}

type ManifestInput struct {
	Key         string `toml:"key"`
	TemplateVar string `toml:"template_var"`
	Required    bool   `toml:"required"`
}

// DomainInputs converts the manifest's protocol-level inputs into their
// domain representation, keeping the conversion in one place for all callers.
func (m Manifest) DomainInputs() []domain.ManifestInput {
	result := make([]domain.ManifestInput, 0, len(m.Inputs))
	for _, input := range m.Inputs {
		result = append(result, domain.ManifestInput{Name: input.Key, TemplateVar: input.TemplateVar, Required: input.Required})
	}
	return result
}

func LoadManifest(fsys fs.FS, lang string) (Manifest, bool, error) {
	manifestPath := path.Join(lang, ManifestFilename)
	content, err := fs.ReadFile(fsys, manifestPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Manifest{}, false, nil
		}
		return Manifest{}, false, fmt.Errorf("failed to read template manifest %s: %w", manifestPath, err)
	}

	manifest, err := decodeManifest(content, manifestPath, lang)
	if err != nil {
		return Manifest{}, false, err
	}

	return manifest, true, nil
}

func decodeManifest(content []byte, manifestPath, lang string) (Manifest, error) {
	if path.Base(manifestPath) != ManifestFilename {
		return Manifest{}, fmt.Errorf("template manifest %s must use reserved filename %s", manifestPath, ManifestFilename)
	}
	if err := rejectLegacyJSON(content, manifestPath); err != nil {
		return Manifest{}, err
	}

	var manifest Manifest
	decoder := toml.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("failed to decode template manifest %s: %w", manifestPath, err)
	}
	if err := validateManifest(manifest, manifestPath, lang); err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}

func validateManifest(manifest Manifest, manifestPath, lang string) error {
	if manifest.Version != ManifestVersion {
		return fmt.Errorf("template manifest %s has unsupported version %d", manifestPath, manifest.Version)
	}
	if manifest.Name != lang {
		return fmt.Errorf("template manifest %s name %q must match template directory %q", manifestPath, manifest.Name, lang)
	}

	seenInputs := make(map[string]struct{}, len(manifest.Inputs))
	for _, input := range manifest.Inputs {
		if _, ok := seenInputs[input.Key]; ok {
			return fmt.Errorf("template manifest %s has duplicate input key %q", manifestPath, input.Key)
		}
		seenInputs[input.Key] = struct{}{}

		if !slices.Contains(allowedManifestTemplateVars, input.TemplateVar) {
			return fmt.Errorf("template manifest %s input %q has unsupported template_var %q", manifestPath, input.Key, input.TemplateVar)
		}
	}

	return nil
}

func rejectLegacyJSON(content []byte, sourcePath string) error {
	trimmed := strings.TrimSpace(string(content))
	if strings.HasPrefix(trimmed, "{") {
		return fmt.Errorf("%s contains legacy JSON; only TOML is supported", sourcePath)
	}
	return nil
}
