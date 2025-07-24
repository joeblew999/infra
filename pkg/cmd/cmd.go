package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/gops"
	"github.com/joeblew999/infra/pkg/store"
	"github.com/joeblew999/infra/web"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "infra",
	Short:   "Infra is a tool for managing infrastructure",
	Long:    `A comprehensive tool for managing infrastructure, including dependencies, services, and more.`,
	Version: "0.0.1",
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: run as a service
		runService(false)
	},
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Run in service mode",
	Run: func(cmd *cobra.Command, args []string) {
		devDocs, _ := cmd.Flags().GetBool("dev-docs")
		runService(devDocs)
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.Flags().Bool("dev-docs", false, "Enable development mode for docs (serve from disk)")
	rootCmd.AddCommand(tofuCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(caddyCmd)

	rootCmd.PersistentFlags().String("mode", "", "Set the operating mode (e.g., cli, service)")
}

var caddyCmd = &cobra.Command{
	Use:                "caddy",
	Short:              "Run caddy commands",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeBinary(store.GetCaddyBinPath(), args...)
	},
}

var tofuCmd = &cobra.Command{
	Use:   "tofu",
	Short: "Run tofu commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeBinary(store.GetTofuBinPath(), args...)
	},
}

var taskCmd = &cobra.Command{
	Use:                "task",
	Short:              "Run task commands",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeBinary(store.GetTaskBinPath(), args...)
	},
}

// Run executes the infra application based on command-line arguments.
func Run() {
	if err := ensureInfraDirectories(); err != nil {
		log.Fatalf("Failed to ensure infra directories: %v", err)
	}

	debug, _ := rootCmd.Flags().GetBool("debug")
	if err := dep.Ensure(debug); err != nil {
		log.Fatalf("Failed to ensure core dependencies: %v", err)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ensureInfraDirectories() error {
	// Create .dep directory
	if err := os.MkdirAll(store.GetDepPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .dep directory: %w", err)
	}
	log.Printf("Ensured directory exists: %s", store.GetDepPath())

	// Create .bin directory
	if err := os.MkdirAll(store.GetBinPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .bin directory: %w", err)
	}
	log.Printf("Ensured directory exists: %s", store.GetBinPath())

	// Create .data directory
	if err := os.MkdirAll(store.GetDataPath(), 0755); err != nil {
		return fmt.Errorf("failed to create .data directory: %w", err)
	}
	log.Printf("Ensured directory exists: %s", store.GetDataPath())

	// Create taskfiles directory
	if err := os.MkdirAll(store.GetTaskfilesPath(), 0755); err != nil {
		return fmt.Errorf("failed to create taskfiles directory: %w", err)
	}
	log.Printf("Ensured directory exists: %s", store.GetTaskfilesPath())

	return nil
}

func runService(devDocs bool) {
	fmt.Println("Running in Service mode...")

	// Check web server port availability
	if !gops.IsPortAvailable(1337) {
		log.Fatalf("Web server port 1337 is already in use. Please free the port and try again.")
	}

	// Check MCP server port availability
	if !gops.IsPortAvailable(8080) {
		log.Fatalf("MCP server port 8080 is already in use. Please free the port and try again.")
	}

	// Start the web server in a goroutine
	go func() {
		if err := web.StartServer(devDocs); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	// Start the MCP server in a goroutine
	// go func() {
	// 	if err := mcp.StartServer(); err != nil {
	// 		log.Fatalf("Failed to start MCP server: %v", err)
	// 	}
	// }()

	log.Println("Service started. Press Ctrl+C to exit.")

	// Set up a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-sigChan

	log.Println("Shutting down service...")
}

func executeBinary(binary string, args ...string) error {
	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
