package create

import (
	"fmt"
	"path/filepath"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

func resolveNewProjectArgs(lang, module, arg string) (projectName string, targetDir string, modulePath string, err error) {
	if lang == langGo && module == "" {
		if projectErr := domain.ValidateProjectName(arg); projectErr != nil {
			if moduleErr := domain.ValidateModulePath(arg); moduleErr == nil {
				name := domain.ProjectNameFromGoModulePath(arg)
				if err := domain.ValidateProjectName(name); err != nil {
					return "", "", "", err
				}
				return name, name, arg, nil
			}
		}
	}

	return arg, arg, module, nil
}

// ResolveNewProjectArgs resolves the project name, target directory, and module path
// from the provided language, module flag, and positional argument.
// For Go projects, it handles both simple project names and full module paths.
// Returns an error if the arguments are invalid or cannot be resolved.
func ResolveNewProjectArgs(lang, module, arg string) (projectName string, targetDir string, modulePath string, err error) {
	return resolveNewProjectArgs(lang, module, arg)
}

// ProjectNameFromGoModulePath extracts the project name from a Go module path.
// For example, "github.com/user/myapp" returns "myapp", and "github.com/user/myapp/v2" also returns "myapp".
func ProjectNameFromGoModulePath(modulePath string) string {
	return domain.ProjectNameFromGoModulePath(modulePath)
}

// ProjectNameFromTargetDir derives the project name from the target directory path.
// It returns the base name of the absolute path. Returns an error if the path cannot be resolved.
func ProjectNameFromTargetDir(targetDir string) (string, error) {
	return projectNameFromTargetDir(targetDir)
}

func projectNameFromTargetDir(targetDir string) (string, error) {
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target directory %q: %w", targetDir, err)
	}
	return filepath.Base(absTarget), nil
}
