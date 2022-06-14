# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.17 AS BUILD

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /

COPY --from=build /app/build/defradb /defradb

EXPOSE 9161
EXPOSE 9171
EXPOSE 9181

CMD ["/defradb", "start"]
