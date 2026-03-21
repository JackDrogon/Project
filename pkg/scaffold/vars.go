package scaffold

import (
	"fmt"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
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
	ProjectName      string
	ProjectNameLower string
	ModulePath       string
	GoVersion        string
	Author           string
	Year             int
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
		ProjectName:      projectName,
		ProjectNameLower: strings.ToLower(projectName),
		ModulePath:       modulePath,
		GoVersion:        detectGoVersion(),
		Author:           author,
		Year:             time.Now().Year(),
	}
}

func resolveTemplateVars(manifest TemplateManifest, opts Options) (TemplateVars, error) {
	vars := NewTemplateVars(opts.ProjectName, opts.ModulePath)
	if len(opts.TemplateInputValues) == 0 {
		return vars, nil
	}

	declaredInputs := make(map[string]TemplateManifestInput, len(manifest.Inputs))
	for _, input := range manifest.Inputs {
		declaredInputs[input.Name] = input
	}

	for name, value := range opts.TemplateInputValues {
		if name == "module_path" {
			return TemplateVars{}, fmt.Errorf("template input %q must be provided via module path options", name)
		}

		input, ok := declaredInputs[name]
		if !ok {
			return TemplateVars{}, fmt.Errorf("template input %q is not declared by template %q", name, manifest.Name)
		}

		if err := applyTemplateInputValue(&vars, input, value); err != nil {
			return TemplateVars{}, err
		}
	}

	return vars, nil
}

func applyTemplateInputValue(vars *TemplateVars, input TemplateManifestInput, value string) error {
	switch input.TemplateVar {
	case "ModulePath":
		return fmt.Errorf("template input %q must be provided via module path options", input.Name)
	case "GoVersion":
		vars.GoVersion = value
	case "Author":
		vars.Author = value
	case "Year":
		year, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("template input %q must be a valid year: %w", input.Name, err)
		}
		vars.Year = year
	default:
		return fmt.Errorf("template input %q has unsupported template_var %q", input.Name, input.TemplateVar)
	}

	return nil
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
