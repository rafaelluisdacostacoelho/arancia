# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0: Build a statically linked binary
# -ldflags="-w -s": Strip debug information to reduce size
# -a: Force rebuilding of packages
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o todo-service \
    ./cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS and timezone data
RUN apk --no-cache add ca-certificates tzdata && \
    # Create non-root user
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    # Create data directory for BoltDB
    mkdir -p /data && \
    chown -R appuser:appuser /data

WORKDIR /home/appuser

# Copy the binary from builder
COPY --from=builder /app/todo-service .

# Change ownership of the binary
RUN chown appuser:appuser todo-service

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

# Run the binary
CMD ["./todo-service"]
