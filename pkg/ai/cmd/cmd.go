package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/ai"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/spf13/cobra"
)

// Register mounts the AI command under the provided parent.
func Register(parent *cobra.Command) {
	parent.AddCommand(NewAICmd())
}

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
	runner := ai.NewGooseRunner()

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

// runMCPInstall handles MCP server installation
func runMCPInstall(args []string) error {
	fmt.Println("ℹ️  MCP servers are now managed by Claude CLI commands")
	fmt.Println("   Use: go run . ai claude mcp add [name] [command]")
	fmt.Println("   Use: go run . ai claude mcp list")
	return nil
}
