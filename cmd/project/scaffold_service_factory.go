package main

import appcreate "github.com/JackDrogon/project/internal/app/create"

var newCreateService = func() *appcreate.Service {
	return appcreate.NewService()
}
