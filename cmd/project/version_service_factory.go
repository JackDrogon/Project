package main

import (
	"github.com/JackDrogon/project/internal/adapters/buildinfo"
	appversion "github.com/JackDrogon/project/internal/app/version"
)

var newVersionService = func() *appversion.Service {
	return appversion.NewService(buildinfo.New())
}
