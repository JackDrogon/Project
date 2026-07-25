# internal/cli

One package per command group (`catalog`, `completion`, `config`, `scaffold`, `version`). Builds Cobra commands; no business logic.

## CONVENTIONS

- Top-level constructors: exported `NewXxxCommand(deps Dependencies) *cobra.Command` (`NewListCommand`, `NewNewCommand`, ...). Nested subcommands: unexported `newXxxCommand` (see config's path/init/validate).
- Each package declares its own `Dependencies` struct of factory funcs (`NewService`, and for scaffold `NewCreator(out io.Writer)`); missing deps panic as a fail-fast wiring guard.
- Commands own their streams: build collaborators from `cmd.OutOrStdout()` / `cmd.ErrOrStderr()` rather than capturing a writer at wiring time.
- `RunE` only maps flags/args into app-service requests or `create.Options` - business rules belong in `internal/app`.
- Keep command files thin and declarative; user-facing strings are pinned by `cmd/project/cli_contract_test.go`.

## PACKAGE NOTES

- `scaffold`: `command_flags.go` centralizes flag definitions and `Changed` tracking (`cmd.Flags().Changed(...)`); resolvers rely on `Changed` to distinguish explicit flags from defaults.
- `catalog`: output format constants `outputFormat*` in `output_spec.go`; the `--toml` bool flag toggles TOML output.
