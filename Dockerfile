# Dockerfile for Fly.io deployment
# This uses a multi-stage build for optimal image size

# Build stage
FROM golang:1.22-alpine AS builder

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

# Production stage - using minimal base image
FROM cgr.dev/chainguard/static:latest

# Add ca-certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Create necessary directories
RUN mkdir -p /app/.data /app/.dep /app/.bin

# Copy binary from builder stage
COPY --from=builder /app/infra /app/infra

# Copy static files
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/.ko.yaml /app/.ko.yaml

# Set working directory
WORKDIR /app

# Create non-root user for security
USER 1000:1000

# Expose port
EXPOSE 1337

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:1337/status || exit 1

# Run the binary
CMD ["./infra", "--mode=service"]