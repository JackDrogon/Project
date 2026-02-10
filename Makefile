# set makefile echo back
ifdef VERBOSE
	V :=
else
	V := @
endif

tag := $(shell git describe --abbrev=0 --always --dirty --tags)
sha := $(shell git rev-parse --short HEAD)
git_tag_sha := $(tag):$(sha)
LDFLAGS="-X 'github.com/JackDrogon/project/pkg/version.GitTagSha=$(git_tag_sha)'"
GOFLAGS=

# COVERAGE=ON to enable coverage
ifeq ($(COVERAGE),ON)
    GOFLAGS += -cover
endif

.PHONY: default
## default: Build project
default: project

.PHONY: build
## build : Build binaries
build: project

.PHONY: bin
## bin : Create bin directory
bin:
	$(V)mkdir -p bin

.PHONY: lint
## lint : Lint codespace
# TODO(Drogon): add golang lint
lint:
	$(V)golangci-lint run

.PHONY: fmt
## fmt : Format all code
fmt:
	$(V)go fmt ./...

.PHONY: test
## test : Run test
test:
	$(V)go test $(shell go list ./...) | grep -F -v '[no test files]' || true

.PHONY: cloc
## cloc : Count lines of code
cloc:
	$(V)tokei -C .

.PHONY: todos
## todos : Print all todos
todos:
	$(V)grep -rnw . -e "TODO" | grep -v '^./pkg/rpc/thrift' | grep -v '^./.git'

.PHONY: help
## help : Print help message
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = ":"} {printf "\033[36m%-23s\033[0m %s\n", $$1, $$2}'



# --------------- ------------------ ---------------
# --------------- User Defined Tasks ---------------

.PHONY: project
## project : Build project
project: bin
	$(V)go build -ldflags $(LDFLAGS) -o bin/project ./cmd/project