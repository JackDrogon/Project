package buildinfo

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const (
	shortRevisionLength = 7
)

var readBuildInfo = debug.ReadBuildInfo

var Tag = "dev"

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Info() string {
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

func (a *Adapter) Verbose() string {
	revision, modified := vcsInfo()
	lines := []string{fmt.Sprintf("Tag:      %s", Tag)}
	if revision != "" {
		lines = append(lines, fmt.Sprintf("Revision: %s", revision))
	}
	lines = append(lines, fmt.Sprintf("Dirty:    %t", modified))
	return strings.Join(lines, "\n")
}

func vcsInfo() (revision string, modified bool) {
	info, ok := readBuildInfo()
	if !ok {
		return "", false
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
			if len(revision) > shortRevisionLength {
				revision = revision[:shortRevisionLength]
			}
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	return revision, modified
}
