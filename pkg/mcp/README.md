# MCP Package

I THINK THIS PKG is ALL WRONG. Need to think more about best way to control the AI, etc -

This package provides tools for managing MCP (Model Context Protocol) servers for Claude Code integration.

## Overview

MCP (Model Context Protocol) allows Claude to extend its capabilities by connecting to external tools and services. This package provides a Go-based manager for installing, configuring, and managing MCP servers.

## Quick Start

### Using the CLI

```bash
# List configured MCP servers
go run ./cmd infra mcp list

# Install MCP servers from configuration
go run ./cmd infra mcp install

# Uninstall specific servers
go run ./cmd infra mcp uninstall github filesystem

# Uninstall all MCP servers
go run ./cmd infra mcp uninstall-all
```

### Configuration

MCP servers are configured in `pkg/mcp/mcp.json`. The configuration format is:

```json
{
  "servers": [
    {
      "name": "github",
      "version": "2025-01-24",
      "repo": "modelcontextprotocol/servers",
      "type": "stdio",
      "command": "node",
      "args": ["build/index.js"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  ]
}
```

## API Usage

```go
package main

import (
    "fmt"
    "github.com/joeblew999/infra/pkg/mcp"
)

func main() {
    // Create manager
    manager, err := mcp.NewManager()
    if err != nil {
        panic(err)
    }
    
    // List configured servers
    servers := manager.List()
    for _, server := range servers {
        fmt.Printf("Server: %s (%s)\n", server.Name, server.Version)
    }
    
    // Install servers from config
    err = manager.LoadConfigFromFile("mcp.json")
    if err != nil {
        panic(err)
    }
}
```

## Available MCP Servers

See [MCP-LIST.md](./MCP-LIST.md) for a list of available MCP servers and their purposes.

## Resources

- [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code/overview)
- [MCP Documentation](https://docs.anthropic.com/en/docs/claude-code/mcp)
- [Official MCP Servers](https://github.com/modelcontextprotocol/servers)