PB = $(wildcard *.proto)
GO = $(PB:.proto=.pb.go)

PROTOC_GEN_GO := $(shell which protoc-gen-go)
PROTOC_GEN_GO_GRPC := $(shell which protoc-gen-go-grpc)
PROTOC_GEN_GO_VTPROTO := $(shell which protoc-gen-go-vtproto)

all: $(GO)

deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto@latest

%.pb.go: %.proto
	protoc \
	--go_out=. --plugin protoc-gen-go="$(PROTOC_GEN_GO)" \
    --go-grpc_out=. --plugin protoc-gen-go-grpc="$(PROTOC_GEN_GO_GRPC)" \
    --go-vtproto_out=. --plugin protoc-gen-go-vtproto="$(PROTOC_GEN_GO_VTPROTO)" \
    --go-vtproto_opt=features=marshal+unmarshal+size \
	$< # This line specifies the input file

clean:
	rm -f *.pb.go
	rm -f *pb_test.go

.PHONY: clean deps debug