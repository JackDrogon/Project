package main

import (
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	"github.com/JackDrogon/project/internal/presenters"
	"github.com/spf13/cobra"
)

type catalogCommandBase struct {
	asTOML bool
}

func (c *catalogCommandBase) bindSharedFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&c.asTOML, "toml", false, "Output as TOML")
}

func (c *catalogCommandBase) newPresenter() (*presenters.Presenter, error) {
	return presenters.NewPresenter(selectedOutputFormat(c.asTOML))
}

func (c *catalogCommandBase) newService() *appcatalog.Service {
	return newCatalogService()
}
