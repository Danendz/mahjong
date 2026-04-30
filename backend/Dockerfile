FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest AS prod
WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]

# Dev stage: full Go toolchain + air for hot reload.
# Source is mounted via docker-compose volume; air rebuilds on file change.
FROM golang:1.26-alpine AS dev
RUN apk add --no-cache git
RUN go install github.com/air-verse/air@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]
