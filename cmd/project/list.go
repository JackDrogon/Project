package main

import (
	"io/fs"

	"github.com/JackDrogon/project/internal/adapters/templatesrc"
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/JackDrogon/project/internal/presenters"
	"github.com/spf13/cobra"
)

var newCatalogService = func() *appcatalog.Service {
	source := templatesrc.New()
	return appcatalog.NewService(source.FS(), nil)
}

// newListCmd creates the "list" subcommand that shows available template languages.
func newListCmd(creator *appcreate.Creator) *cobra.Command {
	var detail bool
	var asTOML bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all supported languages",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			service := newCatalogService()
			format := selectedOutputFormat(asTOML)

			cliPresenter, err := presenters.NewPresenter(format)
			if err != nil {
				return err
			}

			if detail {
				summaries, err := service.ListSummaries()
				if err != nil {
					return err
				}
				return cliPresenter.WriteSummaries(out, summaries)
			}

			langs, err := service.ListLangs()
			if err != nil {
				return err
			}
			return cliPresenter.WriteLangs(out, langs)
		},
	}

	cmd.Flags().BoolVar(&detail, "detail", false, "Show file/template counts and variables")
	cmd.Flags().BoolVar(&asTOML, "toml", false, "Output as TOML")

	return cmd
}

type failingCatalogFS struct{ err error }

func (f failingCatalogFS) Open(string) (fs.File, error)          { return nil, f.err }
func (f failingCatalogFS) ReadDir(string) ([]fs.DirEntry, error) { return nil, f.err }
