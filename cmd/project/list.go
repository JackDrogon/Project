package main

import (
	"io/fs"

	"github.com/JackDrogon/project/internal/adapters/templatesrc"
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
)

var newCatalogService = func() *appcatalog.Service {
	source := templatesrc.New()
	return appcatalog.NewService(source.FS(), nil)
}

type failingCatalogFS struct{ err error }

func (f failingCatalogFS) Open(string) (fs.File, error)          { return nil, f.err }
func (f failingCatalogFS) ReadDir(string) ([]fs.DirEntry, error) { return nil, f.err }
