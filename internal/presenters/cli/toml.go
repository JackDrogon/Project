package cli

import (
	"io"

	"github.com/JackDrogon/project/internal/app/catalog"
	toml "github.com/pelletier/go-toml/v2"
)

type tomlSummaries struct {
	Templates []tomlSummary `toml:"templates"`
}

type tomlSummary struct {
	Name            string   `toml:"name"`
	Description     string   `toml:"description"`
	ManifestVersion int      `toml:"manifest_version"`
	InputNames      []string `toml:"input_names"`
	FileCount       int      `toml:"file_count"`
	TemplateCount   int      `toml:"template_count"`
	Variables       []string `toml:"variables"`
}

type tomlInspection struct {
	Name            string              `toml:"name"`
	Description     string              `toml:"description"`
	ManifestVersion int                 `toml:"manifest_version"`
	Inputs          []tomlManifestInput `toml:"inputs"`
	FileCount       int                 `toml:"file_count"`
	TemplateCount   int                 `toml:"template_count"`
	Variables       []string            `toml:"variables"`
	Mode            string              `toml:"mode"`
	ShownCount      int                 `toml:"shown_count"`
	Files           []tomlFileDetail    `toml:"files"`
}

type tomlManifestInput struct {
	Name        string `toml:"name"`
	TemplateVar string `toml:"template_var"`
}

type tomlFileDetail struct {
	Source     string `toml:"source"`
	Output     string `toml:"output"`
	IsTemplate bool   `toml:"is_template"`
}

func writeTOMLLangs(w io.Writer, langs []string) error {
	content, err := toml.Marshal(struct {
		Languages []string `toml:"languages"`
	}{Languages: langs})
	if err != nil {
		return err
	}
	_, err = w.Write(content)
	return err
}

func writeTOMLSummaries(w io.Writer, summaries []catalog.Summary) error {
	encoded := tomlSummaries{Templates: make([]tomlSummary, 0, len(summaries))}
	for _, summary := range summaries {
		encoded.Templates = append(encoded.Templates, tomlSummary{
			Name:            summary.Name,
			Description:     summary.Description,
			ManifestVersion: summary.ManifestVersion,
			InputNames:      append([]string(nil), summary.InputNames...),
			FileCount:       summary.FileCount,
			TemplateCount:   summary.TemplateCount,
			Variables:       append([]string(nil), summary.Variables...),
		})
	}

	content, err := toml.Marshal(encoded)
	if err != nil {
		return err
	}
	_, err = w.Write(content)
	return err
}

func writeTOMLInspection(w io.Writer, inspection catalog.Inspection) error {
	encoded := tomlInspection{
		Name:            inspection.Name,
		Description:     inspection.Description,
		ManifestVersion: inspection.ManifestVersion,
		FileCount:       inspection.FileCount,
		TemplateCount:   inspection.TemplateCount,
		Variables:       append([]string(nil), inspection.Variables...),
		Mode:            inspection.Mode,
		ShownCount:      inspection.ShownCount,
		Inputs:          make([]tomlManifestInput, 0, len(inspection.Inputs)),
		Files:           make([]tomlFileDetail, 0, len(inspection.Files)),
	}
	for _, input := range inspection.Inputs {
		encoded.Inputs = append(encoded.Inputs, tomlManifestInput{Name: input.Name, TemplateVar: input.TemplateVar})
	}
	for _, file := range inspection.Files {
		encoded.Files = append(encoded.Files, tomlFileDetail{Source: file.Source, Output: file.Output, IsTemplate: file.IsTemplate})
	}

	content, err := toml.Marshal(encoded)
	if err != nil {
		return err
	}
	_, err = w.Write(content)
	return err
}
