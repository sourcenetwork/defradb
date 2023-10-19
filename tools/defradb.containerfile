# syntax=docker/dockerfile:1

# An image to run defradb.

# Stage: PLAYGROUND
FROM docker.io/node:20 AS PLAYGROUND_BUILD
WORKDIR /repo/
COPY playground/ ./
RUN npm install
RUN npm run build

# Stage: BUILD
# Several steps are involved to enable caching and because of the behavior of COPY regarding directories.
FROM docker.io/golang:1.20 AS BUILD
WORKDIR /repo/
COPY go.mod go.sum Makefile ./
RUN make deps:modules
COPY . .
COPY --from=PLAYGROUND_BUILD /repo/dist /repo/playground/dist/
ENV BUILD_TAGS=playground
RUN make build

# Stage: RUN
FROM debian:bookworm-slim
COPY --from=BUILD /repo/build/defradb /defradb

# Documents which ports are normally used.
# To publish the ports: `docker run -p 9181:9181` ...
EXPOSE 9161
EXPOSE 9171
EXPOSE 9181

# Default command provided for convenience.
# e.g. docker run -p 9181:9181 source/defradb  start --url 0.0.0.0:9181
ENTRYPOINT [ "/defradb" ]
