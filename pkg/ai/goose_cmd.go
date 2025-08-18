package ai

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewGooseCmd creates the Goose AI command with all subcommands
func NewGooseCmd() *cobra.Command {
	gooseCmd := &cobra.Command{
		Use:   "goose",
		Short: "Interact with Goose AI",
		Long:  `Direct interface to Goose CLI for AI-powered infrastructure automation`,
	}

	gooseCmd.AddCommand(
		newGooseSessionCmd(),
		newGooseRunCmd(),
		newGooseConfigureCmd(),
		newGooseInfoCmd(),
		newGooseMCPCmd(),
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
			sessionName := ""
			if len(args) > 0 {
				sessionName = args[0]
			}
			return runner.Session(sessionName)
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

func newGooseMCPCmd() *cobra.Command {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP servers for Goose",
		Long:  "Configure and manage MCP servers for Goose AI agent integration",
	}

	mcpCmd.AddCommand(
		newGooseMCPListCmd(),
		newGooseMCPAddCmd(),
	)

	return mcpCmd
}

func newGooseMCPListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List MCP servers",
		Long:  "List MCP servers configured for Goose",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPList()
		},
	}
}

func newGooseMCPAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [name] [command]",
		Short: "Add MCP server",
		Long:  "Add a new MCP server to Goose's configuration",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPAdd(args[0], args[1], args[2:])
		},
	}
}


func newGooseMCPInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install [servers...]",
		Short: "Install MCP servers",
		Long:  "Install MCP servers from the registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPInstall(args)
		},
	}
}

func newGooseMCPUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall [servers...]",
		Short: "Uninstall MCP servers",
		Long:  "Uninstall MCP servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPUninstall(args)
		},
	}
}

func newGooseMCPUninstallAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall-all",
		Short: "Uninstall all MCP servers",
		Long:  "Uninstall all MCP servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCPUninstallAll()
		},
	}
}


// Shared MCP functions
func runMCPList() error {
	fmt.Println("ℹ️  MCP servers are now managed by Claude CLI commands")
	fmt.Println("   Use: go run . ai claude mcp list")
	fmt.Println("   Use: go run . ai claude mcp add [name] [command]")
	return nil
}

func runMCPListLocal() error {
	fmt.Println("ℹ️  MCP servers are now managed by Claude CLI commands")
	fmt.Println("   Use: go run . ai claude mcp list")
	return nil
}

func runMCPUninstall(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please specify server names to uninstall")
	}

	fmt.Printf("ℹ️  Use Claude CLI to remove MCP servers:\n")
	for _, name := range args {
		fmt.Printf("   go run . ai claude mcp remove %s\n", name)
	}
	return nil
}

func runMCPUninstallAll() error {
	fmt.Println("ℹ️  Use Claude CLI to manage MCP servers:")
	fmt.Println("   go run . ai claude mcp list")
	fmt.Println("   go run . ai claude mcp remove [name]")
	return nil
}

func runMCPAdd(name, command string, args []string) error {
	fmt.Printf("ℹ️  Use Claude CLI to add MCP servers:\n")
	fmt.Printf("   go run . ai claude mcp add %s \"%s\"\n", name, command)
	return nil
}