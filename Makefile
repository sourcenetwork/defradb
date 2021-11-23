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

.PHONY: clean
clean:
	go clean cli/defradb/main.go
	rm -f build/defradb

.PHONY: update
update:
	go mod tidy

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: deps
deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0
