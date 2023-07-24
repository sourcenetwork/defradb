# syntax=docker/dockerfile:1

# An image to run defradb.

# Stage: BUILD_WASMER
# Build wasmer to fix linking issue https://github.com/wasmerio/wasmer-go/issues/281
FROM docker.io/rust:1.71 AS BUILD_WASMER
WORKDIR /wasmer/
RUN git clone --depth 1 --branch v4.0.0 https://github.com/wasmerio/wasmer.git .
RUN make build-capi

# Stage: BUILD
# Several steps are involved to enable caching and because of the behavior of COPY regarding directories.
FROM docker.io/golang:1.20 AS BUILD
WORKDIR /repo/
COPY go.mod go.sum Makefile ./
RUN make deps:modules
COPY . .
COPY --from=BUILD_WASMER /wasmer/target/release/libwasmer.so /lib/libwasmer.so
RUN make build

# Stage: RUN
FROM debian:bookworm-slim
COPY --from=BUILD /repo/build/defradb /defradb
COPY --from=BUILD_WASMER /wasmer/target/release/libwasmer.so /lib/libwasmer.so

# Documents which ports are normally used.
# To publish the ports: `docker run -p 9181:9181` ...
EXPOSE 9161
EXPOSE 9171
EXPOSE 9181

# Default command provided for convenience.
# e.g. docker run -p 9181:9181 source/defradb  start --url 0.0.0.0:9181
ENTRYPOINT [ "/defradb" ]
