# cmd/project

Wiring layer only: `main`, root command, ordered command registry, service factories, CLI contract/integration tests. No business logic.

## KEY MECHANICS

- `command_registry.go`: commands self-register in `init()` via `registerOrderedCommand(commandKey*, commandOrder*, builder)`; duplicate keys panic at startup.
- `command_registrations.go`: the single place mapping registry keys to `internal/cli/*` constructors.
- `service_factories.go`: package-level `var` factories (`newCatalogService`, `newCreateService`, ...) that tests replace. Add new services here instead of introducing hidden globals.
- `root_command.go`: `newRootCmd` builds the tree; `PersistentPreRunE` loads the active TOML config once and attaches it to every command's context; `Execute` prints the final error and exits.
- Persistent flags live here: `--config`, `--explain-config`.

## TESTING

- `cli_contract_test.go` pins user-facing output verbatim - any CLI string change must update it deliberately.
- `Execute` failure paths: tests replace the `exitFunc` / `stderrWriter` vars.
- Integration tests restore cwd and injected vars via `t.Cleanup`.
