package main

import (
	"testing"

	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	appversion "github.com/JackDrogon/project/internal/app/version"
	"github.com/JackDrogon/project/internal/testsupport/catalogfixture"
)

func useCatalogServiceFactory(t *testing.T, factory func() *appcatalog.Service) {
	t.Helper()
	oldFactory := newCatalogService
	newCatalogService = factory
	t.Cleanup(func() {
		newCatalogService = oldFactory
	})
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

func useVersionServiceFactoryWith(t *testing.T, factory func() *appversion.Service) {
	t.Helper()
	oldFactory := newVersionService
	newVersionService = factory
	t.Cleanup(func() {
		newVersionService = oldFactory
	})
}
