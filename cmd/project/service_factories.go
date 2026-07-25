package main

import (
	"github.com/JackDrogon/project/internal/adapters/buildinfo"
	"github.com/JackDrogon/project/internal/adapters/templatesrc"
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	appcompletion "github.com/JackDrogon/project/internal/app/completion"
	appconfig "github.com/JackDrogon/project/internal/app/config"
	appcreate "github.com/JackDrogon/project/internal/app/create"
	appversion "github.com/JackDrogon/project/internal/app/version"
)

func newCatalogService() *appcatalog.Service {
	source := templatesrc.New()
	return appcatalog.NewService(source.FS(), nil)
}

func newCreateService() *appcreate.Service {
	return appcreate.NewService()
}

func newConfigService() *appconfig.Service {
	return appconfig.NewService()
}

func newVersionService() *appversion.Service {
	return appversion.NewService(buildinfo.New())
}

func newCompletionService() *appcompletion.Service {
	return appcompletion.NewService()
}
