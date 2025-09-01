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
	Short:   "Infrastructure management system with goreman supervision",
	Long:    `Infrastructure Management System

QUICK START:
  infra          Start all services (NATS, Caddy, Bento, Deck API, Web Server)
  infra shutdown Stop all services

INFRASTRUCTURE COMMANDS:
  service        Run infrastructure services with goreman supervision  
  shutdown       Kill running service processes
  status         Check deployment status and health
  deploy         Deploy application using idempotent workflow

DEVELOPMENT COMMANDS:
  config         Print current configuration
  deck           Deck visualization tools
  dep            Manage binary dependencies  
  gozero         Go-zero microservices operations
  init           Initialize new project

ADVANCED COMMANDS:
  api-check      Check API compatibility between commits
  cli            CLI tool wrappers
  completion     Generate shell autocompletion

Use "infra [command] --help" for detailed information about any command.`,
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
	
	// Set custom help template that shows only our organized structure
	rootCmd.SetHelpTemplate(`{{.Long}}

Usage:
  {{.UseLine}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
`)
}