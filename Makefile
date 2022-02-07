default:
	go run cli/defradb/main.go

.PHONY: install
install:
	go install cli/defradb/main.go

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

.PHONY: deps\:golangci-lint
deps\:golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin v1.43.0

.PHONY: deps\:go-acc
deps\:go-acc:
	go install github.com/ory/go-acc@latest

.PHONY: deps
deps: deps\:golangci-lint deps\:go-acc
	go mod download

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
	go test ./... -race -bench=.

# This also takes integration tests into account.
.PHONY: test\:coverage-full
test\:coverage-full: deps\:go-acc
	go-acc ./... --output=coverage-full.txt --covermode=atomic
	go tool cover -func coverage-full.txt | grep total | awk '{print $$3}'

# This only covers how much of the package is tested by itself (unit test).
.PHONY: test\:coverage-quick
test\:coverage-quick:
	go test ./... -race -coverprofile=coverage-quick.txt -covermode=atomic
	go tool cover -func coverage-quick.txt | grep total | awk '{print $$3}'

.PHONY: validate\:codecov
validate\:codecov:
	curl --data-binary @codecov.yml https://codecov.io/validate

.PHONY: lint
lint:
	golangci-lint run --config .golangci.sourceinc.yaml

.PHONY: lint\:todo
lint\:todo:
	rg "nolint" -g '!{Makefile}'

.PHONY: lint\:list
lint\:list:
	golangci-lint linters --config .golangci.sourceinc.yaml
