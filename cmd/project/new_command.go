package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	lang string

	newCmd = &cobra.Command{
		Use:   "new [project_name]",
		Short: "Create new project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			createProject(lang, args[0])
		},
	}
)

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&lang, "lang", "l", "", "Programming language for the project")
	newCmd.MarkFlagRequired("lang")
}

// createProject creates a new project with the given language
func createProject(lang string, projectName string) {
	fmt.Printf("Creating project with language: %s, project name: %s\n", lang, projectName)

	// Step 1: Check if the language is supported
	langTemplateDir := fmt.Sprintf("templates/%s", lang)
	if _, err := templates.ReadDir(langTemplateDir); err != nil {
		fmt.Printf("Unsupported language: %s\n", lang)
		os.Exit(1)
	}

	// Step 2: Create the project by copying the template files
	if err := copyEmbedDir(templates, langTemplateDir, projectName); err != nil {
		fmt.Printf("Failed to copy template files: %v\n", err)
		os.Exit(1)
	}

	// Step 3: git init
	cmd := exec.Command("git", "init")
	cmd.Dir = projectName
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to git init: %v\n", err)
		os.Exit(1)
	}

	// Step 4: git add .
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = projectName
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to git add: %v\n", err)
		os.Exit(1)
	}

	// Step 5: git commit
	cmd = exec.Command("git", "commit", "-s", "-m", "Initial commit")
	cmd.Dir = projectName
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to git commit: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Project created successfully")
}

// add different languages hooks

// copyEmbedDir recursively copies a directory from embed.FS to the local filesystem
func copyEmbedDir(fsys embed.FS, srcDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	entries, err := fsys.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())
		fmt.Println(entry.Name())

		if entry.IsDir() {
			if err := copyEmbedDir(fsys, srcPath, destPath); err != nil {
				return err
			}
			continue
		}

		content, err := fsys.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return err
		}
	}
	return nil
}
