package main

import (
	"fmt"

	"github.com/JackDrogon/project/pkg/scaffold"
	"github.com/spf13/cobra"
)

// newListCmd creates the "list" subcommand that shows available template languages.
func newListCmd(creator *scaffold.Creator) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list all supported languages",
		RunE: func(cmd *cobra.Command, args []string) error {
			langs, err := creator.ListLangs()
			if err != nil {
				return err
			}
			for _, lang := range langs {
				fmt.Println(lang)
			}
			return nil
		},
	}
}
