package dep

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/joeblew999/infra/pkg/dep/collection"
)

// Cmd is the main dep command that gets added to the root CLI
var Cmd = &cobra.Command{
	Use:   "dep",
	Short: "Manage binary dependencies",
	Long:  `Manage binary dependencies including installation, updates, and removal.`,
}

// ========================
// Core Dependency Commands
// ========================

var depInstallCmd = &cobra.Command{
	Use:   "install [binary]",
	Short: "Install binary dependencies",
	Long:  `Install all configured binary dependencies or a specific binary.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		debug, _ := cmd.Flags().GetBool("debug")

		if len(args) == 0 {
			// Install all binaries
			fmt.Println("Installing all configured binaries...")
			if err := Ensure(debug); err != nil {
				fmt.Fprintf(os.Stderr, "Error installing binaries: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ All binaries installed successfully")
		} else {
			// Install specific binary
			binaryName := args[0]
			fmt.Printf("Installing %s...\n", binaryName)
			if err := InstallBinary(binaryName, debug); err != nil {
				fmt.Fprintf(os.Stderr, "Error installing %s: %v\n", binaryName, err)
				os.Exit(1)
			}
			fmt.Printf("✓ %s installed successfully\n", binaryName)
		}
	},
}

var depRemoveCmd = &cobra.Command{
	Use:   "remove [binary]",
	Short: "Remove a binary",
	Long:  `Remove an installed binary dependency.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		binaryName := args[0]
		fmt.Printf("Removing %s...\n", binaryName)
		if err := Remove(binaryName); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing %s: %v\n", binaryName, err)
			os.Exit(1)
		}
		fmt.Printf("✓ %s removed successfully\n", binaryName)
	},
}

var depStatusCmd = &cobra.Command{
	Use:   "status [binary]",
	Short: "Show status of installed binaries",
	Long:  `Show installation status and version information for all binaries or a specific binary.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := ShowStatus(); err != nil {
				fmt.Fprintf(os.Stderr, "Error showing status: %v\n", err)
				os.Exit(1)
			}
		} else {
			binaryName := args[0]
			if err := ShowBinaryStatus(binaryName); err != nil {
				fmt.Fprintf(os.Stderr, "Error showing status for %s: %v\n", binaryName, err)
				os.Exit(1)
			}
		}
	},
}

var depUpgradeCmd = &cobra.Command{
	Use:   "upgrade [binary]",
	Short: "Upgrade binaries to latest versions",
	Long:  `Upgrade all binaries or a specific binary to their latest versions.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		debug, _ := cmd.Flags().GetBool("debug")

		if len(args) == 0 {
			// Upgrade all binaries
			fmt.Println("Upgrading all binaries...")
			if err := UpgradeAll(debug); err != nil {
				fmt.Fprintf(os.Stderr, "Error upgrading binaries: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ All binaries upgraded successfully")
		} else {
			// Upgrade specific binary
			binaryName := args[0]
			fmt.Printf("Upgrading %s...\n", binaryName)
			if err := UpgradeBinary(binaryName, debug); err != nil {
				fmt.Fprintf(os.Stderr, "Error upgrading %s: %v\n", binaryName, err)
				os.Exit(1)
			}
			fmt.Printf("✓ %s upgraded successfully\n", binaryName)
		}
	},
}

var depListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured binaries",
	Long:  `List all configured binary dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		binaries, err := LoadConfigForTest()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Configured binaries:")
		for _, binary := range binaries {
			fmt.Printf("  %s - %s\n", binary.Name, binary.Repo)
		}
	},
}

// ================================
// Managed Binary Distribution System
// ================================

var depCollectCmd = &cobra.Command{
	Use:   "collect [binary] [version]",
	Short: "Collect binaries for all platforms",
	Long: `Collect a binary from its original source for all configured platforms.
This downloads the binary for all supported platforms and stores them locally
in preparation for creating managed releases.

Examples:
  go run . dep collect flyctl v0.3.162
  go run . dep collect caddy v2.10.0`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runCollectCommand(args[0], args[1])
	},
}

var depCollectAllCmd = &cobra.Command{
	Use:   "collect-all",
	Short: "Collect all configured binaries for all platforms",
	Long: `Collect all binaries defined in dep.json for all configured platforms.
This is useful for bulk collection before creating managed releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🔄 Collection system not implemented yet\n")
		fmt.Printf("Would collect all configured binaries for all platforms\n")
	},
}

var depReleaseCmd = &cobra.Command{
	Use:   "release [binary] [version]",
	Short: "Publish collected binary to managed release",
	Long: `Publish a collected binary to your managed GitHub release.
This uploads all platform variants as release assets.

Examples:
  go run . dep release flyctl v0.3.162
  go run . dep release caddy v2.10.0`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📤 Release system not implemented yet\n")
		fmt.Printf("Would publish %s %s to managed release\n", args[0], args[1])
	},
}

var depReleaseAllCmd = &cobra.Command{
	Use:   "release-all",
	Short: "Publish all collected binaries to managed releases",
	Long: `Publish all collected binaries to your managed GitHub releases.
This creates releases and uploads assets for all collected binaries.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📤 Release system not implemented yet\n")
		fmt.Printf("Would publish all collected binaries to managed releases\n")
	},
}

var depCollectionStatusCmd = &cobra.Command{
	Use:   "collection-status",
	Short: "Show collection and release status",
	Long: `Show the status of collected and published binaries.
This displays what has been collected locally and what has been published
to managed releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📊 Collection status not implemented yet\n")
		fmt.Printf("Would show collection and release status\n")
	},
}

// ========================
// Command Registration
// ========================

func init() {
	// Core dependency commands
	Cmd.AddCommand(depInstallCmd)
	Cmd.AddCommand(depRemoveCmd)
	Cmd.AddCommand(depListCmd)
	Cmd.AddCommand(depStatusCmd)
	Cmd.AddCommand(depUpgradeCmd)
	
	// Managed distribution system commands
	Cmd.AddCommand(depCollectCmd)
	Cmd.AddCommand(depCollectAllCmd)
	Cmd.AddCommand(depReleaseCmd)
	Cmd.AddCommand(depReleaseAllCmd)
	Cmd.AddCommand(depCollectionStatusCmd)

	// Add debug flag to relevant commands
	depInstallCmd.Flags().Bool("debug", false, "Enable debug output")
	depUpgradeCmd.Flags().Bool("debug", false, "Enable debug output")
}

// ========================
// Collection Command Implementation
// ========================

func runCollectCommand(binaryName, version string) {
	ctx := context.Background()
	
	fmt.Printf("🔄 Starting cross-platform collection of %s %s...\n", binaryName, version)
	fmt.Printf("📁 Collection directory: %s\n", ".dep/.collection")
	fmt.Printf("🔧 Target platforms: darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64, windows-arm64\n\n")
	
	// Create cross-platform collector
	config := collection.DefaultConfig()
	collector, err := collection.NewCrossPlatformCollector(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error creating collector: %v\n", err)
		os.Exit(1)
	}
	
	// Start collection
	startTime := time.Now()
	result, err := collector.CollectBinary(ctx, binaryName, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error collecting binary: %v\n", err)
		os.Exit(1)
	}
	
	// Display results
	displayCollectionResult(result, time.Since(startTime))
}

func displayCollectionResult(result *collection.CollectionResult, duration time.Duration) {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("📊 Collection Results for %s %s\n", result.Binary, result.Version)
	fmt.Printf(strings.Repeat("=", 60) + "\n")
	
	successCount := 0
	failureCount := 0
	
	fmt.Printf("%-15s %-8s %-10s %-12s %s\n", "PLATFORM", "STATUS", "SIZE", "DURATION", "ERROR")
	fmt.Printf(strings.Repeat("-", 70) + "\n")
	
	for platform, platformResult := range result.Platforms {
		status := "❌ FAILED"
		sizeStr := "-"
		errorStr := ""
		
		if platformResult.Success {
			status = "✅ SUCCESS"
			sizeStr = formatBytesForCollection(platformResult.Size)
			successCount++
		} else {
			failureCount++
			if platformResult.Error != "" {
				errorStr = platformResult.Error
				if len(errorStr) > 40 {
					errorStr = errorStr[:37] + "..."
				}
			}
		}
		
		durationStr := platformResult.Duration.Round(time.Millisecond).String()
		
		fmt.Printf("%-15s %-8s %-10s %-12s %s\n", 
			platform, status, sizeStr, durationStr, errorStr)
	}
	
	fmt.Printf(strings.Repeat("-", 70) + "\n")
	fmt.Printf("📈 Summary: %d success, %d failed, %s total\n", 
		successCount, failureCount, duration.Round(time.Second))
	
	if result.Success {
		fmt.Printf("✅ Collection completed successfully!\n")
		if result.Manifest != nil {
			fmt.Printf("📄 Manifest saved with %d platforms\n", len(result.Manifest.Platforms))
		}
	} else {
		fmt.Printf("❌ Collection failed for all platforms\n")
	}
	
	if len(result.Errors) > 0 {
		fmt.Printf("\n🔍 Errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  • %s\n", err)
		}
	}
	
	fmt.Printf("\n🚀 Next steps:\n")
	fmt.Printf("  • Review collected binaries in .dep/.collection/%s/%s/\n", result.Binary, result.Version)
	fmt.Printf("  • Run 'go run . dep release %s %s' to publish to managed release\n", result.Binary, result.Version)
}

func formatBytesForCollection(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

