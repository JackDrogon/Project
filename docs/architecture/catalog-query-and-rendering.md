# Catalog Query and Rendering Architecture

This document explains the current architecture behind the `project list` and `project inspect` pipeline.

It is intended for maintainers working in these areas:

- `cmd/project/`
- `internal/app/catalog/`
- `internal/presenters/`

The design intentionally separates command parsing, query construction, template analysis, view building, policy evaluation, and output rendering.

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

These builders centralize:

- flag compatibility rules
- query construction
- output spec construction

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

### Query execution layer: `internal/app/catalog/executor.go`

`QueryExecutor` owns the details of executing typed catalog queries.

Current responsibilities:

- run `SummaryQuery`
- run `InspectionQuery`
- coordinate analyzer + projectors

This keeps query execution details out of the facade.

### Analysis layer: `internal/app/catalog/analyzer.go`

`Analyzer` reads template manifests and files and produces a stable intermediate model:

- `Analysis`

`Analysis` is the canonical internal representation used to derive both summary and inspection views.

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

The current code intentionally exposes two main injection seams.

### Catalog injection

Use:

- `NewServiceWithDeps(...)`

to replace:

- repo asset registry
- inspect mode policy
- governance policy

This is useful for focused testing and for changing catalog domain behavior without rewriting the facade.

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

## Design constraints to preserve

When modifying this system, preserve these constraints:

1. Commands stay thin
2. Query objects own query semantics
3. Policies and registries own domain rules
4. Presenters own rendering only
5. Specs are explicit
6. Default behavior remains easy to construct
7. Injection seams stay testable

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

## Current status

At the time of writing, the catalog/rendering pipeline already supports:

- typed summary queries
- typed inspection queries
- analyzer/builders/projectors
- injected domain policies
- injected presenter factories
- command-level spec builders
- multiple text rendering styles

Update this document whenever responsibilities move between the command, catalog, and presenter layers.
