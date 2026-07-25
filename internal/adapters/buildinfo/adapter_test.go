package buildinfo

import (
	"runtime/debug"
	"strings"
	"testing"
)

// stubbedAdapter builds an adapter over a fixed tag and build info. Nothing
// global is swapped, so every test in this file can run in parallel.
func stubbedAdapter(tag string, info *debug.BuildInfo, ok bool) *Adapter {
	return newWithBuildInfo(tag, func() (*debug.BuildInfo, bool) {
		return info, ok
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	adapter := New()
	if adapter == nil {
		t.Fatal("New() = nil")
	}
	if adapter.tag != Tag {
		t.Fatalf("New().tag = %q, want %q", adapter.tag, Tag)
	}
}

func TestInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{
			name: "without build info",
			ok:   false,
			want: "v1.2.3",
		},
		{
			name: "with revision and dirty flag",
			info: &debug.BuildInfo{Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "1234567890abcdef"},
				{Key: "vcs.modified", Value: "true"},
			}},
			ok:   true,
			want: "v1.2.3:1234567-dirty",
		},
		{
			name: "clean revision omits dirty suffix",
			info: &debug.BuildInfo{Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abcdef0123"},
				{Key: "vcs.modified", Value: "false"},
			}},
			ok:   true,
			want: "v1.2.3:abcdef0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := stubbedAdapter("v1.2.3", tt.info, tt.ok).Info(); got != tt.want {
				t.Fatalf("Info() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVerbose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		info        *debug.BuildInfo
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:        "without revision",
			info:        &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.modified", Value: "false"}}},
			wantContain: []string{"Tag:      v9.9.9", "Dirty:    false"},
			wantAbsent:  []string{"Revision:"},
		},
		{
			name: "with revision",
			info: &debug.BuildInfo{Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abcdef0"},
				{Key: "vcs.modified", Value: "true"},
			}},
			wantContain: []string{"Tag:      v9.9.9", "Revision: abcdef0", "Dirty:    true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := stubbedAdapter("v9.9.9", tt.info, true).Verbose()
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("Verbose() = %q, want contains %q", got, want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("Verbose() = %q, want no %q", got, absent)
				}
			}
		})
	}
}
