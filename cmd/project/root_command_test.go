package main

import (
	"bytes"
	"testing"
	"testing/fstest"

	appcreate "github.com/JackDrogon/project/internal/app/create"
)

func TestNewRootCmd(t *testing.T) {
	fsys := fstest.MapFS{
		"go/Makefile": {Data: []byte("build:")},
	}

	creator := appcreate.NewCreator(fsys, &bytes.Buffer{})

	// Basic smoke test: verify the command tree can be built without panicking
	cmd := newRootCmd(creator)
	if cmd.Use != "project" {
		t.Errorf("root command Use = %q, want %q", cmd.Use, "project")
	}

	// Verify expected subcommands are registered
	subCmds := cmd.Commands()
	wantCmds := map[string]bool{"new": false, "init": false, "list": false, "inspect": false, "version": false, "completion": false}
	for _, sub := range subCmds {
		if _, ok := wantCmds[sub.Name()]; ok {
			wantCmds[sub.Name()] = true
		}
	}
	for name, found := range wantCmds {
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}
