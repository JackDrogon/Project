package main

import (
	"testing"

	appversion "github.com/JackDrogon/project/internal/app/version"
)

type stubVersionProvider struct {
	info    string
	verbose string
}

func (s stubVersionProvider) Info() string {
	return s.info
}

func (s stubVersionProvider) Verbose() string {
	return s.verbose
}

func useVersionServiceFactoryWith(t *testing.T, factory func() *appversion.Service) {
	t.Helper()
	oldFactory := newVersionService
	newVersionService = factory
	t.Cleanup(func() {
		newVersionService = oldFactory
	})
}
