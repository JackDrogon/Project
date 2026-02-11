# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A CLI scaffolding tool (`project`) that creates new projects from embedded templates. Built with Go and Cobra.

## Build & Development Commands

```bash
just build          # Build binary to bin/project (with version ldflags)
just install        # Install binary to $GOPATH/bin
just lint           # Run golangci-lint (default config, no .golangci.yml)
just fmt            # Format all Go code
just test           # Run all tests
just test ./cmd/project/...  # Run tests for a specific package
just test-v         # Run tests with verbose output
just cover          # Generate coverage.html report
just pre-commit     # fmt → lint → test (run before committing)
just run <args>     # Build and run with arguments
just tidy           # go mod tidy
```

Version info is injected at build time via `-ldflags` into `pkg/version.GitTagSha`.

## Architecture

### Dependency injection pattern for embed.FS

Templates are embedded in `pkg/templates/embed.go` via `//go:embed all:cpp all:go`. The `embed.FS` is passed through the command tree via explicit constructor parameters — not global state:

```
main.go → Execute(templates.FS) → newRootCmd(fs) → newNewCmd(fs), newListCmd(fs)
```

### Command files (all in `cmd/project/`, package `main`)

- `main.go` — Entry point; calls `Execute(templates.FS)`
- `root_command.go` — Cobra root command; wires subcommands
- `new_command.go` — `project new -l <lang> <name>`: copies embedded template to new directory, renders `.tmpl` files, runs `git init/add/commit`
- `list_command.go` — `project list`: lists available template languages
- `version_command.go` — `project version`: prints git tag/sha
- `template_vars.go` — `TemplateVars` struct (ProjectName, ModulePath, Author, Year) used in template rendering
- `validate.go` — Project name validation (must start with letter, `[a-zA-Z0-9._-]`, max 255 chars)

### Template system

- Templates live in `pkg/templates/` organized by language (`cpp/`, `go/`)
- Files ending in `.tmpl` are rendered with `text/template` and have the suffix stripped (e.g., `go.mod.tmpl` → `go.mod`)
- Non-`.tmpl` files and files with invalid template syntax are copied as-is
- Available template variables: `{{.ProjectName}}`, `{{.ModulePath}}`, `{{.Author}}`, `{{.Year}}`

### Flags for `project new`

`-l/--lang` (required), `-m/--module`, `--force`, `--signoff`, `-n/--dry-run`

### Testing approach

Tests use `testing/fstest.MapFS` to mock the embedded filesystem and `t.TempDir()` for temporary directories. Table-driven test pattern throughout.
