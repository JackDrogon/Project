package scaffold

import (
	"io"

	appconfig "github.com/JackDrogon/project/internal/app/config"
	appcreate "github.com/JackDrogon/project/internal/app/create"
)

type Dependencies struct {
	// NewCreator builds the creator bound to the writer the command owns, so
	// scaffold progress output follows the command's out stream.
	NewCreator func(out io.Writer) *appcreate.Creator
	NewService func() *appcreate.Service
	// Config is filled in by the root command before any subcommand runs.
	// A nil Config means "no config file", which is what tests that do not
	// exercise config resolution want.
	Config *appconfig.Resolved
}

func (d Dependencies) activeConfig() appconfig.ActiveConfig {
	if d.Config == nil {
		return appconfig.ActiveConfig{}
	}

	return d.Config.Active
}

func (d Dependencies) newCreator(out io.Writer) *appcreate.Creator {
	if d.NewCreator == nil {
		panic("scaffold dependencies require NewCreator")
	}

	return d.NewCreator(out)
}

func (d Dependencies) newService() *appcreate.Service {
	if d.NewService == nil {
		panic("scaffold dependencies require NewService")
	}

	return d.NewService()
}
