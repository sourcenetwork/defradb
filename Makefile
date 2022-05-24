default:
	go run cli/defradb/main.go

.PHONY: install
install:
	go install ./cli/defradb/

.PHONY: build
build:
	go build -o build/defradb cli/defradb/main.go

.PHONY: multi-build
multi-build:
	echo "Compiling for multiple OS and Platforms"
	GOOS=linux GOARCH=arm go build -o build/defradb-linux-arm cli/defradb/main.go
	GOOS=linux GOARCH=arm64 go build -o build/defradb-linux-arm64 cli/defradb/main.go
	GOOS=freebsd GOARCH=386 go build -o build/defradb-freebsd-386 cli/defradb/main.go

.PHONY: start
start: build
	./build/defradb start

.PHONY: dump
dump: build
	./build/defradb client dump

.PHONY: deps\:lint
deps\:lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin latest

.PHONY: deps\:coverage
deps\:coverage:
	go install github.com/ory/go-acc@latest

.PHONY: deps\:bench
deps\:bench:
	go install golang.org/x/perf/cmd/benchstat@latest

.PHONY: deps\:golines
deps\:golines:
	go install github.com/segmentio/golines@latest

.PHONY: deps\:chglog
deps\:chglog:
	go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest

.PHONY: deps\:modules
deps\:modules:
	go mod download

.PHONY: deps\:ci
deps\:ci:
	curl -fLSs https://raw.githubusercontent.com/CircleCI-Public/circleci-cli/master/install.sh | DESTDIR=${HOME}/bin bash

.PHONY: deps
deps: deps\:lint deps\:coverage deps\:bench deps\:golines deps\:chglog deps\:modules deps\:ci

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	go clean cli/defradb/main.go
	rm -f build/defradb

.PHONY: clean\:test
clean\:test:
	go clean -testcache

.PHONY: test
test:
	go test ./... -race

.PHONY: test\:clean
test\:clean: clean\:test test

.PHONY: test\:bench
test\:bench:
	make -C ./tests/bench/ bench

.PHONY: test\:bench-short
test\:bench-short:
	make -C ./tests/bench/ bench:short

# This also takes integration tests into account.
.PHONY: test\:coverage-full
test\:coverage-full: deps\:coverage
	go-acc ./... --output=coverage-full.txt --covermode=atomic
	go tool cover -func coverage-full.txt | grep total | awk '{print $$3}'

# Usage: make test:coverage-html path="{pathToPackage}"
.PHONY: test\:coverage-html
test\:coverage-html:
	go test -v $(path) -race -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm ./coverage.out

# This only covers how much of the package is tested by itself (unit test).
.PHONY: test\:coverage-quick
test\:coverage-quick:
	go test ./... -race -coverprofile=coverage-quick.txt -covermode=atomic
	go tool cover -func coverage-quick.txt | grep total | awk '{print $$3}'

.PHONY: test\:changes
test\:changes:
	env DEFRA_DETECT_DATABASE_CHANGES=true go test ./... -p 1

.PHONY: validate\:codecov
validate\:codecov:
	curl --data-binary @.github/codecov.yml https://codecov.io/validate

.PHONY: validate\:circleci
validate\:circleci:
	circleci config validate

.PHONY: lint
lint:
	golangci-lint run --config tools/configs/golangci.yaml

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
