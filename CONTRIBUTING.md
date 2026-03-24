# Contributing to project

Thanks for your interest in improving `project`.

## Getting Started

1. Fork the repository.
2. Clone your fork.
3. Install the project prerequisites listed in `README.md`.
4. Run the local quality checks before opening a pull request:

```bash
just pre-commit
```

If you want to run spelling checks locally, install [`typos`](https://github.com/crate-ci/typos) first.

## Development Workflow

- Keep changes focused and easy to review.
- Add or update tests when behavior changes.
- Update docs when command behavior, template output, or developer workflows change.
- If you touch template files under `internal/adapters/templatesrc/`, run `just generate` so embedded permission metadata stays in sync.

## Pull Requests

Before opening a pull request:

1. Make sure your branch is up to date with the default branch.
2. Run `just pre-commit` successfully.
3. Explain the user-facing impact and why the change is needed.
4. Include tests or a clear explanation for why tests are not needed.

## Code Style

- Follow standard Go conventions and `gofmt` output.
- Prefer small, explicit functions over clever abstractions.
- Keep CLI wiring in `cmd/project` and reusable logic in `internal/` packages.
- Add comments only when they explain non-obvious intent.

## Reporting Bugs

When reporting a bug, include:

- What you expected to happen
- What actually happened
- Steps to reproduce the issue
- Your Go version and operating system, if relevant
