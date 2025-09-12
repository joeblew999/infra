package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/spf13/cobra"
)

// Build-time variables injected via ldflags
var (
	GitHash   = "dev"
	BuildTime = "unknown"
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
  dep            Manage binary dependencies  
  init           Initialize new project

ADVANCED COMMANDS:
  api-check      Check API compatibility between commits
  cli            CLI tool wrappers
  completion     Generate shell autocompletion

Use "infra [command] --help" for detailed information about any command.`,
	Version: getVersionString(),
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Flags().GetString("env")
		RunService(false, false, false, false, env) // Always start all services
	},
}

// SetBuildInfo sets build information for display in web pages
func SetBuildInfo(gitHash, buildTime string) {
	GitHash = gitHash
	BuildTime = buildTime
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// For direct execution, inject git hash at runtime if not set via ldflags
	if GitHash == "dev" {
		if commit := config.GetRuntimeGitHash(); commit != "" {
			GitHash = commit
		}
	}
	if BuildTime == "unknown" {
		BuildTime = time.Now().UTC().Format(time.RFC3339)
	}
	
	// Set build info for display in web pages
	config.SetBuildInfo(GitHash, BuildTime)

	// Skip directory creation and dependency installation in production environments (Fly.io or Docker container)
	isContainer := isRunningInContainer()
	if os.Getenv("FLY_APP_NAME") == "" && !isContainer {
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

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}


// isRunningInContainer checks if we're running inside a Docker container
// by looking for the .dockerenv file that Docker creates
func isRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

// Removed: getRuntimeGitHash now centralized in pkg/build.GetRuntimeGitHash()

// getVersionString returns version info using build info (DRY)
func getVersionString() string {
	// Ensure build info is set (handles runtime injection for direct execution)
	if GitHash == "dev" {
		if runtimeHash := config.GetRuntimeGitHash(); runtimeHash != "" {
			GitHash = runtimeHash
		}
	}
	
	// Use centralized config package for version formatting
	config.SetBuildInfo(GitHash, BuildTime)
	return config.GetFullVersionString()
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