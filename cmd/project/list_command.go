package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

const (
	outputFormatText = "text"
	outputFormatJSON = "json"
	outputFormatYAML = "yaml"

	inspectModeAll    = "all"
	inspectModeRender = "render"
	inspectModeCopy   = "copy"
)

type inspectOutput struct {
	Name          string                  `json:"name"`
	FileCount     int                     `json:"file_count"`
	TemplateCount int                     `json:"template_count"`
	Variables     []string                `json:"variables"`
	Mode          string                  `json:"mode"`
	ShownCount    int                     `json:"shown_count"`
	Files         []scaffold.TemplateFile `json:"files"`
}

// newListCmd creates the "list" subcommand that shows available template languages.
func newListCmd(creator *scaffold.Creator) *cobra.Command {
	var detail bool
	var asJSON bool
	var asYAML bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all supported languages",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			format, err := selectedOutputFormat(asJSON, asYAML)
			if err != nil {
				return err
			}

			if detail {
				summaries, err := creator.ListTemplateSummaries()
				if err != nil {
					return err
				}

				switch format {
				case outputFormatJSON:
					enc := json.NewEncoder(out)
					enc.SetIndent("", "  ")
					return enc.Encode(summaries)
				case outputFormatYAML:
					return writeYAMLSummaries(out, summaries)
				}

				return writeTextSummaries(out, summaries)
			}

			langs, err := creator.ListLangs()
			if err != nil {
				return err
			}

			switch format {
			case outputFormatJSON:
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(langs)
			case outputFormatYAML:
				return writeYAMLLangs(out, langs)
			}

			return writeTextLangs(out, langs)
		},
	}

	cmd.Flags().BoolVar(&detail, "detail", false, "Show file/template counts and variables")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&asYAML, "yaml", false, "Output as YAML")

	return cmd
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

func selectedOutputFormat(asJSON, asYAML bool) (string, error) {
	if asJSON && asYAML {
		return "", fmt.Errorf("--json and --yaml cannot be used together")
	}

	if asJSON {
		return outputFormatJSON, nil
	}
	if asYAML {
		return outputFormatYAML, nil
	}

	return outputFormatText, nil
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
		Name:          details.Name,
		FileCount:     details.FileCount,
		TemplateCount: details.TemplateCount,
		Variables:     details.Variables,
		Mode:          normalized,
		ShownCount:    len(filtered),
		Files:         filtered,
	}, nil
}

func writeYAMLLangs(w io.Writer, langs []string) error {
	var b strings.Builder
	for _, lang := range langs {
		fmt.Fprintf(&b, "- %s\n", yamlQuote(lang))
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func writeYAMLSummaries(w io.Writer, summaries []scaffold.TemplateSummary) error {
	var b strings.Builder
	for _, summary := range summaries {
		fmt.Fprintf(&b, "- name: %s\n", yamlQuote(summary.Name))
		fmt.Fprintf(&b, "  file_count: %d\n", summary.FileCount)
		fmt.Fprintf(&b, "  template_count: %d\n", summary.TemplateCount)

		if len(summary.Variables) == 0 {
			fmt.Fprintln(&b, "  variables: []")
			continue
		}

		fmt.Fprintln(&b, "  variables:")
		for _, v := range summary.Variables {
			fmt.Fprintf(&b, "    - %s\n", yamlQuote(v))
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func writeYAMLInspectOutput(w io.Writer, output inspectOutput) error {
	var b strings.Builder
	fmt.Fprintf(&b, "name: %s\n", yamlQuote(output.Name))
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

func writeTextLangs(w io.Writer, langs []string) error {
	var b strings.Builder
	for _, lang := range langs {
		fmt.Fprintln(&b, lang)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func writeTextSummaries(w io.Writer, summaries []scaffold.TemplateSummary) error {
	var b strings.Builder
	for _, summary := range summaries {
		vars := "(none)"
		if len(summary.Variables) > 0 {
			vars = strings.Join(summary.Variables, ", ")
		}
		fmt.Fprintf(&b, "%s\tfiles=%d\ttemplates=%d\tvars=[%s]\n", summary.Name, summary.FileCount, summary.TemplateCount, vars)
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

	fmt.Fprintf(&b, "name: %s\n", output.Name)
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

func yamlQuote(s string) string {
	return strconv.Quote(s)
}
