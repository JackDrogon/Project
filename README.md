# Project

A CLI scaffolding tool that creates new projects from embedded templates. All templates are compiled into a single binary — no external files needed at runtime.

## Supported Languages

- **Go** — structured CLI scaffold with `cmd/app`, `internal/`, tests, `.golangci.yml`, GitHub Actions CI, GoReleaser config, Codecov config, contributor docs, `README.md`, and `justfile`
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
project new -l cpp myapp
```

This will:
1. Copy template files into the `myapp/` directory
2. Render template variables (e.g., project name, module path) in `.tmpl` files
3. Run git setup based on `--git` mode (default `init+commit`)

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--lang` | `-l` | Programming language (required) |
| `--module` | `-m` | Module path, e.g., `github.com/user/project` (defaults to project name) |
| `--force` | | Remove and recreate existing project directory |
| `--git` | | Git workflow: `none`, `init-only`, `init+commit` |
| `--signoff` | | Add `Signed-off-by` trailer to the initial commit |
| `--dry-run` | `-n` | Preview files without creating them |
| `--no-git` | | Deprecated alias for `--git none` |

### Examples

```bash
# Create a Go project with a custom module path
project new -l go myapp -m github.com/myorg/myapp

# Preview what files would be created
project new -l go myapp -n

# Create files without initializing git
project new -l go myapp --no-git

# Initialize git repo without creating an initial commit
project new -l go myapp --git init-only

# Overwrite an existing directory
project new -l go myapp --force

# List available languages
project list

# List with metadata
project list --detail

# List in YAML
project list --detail --yaml

# Inspect one language template
project inspect go

# Inspect only rendered files
project inspect go --mode render

# Show version
project version
```

## Template Variables

Templates (`.tmpl` files) support the following variables via Go's `text/template`:

| Variable | Description | Default |
|----------|-------------|---------|
| `{{.ProjectName}}` | Name passed to `project new` | — |
| `{{.ModulePath}}` | From `--module` flag | Same as ProjectName |
| `{{.Author}}` | System username | `"author"` |
| `{{.Year}}` | Current year | — |

Only files ending in `.tmpl` are rendered, and the suffix is stripped (e.g., `go.mod.tmpl` → `go.mod`). Non-`.tmpl` files are copied as-is. Invalid `.tmpl` syntax returns an error.

## Template Discovery

- `project list` prints available language names.
- `project list --detail` prints file count, template count, and template variables per language.
- `project list --json` prints machine-readable output.
- `project list --yaml` prints machine-readable YAML output.
- `project inspect <lang>` shows per-file mappings (`source -> output`) and whether each file is rendered or copied.
- `project inspect <lang> --mode render|copy` filters files by render/copy behavior.
- `project inspect <lang> --json|--yaml` prints structured output.

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

1. Create a directory under `pkg/templates/` with the language name (e.g., `pkg/templates/rust/`)
2. Add template files; use `.tmpl` suffix for files that need variable substitution
3. Update the `//go:embed` directive in `pkg/templates/embed.go` to include the new directory:
   ```go
   //go:embed all:cpp all:go all:rust
   var FS embed.FS
   ```

## Project Name Rules

Project names must:
- Start with a letter (`a-z`, `A-Z`)
- Contain only letters, digits, `.`, `_`, or `-`
- Be at most 255 characters
