package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

const (
	inspectModeAll    = "all"
	inspectModeRender = "render"
	inspectModeCopy   = "copy"
)

type inspectOutput struct {
	Name            string                           `json:"name"`
	Description     string                           `json:"description"`
	ManifestVersion int                              `json:"manifest_version"`
	Inputs          []scaffold.TemplateManifestInput `json:"inputs"`
	FileCount       int                              `json:"file_count"`
	TemplateCount   int                              `json:"template_count"`
	Variables       []string                         `json:"variables"`
	Mode            string                           `json:"mode"`
	ShownCount      int                              `json:"shown_count"`
	Files           []scaffold.TemplateFile          `json:"files"`
}

func newInspectCmd(creator *scaffold.Creator) *cobra.Command {
	var asJSON bool
	var asYAML bool
	var mode string

	cmd := &cobra.Command{
		Use:   "inspect [lang]",
		Short: "Inspect one template language",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			details, err := creator.InspectLang(args[0])
			if err != nil {
				return err
			}

			output, err := buildInspectOutput(details, mode)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			format, err := selectedOutputFormat(asJSON, asYAML)
			if err != nil {
				return err
			}

			switch format {
			case outputFormatJSON:
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(output)
			case outputFormatYAML:
				return writeYAMLInspectOutput(out, output)
			}

			return writeTextInspectOutput(out, output)
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&asYAML, "yaml", false, "Output as YAML")
	cmd.Flags().StringVar(&mode, "mode", inspectModeAll, "File mode: all, render, copy")
	return cmd
}

func buildInspectOutput(details scaffold.TemplateDetails, mode string) (inspectOutput, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		normalized = inspectModeAll
	}

	filtered := make([]scaffold.TemplateFile, 0, len(details.Files))
	for _, file := range details.Files {
		switch normalized {
		case inspectModeAll:
			filtered = append(filtered, file)
		case inspectModeRender:
			if file.IsTemplate {
				filtered = append(filtered, file)
			}
		case inspectModeCopy:
			if !file.IsTemplate {
				filtered = append(filtered, file)
			}
		default:
			return inspectOutput{}, fmt.Errorf("invalid --mode %q: must be one of %s, %s, %s", mode, inspectModeAll, inspectModeRender, inspectModeCopy)
		}
	}

	return inspectOutput{
		Name:            details.Name,
		Description:     details.Description,
		ManifestVersion: details.ManifestVersion,
		Inputs:          append([]scaffold.TemplateManifestInput(nil), details.Inputs...),
		FileCount:       details.FileCount,
		TemplateCount:   details.TemplateCount,
		Variables:       details.Variables,
		Mode:            normalized,
		ShownCount:      len(filtered),
		Files:           filtered,
	}, nil
}

func writeYAMLInspectOutput(w io.Writer, output inspectOutput) error {
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", yamlQuote(output.Name))
	fmt.Fprintf(&b, "description: %s\n", yamlQuote(output.Description))
	fmt.Fprintf(&b, "manifest_version: %d\n", output.ManifestVersion)

	if len(output.Inputs) == 0 {
		fmt.Fprintln(&b, "inputs: []")
	} else {
		fmt.Fprintln(&b, "inputs:")
		for _, input := range output.Inputs {
			fmt.Fprintf(&b, "  - name: %s\n", yamlQuote(input.Name))
			fmt.Fprintf(&b, "    template_var: %s\n", yamlQuote(input.TemplateVar))
		}
	}

	fmt.Fprintf(&b, "file_count: %d\n", output.FileCount)
	fmt.Fprintf(&b, "template_count: %d\n", output.TemplateCount)

	if len(output.Variables) == 0 {
		fmt.Fprintln(&b, "variables: []")
	} else {
		fmt.Fprintln(&b, "variables:")
		for _, v := range output.Variables {
			fmt.Fprintf(&b, "  - %s\n", yamlQuote(v))
		}
	}

	fmt.Fprintf(&b, "mode: %s\n", yamlQuote(output.Mode))
	fmt.Fprintf(&b, "shown_count: %d\n", output.ShownCount)

	if len(output.Files) == 0 {
		fmt.Fprintln(&b, "files: []")
		_, err := io.WriteString(w, b.String())
		return err
	}

	fmt.Fprintln(&b, "files:")
	for _, file := range output.Files {
		fmt.Fprintf(&b, "  - source: %s\n", yamlQuote(file.Source))
		fmt.Fprintf(&b, "    output: %s\n", yamlQuote(file.Output))
		fmt.Fprintf(&b, "    is_template: %t\n", file.IsTemplate)
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func writeTextInspectOutput(w io.Writer, output inspectOutput) error {
	var b strings.Builder
	vars := "(none)"
	if len(output.Variables) > 0 {
		vars = strings.Join(output.Variables, ", ")
	}
	inputs := "(none)"
	if len(output.Inputs) > 0 {
		parts := make([]string, 0, len(output.Inputs))
		for _, input := range output.Inputs {
			parts = append(parts, fmt.Sprintf("%s->%s", input.Name, input.TemplateVar))
		}
		inputs = strings.Join(parts, ", ")
	}
	description := output.Description
	if description == "" {
		description = "(none)"
	}

	fmt.Fprintf(&b, "name: %s\n", output.Name)
	fmt.Fprintf(&b, "description: %s\n", description)
	fmt.Fprintf(&b, "manifest_version: %d\n", output.ManifestVersion)
	fmt.Fprintf(&b, "inputs: %s\n", inputs)
	fmt.Fprintf(&b, "files: %d\n", output.FileCount)
	fmt.Fprintf(&b, "templates: %d\n", output.TemplateCount)
	fmt.Fprintf(&b, "variables: %s\n", vars)
	fmt.Fprintf(&b, "mode: %s\n", output.Mode)
	fmt.Fprintf(&b, "shown: %d\n", output.ShownCount)

	for _, file := range output.Files {
		mode := "copy"
		if file.IsTemplate {
			mode = "render"
		}
		fmt.Fprintf(&b, "- %s -> %s (%s)\n", file.Source, file.Output, mode)
	}

	_, err := io.WriteString(w, b.String())
	return err
}
