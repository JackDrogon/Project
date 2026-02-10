package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/JackDrogon/project/pkg/utils"

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
	files, err := templates.ReadDir(langTemplateDir)
	if err != nil {
		fmt.Printf("Unsupported language: %s\n", lang)
		os.Exit(1)
	}

	// Step 2: Create the project by copying the template files
	os.Mkdir(projectName, 0755)
	for _, file := range files {
		fmt.Println(file.Name())
		srcPath := fmt.Sprintf("%s/%s", langTemplateDir, file.Name())
		destPath := fmt.Sprintf("%s/%s", projectName, file.Name())
		err := utils.CopyFile(srcPath, destPath)
		if err != nil {
			fmt.Printf("Failed to copy file: %v\n", err)
			os.Exit(1)
		}
	}

	// Step 3: git init
	err = os.Chdir(projectName)
	if err != nil {
		fmt.Printf("Failed to switch to project directory: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("git", "init")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to git init: %v\n", err)
		os.Exit(1)
	}

	// Step 4: git add .
	cmd = exec.Command("git", "add", ".")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to git add: %v\n", err)
		os.Exit(1)
	}

	// Step 5: git commit
	cmd = exec.Command("git", "commit", "-m", "-s", "Initial commit")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to git commit: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Project created successfully")
}

// add different languages hooks
