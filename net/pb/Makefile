PB = $(wildcard *.proto)
GO = $(PB:.proto=.pb.go)

all: $(GO)

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