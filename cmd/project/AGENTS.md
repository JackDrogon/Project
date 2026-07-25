# cmd/project

Wiring layer only: `main`, root command, the static subcommand list, service factories, CLI contract/integration tests. No business logic.

## KEY MECHANICS

- `dependencies.go`: the `dependencies` struct is the single assembly point; `newDependencies()` wires production collaborators. Everything downstream receives it as an argument - no package-level mutable state.
- `commands.go`: `subcommands(deps)` returns the top-level commands as one static slice. Adding a command means adding one line; there is no registry and no `init()`.
- `service_factories.go`: plain constructor funcs (`newCatalogService`, `newCreateService`, ...). Add new services here and expose them through `dependencies`.
- `root_command.go`: `newRootCmd(deps)` builds the tree; `PersistentPreRunE` loads the active TOML config once and attaches it to every command's context. The root sets `SilenceErrors`/`SilenceUsage` because cobra sends usage to the *out* stream.
- `main.go`: `main` is the only place that reads `os.Args`, touches `os.Stdout`/`os.Stderr`, or calls `os.Exit`; `run(deps, args, stdout, stderr) int` holds the logic and returns the exit code.
- Persistent flags live here: `--config`, `--explain-config`.

## TESTING

- `cli_contract_test.go` pins user-facing output verbatim - any CLI string change must update it deliberately.
- Tests build a `dependencies` value (`newTestDependencies`, `catalogTestDependencies`, `versionTestDependencies`) and override only the fields they care about; nothing needs `t.Cleanup` because nothing is global.
- `requireSubcommand` / `requireSubcommandWithDeps` return a parentless command so `Execute` runs it directly instead of rerouting through the root.
- Integration tests still restore the working directory via `t.Cleanup`.
