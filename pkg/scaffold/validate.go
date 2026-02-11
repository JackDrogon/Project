package scaffold

import (
	"fmt"
	"regexp"
)

// validProjectName matches names starting with a letter, containing only
// alphanumerics, dots, hyphens, and underscores.
var validProjectName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)

const maxProjectNameLen = 255

// ValidateProjectName checks that name is a safe, valid project/directory name.
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name must not be empty")
	}
	if len(name) > maxProjectNameLen {
		return fmt.Errorf("project name must be at most %d characters, got %d", maxProjectNameLen, len(name))
	}
	if !validProjectName.MatchString(name) {
		return fmt.Errorf("project name %q is invalid: must start with a letter and contain only [a-zA-Z0-9._-]", name)
	}
	return nil
}
