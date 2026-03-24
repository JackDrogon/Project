# CLI Query and Create Architecture

This document explains the current architecture behind two major CLI flows:

- catalog-oriented commands like `project list` and `project inspect`
- create-oriented commands like `project new` and `project init`

It is intended for maintainers working in these areas:

- `cmd/project/`
- `internal/app/catalog/`
- `internal/presenters/`

The design intentionally separates command parsing, request/query construction, application orchestration, policy evaluation, and output rendering.

## End-to-end flow

The catalog-oriented CLI flow is:

1. Cobra commands parse flags and args in `cmd/project/`
2. Command spec builders turn flags into typed query/spec objects
3. `internal/app/catalog` executes those queries through a public facade
4. Analysis results are projected into summaries or inspections
5. `internal/presenters` renders those views according to explicit output/view specs

This keeps each layer focused:

- `cmd/project`: CLI orchestration only
- `internal/app/catalog`: application logic, query execution, domain rules
- `internal/presenters`: rendering only

The create-oriented CLI flow is:

1. Cobra commands parse flags and args in `cmd/project/`
2. Scaffold command spec builders turn flags into typed create/spec objects
3. `internal/app/create` resolves runtime, precedence, target, and execution settings
4. The create facade builds a `ScaffoldSpec`
5. The creator executes the spec and optionally writes replay output

This keeps the create pipeline focused:

- `cmd/project`: CLI orchestration only
- `internal/app/create`: request/spec building, runtime resolution, execution orchestration
- `internal/scaffold`: final create request model and lower-level domain behavior

## Layer responsibilities

### Command layer: `cmd/project/`

The command layer should remain thin.

Its responsibilities are:

- define Cobra flags and positional arguments
- validate CLI-only flag combinations
- build typed query/spec objects
- call the catalog facade
- pass results to presenters

Important files:

- `catalog_command_spec.go`
- `catalog_output_spec.go`
- `list_catalog_command.go`
- `inspect_catalog_command.go`

#### Command spec builders

The command layer currently uses explicit builders:

- `listCommandSpecBuilder`
- `inspectCommandSpecBuilder`
- `newScaffoldCommandSpecBuilder`
- `initScaffoldCommandSpecBuilder`

These builders centralize:

- flag compatibility rules
- query construction
- output spec construction
- scaffold spec construction

That keeps command `RunE` handlers declarative instead of procedural.

### Catalog facade layer: `internal/app/catalog/service.go`

`Service` is the public facade for catalog operations.

It exposes the stable API used by the command layer:

- `ListLangs()`
- `ListSummaries()`
- `QuerySummaries(...)`
- `Inspect(...)`
- `QueryInspection(...)`

The facade should not own detailed query execution rules.

### Create facade layer: `internal/app/create/service.go`

`Service` is also the public facade for create-oriented operations.

It exposes the stable API used by `new` and `init` command handlers:

- `BuildNewOptions(...)`
- `BuildInitOptions(...)`
- `BuildNewSpec(...)`
- `BuildInitSpec(...)`
- `ExecuteScaffoldSpec(...)`

The create facade should coordinate request/spec building and execution, but not hold every resolution rule inline.

### Query execution layer: `internal/app/catalog/executor.go`

`QueryExecutor` owns the details of executing typed catalog queries.

Current responsibilities:

- run `SummaryQuery`
- run `InspectionQuery`
- coordinate analyzer + projectors

This keeps query execution details out of the facade.

### Create resolution and execution layer: `internal/app/create/`

The create pipeline uses explicit request/spec and resolver objects instead of embedding all rules in command handlers.

Important pieces:

- `NewRequest`
- `InitRequest`
- `ScaffoldSpec`
- `ScaffoldSettingsResolver`
- `NewTargetResolver`
- `InitTargetResolver`

This lets `new` and `init` share a unified execution path while keeping mode-specific target resolution separate.

### Analysis layer: `internal/app/catalog/analyzer.go`

`Analyzer` reads template manifests and files and produces a stable intermediate model:

- `Analysis`

`Analysis` is the canonical internal representation used to derive both summary and inspection views.

### Create request/spec layer: `internal/app/create/service.go`

The create flow does not use catalog-style analysis objects, but it does use an explicit request/spec pipeline:

- CLI flags map into `NewRequest` or `InitRequest`
- service methods derive `Options`
- service methods then wrap those into `ScaffoldSpec`

This keeps command parsing separate from final create execution.

### View building / projection layer

`Analysis` exposes:

- `AnalysisViewBuilder`

Concrete builders:

- `SummaryBuilder`
- `InspectionBuilder`

Projectors consume those builders:

- `summaryProjector`
- `inspectionProjector`

This makes the flow explicit:

- analyze once
- build a requested view
- render later

### Domain rules layer

Catalog rules are modeled as injectable dependencies rather than package-global helper logic.

The current rule objects are:

- `RepoAssetRegistry`
- `InspectModePolicy`
- `GovernancePolicy`

They are grouped by:

- `Dependencies`

And injected through:

- `NewServiceWithDeps(...)`

#### Rule ownership

`RepoAssetRegistry`

- asset-name to path mapping
- known repo asset enumeration
- file grouping into repo/language buckets
- repo asset derivation from analyzed files

`InspectModePolicy`

- matching an `InspectMode` to an `InspectionFile`

`GovernancePolicy`

- ranking governance tiers
- deriving governance tier from an `Inspection`

### Create resolver layer

Create-specific resolution rules are modeled as explicit resolvers instead of being left as inline command logic.

Current resolver objects:

- `ScaffoldSettingsResolver`
- `NewTargetResolver`
- `InitTargetResolver`

They are grouped by:

- `Dependencies`

And injected through:

- `NewServiceWithDeps(...)`

#### Resolver ownership

`ScaffoldSettingsResolver`

- resolve shared values like lang, module path, signoff, git mode, and template input values
- merge CLI overrides with replay/default behavior

`NewTargetResolver`

- resolve `project new` target/project/module behavior
- apply force/replay fallback rules

`InitTargetResolver`

- resolve `project init` target/project/module behavior
- apply current-directory and replay fallback rules

### Presenter layer: `internal/presenters/`

The presenter layer owns rendering, not application logic.

#### Output and view specs

Formatting decisions are captured by:

- `OutputSpec`
- `SummaryViewSpec`
- `InspectionViewSpec`

This makes output selection explicit instead of scattering layout decisions across commands.

#### Formatter construction

Formatter construction is handled through:

- `FormatterFactory`
- `TextFormatterRegistry`

Current default implementations:

- `defaultFormatterFactory`
- `defaultTextFormatterRegistry`

This allows text vs TOML format selection, summary table/compact layouts, and inspection compact/default layouts without pushing render branching back into command handlers.

#### Writer families

Text rendering is split into separate writer families:

- summary writers
- inspection writers

That prevents summary and inspection layout variants from collapsing into one branch-heavy formatter.

## Dependency injection seams

The current code intentionally exposes three major injection seams.

### Catalog injection

Use:

- `NewServiceWithDeps(...)`

to replace:

- repo asset registry
- inspect mode policy
- governance policy

This is useful for focused testing and for changing catalog domain behavior without rewriting the facade.

### Create injection

Use:

- `NewServiceWithDeps(...)`

to replace:

- scaffold settings resolution
- new target resolution
- init target resolution

This is useful for focused testing and for experimenting with different request/spec resolution policies without rewriting the command layer.

### Presenter injection

Use:

- `NewPresenterWithFactory(...)`

to replace formatter construction in tests or specialized integrations.

This is useful when verifying custom factories or registries are honored.

## How to extend the system

### Add a new summary filter or sort mode

Preferred path:

1. Extend `SummaryQuery`
2. Update query validation/application logic
3. Update the command spec builder in `cmd/project`
4. Add focused catalog tests first, then command tests

Avoid adding sort/filter rules directly inside Cobra handlers.

### Add a new inspection mode

Preferred path:

1. Extend `InspectMode`
2. Update `ParseInspectMode(...)`
3. Update `InspectModePolicy`
4. Add catalog tests for the new mode

Avoid embedding mode-specific branching directly in presenters.

### Add a new repo-level governance rule

Preferred path:

1. Extend the injected policy or registry object
2. Pass the new behavior through `Dependencies`
3. Add tests proving injected behavior changes results

Avoid reintroducing package-global rule helpers.

### Add a new rendering style

Preferred path:

1. Extend `SummaryViewSpec` or `InspectionViewSpec`
2. Update `TextFormatterRegistry` / factory behavior
3. Add focused presenter tests
4. Update command spec builders if the style is user-selectable

Avoid putting rendering decisions back into the command layer.

### Add a new create-time precedence or target rule

Preferred path:

1. Extend the relevant create resolver
2. Inject the behavior through `Dependencies`
3. Add focused create-level tests for request/spec construction
4. Update command spec builders only if CLI mapping changes

Avoid adding replay/override/target logic directly in Cobra handlers.

## Design constraints to preserve

When modifying this system, preserve these constraints:

1. Commands stay thin
2. Query objects own query semantics
3. Request/spec objects own create-time execution intent
4. Policies, registries, and resolvers own domain rules
5. Presenters own rendering only
6. Specs are explicit
7. Default behavior remains easy to construct
8. Injection seams stay testable

## Testing guidance

Prefer testing at three levels:

### 1. Catalog-level focused tests

Use these to verify:

- query semantics
- governance behavior
- injected dependency behavior

### 2. Presenter-level focused tests

Use these to verify:

- output spec interpretation
- formatter factory behavior
- custom registry/factory injection

### 3. Command-level tests

Use these to verify:

- flag mapping into command spec builders
- CLI contract stability

### 4. Create-level focused tests

Use these to verify:

- replay precedence behavior
- request/spec construction
- resolver injection behavior
- scaffold execution orchestration

## Current status

At the time of writing, the catalog/rendering pipeline already supports:

- typed summary queries
- typed inspection queries
- analyzer/builders/projectors
- injected domain policies
- injected presenter factories
- command-level spec builders
- multiple text rendering styles

The create/new/init pipeline already supports:

- typed new/init requests
- scaffold spec builders
- explicit scaffold specs
- extracted create resolvers
- injected create dependencies
- focused request/spec/DI tests

Update this document whenever responsibilities move between the command, catalog, and presenter layers.
