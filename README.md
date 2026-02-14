# Project

A CLI scaffolding tool that creates new projects from embedded templates. All templates are compiled into a single binary — no external files needed at runtime.

## Supported Languages

- **Go** — `go.mod`, `main.go`, `.gitignore`, `README.md`, `Makefile`
- **C++** — `CMakeLists.txt`, `Makefile`, `src/main.cc`, `include/`, `dev-tools/` (cpplint, formatting scripts)

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
3. Run `git init && git add . && git commit -m "Initial commit"`

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--lang` | `-l` | Programming language (required) |
| `--module` | `-m` | Module path, e.g., `github.com/user/project` (defaults to project name) |
| `--force` | | Remove and recreate existing project directory |
| `--signoff` | | Add `Signed-off-by` trailer to the initial commit |
| `--dry-run` | `-n` | Preview files without creating them |

### Examples

```bash
# Create a Go project with a custom module path
project new -l go myapp -m github.com/myorg/myapp

# Preview what files would be created
project new -l go myapp -n

# Overwrite an existing directory
project new -l go myapp --force

# List available languages
project list

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

Files ending in `.tmpl` have the suffix stripped after rendering (e.g., `go.mod.tmpl` → `go.mod`). Files that are not valid Go templates are copied as-is.

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
