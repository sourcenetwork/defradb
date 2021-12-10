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

.PHONY: deps
deps:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0

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

.PHONY: lint
lint:
	golangci-lint run --config .golangci.sourceinc.yaml

.PHONY: lint\:todo
lint\:todo:
	rg "nolint" -g '!{Makefile}'

.PHONY: lint\:list
lint\:list:
	golangci-lint linters --config .golangci.sourceinc.yaml
