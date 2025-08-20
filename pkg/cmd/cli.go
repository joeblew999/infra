package cmd

import (
	"fmt"
	"strconv"

	"github.com/joeblew999/infra/pkg/bento"
	"github.com/joeblew999/infra/pkg/caddy"
	"github.com/joeblew999/infra/pkg/conduit"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/nats"
	"github.com/joeblew999/infra/pkg/pocketbase"
	"github.com/spf13/cobra"
)

var caddyCmd = &cobra.Command{
	Use:                "caddy",
	Short:              "Run caddy commands with environment-aware SSL",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for special environment-aware commands
		if len(args) > 0 {
			switch args[0] {
			case "proxy":
				return handleCaddyProxy(args[1:])
			case "serve":
				return handleCaddyServe(args[1:])
			default:
				// Pass through to regular caddy binary
				return ExecuteBinary(config.GetCaddyBinPath(), args...)
			}
		}
		return ExecuteBinary(config.GetCaddyBinPath(), args...)
	},
}

var tofuCmd = &cobra.Command{
	Use:                "tofu",
	Short:              "Run tofu commands",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteBinary(config.GetTofuBinPath(), args...)
	},
}

var taskCmd = &cobra.Command{
	Use:                "task",
	Short:              "Run task commands",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteBinary(config.GetTaskBinPath(), args...)
	},
}

var koCmd = &cobra.Command{
	Use:                "ko",
	Short:              "Run ko commands",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteBinary(config.GetKoBinPath(), args...)
	},
}

var flyctlCmd = &cobra.Command{
	Use:                "flyctl",
	Short:              "Run flyctl commands",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteBinary(config.GetFlyctlBinPath(), args...)
	},
}

// handleCaddyProxy handles environment-aware reverse proxy setup
func handleCaddyProxy(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: caddy proxy <from-port> <to-port>")
	}
	
	fromPort, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid from-port: %v", err)
	}
	
	toPort, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid to-port: %v", err)
	}
	
	runner := caddy.New()
	from := fmt.Sprintf(":%d", fromPort)
	to := fmt.Sprintf("localhost:%d", toPort)
	
	if config.ShouldUseHTTPS() {
		from = fmt.Sprintf("localhost:%d", fromPort)
	}
	
	var url string
	if config.ShouldUseHTTPS() {
		url = fmt.Sprintf("https://localhost:%d", fromPort)
	} else {
		url = fmt.Sprintf("http://localhost:%d", fromPort)
	}
	
	fmt.Printf("Starting Caddy reverse proxy: %s -> %s (HTTPS: %v)\n", 
		from, to, config.ShouldUseHTTPS())
	fmt.Printf("🌐 URL: %s\n", url)
	
	return runner.ReverseProxy(from, to)
}

// handleCaddyServe handles environment-aware file server
func handleCaddyServe(args []string) error {
	root := "."
	port := 8080
	
	if len(args) > 0 {
		root = args[0]
	}
	if len(args) > 1 {
		var err error
		port, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid port: %v", err)
		}
	}
	
	runner := caddy.New()
	
	var url string
	if config.ShouldUseHTTPS() {
		url = fmt.Sprintf("https://localhost:%d", port)
	} else {
		url = fmt.Sprintf("http://localhost:%d", port)
	}
	
	fmt.Printf("Starting Caddy file server: %s on port %d (HTTPS: %v)\n", 
		root, port, config.ShouldUseHTTPS())
	fmt.Printf("🌐 URL: %s\n", url)
	
	return runner.FileServer(root, port)
}

// cliCmd is the parent command for all CLI tool wrappers
var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "CLI tool wrappers",
	Long:  `Access to various CLI tools and utilities.`,
}

// RunCLI adds the CLI parent command and management commands to the root.
func RunCLI() {
	// Add tool wrappers under 'cli' parent command
	cliCmd.AddCommand(tofuCmd)
	cliCmd.AddCommand(taskCmd)
	cliCmd.AddCommand(caddyCmd)
	cliCmd.AddCommand(koCmd)
	cliCmd.AddCommand(flyctlCmd)
	cliCmd.AddCommand(nats.NewNATSCmd())
	cliCmd.AddCommand(bento.NewBentoCmd())
	
	// Add development/build workflow tools
	AddWorkflowsToCLI(cliCmd)
	
	// Add AI commands
	AddAIToCLI(cliCmd)
	
	// Add debug and translation tools
	AddDebugToCLI(cliCmd)
	AddTokiToCLI(cliCmd)
	
	// Add CLI parent command to root
	rootCmd.AddCommand(cliCmd)
	
	// Keep essential management commands at root level
	rootCmd.AddCommand(dep.Cmd)
	rootCmd.AddCommand(config.Cmd)
	
	// Move these to cli namespace
	cliCmd.AddCommand(pocketbase.Cmd)
	cliCmd.AddCommand(conduit.Cmd)
}
