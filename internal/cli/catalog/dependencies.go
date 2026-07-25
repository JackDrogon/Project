package catalog

import (
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	appconfig "github.com/JackDrogon/project/internal/app/config"
)

type Dependencies struct {
	NewService func() *appcatalog.Service
	// Config is filled in by the root command before any subcommand runs.
	// A nil Config means "no config file", which is what tests that do not
	// exercise config resolution want.
	Config *appconfig.Resolved
}

func (d Dependencies) newService() *appcatalog.Service {
	if d.NewService == nil {
		panic("catalog dependencies require NewService")
	}

	return d.NewService()
}

func (d Dependencies) activeConfig() appconfig.ActiveConfig {
	if d.Config == nil {
		return appconfig.ActiveConfig{}
	}

	return d.Config.Active
}
