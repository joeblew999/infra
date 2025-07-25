package cmd

import (
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "infra",
	Short:   "Infra is a tool for managing infrastructure",
	Long:    `A comprehensive tool for managing infrastructure, including dependencies, services, and more.`,
	Version: "0.0.1",
	Run: func(cmd *cobra.Command, args []string) {
		mode, _ := cmd.Flags().GetString("mode")
		switch mode {
		case "cli":
			// Cobra will handle the subcommands
		case "service":
			RunService(false, mode) // Pass mode to RunService
		default:
			// Default to service mode if no mode or invalid mode is specified
			RunService(false, "service") // Default to "service" mode
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := EnsureInfraDirectories(); err != nil {
		log.Error("Failed to ensure infra directories", "error", err)
		os.Exit(1)
	}

	debug, _ := rootCmd.Flags().GetBool("debug")
	if err := dep.Ensure(debug); err != nil {
		log.Error("Failed to ensure core dependencies", "error", err)
		os.Exit(1)
	}

	// Add subcommands from other files
	RunCLI()
	// RunService() is called directly from rootCmd.Run

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("mode", "", "Set the operating mode (e.g., cli, service)")
}
