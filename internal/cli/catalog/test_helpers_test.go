package catalog

import (
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

// newTestDependenciesWithTOML builds dependencies carrying the config the root
// command would have resolved from the given TOML document.
func newTestDependenciesWithTOML(t *testing.T, factory func() *appcatalog.Service, content string) Dependencies {
	t.Helper()
	return Dependencies{NewService: factory, Config: resolvedFromTOML(t, content)}
}

func newCommandTestCatalogService() *appcatalog.Service {
	return catalogfixture.NewService()
}

func newFailingCatalogService(t *testing.T) *appcatalog.Service {
	t.Helper()
	return catalogfixture.NewFailingService()
}

// resolvedFromTOML loads a TOML config document through the app config loader
// and returns it in the shape the root command hands to a subcommand.
func resolvedFromTOML(t *testing.T, content string) *appconfig.Resolved {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

	options := appconfig.LoadOptions{ExplicitPath: path}
	active, err := appconfig.NewService().LoadActiveConfig(options)
	if err != nil {
		t.Fatalf("LoadActiveConfig() error = %v", err)
	}

	return &appconfig.Resolved{Active: active, Options: options}
}
