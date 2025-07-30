# MCP Server List

This file lists useful MCP (Model Context Protocol) servers for extending Claude's capabilities.

## Debug Servers

### Go Debug (Delve DAP)
- **Repository**: https://github.com/go-delve/mcp-dap-server
- **Purpose**: Debug Go applications using Delve's Debug Adapter Protocol
- **Status**: Available but may need configuration
- **Use Case**: Interactive debugging sessions with Claude for Go development

## Official MCP Servers

### GitHub Integration
- **Repository**: modelcontextprotocol/servers
- **Purpose**: GitHub API integration for repository management
- **Commands**: Repository search, file operations, issue management

### File System
- **Repository**: modelcontextprotocol/servers
- **Purpose**: Local file system operations
- **Commands**: File reading, directory listing, file operations

### Web Fetch
- **Repository**: modelcontextprotocol/servers
- **Purpose**: HTTP requests and web content fetching
- **Commands**: GET/POST requests, web scraping, API calls

## Configuration Examples

Add these to your `mcp.json`:

```json
{
  "servers": [
    {
      "name": "go-debug",
      "version": "latest",
      "repo": "go-delve/mcp-dap-server",
      "type": "stdio",
      "command": "node",
      "args": ["dist/index.js"]
    }
  ]
}
```