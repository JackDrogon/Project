package catalog

import (
	"testing"

	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	"github.com/JackDrogon/project/internal/testsupport/catalogfixture"
)

func newTestDependencies(factory func() *appcatalog.Service) Dependencies {
	return Dependencies{NewService: factory}
}

func newCommandTestCatalogService() *appcatalog.Service {
	return catalogfixture.NewService()
}

func newFailingCatalogService(t *testing.T) *appcatalog.Service {
	t.Helper()
	return catalogfixture.NewFailingService()
}
