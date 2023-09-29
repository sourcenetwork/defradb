PB = $(wildcard *.proto)
GO = $(PB:.proto=.pb.go)

all: $(GO)

install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto@latest

%.pb.go: %.proto
	protoc \
	--go_out=. --plugin protoc-gen-go="${GOBIN}/protoc-gen-go" \
    --go-grpc_out=. --plugin protoc-gen-go-grpc="${GOBIN}/protoc-gen-go-grpc" \
    --go-vtproto_out=. --plugin protoc-gen-go-vtproto="${GOBIN}/protoc-gen-go-vtproto" \
    --go-vtproto_opt=features=marshal+unmarshal+size \
	$<

clean:
	rm -f *.pb.go
	rm -f *pb_test.go

.PHONY: clean