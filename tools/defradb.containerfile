# syntax=docker/dockerfile:1

# An image to run defradb.

# Stage: BUILD
# Several steps are involved to enable caching and because of the behavior of COPY regarding directories.
FROM docker.io/golang:1.20 AS BUILD
WORKDIR /repo/
COPY go.mod go.sum Makefile ./
RUN make deps:modules
COPY . .
# manually copy libwasmer.so to fix linking issue https://github.com/wasmerio/wasmer-go/issues/281
# NOTE: when building on arm64 change linux-amd64 to linux-aarch64
RUN cp /go/pkg/mod/github.com/wasmerio/wasmer-go@v1.0.4/wasmer/packaged/lib/linux-amd64/libwasmer.so /lib/libwasmer.so
RUN make build

# Stage: RUN
FROM debian:bookworm-slim
COPY --from=BUILD /repo/build/defradb /defradb
COPY --from=BUILD /lib/libwasmer.so /lib/libwasmer.so

# Documents which ports are normally used.
# To publish the ports: `docker run -p 9181:9181` ...
EXPOSE 9161
EXPOSE 9171
EXPOSE 9181

# Default command provided for convenience.
# e.g. docker run -p 9181:9181 source/defradb  start --url 0.0.0.0:9181
ENTRYPOINT [ "/defradb" ]
