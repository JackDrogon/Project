package main

import (
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	appconfig "github.com/JackDrogon/project/internal/app/config"
	appversion "github.com/JackDrogon/project/internal/app/version"
	"github.com/JackDrogon/project/internal/testsupport/catalogfixture"
)

// catalogTestDependencies pins the catalog service to the shared fixture so
// list/inspect output does not depend on the embedded templates.
func catalogTestDependencies() dependencies {
	deps := newDependencies()
	deps.newCatalogService = newCommandTestCatalogService
	return deps
}

func versionTestDependencies(factory func() *appversion.Service) dependencies {
	deps := newDependencies()
	deps.newVersionService = factory
	return deps
}

func newCommandTestCatalogService() *appcatalog.Service {
	return catalogfixture.NewService()
}

type stubVersionProvider struct {
	info    string
	verbose string
}

func (s stubVersionProvider) Info() string {
	return s.info
}

func (s stubVersionProvider) Verbose() string {
	return s.verbose
}

func stubVersionServiceFactory() func() *appversion.Service {
	return func() *appversion.Service {
		return appversion.NewService(stubVersionProvider{
			info:    "short-version",
			verbose: "Tag:      short-version\nDirty:    false",
		})
	}
}

// userConfigServiceFactory points config discovery at a temp home so tests
// never read the developer's real user config directory.
func userConfigServiceFactory(configHome string) func() *appconfig.Service {
	return func() *appconfig.Service {
		return appconfig.NewServiceWithDeps(appconfig.Dependencies{
			UserConfigDir: func() (string, error) { return configHome, nil },
		})
	}
}
