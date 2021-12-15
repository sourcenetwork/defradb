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

.PHONY: deps\:circle-ci
deps\:circle-ci:
	go mod download
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin v1.43.0

.PHONY: deps
deps: deps\:circle-ci

.PHONY: clean
clean:
	go clean cli/defradb/main.go
	rm -f build/defradb

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test ./...

.PHONY: test\:bench
test\:bench:
	go test -bench

.PHONY: test\:coverage
test\:coverage:
	go test ./... -coverprofile cover.out
	go tool cover -func cover.out | grep total | awk '{print $$3}'

.PHONY: lint
lint:
	golangci-lint run --config .golangci.sourceinc.yaml

.PHONY: lint\:todo
lint\:todo:
	rg "nolint" -g '!{Makefile}'

.PHONY: lint\:list
lint\:list:
	golangci-lint linters --config .golangci.sourceinc.yaml
