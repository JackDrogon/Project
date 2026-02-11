package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Tag is set via -ldflags at build time (e.g., "v0.1.0").
var Tag = "dev"

// Info returns a formatted version string combining the tag (from ldflags)
// with VCS revision and dirty state (from runtime/debug.BuildInfo).
func Info() string {
	revision, modified := vcsInfo()

	var b strings.Builder
	b.WriteString(Tag)
	if revision != "" {
		b.WriteString(":")
		b.WriteString(revision)
	}
	if modified {
		b.WriteString("-dirty")
	}
	return b.String()
}

// vcsInfo extracts VCS revision and modified state from debug.BuildInfo.
func vcsInfo() (revision string, modified bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
			if len(revision) > 7 {
				revision = revision[:7]
			}
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	return
}

// Verbose returns multi-line version details.
func Verbose() string {
	revision, modified := vcsInfo()
	lines := []string{fmt.Sprintf("Tag:      %s", Tag)}
	if revision != "" {
		lines = append(lines, fmt.Sprintf("Revision: %s", revision))
	}
	lines = append(lines, fmt.Sprintf("Dirty:    %t", modified))
	return strings.Join(lines, "\n")
}
