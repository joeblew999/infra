package cmd

import (
	"github.com/joeblew999/infra/pkg/ai"
	"github.com/spf13/cobra"
)

// NewClaudeCmd creates the Claude AI command with all subcommands
func NewClaudeCmd() *cobra.Command {
	claudeCmd := &cobra.Command{
		Use:   "claude",
		Short: "Interact with Claude AI",
		Long:  `Direct interface to Claude Code CLI for AI-powered development assistance`,
	}

	claudeCmd.AddCommand(
		newClaudeSessionCmd(),
		newClaudeRunCmd(),
		newClaudeConfigureCmd(),
		newClaudeInfoCmd(),
		newClaudeMCPCmd(),
	)

	return claudeCmd
}

func newClaudeSessionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "session [session-name]",
		Aliases: []string{"s"},
		Short:   "Start or resume interactive Claude session",
		Long:    `Start a new interactive Claude session or resume an existing one`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := ai.NewClaudeRunner()
			sessionName := ""
			if len(args) > 0 {
				sessionName = args[0]
			}
			return runner.Session(sessionName)
		},
	}
}

func newClaudeRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [file]",
		Short: "Execute Claude commands from file or stdin",
		Long:  `Execute Claude automation commands from an instruction file or read from stdin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := ai.NewClaudeRunner()
			if len(args) > 0 {
				return runner.RunFile(args[0])
			}
			return runner.RunStdin()
		},
	}
}

func newClaudeConfigureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure Claude AI provider settings",
		Long:  `Configure Claude with AI provider credentials and preferences using the Claude CLI`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := ai.NewClaudeRunner()
			return runner.Configure()
		},
	}
}

func newClaudeInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Display Claude configuration and system information",
		Long:  `Show current Claude configuration, version, and system paths`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ai.DisplayClaudeInfo()
		},
	}
}

func newClaudeMCPCmd() *cobra.Command {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP servers for Claude",
		Long:  "Configure and manage MCP servers for Claude AI agent integration",
	}

	mcpCmd.AddCommand(
		newClaudeMCPListCmd(),
		newClaudeMCPAddCmd(),
		newClaudeMCPRemoveCmd(),
		newClaudeMCPPresetListCmd(),
		newClaudeMCPPresetInstallCmd(),
		newClaudeMCPPresetInstallAllCmd(),
	)

	return mcpCmd
}

func newClaudeMCPListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List MCP servers for Claude",
		Long:  "Show MCP servers configured for Claude",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPListClaude()
		},
	}
}

func newClaudeMCPAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [name] [command]",
		Short: "Add MCP server to Claude",
		Long: `Add a new MCP server to Claude's configuration

Examples:
  # Add filesystem MCP server
  go run . cli ai claude mcp add filesystem "bun x @modelcontextprotocol/server-filesystem"
  
  # Add DAP server with SSE transport  
  go run . cli ai claude mcp add mcp-dap-server "http://localhost:8080"
  
  # Add GitHub MCP server
  go run . cli ai claude mcp add github "bun x @modelcontextprotocol/server-github"`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPAddClaude(args[0], args[1])
		},
	}
}

func newClaudeMCPRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove MCP server from Claude",
		Long:  "Remove an MCP server from Claude's configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPRemoveClaude(args[0])
		},
	}
}

func newClaudeMCPPresetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preset-list",
		Short: "List available preset MCP servers",
		Long:  "List all available preset MCP servers defined in claude-mcp-default.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := ai.NewClaudeRunner()
			return runner.PresetList()
		},
	}
}

func newClaudeMCPPresetInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preset-install [server-name]",
		Short: "Install specific preset MCP server",
		Long: `Install a specific MCP server from the preset configuration

Use 'preset-list' to see available servers.

Examples:
  # Install GitHub MCP server
  go run . cli ai claude mcp preset-install github
  
  # Install filesystem MCP server  
  go run . cli ai claude mcp preset-install filesystem`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := ai.NewClaudeRunner()
			return runner.InstallMCPByName(args[0])
		},
	}
}

func newClaudeMCPPresetInstallAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preset-install-all",
		Short: "Install all preset MCP servers",
		Long:  "Install all preset MCP servers defined in claude-mcp-default.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := ai.NewClaudeRunner()
			return runner.InstallDefaultMCP()
		},
	}
}

func runMCPListClaude() error {
	runner := ai.NewClaudeRunner()
	return runner.MCPList()
}

func runMCPAddClaude(name, command string) error {
	runner := ai.NewClaudeRunner()
	return runner.MCPAdd(name, command)
}

func runMCPRemoveClaude(name string) error {
	runner := ai.NewClaudeRunner()
	return runner.MCPRemove(name)
}
