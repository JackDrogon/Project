package main

import (
	"bytes"
	"strings"
	"testing"

	projectversion "github.com/JackDrogon/project/pkg/version"
)

func TestVersionCmd_DefaultAndVerbose(t *testing.T) {
	oldTag := projectversion.Tag
	projectversion.Tag = "test-tag"
	t.Cleanup(func() {
		projectversion.Tag = oldTag
	})

	t.Run("default", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := newVersionCmd()
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
		cmd := newVersionCmd()
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
