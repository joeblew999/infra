package cmd

import (
	"fmt"
	"strconv"

	"github.com/joeblew999/infra/pkg/caddy"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/spf13/cobra"
)

// GetCaddyCmd returns the caddy command for CLI integration
func GetCaddyCmd() *cobra.Command {
	return caddyCmd
}

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
				return executeBinary(config.GetCaddyBinPath(), args...)
			}
		}
		return executeBinary(config.GetCaddyBinPath(), args...)
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
	portFromStr := strconv.Itoa(fromPort)
	portToStr := strconv.Itoa(toPort)

	from := fmt.Sprintf(":%d", fromPort)
	to := config.FormatLocalHostPort(portToStr)

	if config.ShouldUseHTTPS() {
		from = config.FormatLocalHostPort(portFromStr)
	}

	var url string
	if config.ShouldUseHTTPS() {
		url = config.FormatLocalHTTPS(portFromStr)
	} else {
		url = config.FormatLocalHTTP(portFromStr)
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
	portStr := strconv.Itoa(port)
	if config.ShouldUseHTTPS() {
		url = config.FormatLocalHTTPS(portStr)
	} else {
		url = config.FormatLocalHTTP(portStr)
	}

	fmt.Printf("Starting Caddy file server: %s on port %d (HTTPS: %v)\n",
		root, port, config.ShouldUseHTTPS())
	fmt.Printf("🌐 URL: %s\n", url)

	return runner.FileServer(root, port)
}

// executeBinary executes a binary with the given arguments
func executeBinary(binary string, args ...string) error {
	runner := caddy.New()
	return runner.Run(args...)
}
