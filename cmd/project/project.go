package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/JackDrogon/project/pkg/version"

	"github.com/spf13/cobra"
)

//go:embed templates
var templates embed.FS

var rootCmd = &cobra.Command{
	Use:   "project",
	Short: "project is a tool to create new project",
}

var newCmd = &cobra.Command{
	Use:   "new [language]",
	Short: "create new project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		createProject(args[0])
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all support languages",
	Run: func(cmd *cobra.Command, args []string) {
		listLangs()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.GitTagSha)
	},
}

func init() {
	rootCmd.AddCommand(newCmd, listCmd, versionCmd)
}

func listLangs() {
	langs, err := templates.ReadDir("templates")
	if err != nil {
		panic(err)
	}

	for _, lang := range langs {
		fmt.Println(lang.Name())
	}
}

func createProject(lang string) {
	langTemplateDir := fmt.Sprintf("templates/%s", lang)
	files, err := templates.ReadDir(langTemplateDir)
	if err != nil {
		fmt.Printf("unsupported language: %s\n", lang)
		os.Exit(1)
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
