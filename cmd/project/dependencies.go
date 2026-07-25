package main

import (
	"io"

	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	appcompletion "github.com/JackDrogon/project/internal/app/completion"
	appconfig "github.com/JackDrogon/project/internal/app/config"
	appcreate "github.com/JackDrogon/project/internal/app/create"
	appversion "github.com/JackDrogon/project/internal/app/version"
)

// dependencies carries every collaborator the command tree needs. Assembly
// happens once in main and is passed down explicitly, so no part of this
// package reads package-level state and tests never swap globals.
type dependencies struct {
	newCreator           func(out io.Writer) *appcreate.Creator
	newCatalogService    func() *appcatalog.Service
	newCreateService     func() *appcreate.Service
	newConfigService     func() *appconfig.Service
	newVersionService    func() *appversion.Service
	newCompletionService func() *appcompletion.Service
}

// newDependencies wires the production collaborators.
func newDependencies() dependencies {
	return dependencies{
		newCreator:           newCommandCreator,
		newCatalogService:    newCatalogService,
		newCreateService:     newCreateService,
		newConfigService:     newConfigService,
		newVersionService:    newVersionService,
		newCompletionService: newCompletionService,
	}
}
