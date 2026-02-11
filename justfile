#!/usr/bin/env -S just --justfile

# Justfile for project — CLI scaffolding tool
# Run `just` or `just --list` to see available recipes

tag := `git describe --abbrev=0 --always --dirty --tags`
sha := `git rev-parse --short HEAD`
git_tag_sha := tag + ":" + sha
ldflags := "-X 'github.com/JackDrogon/project/pkg/version.GitTagSha=" + git_tag_sha + "'"

# ─────────────────────────────────────────────────────────────────────
# Aliases (shortcuts for frequent tasks)
# ─────────────────────────────────────────────────────────────────────

alias b := build
alias t := test
alias l := lint
alias f := fmt
alias c := cover
alias r := run

# ─────────────────────────────────────────────────────────────────────
# Help
# ─────────────────────────────────────────────────────────────────────

# Show all available recipes
[private]
default:
    @just --list --unsorted

# ═════════════════════════════════════════════════════════════════════
#  Build
# ═════════════════════════════════════════════════════════════════════

# Build binary to bin/project
[group('build')]
build:
    @mkdir -p bin
    go build -ldflags "{{ldflags}}" -o bin/project ./cmd/project

# Build with coverage instrumentation
[group('build')]
build-cover:
    @mkdir -p bin
    go build -ldflags "{{ldflags}}" -cover -o bin/project ./cmd/project

# Remove build artifacts
[group('build')]
clean:
    rm -rf bin

# ═════════════════════════════════════════════════════════════════════
#  Code Quality
# ═════════════════════════════════════════════════════════════════════

# Run golangci-lint
[group('quality')]
lint:
    golangci-lint run

# Format all Go code
[group('quality')]
fmt:
    go fmt ./...

# Run go vet
[group('quality')]
vet:
    go vet ./...

# Pre-commit: format → lint → test — run before every commit
[group('quality')]
pre-commit: fmt lint test
    @echo "All checks passed."

# ═════════════════════════════════════════════════════════════════════
#  Test
# ═════════════════════════════════════════════════════════════════════

# Run tests (e.g., just test, just test ./cmd/project/...)
[group('test')]
test pkg='./...':
    go test {{pkg}}

# Run tests with verbose output
[group('test')]
test-v pkg='./...':
    go test -v {{pkg}}

# Run tests with coverage report
[group('test')]
cover:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report: coverage.html"

# ═════════════════════════════════════════════════════════════════════
#  Run
# ═════════════════════════════════════════════════════════════════════

# Build and run project (e.g., just run new -l go myapp)
[group('run')]
run *args: build
    ./bin/project {{args}}

# ═════════════════════════════════════════════════════════════════════
#  Dependencies
# ═════════════════════════════════════════════════════════════════════

# Run go mod tidy
[group('deps')]
tidy:
    go mod tidy

# Show dependency graph
[group('deps')]
deps:
    go mod graph

# ═════════════════════════════════════════════════════════════════════
#  Maintenance & Info
# ═════════════════════════════════════════════════════════════════════

# Count lines of code (requires tokei)
[group('maintenance')]
loc:
    tokei --sort code

# Print all TODOs in codebase
[group('maintenance')]
todos:
    grep -rnw . -e "TODO" | grep -v '^./.git'

# Show concise git log
[group('maintenance')]
log n='20':
    git log --oneline --graph --decorate -n {{n}}

# Show Go toolchain info
[group('maintenance')]
info:
    @echo "── Go Toolchain ──"
    go version
    @echo ""
    @echo "── Module ──"
    head -1 go.mod
