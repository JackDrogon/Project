package buildinfo

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const (
	shortRevisionLength = 7
)

// Tag is the release tag, injected at link time with
// -X .../buildinfo.Tag=$(git describe). It has to be a package-level variable
// for -X to reach it, but it is written once by the linker and only read here:
// New snapshots it, so nothing mutates it at run time.
var Tag = "dev"

// readBuildInfoFunc is debug.ReadBuildInfo, held as a field rather than a
// package variable so tests stub it per-adapter instead of swapping a global.
type readBuildInfoFunc func() (*debug.BuildInfo, bool)

type Adapter struct {
	tag           string
	readBuildInfo readBuildInfoFunc
}

func New() *Adapter {
	return &Adapter{tag: Tag, readBuildInfo: debug.ReadBuildInfo}
}

// newWithBuildInfo builds an adapter around a fixed tag and a substitute
// build-info reader.
func newWithBuildInfo(tag string, read readBuildInfoFunc) *Adapter {
	if read == nil {
		read = debug.ReadBuildInfo
	}

	return &Adapter{tag: tag, readBuildInfo: read}
}

func (a *Adapter) Info() string {
	revision, modified := a.vcsInfo()

	var b strings.Builder
	b.WriteString(a.tag)
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
	revision, modified := a.vcsInfo()
	lines := []string{fmt.Sprintf("Tag:      %s", a.tag)}
	if revision != "" {
		lines = append(lines, fmt.Sprintf("Revision: %s", revision))
	}
	lines = append(lines, fmt.Sprintf("Dirty:    %t", modified))
	return strings.Join(lines, "\n")
}

func (a *Adapter) vcsInfo() (revision string, modified bool) {
	info, ok := a.readBuildInfo()
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
