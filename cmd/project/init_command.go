package main

import (
	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

func newInitCmd(creator *appcreate.Creator) *cobra.Command {
	var shared scaffoldCommandFlags

	cmd := &cobra.Command{
		Use:   "init [target_dir]",
		Short: "Initialize project in current or target directory",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := newCreateService()
			opts, err := svc.BuildInitOptions(shared.initRequest(cmd, args))
			if err != nil {
				return err
			}

			return svc.ScaffoldAndMaybeWriteReplay(creator, shared.toAppFlags(), appcreate.CommandInit, opts)
		},
	}

	bindScaffoldCommandFlags(cmd, &shared)

	return cmd
}

func projectNameFromTargetDir(targetDir string) (string, error) {
	return appcreate.ProjectNameFromTargetDir(targetDir)
}
