FROM golang:1.26-alpine

WORKDIR /app

# Install air for hot reload
RUN go install github.com/air-verse/air@latest

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN go build -o server ./cmd/server

EXPOSE 8080

CMD ["./server"]
