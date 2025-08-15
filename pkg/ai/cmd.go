package ai

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/dep"
	"github.com/spf13/cobra"
)

func NewAICmd() *cobra.Command {
	aiCmd := &cobra.Command{
		Use:   "ai",
		Short: "AI-powered infrastructure management",
		Long:  `Commands for AI-powered infrastructure analysis, optimization, and automation using Goose and Claude`,
	}

	aiCmd.AddCommand(
		NewGooseCmd(),
		NewClaudeCmd(),
	)

	return aiCmd
}

func configureAITools() error {
	fmt.Println("🔧 Configuring AI tools...")

	// Ensure Goose is installed (idempotent)
	fmt.Println("\n🦆 Ensuring Goose is available...")
	runner := NewGooseRunner()

	// Ensure Claude is available (idempotent)
	fmt.Println("\n🤖 Ensuring Claude is available...")
	claudePath, err := dep.Get("claude")
	if err != nil {
		fmt.Println("   Claude not found, attempting installation...")
		if installErr := dep.InstallBinary("claude", false); installErr != nil {
			fmt.Printf("   ⚠️  Could not auto-install Claude: %v\n", installErr)
			fmt.Println("   You may need to install Claude manually or ensure it's in PATH")
		} else {
			fmt.Println("   ✅ Claude installed successfully")
		}
	} else {
		fmt.Printf("   ✅ Claude available at: %s\n", claudePath)
	}

	// Check current configuration
	fmt.Println("\n📋 Current Goose configuration:")
	if err := runner.Info(); err != nil {
		fmt.Printf("⚠️  Could not get Goose info: %v\n", err)
		fmt.Println("\n🚀 Run this to configure Goose:")
		fmt.Println("  go run . ai goose configure")
	}

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

	// Install MCP servers if config exists
	fmt.Println("\n🔌 Setting up MCP servers...")
	if err := runMCPInstall([]string{}); err != nil {
		fmt.Printf("⚠️  MCP setup encountered issues: %v\n", err)
	}

	fmt.Println("\n✅ AI tools configuration complete!")
	fmt.Println("   Run 'go run . ai goose session' to start an interactive session")

	return nil
}

// MCP command handlers - moved here to avoid duplication

func runMCPInstall(args []string) error {
	manager, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
	}

	// Load our configuration
	configPath := "pkg/ai/claude-mcp-default.json"
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

func runMCPListClaude() error {
	manager, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to create MCP manager: %w", err)
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