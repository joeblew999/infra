package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	rootcmd "github.com/joeblew999/infra/pkg/cmd"
	serviceruntime "github.com/joeblew999/infra/pkg/service/runtime"
	"github.com/spf13/cobra"
)

// GetServiceCmd returns the service command structure
func GetServiceCmd() *cobra.Command {
	serviceCmd := &cobra.Command{
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
				Preflight:    rootcmd.RunDevelopmentPreflightIfNeeded,
			}

			return serviceruntime.Start(opts)
		},
	}

	serviceCmd.Flags().Bool("no-mox", true, "Skip mox mail server (enable with --no-mox=false)")
	serviceCmd.Flags().String("env", "production", "Environment (production/development)")

	return serviceCmd
}

// GetAPICheckCmd returns the API check command
func GetAPICheckCmd() *cobra.Command {
	apiCheckCmd := &cobra.Command{
		Use:   "api-check",
		Short: "Check API compatibility between commits",
		Long: `Check API compatibility between two Git commits using apidiff.
This command helps ensure that public APIs remain backward compatible.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			oldCommit, _ := cmd.Flags().GetString("old")
			newCommit, _ := cmd.Flags().GetString("new")

			return runAPICompatibilityCheck(oldCommit, newCommit)
		},
	}

	apiCheckCmd.Flags().String("old", "HEAD~1", "Old commit to compare against")
	apiCheckCmd.Flags().String("new", "HEAD", "New commit to compare")

	return apiCheckCmd
}

// GetShutdownCmd returns the shutdown command
func GetShutdownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shutdown",
		Short: "Kill running service processes",
		Long:  "Find and kill all running service processes (goreman-supervised and standalone)",
		Run: func(cmd *cobra.Command, args []string) {
			serviceruntime.Shutdown()
		},
	}
}

// GetContainerCmd returns the container command
func GetContainerCmd() *cobra.Command {
	containerCmd := &cobra.Command{
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

	containerCmd.Flags().String("env", "production", "Environment (production/development)")

	return containerCmd
}

// runAPICompatibilityCheck runs the API compatibility check logic
func runAPICompatibilityCheck(oldCommit, newCommit string) error {
	if oldCommit == "" {
		oldCommit = "HEAD~1"
	}
	if newCommit == "" {
		newCommit = "HEAD"
	}

	// Check if this is the first commit - if so, skip API check
	if err := exec.Command("git", "rev-parse", "--verify", oldCommit).Run(); err != nil {
		fmt.Printf("No previous commit found (%s), skipping API compatibility check\n", oldCommit)
		return nil
	}

	fmt.Printf("Checking API compatibility between %s and %s\n", oldCommit, newCommit)

	// Create temporary directories
	oldDir, err := os.MkdirTemp("", "api-check-old-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(oldDir)

	newDir, err := os.MkdirTemp("", "api-check-new-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(newDir)

	// Add worktrees
	if err := exec.Command("git", "worktree", "add", oldDir, oldCommit).Run(); err != nil {
		return fmt.Errorf("failed to checkout %s: %w", oldCommit, err)
	}
	defer exec.Command("git", "worktree", "remove", oldDir, "--force").Run()

	if err := exec.Command("git", "worktree", "add", newDir, newCommit).Run(); err != nil {
		exec.Command("git", "worktree", "remove", oldDir, "--force").Run()
		return fmt.Errorf("failed to checkout %s: %w", newCommit, err)
	}
	defer exec.Command("git", "worktree", "remove", newDir, "--force").Run()

	// Find all Go packages
	packages, err := findGoPackages()
	if err != nil {
		return fmt.Errorf("failed to find Go packages: %w", err)
	}

	breakingChanges := false
	for _, pkg := range packages {
		oldPkgPath := filepath.Join(oldDir, pkg)
		newPkgPath := filepath.Join(newDir, pkg)

		if _, err := os.Stat(oldPkgPath); os.IsNotExist(err) {
			continue
		}
		if _, err := os.Stat(newPkgPath); os.IsNotExist(err) {
			continue
		}

		fmt.Printf("Checking package %s...\n", pkg)
		cmd := exec.Command("apidiff", oldPkgPath, newPkgPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Breaking changes detected in %s\n", pkg)
			breakingChanges = true
		} else {
			fmt.Printf("✅ %s is API compatible\n", pkg)
		}
	}

	if breakingChanges {
		return fmt.Errorf("breaking changes detected")
	}

	fmt.Println("✅ No breaking changes detected")
	return nil
}

// findGoPackages finds all Go packages in the current directory
func findGoPackages() ([]string, error) {
	var packages []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git") {
			return filepath.SkipDir
		}

		if filepath.Ext(path) == ".go" {
			dir := filepath.Dir(path)
			found := false
			for _, pkg := range packages {
				if pkg == dir {
					found = true
					break
				}
			}
			if !found {
				packages = append(packages, dir)
			}
		}

		return nil
	})

	return packages, err
}
