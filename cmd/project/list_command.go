package main

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/spf13/cobra"
)

// newListCmd creates the "list" subcommand that shows available template languages.
func newListCmd(templateFS embed.FS) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list all supported languages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listLangs(templateFS)
		},
	}
}

func listLangs(templateFS fs.FS) error {
	langs, err := fs.ReadDir(templateFS, ".")
	if err != nil {
		return fmt.Errorf("failed to read templates: %w", err)
	}

	for _, lang := range langs {
		fmt.Println(lang.Name())
	}
	return nil
}
