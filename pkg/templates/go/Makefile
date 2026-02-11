# set makefile echo back
ifdef VERBOSE
	V :=
else
	V := @
endif

tag := $(shell git describe --abbrev=0 --always --dirty --tags)
sha := $(shell git rev-parse --short HEAD)
git_tag_sha := $(tag):$(sha)
LDFLAGS="-X '{{.ModulePath}}/pkg/version.GitTagSha=$(git_tag_sha)'"
GOFLAGS=

# COVERAGE=ON to enable coverage
ifeq ($(COVERAGE),ON)
    GOFLAGS += -cover
endif

.PHONY: default
## default: Build project
default: build

.PHONY: build
## build : Build binaries
build: bin
	$(V)go build -ldflags $(LDFLAGS) $(GOFLAGS) -o bin/{{.ProjectName}} .

.PHONY: bin
## bin : Create bin directory
bin:
	$(V)mkdir -p bin

.PHONY: lint
## lint : Lint codespace
lint:
	$(V)golangci-lint run

.PHONY: fmt
## fmt : Format all code
fmt:
	$(V)go fmt ./...

.PHONY: test
## test : Run test
test:
	$(V)go test ./...

.PHONY: cloc
## cloc : Count lines of code
cloc:
	$(V)tokei -C .

.PHONY: todos
## todos : Print all todos
todos:
	$(V)grep -rnw . -e "TODO" | grep -v '^./.git'

.PHONY: help
## help : Print help message
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = ":"} {printf "\033[36m%-23s\033[0m %s\n", $$1, $$2}'
