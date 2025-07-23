package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joeblew999/infra/web"
	// "path/to/your/dep" // Placeholder for dep package import
)

// Run executes the infra application based on command-line arguments.
// It serves as the main entry point for both CLI and service modes.
func Run() {
	mode := flag.String("mode", "service", "Operational mode: 'cli' or 'service'")
	flag.Parse()

	// Placeholder for dep integration
	// err := dep.Ensure()
	// if err != nil {
	// 	log.Fatalf("Failed to ensure dependencies: %v", err)
	// }

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
	<-	sigChan

	log.Println("Shutting down service...")
}
