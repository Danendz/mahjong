FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

# Dev stage: full Go toolchain + air for hot reload.
# Source is mounted via docker-compose volume; air rebuilds on file change.
# Build with `docker build --target dev .` (or via docker-compose, which
# targets it explicitly). Don't ship this image to production — it has no
# source baked in and expects a volume mount.
FROM golang:1.26-alpine AS dev
RUN apk add --no-cache git
RUN go install github.com/air-verse/air@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# Production stage (default target). A plain `docker build .` ends here.
FROM alpine:latest AS prod
WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
