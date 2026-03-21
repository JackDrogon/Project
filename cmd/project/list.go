package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

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
		fmt.Fprintf(&b, "  description: %s\n", yamlQuote(summary.Description))
		fmt.Fprintf(&b, "  manifest_version: %d\n", summary.ManifestVersion)

		if len(summary.InputNames) == 0 {
			fmt.Fprintln(&b, "  input_names: []")
		} else {
			fmt.Fprintln(&b, "  input_names:")
			for _, name := range summary.InputNames {
				fmt.Fprintf(&b, "    - %s\n", yamlQuote(name))
			}
		}

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
