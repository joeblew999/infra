# pkg/ai - Claude CLI Control Package

The `pkg/ai` package provides programmatic control over Claude CLI operations within the infrastructure management system.

## Overview

This package acts as the bridge between your Go application and Claude Code, enabling AI-driven infrastructure management through:
- **Programmatic Claude CLI execution**
- **AI-powered configuration management**
- **Intelligent automation workflows**
- **Real-time AI feedback loops**

## Architecture

The package follows the same self-similar design pattern as the rest of the infrastructure:
- **Development**: Direct Claude CLI usage via `go run .`
- **CI/CD**: Same Claude CLI via GitHub Actions
- **Production**: Orchestrated Claude CLI operations

## Usage Patterns

### Basic Claude CLI Control
```go
import "github.com/joeblew999/infra/pkg/ai"

// Execute Claude CLI commands
ai := pkgai.NewClaudeController()
result, err := ai.RunCommand("claude mcp list")
```

### AI-Driven Infrastructure Analysis
```go
// Use Claude to analyze infrastructure state
analysis, err := ai.AnalyzeInfrastructure(currentState)
```

### Configuration Management
```go
// Let Claude optimize configurations
optimizedConfig, err := ai.OptimizeConfig(currentConfig)
```

## CLI Integration

The package integrates with the unified CLI system:

```bash
# Development mode
claude dev

# AI-powered analysis
claude ai-analyze

# Configuration optimization
claude ai-optimize

# MCP server management
claude mcp list
```

## Environment Setup

### Required Tools
- **Claude Code**: `npm install -g @anthropic-ai/claude-code`
- **GitHub CLI**: `gh auth login` (for GitHub integration)
- **Go 1.24.5+**: For Go-based control

### Environment Variables
```bash
export GITHUB_TOKEN=your_token_here
export ANTHROPIC_API_KEY=your_key_here
```

## MCP Server Management

The package provides direct control over MCP servers:

```bash
# List configured MCP servers
claude mcp list

# Add a new MCP server
claude mcp add github "npx @modelcontextprotocol/server-github" \
  --env GITHUB_TOKEN="$GITHUB_TOKEN"

# Remove an MCP server
claude mcp remove github

# Custom MCP configuration
claude --mcp-config ./.mcp.json --strict-mcp-config
```

## Configuration Files

### Project-Specific Configuration
- `.claude.json`: Project-specific settings
- `.mcp.json`: MCP server configuration
- `.env`: Environment variables

### Global Configuration
- `~/.claude.json`: Global Claude settings
- `~/.claude/config.json`: Global configuration

## AI Workflows

### 1. Infrastructure Analysis
```go
// Analyze current infrastructure state
analysis := ai.AnalyzeDeployment()
```

### 2. Configuration Generation
```go
// Generate configurations using AI
config := ai.GenerateConfig(requirements)
```

### 3. Error Diagnosis
```go
// AI-powered error diagnosis
diagnosis := ai.DiagnoseError(error)
```

### 4. Performance Optimization
```go
// Optimize performance using AI insights
optimization := ai.OptimizePerformance(metrics)
```

## Integration Examples

### GitHub Integration
```go
// Use Claude to manage GitHub operations
gh := ai.NewGitHubController()
pr, err := gh.CreatePR("feature-branch", "main")
```

### Infrastructure Automation
```go
// AI-driven infrastructure automation
automation := ai.NewInfrastructureAutomation()
err := automation.DeployStack(config)
```

## Development Workflow

### Local Development
```bash
# Start development with AI assistance
go run . dev

# Use Claude for code analysis
go run . ai-analyze

# Optimize with AI
go run . ai-optimize
```

### CI/CD Integration
```yaml
# GitHub Actions workflow with AI
- name: AI Infrastructure Analysis
  run: |
    claude ai-analyze
    claude ai-optimize
```

## Best Practices

### 1. Token Management
- Store tokens in environment variables
- Use project-specific tokens when possible
- Never commit tokens to version control

### 2. Rate Limiting
- Implement proper rate limiting
- Cache AI responses when appropriate
- Use exponential backoff for retries

### 3. Security
- Validate all AI-generated configurations
- Review AI suggestions before applying
- Use least-privilege tokens

## Troubleshooting

### Common Issues

**MCP Server Connection Failed**
```bash
# Check MCP server status
claude mcp list

# Reconfigure MCP server
claude mcp remove github
claude mcp add github "npx @modelcontextprotocol/server-github" \
  --env GITHUB_TOKEN="$GITHUB_TOKEN"
```

**Authentication Issues**
```bash
# Verify GitHub token
gh auth status

# Test GitHub API
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/user
```

**Configuration Path Issues**
```bash
# Use custom configuration
claude --settings ./custom-settings.json
claude --mcp-config ./custom-mcp.json
```

## API Reference

### ClaudeController
```go
type ClaudeController struct {
    // Control Claude CLI operations
}

func (c *ClaudeController) RunCommand(cmd string) (string, error)
func (c *ClaudeController) AnalyzeInfrastructure(state interface{}) (Analysis, error)
func (c *ClaudeController) GenerateConfig(requirements interface{}) (Config, error)
```

### MCPManager
```go
type MCPManager struct {
    // Manage MCP servers
}

func (m *MCPManager) ListServers() ([]Server, error)
func (m *MCPManager) AddServer(name string, config ServerConfig) error
func (m *MCPManager) RemoveServer(name string) error
```

## Future Enhancements

- **Multi-model support** (Gemini, GPT-4, etc.)
- **Advanced AI workflows**
- **Custom AI agents**
- **Real-time collaboration features**
- **Integration with other AI tools**

## Contributing

1. Ensure Claude CLI is properly configured
2. Test all AI workflows locally
3. Follow the self-similar design pattern
4. Document all new AI capabilities

---

**Note**: This package is designed to work seamlessly with the existing infrastructure management system, maintaining the same patterns and workflows across development, CI/CD, and production environments.