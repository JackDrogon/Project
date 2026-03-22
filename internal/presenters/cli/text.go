package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/JackDrogon/project/internal/app/catalog"
)

func WriteTextLangs(w io.Writer, langs []string) error {
	var b strings.Builder
	for _, lang := range langs {
		fmt.Fprintln(&b, lang)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func WriteTextSummaries(w io.Writer, summaries []catalog.Summary) error {
	var b strings.Builder
	for _, summary := range summaries {
		vars := "(none)"
		if len(summary.Variables) > 0 {
			vars = strings.Join(summary.Variables, ", ")
		}

		inputs := "(none)"
		if len(summary.InputNames) > 0 {
			inputs = strings.Join(summary.InputNames, ", ")
		}

		description := summary.Description
		if description == "" {
			description = "(none)"
		}

		fmt.Fprintf(&b, "%s\tdesc=%q\tmanifest=v%d\tinputs=[%s]\tfiles=%d\ttemplates=%d\tvars=[%s]\n", summary.Name, description, summary.ManifestVersion, inputs, summary.FileCount, summary.TemplateCount, vars)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func WriteTextInspection(w io.Writer, inspection catalog.Inspection) error {
	var b strings.Builder
	vars := "(none)"
	if len(inspection.Variables) > 0 {
		vars = strings.Join(inspection.Variables, ", ")
	}
	inputs := "(none)"
	if len(inspection.Inputs) > 0 {
		parts := make([]string, 0, len(inspection.Inputs))
		for _, input := range inspection.Inputs {
			parts = append(parts, fmt.Sprintf("%s->%s", input.Name, input.TemplateVar))
		}
		inputs = strings.Join(parts, ", ")
	}
	description := inspection.Description
	if description == "" {
		description = "(none)"
	}

	fmt.Fprintf(&b, "name: %s\n", inspection.Name)
	fmt.Fprintf(&b, "description: %s\n", description)
	fmt.Fprintf(&b, "manifest_version: %d\n", inspection.ManifestVersion)
	fmt.Fprintf(&b, "inputs: %s\n", inputs)
	fmt.Fprintf(&b, "files: %d\n", inspection.FileCount)
	fmt.Fprintf(&b, "templates: %d\n", inspection.TemplateCount)
	fmt.Fprintf(&b, "variables: %s\n", vars)
	fmt.Fprintf(&b, "mode: %s\n", inspection.Mode)
	fmt.Fprintf(&b, "shown: %d\n", inspection.ShownCount)

	for _, file := range inspection.Files {
		mode := "copy"
		if file.IsTemplate {
			mode = "render"
		}
		fmt.Fprintf(&b, "- %s -> %s (%s)\n", file.Source, file.Output, mode)
	}

	_, err := io.WriteString(w, b.String())
	return err
}
