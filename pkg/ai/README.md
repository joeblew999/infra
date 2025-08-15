# pkg/ai - AI-Powered Infrastructure Management

The `pkg/ai` package provides unified access to AI tools for infrastructure management, analysis, and automation. It integrates both **Claude** and **Goose** AI agents through a consistent CLI interface.

## Overview

This package bridges your Go application with AI tools, enabling:
- **Interactive AI sessions** for infrastructure tasks
- **Automated analysis and optimization** of configurations
- **AI-powered automation workflows** 
- **Unified command interface** for multiple AI tools

## Supported AI Tools

### Goose (Block)
**Interactive AI agent for development and automation**
- Repository: [block/goose](https://github.com/block/goose)
- Version: v1.3.1 (managed by dep system)
- Capabilities: Interactive sessions, file automation, web interface

### Claude (Anthropic)
**AI assistant for code and infrastructure analysis**
- Repository: [anthropics/claude-code](https://github.com/anthropics/claude-code)
- Version: Latest (managed by dep system)
- Capabilities: MCP servers, code analysis, GitHub integration

## Installation

Both AI tools are automatically installed through the dependency system:

```bash
# Install both Claude and Goose
go run . dep ensure

# Or install individually
go run . dep install claude
go run . dep install goose
```

## Quick Start

### Configure AI Tools
```bash
# Configure all AI tools at once
go run . ai configure

# Configure Goose specifically
go run . ai goose configure

# Check Claude configuration
go run . ai config
```

### Basic Usage
```bash
# Start interactive Goose session
go run . ai goose session

# Analyze infrastructure with AI
go run . ai analyze infrastructure

# Optimize configurations
go run . ai optimize configs

# Manage MCP servers for AI agents
go run . ai mcp list
go run . ai mcp install

# Start Goose web interface
go run . ai goose web
```

## Command Reference

### Core AI Commands

#### `go run . ai configure`
Configure all AI tools and check system status.

#### `go run . ai analyze [target]`
AI-powered infrastructure analysis.
- `infrastructure` - Analyze overall infrastructure
- `configs` - Analyze configuration files
- `logs` - Analyze system logs
- `metrics` - Analyze performance metrics

#### `go run . ai optimize [component]`
AI-powered optimization recommendations.
- `configs` - Optimize configuration files
- `performance` - Performance optimization
- `security` - Security improvements

### Goose Commands

#### `go run . ai goose session [name]`
Start or resume interactive Goose session.
```bash
# Start default session
go run . ai goose session

# Start named session
go run . ai goose session infrastructure-review
```

#### `go run . ai goose run [file]`
Execute Goose automation from file or stdin.
```bash
# Run from file
go run . ai goose run automation.md

# Run from stdin
echo "Analyze the current directory" | go run . ai goose run
```

#### `go run . ai goose info`
Display Goose configuration and system information.

#### `go run . ai goose web`
Start Goose web interface for browser-based interaction.

### MCP Commands

#### `go run . ai mcp list`
List both local MCP configuration and Claude's running servers.

#### `go run . ai mcp list-local`
Show local MCP server configuration from pkg/ai/mcp.json.

#### `go run . ai mcp list-claude`
Show MCP servers currently running in Claude.

#### `go run . ai mcp install`
Install MCP servers from configuration file.

#### `go run . ai mcp uninstall [server...]`
Remove specific MCP servers.

#### `go run . ai mcp uninstall-all`
Remove all MCP servers.

### Claude Commands

#### `go run . ai config`
Show Claude configuration paths and settings.

## Configuration

### Environment Variables
```bash
# Required for most AI providers
export ANTHROPIC_API_KEY=your_anthropic_key_here

# Optional for additional providers
export OPENAI_API_KEY=your_openai_key_here
export GITHUB_TOKEN=your_github_token_here
```

### Configuration Files

#### Goose Configuration
- **Config**: `~/.config/goose/config.yaml`
- **Sessions**: `~/.local/share/goose/sessions`
- **Logs**: `~/.local/state/goose/logs`

#### Claude Configuration
- **Settings**: `~/.claude/settings.json`
- **Project config**: `.claude.json` (optional)
- **MCP config**: `.mcp.json` (optional)

#### MCP Configuration
- **Default config**: `pkg/ai/mcp.json`
- **Config templates**: `pkg/ai/config/mcp-servers.json`
- **Available servers**: See supported MCP server types below

## Programming Interface

### GooseRunner
Direct programmatic access to Goose functionality:

```go
import "github.com/joeblew999/infra/pkg/ai"

// Create runner
runner := ai.NewGooseRunner()

// Start interactive session
err := runner.Session("my-session")

// Run from file
err := runner.RunFile("automation.md")

// Get system info
err := runner.Info()

// Start web interface
err := runner.Web()
```

### Available Runner Methods
```go
// Session management
Session(name string) error
RunFile(filename string) error
RunStdin() error

// Configuration
Configure() error
Info() error
Version() (string, error)

// Interfaces
Web() error
ListSessions() error

// Advanced features
Schedule(action string, args ...string) error
Benchmark() error
MCP(serverName string, args ...string) error
Recipe(action string, args ...string) error
Update() error
```

## Usage Examples

### Infrastructure Analysis Workflow
```bash
# 1. Configure AI tools
go run . ai configure

# 2. Analyze current infrastructure
go run . ai analyze infrastructure

# 3. Get optimization recommendations
go run . ai optimize performance

# 4. Start interactive session for implementation
go run . ai goose session infrastructure-improvements
```

### Automated Configuration Review
```bash
# Create automation file
cat > config-review.md << EOF
# Infrastructure Configuration Review

Please analyze all configuration files in this project:
1. Check for security best practices
2. Identify performance bottlenecks  
3. Suggest optimization opportunities
4. Generate improved configurations

Focus on:
- Docker configurations
- Service configurations
- Network settings
- Resource limits
EOF

# Execute with Goose
go run . ai goose run config-review.md
```

### Web Interface Development
```bash
# Start Goose web interface
go run . ai goose web

# Access at http://localhost:8080 (or configured port)
# Use for browser-based AI interaction
```

## Architecture

The package follows the infrastructure's self-similar design pattern:

```
pkg/ai/
├── cmd.go                # All AI commands (Goose, Claude, MCP)
├── runner.go             # GooseRunner implementation
├── mcp_manager.go        # MCP server management
├── mcp_installer.go      # MCP server installation
├── mcp_types.go          # MCP data structures  
├── mcp_consts.go         # MCP constants and paths
├── mcp.json              # Default MCP server config
├── config/               # Configuration templates
└── README.md             # This documentation

Integration:
pkg/cmd/ai.go             # CLI integration point
```

## Troubleshooting

### Goose Not Found
```bash
# Reinstall goose
go run . dep remove goose
go run . dep install goose

# Check installation
go run . ai goose info
```

### Configuration Issues
```bash
# Check current setup
go run . ai configure

# Reconfigure Goose
go run . ai goose configure

# Check Claude setup
go run . ai config
```

### Provider Authentication
```bash
# Check environment variables
echo $ANTHROPIC_API_KEY
echo $OPENAI_API_KEY

# Test Claude authentication
go run . ai mcp

# Test Goose functionality
go run . ai goose info
```

### Binary Path Issues
The package automatically uses dependency-managed binaries:
- Goose: Uses `dep.Get("goose")` for path resolution
- Claude: Uses `config.GetClaudeBinPath()` for path resolution

If you encounter path issues, ensure dependencies are installed:
```bash
go run . dep ensure
```

## Best Practices

### 1. Session Management
- Use descriptive session names: `infrastructure-review`, `security-audit`
- Save important sessions for future reference
- Use `go run . ai goose session --help` for session options

### 2. Automation Files
- Use Markdown format for Goose automation files
- Include clear objectives and context
- Break complex tasks into steps
- Save reusable automations in version control

### 3. Security
- Never commit API keys to version control
- Use environment variables for credentials
- Validate AI-generated configurations before applying
- Review all AI recommendations before implementation

### 4. Performance
- Use specific analysis targets: `go run . ai analyze configs` vs generic analysis
- Cache AI responses for repeated operations
- Use batch operations when possible

## Integration with Infrastructure System

The AI package integrates seamlessly with the infrastructure system:

```bash
# Service mode with AI assistance
go run . service --ai-enabled

# Deployment with AI analysis
go run . workflows deploy --analyze

# Configuration generation
go run . ai optimize configs > optimized-configs.yaml
```

## Future Enhancements

- [ ] **Multi-model support** (Gemini, GPT-4, local models)
- [ ] **Advanced workflow templates**
- [ ] **Integration with monitoring systems**
- [ ] **Automated deployment pipelines**
- [ ] **Custom AI agent development**
- [ ] **Real-time collaboration features**

## Contributing

1. Ensure both Claude and Goose are properly configured
2. Test all AI workflows locally before committing
3. Follow the self-similar design pattern
4. Document new AI capabilities and workflows
5. Add examples for new functionality

---

**Note**: This package maintains consistency with the infrastructure system's patterns, providing AI-powered enhancements without disrupting existing workflows.