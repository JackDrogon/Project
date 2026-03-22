package scaffold

import (
	"fmt"
	"path"
	"regexp"
)

var goMajorVersionSuffix = regexp.MustCompile(`^v[2-9][0-9]*$`)

func ResolveGitMode(req CreateRequest) (GitMode, error) {
	if req.NoGit {
		if req.GitMode != "" && req.GitMode != GitModeNone {
			return "", fmt.Errorf("conflicting git options: --no-git cannot be combined with --git=%s", req.GitMode)
		}
		return GitModeNone, nil
	}

	if req.GitMode == "" {
		return GitModeInitCommit, nil
	}

	switch req.GitMode {
	case GitModeNone, GitModeInitOnly, GitModeInitCommit:
		return req.GitMode, nil
	default:
		return "", fmt.Errorf("invalid git mode %q: must be one of %s, %s, %s", req.GitMode, GitModeNone, GitModeInitOnly, GitModeInitCommit)
	}
}

func ProjectNameFromGoModulePath(modulePath string) string {
	name := path.Base(modulePath)
	if !goMajorVersionSuffix.MatchString(name) {
		return name
	}

	parent := path.Base(path.Dir(modulePath))
	if parent == "." || parent == "/" || parent == "" {
		return name
	}

	return parent
}

func DefaultModulePath(req CreateRequest) string {
	if req.ModulePath != "" {
		return req.ModulePath
	}

	return req.ProjectName
}
