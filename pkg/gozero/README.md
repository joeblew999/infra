# pkg/gozero

Go-zero wrapper for the infra project, providing consistent patterns for go-zero code generation.

## Overview

This package wraps the `goctl` command-line tool to provide high-level functions for generating go-zero API services that follow infra patterns. It ensures generated code conforms to our standards and integrates seamlessly with the infra architecture.

You can generate Go, iOS, Android, Kotlin, Dart, TypeScript, JavaScript from .api files with goctl.

## Features

- **MCP API Generation**: Creates go-zero API services optimized for Model Context Protocol (MCP) integration
- **Standard API Generation**: Creates basic go-zero API services from endpoint definitions
- **Infra Patterns**: Follows established infra conventions for naming, structure, and configuration
- **Built-in Testing**: Includes comprehensive tests to verify goctl integration

## Prerequisites

The `goctl` binary must be installed via the infra dependency system:

```bash
# Install goctl
go run . dep local install goctl --env development

# Verify installation
.dep/goctl --version
```

## Usage

### Basic Service Creation

```go
package main

import (
    "context"
    "github.com/joeblew999/infra/pkg/gozero"
)

func main() {
    // Create service wrapper
    service := gozero.NewService(false) // false = no debug output

    // Create MCP-compatible API service
    ctx := context.Background()
    err := service.CreateMCPAPI(ctx, "my-service", "My MCP API Service", "./output")
    if err != nil {
        panic(err)
    }
}
```

### Custom Endpoint API

```go
// Define custom endpoints
endpoints := []gozero.Endpoint{
    {
        Name:         "CreateUser",
        Method:       "POST",
        Path:         "/users",
        RequestType:  "CreateUserRequest",
        ResponseType: "CreateUserResponse",
        Handler:      "CreateUserHandler",
    },
    {
        Name:         "GetUser", 
        Method:       "GET",
        Path:         "/users/:id",
        ResponseType: "GetUserResponse",
        Handler:      "GetUserHandler",
    },
}

// Generate API service
err := service.CreateStandardAPI(ctx, "user-service", endpoints, "./output")
```

### Direct goctl Commands

```go
// Create low-level runner for direct goctl access
runner := gozero.NewGoZeroRunner(true) // true = debug output
runner.SetWorkDir("./my-project")

// Run specific goctl commands
err := runner.ApiGenerate("user.api", "./generated")
err = runner.ApiSwagger("user.api", "./docs")
err = runner.ModelGenerate("mysql://user:pass@localhost/db", "users", "./models")
```

## Generated Project Structure

When creating API services, the following structure is generated:

```
service-name/
├── service-name.api          # API definition file
├── service-name.go          # Main service entry point
├── service-name.json        # Swagger documentation
├── go.mod                   # Go module file
├── go.sum                   # Go dependencies
├── etc/
│   └── service_name_api.yaml # Configuration file
└── internal/
    ├── config/
    │   └── config.go        # Configuration structures
    ├── handler/
    │   ├── mcphandler.go    # MCP request handler
    │   └── healthhandler.go # Health check handler
    ├── logic/
    │   └── [logic files]    # Business logic
    ├── svc/
    │   └── servicecontext.go # Service context
    └── types/
        └── types.go         # Request/response types
```

## MCP Integration

The MCP API generation creates services that implement the Model Context Protocol for Claude Code integration:

- **JSON-RPC 2.0**: Standard MCP request/response format
- **Health Endpoint**: `/health` for service monitoring
- **MCP Endpoint**: `/mcp` for tool execution requests

## Testing

The package includes comprehensive tests:

```bash
# Run all tests
go test -v .

# Run specific test
go test -v -run TestService_CreateMCPAPI .
```

Tests verify:
- ✅ goctl binary installation and execution
- ✅ MCP API service generation end-to-end
- ✅ Project structure expectations

## Configuration

The package uses the infra configuration system:
- Binary paths: `config.Get(config.BinaryGoctl)` (strongly typed, garble-proof)
- Logging: Uses `pkg/log` for structured output
- Error handling: Follows infra error patterns

## Architecture Notes

- **Service Name Normalization**: Hyphens in service names are converted to underscores for go-zero compatibility
- **Working Directory**: Operations are isolated to specified directories
- **Module Initialization**: Automatically runs `go mod init` and `go mod tidy` for generated projects

## Development

The code generation uses string templates but should be refactored to use embedded templates for better maintainability:

```go
//go:embed templates/*.api
var templates embed.FS
```

This would make the API templates more manageable and less error-prone.

## Dependencies

- `github.com/joeblew999/infra/pkg/config` - Configuration management
- `github.com/joeblew999/infra/pkg/log` - Structured logging
- `goctl` binary (via dep system) - go-zero code generation tool