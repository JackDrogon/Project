package scaffold

import (
	"fmt"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	currentUser      = user.Current
	execGoCommand    = exec.Command
	runtimeVersion   = runtime.Version
	goVersionPattern = regexp.MustCompile(`go(\d+)\.(\d+)`)
)

// TemplateVars holds the variables available for template rendering.
type TemplateVars struct {
	ProjectName string
	ModulePath  string
	GoVersion   string
	Author      string
	Year        int
}

// NewTemplateVars creates a TemplateVars with sensible defaults.
func NewTemplateVars(projectName, modulePath string) TemplateVars {
	if modulePath == "" {
		modulePath = projectName
	}

	author := "author"
	if u, err := currentUser(); err == nil && u.Username != "" {
		author = u.Username
	}

	return TemplateVars{
		ProjectName: projectName,
		ModulePath:  modulePath,
		GoVersion:   detectGoVersion(),
		Author:      author,
		Year:        time.Now().Year(),
	}
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
