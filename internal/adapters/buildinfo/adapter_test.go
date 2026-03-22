package buildinfo

import (
	"runtime/debug"
	"strings"
	"testing"
)

func stubBuildInfo(t *testing.T, info *debug.BuildInfo, ok bool) {
	t.Helper()
	old := readBuildInfo
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return info, ok
	}
	t.Cleanup(func() {
		readBuildInfo = old
	})
}

func TestNew(t *testing.T) {
	adapter := New()
	if adapter == nil {
		t.Fatal("New() = nil")
	}
}

func TestInfo(t *testing.T) {
	oldTag := Tag
	Tag = "v1.2.3"
	t.Cleanup(func() {
		Tag = oldTag
	})

	adapter := New()

	t.Run("without build info", func(t *testing.T) {
		stubBuildInfo(t, nil, false)
		if got := adapter.Info(); got != "v1.2.3" {
			t.Fatalf("Info() = %q, want %q", got, "v1.2.3")
		}
	})

	t.Run("with revision and dirty flag", func(t *testing.T) {
		stubBuildInfo(t, &debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890abcdef"},
			{Key: "vcs.modified", Value: "true"},
		}}, true)

		if got := adapter.Info(); got != "v1.2.3:1234567-dirty" {
			t.Fatalf("Info() = %q, want %q", got, "v1.2.3:1234567-dirty")
		}
	})
}

func TestVerbose(t *testing.T) {
	oldTag := Tag
	Tag = "v9.9.9"
	t.Cleanup(func() {
		Tag = oldTag
	})

	adapter := New()

	t.Run("without revision", func(t *testing.T) {
		stubBuildInfo(t, &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.modified", Value: "false"}}}, true)
		got := adapter.Verbose()
		if strings.Contains(got, "Revision:") {
			t.Fatalf("Verbose() = %q, want no revision line", got)
		}
		if !strings.Contains(got, "Tag:      v9.9.9") || !strings.Contains(got, "Dirty:    false") {
			t.Fatalf("Verbose() = %q, want tag and dirty lines", got)
		}
	})

	t.Run("with revision", func(t *testing.T) {
		stubBuildInfo(t, &debug.BuildInfo{Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef0"},
			{Key: "vcs.modified", Value: "true"},
		}}, true)
		got := adapter.Verbose()
		if !strings.Contains(got, "Revision: abcdef0") || !strings.Contains(got, "Dirty:    true") {
			t.Fatalf("Verbose() = %q, want revision and dirty lines", got)
		}
	})
}
