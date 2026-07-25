package create

import (
	"context"
	"fmt"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strings"
	"time"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

const (
	defaultAuthorName = "author"
)

var (
	currentUser      = user.Current
	execGoCommand    = exec.CommandContext
	runtimeVersion   = runtime.Version
	goVersionPattern = regexp.MustCompile(`go(\d+)\.(\d+)`)
)

func NewTemplateVars(ctx context.Context, projectName, modulePath string) domain.TemplateVars {
	return newDefaultTemplateVars(ctx, projectName, modulePath)
}

func newDefaultTemplateVars(ctx context.Context, projectName, modulePath string) domain.TemplateVars {
	author := defaultAuthorName
	if u, err := currentUser(); err == nil && u.Username != "" {
		author = u.Username
	}

	return domain.NewTemplateVars(projectName, modulePath, detectGoVersion(ctx), author, time.Now().Year())
}

// detectGoVersion prefers the toolchain on PATH and falls back to the version
// this binary was built with, so a cancelled `go env` degrades instead of
// failing the scaffold.
func detectGoVersion(ctx context.Context) string {
	if version, err := localGoVersion(ctx); err == nil {
		return version
	}

	version, err := parseGoLanguageVersion(runtimeVersion())
	if err == nil {
		return version
	}

	return strings.TrimPrefix(strings.TrimSpace(runtimeVersion()), "go")
}

func localGoVersion(ctx context.Context) (string, error) {
	output, err := execGoCommand(ctx, "go", "env", "GOVERSION").Output()
	if err != nil {
		return "", err
	}

	return parseGoLanguageVersion(string(output))
}

func parseGoLanguageVersion(raw string) (string, error) {
	matches := goVersionPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if len(matches) != 3 {
		return "", fmt.Errorf("unsupported Go version format %q", strings.TrimSpace(raw))
	}

	return matches[1] + "." + matches[2], nil
}
