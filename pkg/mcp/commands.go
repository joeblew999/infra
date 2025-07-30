package mcp

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// AddCommands adds MCP management commands to the CLI
func AddCommands(rootCmd *cobra.Command) {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP (Model Context Protocol) servers",
		Long:  "Install, configure, and manage MCP servers for Claude Code integration",
	}

	installCmd := &cobra.Command{
		Use:   "install [server...]",
		Short: "Install MCP servers from configuration",
		Long:  "Install MCP servers from pkg/mcp/mcp.json or specified servers",
		RunE:  runInstall,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List both local config and Claude running servers",
		Long:  "Show both local configuration and actual Claude MCP server status",
		RunE:  runList,
	}

	listLocalCmd := &cobra.Command{
		Use:   "list-local",
		Short: "List local MCP configuration",
		Long:  "List MCP servers configured in pkg/mcp/mcp.json",
		RunE:  runListLocal,
	}

	listClaudeCmd := &cobra.Command{
		Use:   "list-claude",
		Short: "List actual Claude MCP servers",
		Long:  "Show MCP servers currently running in Claude",
		RunE:  runListClaude,
	}

	uninstallCmd := &cobra.Command{
		Use:   "uninstall [server...]",
		Short: "Uninstall MCP servers",
		Long:  "Remove specified MCP servers from configuration",
		RunE:  runUninstall,
	}

	uninstallAllCmd := &cobra.Command{
		Use:   "uninstall-all",
		Short: "Uninstall all MCP servers",
		Long:  "Remove all MCP servers from Claude configuration",
		RunE:  runUninstallAll,
	}

	mcpCmd.AddCommand(installCmd, listCmd, listLocalCmd, listClaudeCmd, uninstallCmd, uninstallAllCmd)
	rootCmd.AddCommand(mcpCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	manager, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
	}

	// Load our configuration
	configPath := "pkg/mcp/mcp.json"
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

func runList(cmd *cobra.Command, args []string) error {
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

func runListLocal(cmd *cobra.Command, args []string) error {
	manager, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
	}

	servers := manager.List()
	if len(servers) == 0 {
		fmt.Println("No MCP servers configured locally")
		return nil
	}

	fmt.Println("Local MCP configuration (pkg/mcp/mcp.json):")
	for _, server := range servers {
		fmt.Printf("  - %s (%s)\n", server.Name, server.Version)
		fmt.Printf("    Command: %s %s\n", server.Command, strings.Join(server.Args, " "))
		fmt.Printf("    Repo: %s\n", server.Repo)
	}

	return nil
}

func runListClaude(cmd *cobra.Command, args []string) error {
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

func runUninstall(cmd *cobra.Command, args []string) error {
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

func runUninstallAll(cmd *cobra.Command, args []string) error {
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