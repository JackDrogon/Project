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

				for _, summary := range summaries {
					vars := "(none)"
					if len(summary.Variables) > 0 {
						vars = strings.Join(summary.Variables, ", ")
					}

					if _, err := fmt.Fprintf(out, "%s\tfiles=%d\ttemplates=%d\tvars=[%s]\n", summary.Name, summary.FileCount, summary.TemplateCount, vars); err != nil {
						return err
					}
				}
				return nil
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

			for _, lang := range langs {
				if _, err := fmt.Fprintln(out, lang); err != nil {
					return err
				}
			}
			return nil
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

			vars := "(none)"
			if len(output.Variables) > 0 {
				vars = strings.Join(output.Variables, ", ")
			}

			if _, err := fmt.Fprintf(out, "name: %s\n", output.Name); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "files: %d\n", output.FileCount); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "templates: %d\n", output.TemplateCount); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "variables: %s\n", vars); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "mode: %s\n", output.Mode); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "shown: %d\n", output.ShownCount); err != nil {
				return err
			}

			for _, file := range output.Files {
				mode := "copy"
				if file.IsTemplate {
					mode = "render"
				}
				if _, err := fmt.Fprintf(out, "- %s -> %s (%s)\n", file.Source, file.Output, mode); err != nil {
					return err
				}
			}

			return nil
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
	for _, lang := range langs {
		if _, err := fmt.Fprintf(w, "- %s\n", yamlQuote(lang)); err != nil {
			return err
		}
	}
	return nil
}

func writeYAMLSummaries(w io.Writer, summaries []scaffold.TemplateSummary) error {
	for _, summary := range summaries {
		if _, err := fmt.Fprintf(w, "- name: %s\n", yamlQuote(summary.Name)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  file_count: %d\n", summary.FileCount); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  template_count: %d\n", summary.TemplateCount); err != nil {
			return err
		}

		if len(summary.Variables) == 0 {
			if _, err := fmt.Fprintln(w, "  variables: []"); err != nil {
				return err
			}
			continue
		}

		if _, err := fmt.Fprintln(w, "  variables:"); err != nil {
			return err
		}
		for _, v := range summary.Variables {
			if _, err := fmt.Fprintf(w, "    - %s\n", yamlQuote(v)); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeYAMLInspectOutput(w io.Writer, output inspectOutput) error {
	if _, err := fmt.Fprintf(w, "name: %s\n", yamlQuote(output.Name)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "file_count: %d\n", output.FileCount); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "template_count: %d\n", output.TemplateCount); err != nil {
		return err
	}

	if len(output.Variables) == 0 {
		if _, err := fmt.Fprintln(w, "variables: []"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(w, "variables:"); err != nil {
			return err
		}
		for _, v := range output.Variables {
			if _, err := fmt.Fprintf(w, "  - %s\n", yamlQuote(v)); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintf(w, "mode: %s\n", yamlQuote(output.Mode)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "shown_count: %d\n", output.ShownCount); err != nil {
		return err
	}

	if len(output.Files) == 0 {
		_, err := fmt.Fprintln(w, "files: []")
		return err
	}

	if _, err := fmt.Fprintln(w, "files:"); err != nil {
		return err
	}
	for _, file := range output.Files {
		if _, err := fmt.Fprintf(w, "  - source: %s\n", yamlQuote(file.Source)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "    output: %s\n", yamlQuote(file.Output)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "    is_template: %t\n", file.IsTemplate); err != nil {
			return err
		}
	}

	return nil
}

func yamlQuote(s string) string {
	return strconv.Quote(s)
}
