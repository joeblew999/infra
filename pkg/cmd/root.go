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
		env, _ := cmd.Flags().GetString("env")
		RunService(false, false, false, env) // Always start all services
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Skip directory creation and dependency installation in production Fly.io environment
	if os.Getenv("FLY_APP_NAME") == "" {
		if err := EnsureInfraDirectories(); err != nil {
			log.Error("Failed to ensure infra directories", "error", err)
			os.Exit(1)
		}

		debug, _ := rootCmd.Flags().GetBool("debug")
		if err := dep.Ensure(debug); err != nil {
			log.Error("Failed to ensure core dependencies", "error", err)
			os.Exit(1)
		}
	} else {
		log.Info("Running in Fly.io production environment - skipping dependency installation")
	}

	// Add organized command structure
	RunCLI()        // Adds 'cli' namespace with tools
	RunWorkflows()  // Adds core infrastructure commands (deploy, status, init)
	RunDeck()       // Deck commands (if any)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}


func init() {
	rootCmd.PersistentFlags().String("env", "production", "Environment: production or development")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
}