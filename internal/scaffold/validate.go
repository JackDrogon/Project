package scaffold

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	validProjectName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)
	validModulePath  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~/-]*$`)
)

const maxProjectNameLen = 255

func ValidateProjectName(name string) error {
	if name == "" {
		return errors.New("project name must not be empty")
	}
	if len(name) > maxProjectNameLen {
		return fmt.Errorf("project name must be at most %d characters, got %d", maxProjectNameLen, len(name))
	}
	if !validProjectName.MatchString(name) {
		return fmt.Errorf("project name %q is invalid: must start with a letter and contain only [a-zA-Z0-9._-]", name)
	}
	return nil
}

func ValidateModulePath(modulePath string) error {
	if modulePath == "" {
		return nil
	}
	if strings.HasPrefix(modulePath, "/") || strings.HasSuffix(modulePath, "/") {
		return fmt.Errorf("module path %q is invalid: must not start or end with '/'", modulePath)
	}
	if strings.Contains(modulePath, "//") {
		return fmt.Errorf("module path %q is invalid: must not contain empty path segments", modulePath)
	}
	if !validModulePath.MatchString(modulePath) {
		return fmt.Errorf("module path %q is invalid: must contain only [A-Za-z0-9._~/-]", modulePath)
	}

	for segment := range strings.SplitSeq(modulePath, "/") {
		if segment == "." || segment == ".." {
			return fmt.Errorf("module path %q is invalid: must not contain '.' or '..' path segments", modulePath)
		}
		if strings.HasPrefix(segment, ".") || strings.HasSuffix(segment, ".") {
			return fmt.Errorf("module path %q is invalid: path segments must not start or end with '.'", modulePath)
		}
	}

	return nil
}
