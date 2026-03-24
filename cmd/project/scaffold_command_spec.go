package main

import (
	appcreate "github.com/JackDrogon/project/internal/app/create"
	"github.com/spf13/cobra"
)

type scaffoldCommandSpecBuilder interface {
	Build(*appcreate.Service, *cobra.Command, []string) (appcreate.ScaffoldSpec, error)
}

type newScaffoldCommandSpecBuilder struct {
	flags *scaffoldCommandFlags
	force *bool
}

type initScaffoldCommandSpecBuilder struct {
	flags *scaffoldCommandFlags
}

func (b newScaffoldCommandSpecBuilder) Build(service *appcreate.Service, cmd *cobra.Command, args []string) (appcreate.ScaffoldSpec, error) {
	return service.BuildNewSpec(b.flags.newRequest(cmd, valueOrFalse(b.force), args))
}

func (b initScaffoldCommandSpecBuilder) Build(service *appcreate.Service, cmd *cobra.Command, args []string) (appcreate.ScaffoldSpec, error) {
	return service.BuildInitSpec(b.flags.initRequest(cmd, args))
}

func valueOrFalse(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}
