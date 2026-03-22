package create

import (
	"fmt"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strings"
	"time"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

var (
	currentUser      = user.Current
	execGoCommand    = exec.Command
	runtimeVersion   = runtime.Version
	goVersionPattern = regexp.MustCompile(`go(\d+)\.(\d+)`)
)

func NewTemplateVars(projectName, modulePath string) domain.TemplateVars {
	return newDefaultTemplateVars(projectName, modulePath)
}

func newDefaultTemplateVars(projectName, modulePath string) domain.TemplateVars {
	author := "author"
	if u, err := currentUser(); err == nil && u.Username != "" {
		author = u.Username
	}

	return domain.NewTemplateVars(projectName, modulePath, detectGoVersion(), author, time.Now().Year())
}

func detectGoVersion() string {
	if version, err := localGoVersion(); err == nil {
		return version
	}

	version, err := parseGoLanguageVersion(runtimeVersion())
	if err == nil {
		return version
	}

	return strings.TrimPrefix(strings.TrimSpace(runtimeVersion()), "go")
}

func localGoVersion() (string, error) {
	output, err := execGoCommand("go", "env", "GOVERSION").Output()
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
