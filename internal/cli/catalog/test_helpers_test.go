package catalog

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	appconfig "github.com/JackDrogon/project/internal/app/config"
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

// withActiveConfigContext loads a TOML config document through the app config
// loader and returns a context carrying the resulting ActiveConfig.
func withActiveConfigContext(t *testing.T, content string) context.Context {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

	active, err := appconfig.NewService().LoadActiveConfig(appconfig.Context{ExplicitPath: path})
	if err != nil {
		t.Fatalf("LoadActiveConfig() error = %v", err)
	}

	return appconfig.WithActiveConfig(context.Background(), active)
}
