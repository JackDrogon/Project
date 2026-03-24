package main

import (
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	"github.com/spf13/cobra"
)

type catalogCommandBase struct {
	asTOML  bool
	compact bool
}

func (c *catalogCommandBase) bindSharedFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&c.asTOML, "toml", false, "Output as TOML")
	cmd.Flags().BoolVar(&c.compact, "compact", false, "Use a more compact human-readable text layout")
}

func (c *catalogCommandBase) newService() *appcatalog.Service {
	return newCatalogService()
}
