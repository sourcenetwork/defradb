# Make DefraDB!
# For compatibility, prerequisites are instead explicit calls to make.

ifndef VERBOSE
MAKEFLAGS+=--no-print-directory
endif

# Detect OS (`Linux`, `Darwin`, `Windows`)
# Note: can use `lsb_release --id --short` for more specfic linux distro information.
OS_GENERAL := Unknown
ifeq ($(OS),Windows_NT)
	OS_GENERAL := Windows
else
	OS_GENERAL := $(shell sh -c 'uname 2>/dev/null || echo Unknown')
endif

# Detect OS specfic package manager if possible (`apt`, `yum`, `pacman`, `brew`, `choco`)
OS_PACKAGE_MANAGER := Unknown
ifeq ($(OS_GENERAL),Linux)
	ifneq ($(shell which apt 2>/dev/null),)
		OS_PACKAGE_MANAGER := apt
	else ifneq ($(shell which yum 2>/dev/null),)
		OS_PACKAGE_MANAGER := yum
	else ifneq ($(shell which pacman 2>/dev/null),)
		OS_PACKAGE_MANAGER := pacman
	else ifneq ($(shell which dnf 2>/dev/null),)
		OS_PACKAGE_MANAGER := dnf
	endif
else ifeq ($(OS_GENERAL),Darwin)
	OS_PACKAGE_MANAGER := brew
else ifeq ($(OS_GENERAL),Windows)
	OS_PACKAGE_MANAGER := choco
endif

# Provide info from git to the version package using linker flags.
ifeq (, $(shell which git))
$(error "No git in $(PATH), version information won't be included")
else
VERSION_GOINFO=$(shell go version)
VERSION_GITCOMMIT=$(shell git rev-parse HEAD)
VERSION_GITCOMMITDATE=$(shell git show -s --format=%cs HEAD)
ifneq ($(shell git symbolic-ref -q --short HEAD),master)
VERSION_GITRELEASE=dev-$(shell git symbolic-ref -q --short HEAD)
else
VERSION_GITRELEASE=$(shell git describe --tags)
endif

$(info ----------------------------------------);
$(info OS = $(OS_GENERAL));
$(info PACKAGE_MANAGER = $(OS_PACKAGE_MANAGER));
$(info GOINFO = $(VERSION_GOINFO));
$(info GITCOMMIT = $(VERSION_GITCOMMIT));
$(info GITCOMMITDATE = $(VERSION_GITCOMMITDATE));
$(info GITRELEASE = $(VERSION_GITRELEASE));
$(info ----------------------------------------);

BUILD_FLAGS=-trimpath -ldflags "\
-X 'github.com/sourcenetwork/defradb/version.GoInfo=$(VERSION_GOINFO)'\
-X 'github.com/sourcenetwork/defradb/version.GitRelease=$(VERSION_GITRELEASE)'\
-X 'github.com/sourcenetwork/defradb/version.GitCommit=$(VERSION_GITCOMMIT)'\
-X 'github.com/sourcenetwork/defradb/version.GitCommitDate=$(VERSION_GITCOMMITDATE)'"
endif

ifdef BUILD_TAGS
BUILD_FLAGS+=-tags $(BUILD_TAGS)
endif

TEST_FLAGS=-race -shuffle=on -timeout 5m

COVERAGE_DIRECTORY=$(PWD)/coverage
COVERAGE_FILE=coverage.txt
COVERAGE_FLAGS=-covermode=atomic -coverpkg=./... -args -test.gocoverdir=$(COVERAGE_DIRECTORY)

PLAYGROUND_DIRECTORY=playground
CHANGE_DETECTOR_TEST_DIRECTORY=tests/change_detector
DEFAULT_TEST_DIRECTORIES=./...

default:
	@go run $(BUILD_FLAGS) cmd/defradb/main.go

.PHONY: install
install:
	@go install $(BUILD_FLAGS) ./cmd/defradb

.PHONY: install\:manpages
install\:manpages:
ifeq ($(OS_GENERAL),Linux)
	cp build/man/* /usr/share/man/man1/
endif
ifneq ($(OS_GENERAL),Linux)
	@echo "Direct installation of Defradb's man pages is not supported on your system."
endif

# Usage:
# 	- make build
# 	- make build path="path/to/defradb-binary"
.PHONY: build
build:
ifeq ($(path),)
	@go build $(BUILD_FLAGS) -o build/defradb cmd/defradb/main.go
else
	@go build $(BUILD_FLAGS) -o $(path) cmd/defradb/main.go
endif

# Usage: make cross-build platforms="{platforms}"
# platforms is specified as a comma-separated list with no whitespace, e.g. "linux/amd64,linux/arm,linux/arm64"
# If none is specified, build for all platforms.
.PHONY: cross-build
cross-build:
	bash tools/scripts/cross-build.sh $(platforms)

.PHONY: start
start:
	@$(MAKE) build
	./build/defradb start

.PHONY: client\:dump
client\:dump:
	./build/defradb client dump

.PHONY: client\:add-schema
client\:add-schema:
	./build/defradb client schema add -f examples/schema/bookauthpub.graphql

.PHONY: deps\:lint-go
deps\:lint-go:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54

.PHONY: deps\:lint-yaml
deps\:lint-yaml:
ifeq (, $(shell which yamllint))
	$(info YAML linter 'yamllint' not found on the system, please install it.)
	$(info Can try using your local package manager: $(OS_PACKAGE_MANAGER))
else
	$(info YAML linter 'yamllint' already installed.)
endif

.PHONY: deps\:lint
deps\:lint:
	@$(MAKE) deps:lint-go && \
	$(MAKE) deps:lint-yaml

.PHONY: deps\:test
deps\:test:
	go install gotest.tools/gotestsum@latest
	rustup target add wasm32-unknown-unknown
	@$(MAKE) -C ./tests/lenses build

.PHONY: deps\:bench
deps\:bench:
	go install golang.org/x/perf/cmd/benchstat@latest

.PHONY: deps\:chglog
deps\:chglog:
	go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest

.PHONY: deps\:modules
deps\:modules:
	go mod download

.PHONY: deps\:mock
deps\:mock:
	go install github.com/vektra/mockery/v2@v2.32.0

.PHONY: deps\:playground
deps\:playground:
	cd $(PLAYGROUND_DIRECTORY) && npm install && npm run build

.PHONY: deps
deps:
	@$(MAKE) deps:modules && \
	$(MAKE) deps:bench && \
	$(MAKE) deps:chglog && \
	$(MAKE) deps:lint && \
	$(MAKE) deps:test && \
	$(MAKE) deps:mock

.PHONY: mock
mock:
	@$(MAKE) deps:mock
	mockery --config="tools/configs/mockery.yaml"

.PHONY: dev\:start
dev\:start:
	@$(MAKE) build
	DEFRA_ENV=dev ./build/defradb start

# Note: In some situations `verify` can modify `go.sum` file, but until a
#       read-only version is available we have to rely on this.
# Here are some relevant issues:
#   - https://github.com/golang/go/issues/31372
#   - https://github.com/cosmos/cosmos-sdk/issues/4165
.PHONY: verify
verify:
	@if go mod verify | grep -q 'all modules verified'; then \
		echo "Success!";                                     \
	else                                                     \
		echo "Failure:";                                     \
		go mod verify;                                       \
		exit 2;                                              \
	fi;

.PHONY: tidy
tidy:
	go mod tidy -go=1.21.3

.PHONY: clean
clean:
	go clean cmd/defradb/main.go
	rm -f build/defradb

.PHONY: clean\:test
clean\:test:
	go clean -testcache

.PHONY: clean\:coverage
clean\:coverage:
	rm -rf $(COVERAGE_DIRECTORY) 
	rm -f $(COVERAGE_FILE)

# Example: `make tls-certs path="~/.defradb/certs"`
.PHONY: tls-certs
tls-certs:
ifeq ($(path),)
	openssl ecparam -genkey -name secp384r1 -out server.key
	openssl req -new -x509 -sha256 -key server.key -out server.crt -days 365
else
	mkdir -p $(path)
	openssl ecparam -genkey -name secp384r1 -out $(path)/server.key
	openssl req -new -x509 -sha256 -key $(path)/server.key -out $(path)/server.crt -days 365
endif

.PHONY: test
test:
	gotestsum --format pkgname -- $(DEFAULT_TEST_DIRECTORIES) $(TEST_FLAGS)

.PHONY: test\:quick
test\:quick:
	gotestsum --format pkgname -- $(DEFAULT_TEST_DIRECTORIES)

# Only build the tests (don't execute them).
.PHONY: test\:build
test\:build:
	gotestsum --format pkgname -- $(DEFAULT_TEST_DIRECTORIES) $(TEST_FLAGS) -run=nope

.PHONY: test\:gql-mutations
test\:gql-mutations:
	DEFRA_MUTATION_TYPE=gql DEFRA_BADGER_MEMORY=true gotestsum --format pkgname -- $(DEFAULT_TEST_DIRECTORIES)

# This action and the test:col-named-mutations (below) runs the test suite with any supporting mutation test
# actions running their mutations via their corresponding named [Collection] call.
#
# For example, CreateDoc will call [Collection.Create], and
# UpdateDoc will call [Collection.Update].
.PHONY: test\:col-named-mutations
test\:col-named-mutations:
	DEFRA_MUTATION_TYPE=collection-named DEFRA_BADGER_MEMORY=true gotestsum --format pkgname -- $(DEFAULT_TEST_DIRECTORIES)

.PHONY: test\:go
test\:go:
	go test $(DEFAULT_TEST_DIRECTORIES) $(TEST_FLAGS)

.PHONY: test\:http
test\:http:
	DEFRA_CLIENT_HTTP=true go test $(DEFAULT_TEST_DIRECTORIES) $(TEST_FLAGS)

.PHONY: test\:cli
test\:cli:
	DEFRA_CLIENT_CLI=true go test $(DEFAULT_TEST_DIRECTORIES) $(TEST_FLAGS)

.PHONY: test\:names
test\:names:
	gotestsum --format testname -- $(DEFAULT_TEST_DIRECTORIES) $(TEST_FLAGS)

.PHONY: test\:verbose
test\:verbose:
	gotestsum --format standard-verbose -- $(DEFAULT_TEST_DIRECTORIES) $(TEST_FLAGS)

.PHONY: test\:watch
test\:watch:
	gotestsum --watch -- $(DEFAULT_TEST_DIRECTORIES)

.PHONY: test\:clean
test\:clean:
	@$(MAKE) clean:test && $(MAKE) test

.PHONY: test\:bench
test\:bench:
	@$(MAKE) -C ./tests/bench/ bench

.PHONY: test\:bench-short
test\:bench-short:
	@$(MAKE) -C ./tests/bench/ bench:short

.PHONY: test\:scripts
test\:scripts:
	@$(MAKE) -C ./tools/scripts/ test

.PHONY: test\:coverage
test\:coverage:
	@$(MAKE) clean:coverage
	mkdir $(COVERAGE_DIRECTORY)
ifeq ($(path),)
	gotestsum --format testname -- ./... $(TEST_FLAGS) $(COVERAGE_FLAGS)
else
	gotestsum --format testname -- $(path) $(TEST_FLAGS) $(COVERAGE_FLAGS)
endif
	go tool covdata textfmt -i=$(COVERAGE_DIRECTORY) -o $(COVERAGE_FILE)

.PHONY: test\:coverage-func
test\:coverage-func:
	@$(MAKE) test:coverage
	go tool cover -func=$(COVERAGE_FILE)

.PHONY: test\:coverage-html
test\:coverage-html:
	@$(MAKE) test:coverage path=$(path)
	go tool cover -html=$(COVERAGE_FILE)
	@$(MAKE) clean:coverage
	

.PHONY: test\:changes
test\:changes:
	gotestsum --format testname -- ./$(CHANGE_DETECTOR_TEST_DIRECTORY)/... -timeout 15m --tags change_detector

.PHONY: validate\:codecov
validate\:codecov:
	curl --data-binary @.github/codecov.yml https://codecov.io/validate

.PHONY: validate\:circleci
validate\:circleci:
	circleci config validate

.PHONY: lint
lint:
	golangci-lint run --config tools/configs/golangci.yaml
	yamllint -c tools/configs/yamllint.yaml .

.PHONY: lint\:fix
lint\:fix:
	golangci-lint run --config tools/configs/golangci.yaml --fix

.PHONY: lint\:todo
lint\:todo:
	rg "nolint" -g '!{Makefile}'

.PHONY: lint\:list
lint\:list:
	golangci-lint linters --config tools/configs/golangci.yaml

.PHONY: chglog
chglog:
	git-chglog -c "tools/configs/chglog/config.yml" --next-tag v0.x.0 -o CHANGELOG.md

.PHONY: docs
docs:
	@$(MAKE) docs\:cli
	@$(MAKE) docs\:manpages

.PHONY: docs\:cli
docs\:cli:
	rm -f docs/website/references/cli/*.md
	go run cmd/genclidocs/main.go -o docs/website/references/cli

.PHONY: docs\:manpages
docs\:manpages:
	go run cmd/genmanpages/main.go -o build/man/

.PHONY: docs\:godoc
docs\:godoc:
	godoc -http=:6060
	# open http://localhost:6060/pkg/github.com/sourcenetwork/defradb/
