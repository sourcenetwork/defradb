# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.18 AS BUILD

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /defradb cmd/defradb/main.go

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /

COPY --from=build /defradb /defradb

EXPOSE 9161
EXPOSE 9171
EXPOSE 9181

CMD ["/defradb", "start"]
