PB = $(wildcard *.proto)
GO = $(PB:.proto=.pb.go)

all: $(GO)

%.pb.go: %.proto
	protoc -I=. --gogofaster_out=\
	plugins=grpc:\
	. $<

clean:
	rm -f *.pb.go
	rm -f *pb_test.go

.PHONY: clean