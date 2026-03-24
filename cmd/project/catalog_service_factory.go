package main

import (
	"github.com/JackDrogon/project/internal/adapters/templatesrc"
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
)

var newCatalogService = func() *appcatalog.Service {
	source := templatesrc.New()
	return appcatalog.NewService(source.FS(), nil)
}
