package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/joeblew999/infra/pkg/ai"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI-powered infrastructure management via Claude CLI",
	Long:  `Control Claude CLI operations for AI-driven infrastructure management using isolated configuration.`,
}

var aiConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show AI configuration",
	Long:  `Display the isolated Claude configuration being used for AI operations`,
	Run: func(cmd *cobra.Command, args []string) {
		showAIConfig()
	},
}

var aiMCPCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP servers",
	Long:  `List and manage MCP servers configured for AI operations`,
	Run: func(cmd *cobra.Command, args []string) {
		listAIMCP()
	},
}

func showAIConfig() {
	configPath := config.GetClaudeConfigPath()
	fmt.Printf("Using isolated config: %s\n", configPath)
	
	cmd := exec.Command(config.GetClaudeBinPath(), "--settings", configPath, "config", "list")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func listAIMCP() {
	configPath := config.GetClaudeConfigPath()
	fmt.Printf("Using isolated MCP config: %s\n", configPath)
	
	cmd := exec.Command(config.GetClaudeBinPath(), "--mcp-config", configPath, "mcp", "list")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}


// RunAI adds all AI-related commands to the root command
func RunAI() {
	// Get the comprehensive AI command from pkg/ai
	aiFullCmd := ai.NewAICmd()
	
	// Add the existing Claude-specific commands to it
	aiFullCmd.AddCommand(aiConfigCmd)
	aiFullCmd.AddCommand(aiMCPCmd)
	
	rootCmd.AddCommand(aiFullCmd)
}