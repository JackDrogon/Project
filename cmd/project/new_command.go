package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

const tmplSuffix = ".tmpl"

// newNewCmd creates the "new" subcommand that scaffolds a project from templates.
func newNewCmd(templateFS embed.FS) *cobra.Command {
	var lang string
	var module string
	var force bool
	var signoff bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "new [project_name]",
		Short: "Create new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createProject(templateFS, lang, args[0], module, force, signoff, dryRun)
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "", "Programming language for the project")
	_ = cmd.MarkFlagRequired("lang")
	cmd.Flags().StringVarP(&module, "module", "m", "", "Module path (e.g. github.com/user/project)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing project directory")
	cmd.Flags().BoolVar(&signoff, "signoff", false, "Add Signed-off-by trailer to the initial commit")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Preview files without creating them")

	return cmd
}

// createProject creates a new project with the given language template.
func createProject(templateFS embed.FS, lang, projectName, modulePath string, force, signoff, dryRun bool) error {
	// Validate project name
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	fmt.Printf("Creating project with language: %s, project name: %s\n", lang, projectName)

	// Check if the language is supported
	langTemplateDir := lang
	if _, err := fs.ReadDir(templateFS, langTemplateDir); err != nil {
		return fmt.Errorf("unsupported language: %s", lang)
	}

	// Build template variables
	vars := newTemplateVars(projectName, modulePath)

	if dryRun {
		fmt.Println("Dry-run mode: no files will be created")
		return previewEmbedDir(templateFS, langTemplateDir, projectName)
	}

	// Check if the target directory already exists
	if info, err := os.Stat(projectName); err == nil && info.IsDir() {
		if !force {
			return fmt.Errorf("directory %q already exists; use --force to overwrite", projectName)
		}
		fmt.Printf("Warning: directory %q already exists, overwriting due to --force\n", projectName)
	}

	// Copy the template files into the project directory
	if err := copyEmbedDir(templateFS, langTemplateDir, projectName, vars); err != nil {
		return fmt.Errorf("failed to copy template files: %w", err)
	}

	// Initialize git repository
	gitCommitArgs := []string{"commit", "-m", "Initial commit"}
	if signoff {
		gitCommitArgs = []string{"commit", "-s", "-m", "Initial commit"}
	}
	gitCmds := [][]string{
		{"init"},
		{"add", "."},
		gitCommitArgs,
	}
	for _, gitArgs := range gitCmds {
		if err := runGit(projectName, gitArgs...); err != nil {
			return err
		}
	}

	fmt.Println("Project created successfully")
	return nil
}

// runGit executes a git command in the given directory.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed: %w\n%s", args[0], err, string(output))
	}
	return nil
}

// renderTemplate applies TemplateVars to content using text/template.
// If parsing fails (e.g. the file is not a Go template), the original
// content is returned unchanged.
func renderTemplate(content []byte, vars TemplateVars) ([]byte, error) {
	tmpl, err := template.New("").Option("missingkey=zero").Parse(string(content))
	if err != nil {
		// Not a valid template — return as-is
		return content, nil
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return content, nil
	}
	return buf.Bytes(), nil
}

// copyEmbedDir recursively copies a directory from an embedded filesystem
// to the local filesystem, rendering template variables in file contents.
func copyEmbedDir(fsys fs.FS, srcDir, destDir string, vars TemplateVars) error {
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
		// Strip .tmpl suffix so "go.mod.tmpl" becomes "go.mod"
		destName := strings.TrimSuffix(entry.Name(), tmplSuffix)
		destPath := filepath.Join(destDir, destName)
		fmt.Printf("  create %s\n", destPath)

		if entry.IsDir() {
			if err := copyEmbedDir(fsys, srcPath, destPath, vars); err != nil {
				return err
			}
			continue
		}

		content, err := fs.ReadFile(fsys, srcPath)
		if err != nil {
			return err
		}

		rendered, err := renderTemplate(content, vars)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", srcPath, err)
		}

		if err := os.WriteFile(destPath, rendered, 0644); err != nil {
			return err
		}
	}
	return nil
}

// previewEmbedDir prints what files would be created without writing anything.
func previewEmbedDir(fsys fs.FS, srcDir, destDir string) error {
	entries, err := fs.ReadDir(fsys, srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := path.Join(srcDir, entry.Name())
		destName := strings.TrimSuffix(entry.Name(), tmplSuffix)
		destPath := filepath.Join(destDir, destName)

		if entry.IsDir() {
			fmt.Printf("  create %s/\n", destPath)
			if err := previewEmbedDir(fsys, srcPath, destPath); err != nil {
				return err
			}
			continue
		}

		fmt.Printf("  create %s\n", destPath)
	}
	return nil
}
