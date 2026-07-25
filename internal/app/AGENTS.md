# internal/app

Application services implementing command behavior: `catalog`, `config`, `create`, `version`. `create` is the largest: resolution, scaffolding orchestration, replay.

## create: RESOLUTION MODEL

- Current API: `Service.BuildNewOptions` / `BuildInitOptions` (plus `BuildNewSpec`/`BuildInitSpec` for dry-run origin reporting).
- `resolver.go` resolves each setting by precedence:
  1. explicit CLI flag (`Changed.X` true)
  2. replay file (when `--replay`)
  3. active config (`[new]`/`[init]` sections)
  4. flag default
- `--no-git` is a deprecated alias folded into git mode: an explicit `--no-git` makes `resolveGitMode` return `""` (skipping replay/config), then domain `scaffold.ResolveGitMode` normalizes NoGit+GitMode and rejects conflicts.
- `replay_codec.go` normalizes NoGit+GitMode into a resolved git mode BEFORE writing a replay - replay files never carry `NoGit`.
- `--config` and `--replay` are mutually exclusive; replay mode ignores the global config entirely.
- `service_compat.go` exported wrappers (`ResolveNewProjectArgs`, ...) exist for cross-package integration tests in `cmd/project` - do not remove them as "dead code".
- Domain rules (name validation, git modes, `TemplateVars`) live in `internal/scaffold`, dependency-free.
- `context.Context` is threaded only where an external process is launched - `Creator.Create`/`BuildDryRunPlan`/`ExecuteScaffoldSpec` pass it to git and to `go env` version detection. Pure in-memory services (`catalog`, `config`, `version`, resolution) deliberately take no context.

## config

- `context.go` attaches `ActiveConfig` to `context.Context`; the root command's `PersistentPreRunE` is the only loader.
- One-shot flags (`--force`, `--dry-run`, `--replay`, `--write-replay`) are rejected in the global config.

## TESTING

- `create/service_replay_test.go` covers the precedence matrix (flag vs replay vs config) - extend it when touching resolution.
