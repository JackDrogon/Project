package main

import (
	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

type scaffoldCommandBase struct {
	creator *appcreate.Creator
	flags   scaffoldCommandFlags
}

func newScaffoldCommandBase(creator *appcreate.Creator) scaffoldCommandBase {
	return scaffoldCommandBase{creator: creator}
}

func (c *scaffoldCommandBase) bindSharedFlags(cmd *cobra.Command) {
	bindScaffoldCommandFlags(cmd, &c.flags)
}

func (c *scaffoldCommandBase) execute(
	cmd *cobra.Command,
	args []string,
	builder scaffoldCommandSpecBuilder,
) error {
	service := newCreateService()
	spec, err := builder.Build(service, cmd, args)
	if err != nil {
		return err
	}

	return service.ExecuteScaffoldSpec(c.creator, spec)
}
