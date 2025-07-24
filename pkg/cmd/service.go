package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joeblew999/infra/pkg/gops"
	"github.com/joeblew999/infra/web"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Run in service mode",
	Run: func(cmd *cobra.Command, args []string) {
		devDocs, _ := cmd.Flags().GetBool("dev-docs")
		RunService(devDocs)
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.Flags().Bool("dev-docs", false, "Enable development mode for docs (serve from disk)")
}

func RunService(devDocs bool) {
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
