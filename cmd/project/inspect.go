package main

import (
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	appcreate "github.com/JackDrogon/project/internal/app/create"
	presenter "github.com/JackDrogon/project/internal/presenters/cli"

	"github.com/spf13/cobra"
)

func newInspectCmd(creator *appcreate.Creator) *cobra.Command {
	var asTOML bool
	var mode string

	cmd := &cobra.Command{
		Use:   "inspect [lang]",
		Short: "Inspect one template language",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := newCatalogService()
			inspection, err := service.Inspect(args[0], mode)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			format := selectedOutputFormat(asTOML)

			switch format {
			case outputFormatTOML:
				return presenter.WriteTOMLInspection(out, inspection)
			}

			return presenter.WriteTextInspection(out, inspection)
		},
	}

	cmd.Flags().BoolVar(&asTOML, "toml", false, "Output as TOML")
	cmd.Flags().StringVar(&mode, "mode", appcatalog.InspectModeAll, "File mode: all, render, copy")
	return cmd
}
