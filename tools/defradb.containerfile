# syntax=docker/dockerfile:1

# An image to run defradb.

# Stage: build
# Several steps are involved to enable caching and because of the behavior of COPY regarding directories.
# WARNING: The build image SHOULD use the same OS version as the run image. Otherwise they may have
# different versions of GLIBC and the dynamic linking of libraries will fail.
# `bookworm` is Debian 12.
FROM docker.io/golang:1.25-bookworm AS build
WORKDIR /repo/
COPY go.mod go.sum Makefile ./
RUN make deps:modules
COPY . .
RUN make deps:playground
ENV BUILD_TAGS=playground,telemetry
RUN make build

# Stage: RUN
FROM debian:bookworm-slim
COPY --from=build /repo/build/defradb /defradb

# Documents which ports are normally used.
# To publish the ports: `docker run -p 9181:9181` ...
EXPOSE 9161
EXPOSE 9171
EXPOSE 9181

# Default command provided for convenience.
# e.g. docker run -p 9181:9181 source/defradb  start --url 0.0.0.0:9181
ENTRYPOINT [ "/defradb" ]
