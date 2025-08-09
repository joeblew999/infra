# Dockerfile for Fly.io deployment
# This uses a multi-stage build for optimal image size

# Build stage
FROM golang:1.24-alpine AS builder

# Install git for version info
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimization flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -extldflags=-static" \
    -trimpath \
    -o infra .

# Production stage - using alpine for shell compatibility
FROM alpine:latest

# Install ca-certificates and wget for health checks
RUN apk add --no-cache ca-certificates wget

# Create directories and user
RUN addgroup -g 1000 app && adduser -D -s /bin/sh -u 1000 -G app app

# Create necessary directories
RUN mkdir -p /app/.data /app/.dep /app/.bin && chown -R app:app /app

# Copy binary from builder stage
COPY --from=builder /app/infra /app/infra

# Copy static files
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/.ko.yaml /app/.ko.yaml

# Set working directory
WORKDIR /app

# Set ownership
RUN chown -R app:app /app

# Switch to non-root user
USER app:app

# Expose port
EXPOSE 1337

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:1337/status || exit 1

# Run the binary
CMD ["./infra", "service"]