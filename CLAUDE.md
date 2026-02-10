# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A CLI scaffolding tool (`project`) that creates new projects from embedded templates. Built with Go and Cobra.

## Build & Development Commands

```bash
make build          # Build binary to bin/project
make lint           # Run golangci-lint
make fmt            # Format all Go code
make test           # Run tests
make VERBOSE=1 ...  # Show command output
```

Version info is injected at build time via `-ldflags` into `pkg/version.GitTagSha`.

## Architecture

- `cmd/project/main.go` — Entry point; embeds `templates/` directory via `//go:embed`
- `cmd/project/root_command.go` — Cobra root command definition
- `cmd/project/new_command.go` — `project new -l <lang> <name>`: creates a project from embedded templates, then runs `git init/add/commit` in the new directory
- `cmd/project/list_command.go` — `project list`: lists available template languages by reading the embedded `templates/` directory
- `cmd/project/version_command.go` — `project version`: prints git tag/sha
- `cmd/project/templates/` — Embedded template files organized by language (`cpp/`, `go/`)
- `pkg/version/version.go` — `GitTagSha` variable, set via ldflags at build time

Key pattern: templates are embedded into the binary via `embed.FS` in `main.go` and accessed as a package-level `templates` variable across all command files in the `main` package.
