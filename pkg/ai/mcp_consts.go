package ai

import "path/filepath"

// MCP package constants
const (
	// DefaultConfigFile is the default MCP configuration file name
	DefaultConfigFile = "claude-mcp-default.json"

	// DefaultConfigDir is the default directory for MCP configurations
	DefaultConfigDir = "pkg/ai"

	// ClaudeConfigDir is the Claude configuration directory
	ClaudeConfigDir = ".config/claude"

	// ClaudeConfigFile is the Claude MCP configuration file name
	ClaudeConfigFile = "claude-mcp-default.json"

	// StatusRunning indicates a server is running
	StatusRunning = "running"

	// StatusError indicates a server has an error
	StatusError = "error"

	// StatusUnknown indicates unknown server status
	StatusUnknown = "unknown"

	// ServerTypeStdio indicates stdio-based MCP server
	ServerTypeStdio = "stdio"

	// ServerTypeHTTP indicates HTTP-based MCP server
	ServerTypeHTTP = "http"

	// ServerTypeSSE indicates SSE-based MCP server
	ServerTypeSSE = "sse"

	// DefaultVersion is the default version for new MCP servers
	DefaultVersion = "latest"

	// DefaultRepo is the default repository for MCP servers
	DefaultRepo = "modelcontextprotocol/servers"
)

// GetDefaultConfigPath returns the default MCP configuration file path
func GetDefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir, DefaultConfigFile)
}

// GetClaudeConfigPath returns the Claude MCP configuration file path
func GetClaudeConfigPath() string {
	return filepath.Join(ClaudeConfigDir, ClaudeConfigFile)
}

// CommonClaudeConfigLocations returns common Claude configuration file locations
func CommonClaudeConfigLocations() []string {
	return []string{
		"~/.claude.json",
		"~/.config/claude/mcp.json",
		"~/.claude/mcp.json",
	}
}