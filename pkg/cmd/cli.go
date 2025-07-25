package cmd

import (
	"fmt"
	"strconv"

	"github.com/joeblew999/infra/pkg/caddy"
	"github.com/joeblew999/infra/pkg/config"
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
	
	fmt.Printf("Starting Caddy file server: %s on port %d (HTTPS: %v)\n", 
		root, port, config.ShouldUseHTTPS())
	
	return runner.FileServer(root, port)
}

// RunCLI adds all CLI-specific commands to the root command.
func RunCLI() {
	rootCmd.AddCommand(tofuCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(caddyCmd)
}
