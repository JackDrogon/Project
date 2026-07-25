# internal/adapters

Integrations behind the app layer. One package per concern.

## PACKAGES

- `buildinfo`: version metadata. `Tag` is injected via ldflags (`just build`); revision and dirty state come from `runtime/debug.BuildInfo`.
- `gitexec`: thin git command wrapper built on `exec.CommandContext`, so cancelling the context kills the child; tests inject fake runners instead of shelling out.
- `protocoltoml`: TOML codecs for the three protocol files - user config (`config.go`), replay (`replay.go`), template manifest (`manifest.go`). Legacy JSON payloads are rejected explicitly.
- `templatefs`: template tree traversal + materialization. `.tmpl` files render via `text/template` and drop the suffix; everything else copies as-is; invalid template syntax fails fast.
- `templatesrc`: embedded templates grouped by language.

## templatesrc RULES

- `embed.go`: `//go:embed all:cpp all:go all:rust` - adding a language template requires extending this directive.
- `permissions_generated.go` is generated (`just generate`) and captures file permission metadata; CI fails when it is stale. Regenerate after ANY template file add/remove/chmod.
- `.project-template-manifest.toml` (schema version 2) declares template metadata/inputs; it is never copied into generated projects.
- Rendering variables come from `scaffold.TemplateVars`: `ProjectName`, `ProjectNameLower`, `ModulePath`, `GoVersion`, `Author`, `Year`.
