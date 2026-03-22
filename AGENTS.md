# AGENTS.md

Guidance for agentic coding assistants working in this repository.

## Project Snapshot

- This repo builds `project`, a Go CLI scaffolding tool built with Cobra.
- The module is `github.com/JackDrogon/project` and the repo targets Go `1.21.7`.
- The binary scaffolds new projects from embedded templates.
- Templates are compiled into the binary and live under `pkg/templates/`.
- The main executable entrypoint is `./cmd/project`.
- The core scaffolding logic lives in `pkg/scaffold`.
- Prefer facts from `justfile` and source code over README summaries when they differ.

## Rules File Status

- No `.cursor/rules/` directory was found when this guide was written.
- No `.cursorrules` file was found when this guide was written.
- No `.github/copilot-instructions.md` file was found when this guide was written.
- If any of those files are added later, treat them as additional repository-specific instructions.

## Repository Facts

- The canonical task runner and command source is the root `justfile`.
- There is no checked-in GitHub Actions workflow in `.github/workflows/` at the time of writing.

## Canonical Development Commands

### Build

- `just build` builds `./cmd/project` to `bin/project` with version ldflags.
- `just install` installs `./cmd/project` with the same ldflags.
- `just build-cover` builds an instrumented binary with `-cover`.
- `just run <args>` builds first, then runs `./bin/project <args>`.
- `just clean` removes `bin/`.

### Format, Lint, and Vet

- `just fmt` runs `go fmt ./...`.
- `just lint` runs `golangci-lint run`.
- `just vet` runs `go vet ./...`.
- `just generate` runs `go generate ./...` to regenerate auto-generated source files.
- `just pre-commit` runs `fmt`, `lint`, and `test` in sequence.

### Test

- `just test` runs `go test ./...`.
- `just test ./cmd/project/...` runs tests for a package subtree.
- `just test ./pkg/scaffold` runs tests for one package.
- `just test-v` runs `go test -v ./...`.
- `just test-v ./pkg/scaffold` runs one package verbosely.
- `just cover` generates `coverage.out` and `coverage.html`.

### Single-Test Execution

- There is no dedicated `just` recipe for one named test.
- For one test function, run `go test -run` directly.
- Example: `go test ./pkg/scaffold -run '^TestCreate_NoGitSkipsGit$'`.
- Example: `go test -v ./cmd/project -run '^TestExecute_ExitsOnError$'`.
- Use anchored regexes (`^...$`) so similarly named tests do not also run.

### Dependency and Info Commands

- `just tidy` runs `go mod tidy`.
- `just deps` prints the module dependency graph.
- `just info` prints Go toolchain info and the module line from `go.mod`.
- `just log n='20'` shows a concise git log.

## Version Injection and Build Metadata

- `just build` and `just install` inject `pkg/version.Tag` via `-ldflags`.
- The tag value comes from `git describe --abbrev=0 --always --tags`.
- The version string is then combined with revision and dirty state from `runtime/debug.BuildInfo`.
- If you touch version documentation or build flags, keep it aligned with `pkg/version/version.go` and `justfile`.

## Codebase Layout

- `cmd/project/` contains Cobra command wiring and CLI-facing behavior.
- `pkg/scaffold/` contains creation, validation, template walking, inspection, and git mode logic.
- `pkg/git/` contains the thin wrapper around git command execution.
- `pkg/version/` contains version string formatting and build-info extraction.
- `pkg/templates/` contains embedded templates grouped by language.

## Architectural Conventions

- Keep CLI parsing and flag wiring in `cmd/project/`.
- Keep reusable logic in `pkg/*` packages, especially `pkg/scaffold`.
- Pass dependencies explicitly instead of introducing new hidden globals.
- Follow the existing testability pattern: inject collaborators where practical.
- Existing examples include injected git runners and replaceable package-level OS/build-info helpers for tests.
- Command constructors follow `newXxxCmd` and return `*cobra.Command`.
- `RunE` handlers should map flags and args into focused helpers or `scaffold.Options`.
- Avoid moving business rules into command files unless the rule is inherently CLI-only.
- Keep command files thin and declarative, with stable user-facing strings because tests assert on them.

## Formatting and Imports

- Run `just fmt` after editing Go files.
- Follow `gofmt` output exactly; do not preserve manual alignment.
- Keep standard library imports first.
- Keep non-stdlib imports after the standard library block.
- Some files split local and third-party imports into separate blocks; prefer the local style of the file you touch.
- Do not churn import ordering in untouched files unless needed for the change.

## Naming Conventions

- Exported identifiers use Go's standard PascalCase.
- Unexported helpers use lowerCamelCase.
- Command builders use `newRootCmd`, `newListCmd`, `newInitCmd`, `newVersionCmd`, and similar names.
- Constants are grouped by concern and often use a shared prefix, such as `outputFormat*` or `inspectMode*`.
- Tests commonly use `TestXxx_Scenario` names for scenario coverage.

## Types and API Design

- Prefer small concrete structs with focused methods, as seen in `Creator` and `Options`.
- Add interfaces only when they remove real coupling or enable testing cleanly.
- Preserve option structs when plumbing multiple related parameters through layers.
- Favor explicit fields and readable control flow over compact but opaque abstractions.

## Error Handling

- Prefer early returns over nested conditionals.
- Wrap contextual errors with `fmt.Errorf("...: %w", err)` when crossing responsibility boundaries.
- Use `errors.Is` for filesystem or OS-specific checks.
- Let Cobra `RunE` return errors upward instead of printing directly.
- The top-level `Execute` function is responsible for printing the final error and exiting.
- Best-effort progress writes often ignore write failures with `_, _ = fmt.Fprintln(...)`.
- Preserve that write-ignore pattern only for non-critical user-facing output.

## Template and Scaffolding Behavior

- Files ending in `.tmpl` are rendered with Go's `text/template` and written without the suffix.
- Non-`.tmpl` files are copied as-is, even if they contain template-like text.
- Invalid `.tmpl` syntax should fail fast rather than silently copy.
- Template rendering uses repository-defined variables such as `ProjectName`, `ModulePath`, `Author`, and `Year`.
- `--signoff` is only valid with `--git=init+commit`, and `--no-git` conflicts with other non-`none` git modes.
- `init` derives `ProjectName` from the target directory and can reuse an existing empty directory.
- When adding a new language template, update `pkg/templates/embed.go` so the directory is embedded.
- If you change scaffold behavior, verify both creation and inspection paths.

## Output Behavior

- `list` and `inspect` reject using `--json` and `--yaml` together.
- YAML output is produced by repository-owned writer helpers rather than a third-party serializer.

## Testing Conventions

- Prefer table-driven tests with `t.Run` for multi-case behavior.
- Use `t.TempDir()` for filesystem-based tests.
- Use `testing/fstest.MapFS` for embedded template fixtures where possible.
- Use `t.Cleanup` to restore global state, working directories, and injected function variables.
- Verify side effects directly: files created, output written, git calls made, and exit behavior triggered.
- Inject fakes instead of shelling out in unit tests.
- This codebase explicitly tests failure paths by replacing package-level helper variables inside tests.

## Change Guidance for Agents

- Use the `justfile` as the canonical source for repo commands.
- Keep diffs surgical and consistent with the touched file.
- Match existing patterns before introducing a new abstraction.
- Update or add tests when behavior changes.
- Do not remove or weaken tests just to make a change pass.
- If you touch command flags or output, inspect related tests under `cmd/project/*_test.go`.
- If you touch scaffolding logic, inspect `pkg/scaffold/*_test.go` and template-related tests.
- If you add, remove, or change permissions of template files under `internal/adapters/templatesrc/`,
  run `just generate` to regenerate the permission metadata, then commit the updated
  `permissions_generated.go` alongside your changes.
- If you touch version behavior, inspect both `pkg/version/*` and command-level version tests.
- Run the narrowest relevant tests first, then expand to `just test` for non-trivial Go changes.
- Run `just fmt` after Go edits and `just lint` when the change could affect style or static analysis.
- Prefer `just pre-commit` before finishing large or risky changes.
