package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

// newListCmd creates the "list" subcommand that shows available template languages.
func newListCmd(creator *scaffold.Creator) *cobra.Command {
	var detail bool
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all supported languages",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if detail {
				summaries, err := creator.ListTemplateSummaries()
				if err != nil {
					return err
				}

				if asJSON {
					enc := json.NewEncoder(out)
					enc.SetIndent("", "  ")
					return enc.Encode(summaries)
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

			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(langs)
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

	return cmd
}

func newInspectCmd(creator *scaffold.Creator) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "inspect [lang]",
		Short: "Inspect one template language",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			details, err := creator.InspectLang(args[0])
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(details)
			}

			vars := "(none)"
			if len(details.Variables) > 0 {
				vars = strings.Join(details.Variables, ", ")
			}

			if _, err := fmt.Fprintf(out, "name: %s\n", details.Name); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "files: %d\n", details.FileCount); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "templates: %d\n", details.TemplateCount); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "variables: %s\n", vars); err != nil {
				return err
			}

			for _, file := range details.Files {
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
	return cmd
}
