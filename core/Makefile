.PHONY: help build run clean test deps stack-up stack-down stack-status

# Default target
help:
	@echo "Infra Stack - Development Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make run          - Build and run the stack (go run . stack up)"
	@echo "  make build        - Build the infra binary"
	@echo "  make stack-up     - Start the service stack"
	@echo "  make stack-down   - Stop the service stack"
	@echo "  make stack-status - Check stack status"
	@echo "  make clean        - Remove built binaries"
	@echo "  make test         - Run tests"
	@echo "  make deps         - Download dependencies"

# Build the main infra binary
build:
	@echo "Building infra..."
	@go build -o infra .

# Run the stack directly (compile on the fly)
run:
	@go run . stack up

# Stack management commands
stack-up:
	@go run . stack up

stack-down:
	@go run . stack down

stack-status:
	@go run . stack status

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f infra
	@rm -rf .core-stack
	@rm -rf .dep

# Run tests
test:
	@go test -v ./...

# Download dependencies
deps:
	@go mod download
	@go mod tidy
