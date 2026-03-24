package main

import (
	"github.com/JackDrogon/project/internal/adapters/buildinfo"
	"github.com/JackDrogon/project/internal/adapters/templatesrc"
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	appcreate "github.com/JackDrogon/project/internal/app/create"
	appversion "github.com/JackDrogon/project/internal/app/version"
)

var newCatalogService = func() *appcatalog.Service {
	source := templatesrc.New()
	return appcatalog.NewService(source.FS(), nil)
}

var newCreateService = func() *appcreate.Service {
	return appcreate.NewService()
}

var newVersionService = func() *appversion.Service {
	return appversion.NewService(buildinfo.New())
}
