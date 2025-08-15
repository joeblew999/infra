package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func NewAICmd() *cobra.Command {
	aiCmd := &cobra.Command{
		Use:   "ai",
		Short: "AI-powered infrastructure management",
		Long:  `Commands for AI-powered infrastructure analysis, optimization, and automation using Goose and Claude`,
	}

	aiCmd.AddCommand(
		newGooseCmd(),
		newClaudeCmd(),
		newMCPCmd(),
		newAnalyzeCmd(),
		newOptimizeCmd(),
		newConfigureCmd(),
	)

	return aiCmd
}

func newGooseCmd() *cobra.Command {
	gooseCmd := &cobra.Command{
		Use:   "goose",
		Short: "Interact with Goose AI agent",
		Long:  `Direct interface to the Goose AI agent for interactive sessions, automation, and project management`,
	}

	gooseCmd.AddCommand(
		newGooseSessionCmd(),
		newGooseRunCmd(),
		newGooseConfigureCmd(),
		newGooseInfoCmd(),
		newGooseWebCmd(),
	)

	return gooseCmd
}

func newGooseSessionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "session [session-name]",
		Aliases: []string{"s"},
		Short:   "Start or resume interactive Goose session",
		Long:    `Start a new interactive Goose session or resume an existing one`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := NewGooseRunner()
			if len(args) > 0 {
				return runner.Session(args[0])
			}
			return runner.Session("")
		},
	}
}

func newGooseRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [file]",
		Short: "Execute Goose commands from file or stdin",
		Long:  `Execute Goose automation commands from an instruction file or read from stdin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := NewGooseRunner()
			if len(args) > 0 {
				return runner.RunFile(args[0])
			}
			return runner.RunStdin()
		},
	}
}

func newGooseConfigureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure Goose AI provider settings",
		Long:  `Configure Goose with AI provider credentials and preferences`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := NewGooseRunner()
			return runner.Configure()
		},
	}
}

func newGooseInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Display Goose configuration and system information",
		Long:  `Show current Goose configuration, version, and system paths`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := NewGooseRunner()
			return runner.Info()
		},
	}
}

func newGooseWebCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "web",
		Short: "Start Goose web interface",
		Long:  `Start the experimental Goose web interface for browser-based interaction`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := NewGooseRunner()
			return runner.Web()
		},
	}
}

func newClaudeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claude",
		Short: "Interact with Claude AI",
		Long:  `Direct interface to Claude Code CLI for AI-powered development assistance`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Placeholder for Claude integration
			fmt.Println("Claude integration - to be implemented")
			return nil
		},
	}
}

func newMCPCmd() *cobra.Command {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP (Model Context Protocol) servers",
		Long:  "Install, configure, and manage MCP servers for AI agent integration",
	}

	mcpCmd.AddCommand(
		newMCPInstallCmd(),
		newMCPListCmd(),
		newMCPListLocalCmd(),
		newMCPListClaudeCmd(),
		newMCPUninstallCmd(),
		newMCPUninstallAllCmd(),
	)

	return mcpCmd
}

func newAnalyzeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [target]",
		Short: "AI-powered infrastructure analysis",
		Long: `Use AI to analyze infrastructure components, configurations, and performance.
Target can be: infrastructure, configs, logs, metrics, or specific service name`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "infrastructure"
			if len(args) > 0 {
				target = args[0]
			}
			return analyzeInfrastructure(target)
		},
	}
}

func newOptimizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "optimize [component]",
		Short: "AI-powered infrastructure optimization",
		Long: `Use AI to optimize infrastructure configurations and performance.
Component can be: configs, performance, security, or specific service name`,
		RunE: func(cmd *cobra.Command, args []string) error {
			component := "configs"
			if len(args) > 0 {
				component = args[0]
			}
			return optimizeInfrastructure(component)
		},
	}
}

func newConfigureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure AI tools and providers",
		Long:  `Configure API keys, preferences, and settings for AI tools (Goose, Claude)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return configureAITools()
		},
	}
}

// Implementation functions

func analyzeInfrastructure(target string) error {
	fmt.Printf("🔍 Analyzing infrastructure: %s\n", target)
	
	runner := NewGooseRunner()
	prompt := fmt.Sprintf(`Analyze the current infrastructure component: %s
	
Please examine:
1. Current configuration and state
2. Potential issues or bottlenecks
3. Security considerations
4. Performance optimization opportunities
5. Best practice recommendations

Provide a detailed analysis with actionable recommendations.`, target)
	
	// Create a temporary analysis file
	tmpFile, err := os.CreateTemp("", "goose-analysis-*.md")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(prompt); err != nil {
		return fmt.Errorf("failed to write prompt to file: %w", err)
	}
	tmpFile.Close()
	
	return runner.RunFile(tmpFile.Name())
}

func optimizeInfrastructure(component string) error {
	fmt.Printf("⚡ Optimizing infrastructure: %s\n", component)
	
	runner := NewGooseRunner()
	prompt := fmt.Sprintf(`Optimize the infrastructure component: %s
	
Please provide:
1. Performance optimization recommendations
2. Configuration improvements
3. Resource efficiency suggestions
4. Security enhancements
5. Specific implementation steps

Generate optimized configurations where applicable.`, component)
	
	// Create a temporary optimization file
	tmpFile, err := os.CreateTemp("", "goose-optimize-*.md")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(prompt); err != nil {
		return fmt.Errorf("failed to write prompt to file: %w", err)
	}
	tmpFile.Close()
	
	return runner.RunFile(tmpFile.Name())
}

func configureAITools() error {
	fmt.Println("🔧 Configuring AI tools...")
	
	// Check current configuration
	runner := NewGooseRunner()
	fmt.Println("\n📋 Current Goose configuration:")
	if err := runner.Info(); err != nil {
		fmt.Printf("⚠️  Could not get Goose info: %v\n", err)
	}
	
	fmt.Println("\n🚀 To configure Goose, run:")
	fmt.Println("  go run . ai goose configure")
	
	fmt.Println("\n📄 Configuration files:")
	fmt.Println("  Goose config: ~/.config/goose/config.yaml")
	fmt.Println("  Claude config: ~/.claude/settings.json")
	
	fmt.Println("\n🔑 Required environment variables:")
	fmt.Println("  ANTHROPIC_API_KEY - for Claude/Goose AI providers")
	fmt.Println("  OPENAI_API_KEY    - for OpenAI providers (optional)")
	
	// Check if config directory exists
	homeDir, err := os.UserHomeDir()
	if err == nil {
		gooseConfigDir := filepath.Join(homeDir, ".config", "goose")
		if _, err := os.Stat(gooseConfigDir); os.IsNotExist(err) {
			fmt.Printf("\n💡 Creating Goose config directory: %s\n", gooseConfigDir)
			if err := os.MkdirAll(gooseConfigDir, 0755); err != nil {
				fmt.Printf("⚠️  Could not create config directory: %v\n", err)
			}
		}
	}
	
	return nil
}

// MCP command implementations

func newMCPInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install [server...]",
		Short: "Install MCP servers from configuration",
		Long:  "Install MCP servers from pkg/ai/mcp.json or specified servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPInstall(args)
		},
	}
}

func newMCPListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List both local config and Claude running servers",
		Long:  "Show both local configuration and actual Claude MCP server status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPList()
		},
	}
}

func newMCPListLocalCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-local",
		Short: "List local MCP configuration",
		Long:  "List MCP servers configured in pkg/ai/mcp.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPListLocal()
		},
	}
}

func newMCPListClaudeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-claude",
		Short: "List actual Claude MCP servers",
		Long:  "Show MCP servers currently running in Claude",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPListClaude()
		},
	}
}

func newMCPUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall [server...]",
		Short: "Uninstall MCP servers",
		Long:  "Remove specified MCP servers from configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPUninstall(args)
		},
	}
}

func newMCPUninstallAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall-all",
		Short: "Uninstall all MCP servers",
		Long:  "Remove all MCP servers from Claude configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPUninstallAll()
		},
	}
}

// MCP command handlers - adapted from the original commands.go

func runMCPInstall(args []string) error {
	manager, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
	}

	// Load our configuration
	configPath := "pkg/ai/mcp.json"
	if _, err := os.Stat(configPath); err == nil {
		if err := manager.LoadConfigFromFile(configPath); err != nil {
			return fmt.Errorf("failed to load MCP config: %w", err)
		}
		fmt.Println("✅ MCP servers installed from configuration")
	} else {
		fmt.Println("ℹ️  No MCP configuration found, skipping")
	}

	return nil
}

func runMCPList() error {
	manager, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
	}

	// Show local configuration
	localServers := manager.List()
	fmt.Println("📝 Local MCP configuration (what can be installed):")
	if len(localServers) == 0 {
		fmt.Println("  No local configuration found")
	} else {
		for _, server := range localServers {
			fmt.Printf("  - %s (%s)\n    Command: %s %s\n", server.Name, server.Version, server.Command, strings.Join(server.Args, " "))
		}
	}

	// Show Claude's actual running servers
	fmt.Println("\n🚀 Claude's currently running MCP servers:")
	claudeStatus, err := manager.GetClaudeStatus()
	if err != nil {
		fmt.Printf("  Could not query Claude: %v\n", err)
		return nil
	}

	if len(claudeStatus) == 0 {
		fmt.Println("  No MCP servers currently running")
		return nil
	}

	for _, server := range claudeStatus {
		status := server.Status
		if server.Error != "" {
			status = fmt.Sprintf("%s (Error: %s)", server.Status, server.Error)
		}
		fmt.Printf("  - %s [%s]\n    Command: %s\n", server.Name, status, server.Command)
	}

	return nil
}

func runMCPListLocal() error {
	manager, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
	}

	servers := manager.List()
	if len(servers) == 0 {
		fmt.Println("No MCP servers configured locally")
		return nil
	}

	fmt.Println("Local MCP configuration (pkg/ai/mcp.json):")
	for _, server := range servers {
		fmt.Printf("  - %s (%s)\n", server.Name, server.Version)
		fmt.Printf("    Command: %s %s\n", server.Command, strings.Join(server.Args, " "))
		fmt.Printf("    Repo: %s\n", server.Repo)
	}

	return nil
}

func runMCPListClaude() error {
	manager, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
	}

	// Show config locations
	locations, err := manager.GetClaudeConfigLocations()
	if err != nil {
		fmt.Printf("Warning: Could not determine Claude config locations: %v\n", err)
	} else {
		fmt.Println("Claude MCP configuration locations:")
		for _, location := range locations {
			fmt.Printf("  - %s\n", location)
		}
		fmt.Println()
	}

	claudeStatus, err := manager.GetClaudeStatus()
	if err != nil {
		return fmt.Errorf("failed to query Claude: %w", err)
	}

	if len(claudeStatus) == 0 {
		fmt.Println("No MCP servers currently running in Claude")
		return nil
	}

	fmt.Println("MCP servers currently running in Claude:")
	for _, server := range claudeStatus {
		status := server.Status
		if server.Error != "" {
			status = fmt.Sprintf("%s (Error: %s)", server.Status, server.Error)
		}
		fmt.Printf("  - %s [%s]\n", server.Name, status)
		fmt.Printf("    Command: %s\n", server.Command)
	}

	return nil
}

func runMCPUninstall(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please specify server names to uninstall")
	}

	manager, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
	}

	if err := manager.Uninstall(args); err != nil {
		return fmt.Errorf("failed to uninstall servers: %w", err)
	}

	fmt.Printf("✅ Uninstalled servers: %v\n", args)
	return nil
}

func runMCPUninstallAll() error {
	manager, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
	}

	servers := manager.List()
	if len(servers) == 0 {
		fmt.Println("ℹ️  No MCP servers to uninstall")
		return nil
	}

	// Collect all server names
	serverNames := make([]string, len(servers))
	for i, server := range servers {
		serverNames[i] = server.Name
	}

	if err := manager.Uninstall(serverNames); err != nil {
		return fmt.Errorf("failed to uninstall all servers: %w", err)
	}

	fmt.Printf("✅ Uninstalled all %d MCP servers\n", len(servers))
	return nil
}