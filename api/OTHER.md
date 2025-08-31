# Go-Zero MCP Server

This is a Model Context Protocol (MCP) server built with the go-zero framework, demonstrating how to create MCP servers using go-zero's code generation capabilities.

## What This Demonstrates

✅ **Go-zero can be used as an MCP server framework**  
✅ **Minimal code required** - only business logic, everything else is generated  
✅ **Infrastructure-focused MCP tools** for system management  

## Quick Start

```bash
# Run the server
./mcpserver

# Test the MCP tools endpoint
curl -X POST http://localhost:8889/mcp/tools/list \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

## Available MCP Tools

- **`system_info`** - Get system information (CPU, memory, disk usage)
- **`list_services`** - List running services on the system  
- **`get_logs`** - Get logs from a specific service or file

## How This Was Built

### 1. Create Project Structure
```bash
# Create new go-zero API project  
cd pkg/ai/go-zero/
goctl api new mcpserver
cd mcpserver
```

### 2. Initialize Go Module
```bash
# Initialize with proper module name
go mod init github.com/joeblew999/infra/pkg/ai/go-zero/mcpserver
go mod tidy
```

### 3. Add to Workspace
```bash
# Add to root go.work file
echo "use ./pkg/ai/go-zero/mcpserver" >> ../../../../go.work
```

### 4. Define MCP API
Edit `mcpserver.api` to define MCP JSON-RPC endpoints:

```api
syntax = "v1"

// MCP JSON-RPC Request
type McpRequest {
    Jsonrpc string `json:"jsonrpc"`
    Method  string `json:"method"`
    Id      int64  `json:"id"`
    Params  string `json:"params,optional"`
}

// MCP JSON-RPC Response  
type McpResponse {
    Jsonrpc string `json:"jsonrpc"`
    Id      int64  `json:"id"`
    Result  string `json:"result,optional"`
    Error   string `json:"error,optional"`
}

// Tool definition
type Tool {
    Name        string `json:"name"`
    Description string `json:"description"`
    Schema      string `json:"schema"`
}

// Tools list response
type ToolsListResponse {
    Tools []Tool `json:"tools"`
}

service mcpserver-api {
    @handler ToolsListHandler
    post /mcp/tools/list (McpRequest) returns (McpResponse)
    
    @handler ToolsCallHandler  
    post /mcp/tools/call (McpRequest) returns (McpResponse)
    
    @handler PromptsListHandler
    post /mcp/prompts/list (McpRequest) returns (McpResponse)
    
    @handler ResourcesListHandler
    post /mcp/resources/list (McpRequest) returns (McpResponse)
}
```

### 5. Generate Code
```bash
# Generate all handlers, types, and boilerplate
goctl api go -api mcpserver.api -dir .
```

### 6. Configure Server Port
Edit `etc/mcpserver-api.yaml`:
```yaml
Name: mcpserver-api
Host: 0.0.0.0
Port: 8889  # Changed from 8888 to avoid conflicts
```

### 7. Implement Business Logic
Only needed to edit `internal/logic/toolslistlogic.go` to:
- Define the available MCP tools
- Implement JSON marshalling 
- Return proper MCP JSON-RPC responses

### 8. Build and Run
```bash
# Build the server
go build -o mcpserver .

# Run the server  
./mcpserver
```

## What Go-Zero Generated vs What I Coded

### Generated Automatically by goctl:
- ✅ All HTTP handlers (`internal/handler/*.go`)
- ✅ Request/response types (`internal/types/types.go`)  
- ✅ URL routing (`internal/handler/routes.go`)
- ✅ Service context (`internal/svc/servicecontext.go`)
- ✅ Configuration structs (`internal/config/config.go`)
- ✅ Main server setup (`mcpserver.go`)

### Coded Manually:
- ✅ API specification (`mcpserver.api`) - Defined MCP endpoints
- ✅ Business logic (`internal/logic/toolslistlogic.go`) - ~40 lines of actual MCP logic

**Total manual code: ~40 lines**  
**Total generated code: ~300+ lines**

## Key Insights

1. **Go-zero is perfect for MCP servers** - JSON-RPC over HTTP works seamlessly
2. **Code generation does the heavy lifting** - Focus on business logic, not boilerplate  
3. **API-first design** - Define your contract in `.api`, implementation follows
4. **Infrastructure tools are natural fit** - MCP + go-zero perfect for DevOps automation

## Next Steps

- Implement `tools/call` handler to execute the tools
- Add more infrastructure management tools
- Integrate with existing infrastructure packages  
- Add prompts and resources endpoints

## Testing the MCP Server

```bash
# Test tools list
curl -X POST http://localhost:8889/mcp/tools/list \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' | jq

# Expected response:
# {
#   "jsonrpc": "2.0", 
#   "id": 1,
#   "result": "{\"tools\":[{\"name\":\"system_info\",...}]}"
# }
```

This demonstrates that **go-zero is an excellent choice for building MCP servers** with minimal effort and maximum productivity! 🚀