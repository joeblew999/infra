package cmd

import (
	serviceruntime "github.com/joeblew999/infra/pkg/service/runtime"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Run in service mode (same as root command)",
	Long:  "Start all infrastructure services with goreman supervision. This is identical to running the root command without arguments.",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, _ := cmd.Flags().GetString("env")
		noMox, _ := cmd.Flags().GetBool("no-mox")

		opts := serviceruntime.Options{
			Mode:         env,
			NoDevDocs:    true,
			NoNATS:       false,
			NoPocketbase: false,
			NoMox:        noMox,
			Preflight:    RunDevelopmentPreflightIfNeeded,
		}

		return serviceruntime.Start(opts)
	},
}

var apiCheckCmd = &cobra.Command{
	Use:   "api-check",
	Short: "Check API compatibility between commits",
	Long: `Check API compatibility between two Git commits using apidiff.
This command helps ensure that public APIs remain backward compatible.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		oldCommit, _ := cmd.Flags().GetString("old")
		newCommit, _ := cmd.Flags().GetString("new")

		if oldCommit == "" {
			oldCommit = "HEAD~1"
		}
		if newCommit == "" {
			newCommit = "HEAD"
		}

		return runAPICompatibilityCheck(oldCommit, newCommit)
	},
}

var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "Kill running service processes",
	Long:  "Find and kill all running service processes (goreman-supervised and standalone)",
	Run: func(cmd *cobra.Command, args []string) {
		serviceruntime.Shutdown()
	},
}

var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "Build and run containerized service with ko and Docker",
	Long: `Build the application with ko and run it in a Docker container.

This command:
- Builds the container image using ko
- Stops any conflicting containers (idempotent)
- Runs the container with proper port mappings
- Mounts data directory for persistence

This provides a containerized alternative to 'go run . service'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		environment, _ := cmd.Flags().GetString("env")
		return serviceruntime.RunContainer(environment)
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(apiCheckCmd)
	rootCmd.AddCommand(shutdownCmd)
	rootCmd.AddCommand(containerCmd)

	apiCheckCmd.Flags().String("old", "HEAD~1", "Old commit to compare against")
	apiCheckCmd.Flags().String("new", "HEAD", "New commit to compare")

	serviceCmd.Flags().Bool("no-mox", false, "Disable mox mail server")
	serviceCmd.Flags().String("env", "production", "Environment (production/development)")

	containerCmd.Flags().String("env", "production", "Environment (production/development)")
}
