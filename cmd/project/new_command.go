package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

// newNewCmd creates the "new" subcommand that scaffolds a project from templates.
func newNewCmd(templateFS embed.FS) *cobra.Command {
	var lang string

	cmd := &cobra.Command{
		Use:   "new [project_name]",
		Short: "Create new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createProject(templateFS, lang, args[0])
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "", "Programming language for the project")
	_ = cmd.MarkFlagRequired("lang")

	return cmd
}

// createProject creates a new project with the given language template.
func createProject(templateFS embed.FS, lang, projectName string) error {
	fmt.Printf("Creating project with language: %s, project name: %s\n", lang, projectName)

	// Check if the language is supported
	langTemplateDir := lang
	if _, err := fs.ReadDir(templateFS, langTemplateDir); err != nil {
		return fmt.Errorf("unsupported language: %s", lang)
	}

	// Copy the template files into the project directory
	if err := copyEmbedDir(templateFS, langTemplateDir, projectName); err != nil {
		return fmt.Errorf("failed to copy template files: %w", err)
	}

	// Initialize git repository
	gitCmds := [][]string{
		{"init"},
		{"add", "."},
		{"commit", "-s", "-m", "Initial commit"},
	}
	for _, gitArgs := range gitCmds {
		if err := runGit(projectName, gitArgs...); err != nil {
			return fmt.Errorf("failed to run git %s: %w", gitArgs[0], err)
		}
	}

	fmt.Println("Project created successfully")
	return nil
}

// runGit executes a git command in the given directory.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

// copyEmbedDir recursively copies a directory from an embedded filesystem
// to the local filesystem.
func copyEmbedDir(fsys fs.FS, srcDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	entries, err := fs.ReadDir(fsys, srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// embed.FS always uses forward slashes
		srcPath := path.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())
		fmt.Println(entry.Name())

		if entry.IsDir() {
			if err := copyEmbedDir(fsys, srcPath, destPath); err != nil {
				return err
			}
			continue
		}

		content, err := fs.ReadFile(fsys, srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return err
		}
	}
	return nil
}
