package main

import (
	"bytes"
	"strings"
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

func useVersionServiceFactory(t *testing.T, provider stubVersionProvider) {
	t.Helper()
	oldFactory := newVersionService
	newVersionService = func() *appversion.Service {
		return appversion.NewService(provider)
	}
	t.Cleanup(func() {
		newVersionService = oldFactory
	})
}

func TestVersionCmd_DefaultAndVerbose(t *testing.T) {
	useVersionServiceFactory(t, stubVersionProvider{
		info:    "test-tag",
		verbose: "Tag:      test-tag\nDirty:    false",
	})

	t.Run("default", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := buildVersionCommand(commandDependencies{})
		cmd.SetOut(&buf)
		cmd.SetArgs(nil)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !strings.Contains(buf.String(), "test-tag") {
			t.Fatalf("output = %q, want contains test tag", buf.String())
		}
	})

	t.Run("verbose", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := buildVersionCommand(commandDependencies{})
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--verbose"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		got := buf.String()
		if !strings.Contains(got, "Tag:") || !strings.Contains(got, "Dirty:") {
			t.Fatalf("output = %q, want verbose fields", got)
		}
	})
}
