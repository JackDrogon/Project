package catalog

import (
	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	"github.com/spf13/cobra"
)

type commandBase struct {
	deps    Dependencies
	asTOML  bool
	compact bool
}

func (c *commandBase) bindSharedFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&c.asTOML, "toml", false, "Output as TOML")
	cmd.Flags().BoolVar(&c.compact, "compact", false, "Use a more compact human-readable text layout")
}

func (c *commandBase) newService() *appcatalog.Service {
	return c.deps.newService()
}
