# pkg/ai - AI-Powered Infrastructure Management

Unified AI tools for infrastructure automation, analysis, and optimization using **Goose** and **Claude** agents plus **MCP** (Model Context Protocol) server management.

## Quick Start

```bash
# Install AI tools
go run . dep install goose claude

# Configure with your API keys
export ANTHROPIC_API_KEY=your_key_here
go run . ai configure

# Start interactive Goose session
go run . ai goose session

# Manage MCP servers for AI agents
go run . ai mcp list
```

## Supported Tools

### Goose (Block AI)
Interactive AI agent for development automation
- **Repository**: [block/goose](https://github.com/block/goose) 
- **Version**: v1.3.1 (auto-installed)
- **Use cases**: Code analysis, automation scripts, interactive problem-solving

### Claude (Anthropic)
AI assistant for code and infrastructure analysis  
- **Repository**: [anthropics/claude-code](https://github.com/anthropics/claude-code)
- **Use cases**: Code review, documentation, GitHub integration

### MCP Servers
**Model Context Protocol** - Extends AI agents with external tool access
- **Examples**: GitHub API, file system, databases, web search
- **Purpose**: Give AI agents structured access to your infrastructure

## Commands

### Goose Automation
```bash
# Interactive sessions
go run . ai goose session [name]         # Start/resume session
go run . ai goose info                   # Check configuration
go run . ai goose web                    # Web interface

# Automation
go run . ai goose run automation.md      # Execute from file
echo "Analyze this repo" | go run . ai goose run  # From stdin
```

### MCP Server Management  
```bash
go run . ai mcp list                     # Show local config + Claude status
go run . ai mcp list-local               # Show pkg/ai/mcp.json config  
go run . ai mcp list-claude              # Show Claude's running servers
go run . ai mcp install                  # Install from config
go run . ai mcp uninstall github         # Remove specific servers
```

### AI Analysis (Experimental)
```bash
go run . ai analyze infrastructure       # Infrastructure analysis
go run . ai optimize configs             # Configuration optimization  
go run . ai configure                    # Setup all AI tools
```

## Configuration

### Environment Variables
```bash
export ANTHROPIC_API_KEY=your_anthropic_key    # Required for most AI providers
export GITHUB_TOKEN=your_github_token          # For MCP GitHub integration
```

### MCP Server Configuration
Edit `pkg/ai/mcp.json` to configure available MCP servers:

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

### File Locations
- **Goose config**: `~/.config/goose/config.yaml`
- **Goose sessions**: `~/.local/share/goose/sessions`  
- **Claude settings**: `~/.claude/settings.json`
- **MCP config**: `pkg/ai/mcp.json`

## Examples

### Infrastructure Analysis Workflow
```bash
# 1. Start Goose session for infrastructure review
go run . ai goose session infra-analysis

# In the session, ask Goose to:
# - Analyze current infrastructure setup
# - Identify potential improvements  
# - Generate optimized configurations
# - Document findings
```

### MCP Server Setup
```bash
# 1. Configure MCP servers for AI agents
go run . ai mcp install

# 2. Verify servers are running  
go run . ai mcp list-claude

# 3. Use in Claude Code with MCP access to:
#    - GitHub repositories  
#    - File system operations
#    - Database queries
```

### Automation Script Example
Create `infrastructure-audit.md`:
```markdown
# Infrastructure Security Audit

Please perform a security audit of this infrastructure:

1. Review all configuration files for security best practices
2. Check for exposed secrets or credentials  
3. Analyze network security configurations
4. Generate a security improvement plan
5. Create implementation checklist

Focus on: Docker configs, service configs, environment variables
```

Execute: `go run . ai goose run infrastructure-audit.md`

## Programming Interface

### GooseRunner
```go
import "github.com/joeblew999/infra/pkg/ai"

// Create runner (uses dependency-managed binary)
runner := ai.NewGooseRunner()

// Interactive session
err := runner.Session("my-session")

// Run automation from file
err := runner.RunFile("automation.md") 

// Get system info
err := runner.Info()
```

### MCP Manager
```go
// Create MCP manager
manager, err := ai.NewManager()

// List configured servers
servers := manager.List()

// Get Claude's running servers
status, err := manager.GetClaudeStatus()
```

## Troubleshooting

### Installation Issues
```bash
# Reinstall tools
go run . dep remove goose claude
go run . dep install goose claude

# Check versions
go run . ai goose info
```

### Configuration Problems
```bash
# Check setup
go run . ai configure

# Test authentication
echo $ANTHROPIC_API_KEY
go run . ai goose info
```

### MCP Server Issues  
```bash
# Check if Node.js is available (required for MCP servers)
node --version

# Verify MCP configuration
go run . ai mcp list-local

# Check Claude's server status
go run . ai mcp list-claude
```

## Architecture

```
pkg/ai/
├── cmd.go                # All commands (Goose, Claude, MCP)
├── runner.go             # Goose automation interface
├── mcp_manager.go        # MCP server management
├── mcp_installer.go      # MCP installation logic  
├── mcp_types.go          # MCP data structures
├── mcp_consts.go         # Constants and paths
├── mcp.json              # Default MCP configuration
└── config/               # Configuration templates
```

## Best Practices

1. **Use descriptive session names**: `security-audit`, `performance-review`
2. **Save automation scripts**: Store reusable `.md` files in version control  
3. **Validate AI output**: Always review AI-generated configs before applying
4. **Environment variables**: Never commit API keys, use env vars
5. **MCP security**: Restrict file system access, use least-privilege tokens

---

**Note**: This package provides AI-powered infrastructure management while maintaining consistency with the existing infrastructure system patterns.