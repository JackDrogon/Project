# AGENTS.md

Go CLI scaffolding tool (`project`) built with Cobra. Module `github.com/JackDrogon/project`, Go `1.25.0`. Templates are embedded into the single binary. Prefer facts from `justfile` and source code over README summaries when they differ.

## STRUCTURE

```
cmd/project/          # main, root command, static subcommand list, service factories, contract tests
internal/cli/         # one package per command group: catalog, completion, config, scaffold, version
internal/app/         # application services: catalog, config, create, version
internal/scaffold/    # dependency-free domain: name validation, git modes, template vars
internal/adapters/    # buildinfo, gitexec, protocoltoml, templatefs, templatesrc
internal/presenters/  # catalog output rendering: text (default) or TOML
internal/testsupport/ # shared test fixtures
acceptance/           # `acceptance`-tagged E2E: real binary + real templates + real toolchains
```

Layer details: [cmd/project/AGENTS.md](cmd/project/AGENTS.md), [internal/cli/AGENTS.md](internal/cli/AGENTS.md), [internal/app/AGENTS.md](internal/app/AGENTS.md), [internal/adapters/AGENTS.md](internal/adapters/AGENTS.md).

## WHERE TO LOOK

| Task | Location |
|------|----------|
| Add/change CLI flags or output | `internal/cli/*` + contract tests in `cmd/project/cli_contract_test.go` |
| Scaffold/resolve/replay logic | `internal/app/create` + domain rules in `internal/scaffold` |
| Add/edit embedded templates | `internal/adapters/templatesrc/` (then `just generate`) |
| Version output | `internal/adapters/buildinfo` + `internal/app/version` |
| `list`/`inspect` output format | `internal/presenters` |
| Add/remove a top-level command | `cmd/project/commands.go` (+ `cmd/project/dependencies.go` for a new service) |

## CONVENTIONS (project-specific)

- `justfile` is the canonical command source.
- Formatting is `gofumpt` (stricter than `go fmt`); run `just fmt` after Go edits.
- `cmd/project/cli_contract_test.go` pins user-facing CLI strings verbatim - update it deliberately when output changes.
- `--toml` is the only machine-readable output flag; there are no `--json`/`--yaml` flags.
- Collaborators are injected, not global: tests build a `dependencies` value and override fields instead of using mocking frameworks.

## COMMANDS

```bash
just build            # → bin/project (ldflags inject internal/adapters/buildinfo.Tag from git describe)
just test [pkg]       # go test; e.g. just test ./internal/app/create
just acceptance       # go test -tags acceptance ./acceptance/... (needs go, cargo, cmake, git)
just lint             # golangci-lint run
just fmt              # gofumpt -l -w .
just pre-commit       # generate → fmt → lint → test → spellcheck
go test ./internal/app/create -run '^TestServiceBuildNewOptions_UsesReplayWhenArgOmitted$'  # single test
```

## NOTES

- Any template file add/remove/chmod under `internal/adapters/templatesrc/` requires `just generate`; CI fails if `permissions_generated.go` is stale.
- CI (`.github/workflows/ci.yml`): `test`, `acceptance`, `lint`, `spellcheck` jobs; Go version is read from `go.mod`.
- Run the narrowest relevant tests first, then `just test`; prefer `just pre-commit` before finishing risky changes.
- `just test` never runs `acceptance/` (build tag gated). After touching templates, `templatefs`, `templatesrc`, or scaffolding/git behavior, also run `just acceptance` - it is the only check that the shipped templates actually produce buildable projects.
- Scaffolding never deletes files: a non-empty destination is always rejected, and `--force` only permits an existing *empty* directory.
