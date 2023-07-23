# syntax=docker/dockerfile:1

# An image with defradb's code and go tooling available, to assemble in a larger container.

FROM docker.io/golang:1.20 AS BUILD

WORKDIR /lib/defradb/

COPY . .
RUN make deps:modules
