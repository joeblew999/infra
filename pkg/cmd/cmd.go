package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/store"
	"github.com/joeblew999/infra/web"
)

// Run executes the infra application based on command-line arguments.
// It serves as the main entry point for both CLI and service modes.
func Run() {
	// Ensure necessary directories exist
	if err := ensureInfraDirectories(); err != nil {
		log.Fatalf("Failed to ensure infra directories: %v", err)
	}

	debug := flag.Bool("debug", false, "Enable debug features (e.g., use gh cli for dep downloads)")
	mode := flag.String("mode", "service", "Operational mode: 'cli' or 'service'")
	flag.Parse()

	// Ensure core dependencies are in place
	if err := dep.Ensure(*debug); err != nil {
		log.Fatalf("Failed to ensure core dependencies: %v", err)
	}

	switch *mode {
	case "cli":
		runCLI()
	case "service":
		runService()
	default:
		fmt.Printf("Invalid mode: %s. Please use 'cli' or 'service'.\n", *mode)
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

func runCLI() {
	fmt.Println("Running in CLI mode...")
	// TODO: Implement CLI command parsing and execution logic here
	// Example:
	// if len(flag.Args()) > 0 {
	// 	command := flag.Args()[0]
	// 	fmt.Printf("Executing CLI command: %s\n", command)
	// } else {
	// 	fmt.Println("No CLI command specified.")
	// }
}

func runService() {
	fmt.Println("Running in Service mode...")

	if err := web.StartServer(); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}

	log.Println("Service started. Press Ctrl+C to exit.")

	// Set up a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-sigChan

	log.Println("Shutting down service...")
}
