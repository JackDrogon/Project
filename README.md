# Project

A CLI scaffolding tool that creates new projects from embedded templates. All templates are compiled into a single binary — no external files needed at runtime.

## Supported Languages

- **Go** — structured CLI scaffold with `cmd/<projectname-lower>`, `internal/`, tests, `.golangci.yml`, GitHub Actions CI, GoReleaser config, Codecov config, contributor docs, `README.md`, and `justfile`
- **C++** — `CMakeLists.txt`, `src/main.cc`, `include/`, `dev-tools/`, `justfile`

Run `project list` to see all available languages.

## Installation

Requires Go 1.21+.

```bash
# Install directly from GitHub
go install github.com/JackDrogon/project/cmd/project@latest
```

### From source

Requires [just](https://github.com/casey/just).

```bash
git clone https://github.com/JackDrogon/project.git
cd project

# Build to ./bin/project
just build

# Or install to $GOPATH/bin directly
just install
```

## Usage

### Create a new project

```bash
project new -l go myapp
project new -l go github.com/myorg/myapp
project new -l cpp myapp
```

For Go projects, you can pass either a local project name or a full module path as the positional argument. When you pass a module path directly, `project` derives the output directory from the repository name: `project new -l go github.com/myorg/myapp` creates `./myapp` with `module github.com/myorg/myapp`, and `project new -l go github.com/myorg/myapp/v2` still creates `./myapp` with `module github.com/myorg/myapp/v2`.

This will:
1. Copy template files into the `myapp/` directory
2. Render template variables (e.g., project name, module path) in `.tmpl` files
3. Run git setup based on `--git` mode (default `init+commit`)

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--lang` | `-l` | Programming language (required unless provided via `--replay`) |
| `--module` | `-m` | Module path, e.g., `github.com/user/project` (defaults to project name, or to the positional Go module path when you pass one directly) |
| `--force` | | Remove and recreate existing project directory |
| `--git` | | Git workflow: `none`, `init-only`, `init+commit` |
| `--signoff` | | Add `Signed-off-by` trailer to the initial commit |
| `--dry-run` | `-n` | Preview files without creating them |
| `--replay` | | Load project configuration from a TOML replay file |
| `--write-replay` | | Write resolved replay to a TOML file after success |
| `--set` | | Set a template-specific input value (`--set key=value`) |
| `--no-git` | | Deprecated alias for `--git none` |

### Examples

```bash
# Create a Go project with a custom module path
project new -l go myapp -m github.com/myorg/myapp

# Create a project using a replay file
project new --replay .project-replay.toml

# Override replay values with CLI flags
project new --replay .project-replay.toml --set go_version=1.22

# Save resolved replay to a file
project new -l go myapp --write-replay .project-replay.toml

# Preview what files would be created
project new -l go myapp -n
```

## Dry-run Execution Plan

The `--dry-run` (`-n`) flag provides a detailed preview of the resolved inputs and the planned file operations:

```text
Creating project with language: go, project name: myapp
Dry-run mode: no files will be created
template: go
description: Production-ready Go CLI starter
target_dir: myapp
resolved inputs:
  project_name: myapp
  module_path: myapp
  go_version: 1.21
  git_mode: init+commit
explicit overrides:
  (none)
actions:
  create myapp/.github/
  copy go/.github/workflows/ci.yml -> myapp/.github/workflows/ci.yml
  render go/README.md.tmpl -> myapp/README.md
  ...
```

## Scaffolding Contracts

### Template Manifest (`.project-template-manifest.toml`)

Templates may include a `.project-template-manifest.toml` manifest at their root to define metadata and input variables. This is a reserved file and is never copied to the target project. The current schema version is `2`.

Example manifest:
```toml
version = 2
name = "go"
description = "Production-ready Go CLI starter"

[[inputs]]
key = "go_version"
template_var = "GoVersion"
required = false
```

### Replay File (`.project-replay.toml`)

Using `--write-replay` produces a TOML file containing the final resolved configuration. This file enables consistent project recreation and can be used with `--replay`.

**Contract Rules:**
- **Precedence**: Explicit CLI flags and `--set` arguments always take precedence over values in a replay file.
- **Derived Defaults**: Runtime-derived values (like the current `year` or system `author`) are not persisted in the replay file unless they were explicitly provided via CLI or replay.

### Third-party YAML Allowlist

The following third-party configuration files are explicitly supported and included in the Go template:

.github/workflows/ci.yml
codecov.yml
.golangci.yml
.goreleaser.yml.tmpl

## Template Variables

Templates (`.tmpl` files) support the following variables via Go's `text/template`:

| Variable | Description | Default |
|----------|-------------|---------|
| `{{.ProjectName}}` | Name passed to `project new`, or the derived directory name from a direct Go module path argument | — |
| `{{.ModulePath}}` | From `--module` flag or a direct Go module path argument | Same as ProjectName |
| `{{.Author}}` | System username | `"author"` |
| `{{.Year}}` | Current year | — |

Only files ending in `.tmpl` are rendered, and the suffix is stripped (e.g., `go.mod.tmpl` → `go.mod`). Non-`.tmpl` files are copied as-is. Invalid `.tmpl` syntax returns an error.

## Template Discovery

- In structured `list --detail` and `inspect` output, `description`, `manifest_version`, `input_names`, and `inputs` come from `.project-template-manifest.toml`, while `variables`, `file_count`, `template_count`, and `files` are derived from inspecting the embedded template tree.
- `project list` prints available language names.
- `project list --detail` prints file count, template count, and template variables per language.
- `project list --toml` prints machine-readable TOML output.
- `project inspect <lang>` shows per-file mappings (`source -> output`) and whether each file is rendered or copied.
- `project inspect <lang> --mode render|copy` filters files by render/copy behavior.
- `project inspect <lang> --toml` prints structured output.

## Shell Completion

Generate shell completion scripts with `project completion <shell>`:

### Bash

```bash
# Current session
source <(project completion bash)

# Persistent (Linux)
project completion bash > /etc/bash_completion.d/project

# Persistent (macOS with Homebrew)
project completion bash > $(brew --prefix)/etc/bash_completion.d/project
```

### Zsh

```bash
# Enable completion if not already
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Install completion
project completion zsh > "${fpath[1]}/_project"

# Start a new shell to take effect
```

### Fish

```bash
# Current session
project completion fish | source

# Persistent
project completion fish > ~/.config/fish/completions/project.fish
```

### PowerShell

```powershell
# Current session
project completion powershell | Out-String | Invoke-Expression

# Persistent — add to your PowerShell profile
project completion powershell > project.ps1
```

## Development

```bash
just build          # Build binary
just test           # Run all tests
just test-v         # Run tests with verbose output
just lint           # Run golangci-lint
just fmt            # Format code
just cover          # Generate coverage report (coverage.html)
just pre-commit     # fmt → lint → test
just run <args>     # Build and run (e.g., just run new -l go myapp)
just tidy           # go mod tidy
```

### Adding a new language template

1. Create a directory under `internal/adapters/templatesrc/` with the language name (e.g., `internal/adapters/templatesrc/rust/`)
2. Add template files; use `.tmpl` suffix for files that need variable substitution
3. Update the `//go:embed` directive in `internal/adapters/templatesrc/embed.go` to include the new directory:
   ```go
   //go:embed all:cpp all:go all:rust
   var FS embed.FS
   ```

## Project Name Rules

Project names must:
- Start with a letter (`a-z`, `A-Z`)
- Contain only letters, digits, `.`, `_`, or `-`
- Be at most 255 characters
