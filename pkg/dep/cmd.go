package dep

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/joeblew999/infra/pkg/dep/collection"
)

// Cmd is the main dep command that gets added to the root CLI
var Cmd = &cobra.Command{
	Use:   "dep",
	Short: "Manage binary dependencies",
	Long: `Manage binary dependencies across development lifecycle phases:

Local Development:
  dep local:install   - Install/sync binaries to configured versions
  dep local:remove    - Remove installed binary
  dep local:list      - List configured binaries  
  dep local:status    - Show installation status
  dep local:sync      - Sync all binaries to configured versions

Collection (Multi-platform):
  dep collect:binary  - Collect binary for all platforms
  dep collect:all     - Collect all binaries
  dep collect:status  - Show collection status

Distribution:
  dep release:binary  - Publish binary to managed releases
  dep release:all     - Publish all collected binaries
  dep release:status  - Show release status`,
}

// ========================
// Phase Group Commands
// ========================

// Local Development Phase
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Local development binary management",
	Long:  `Commands for managing binaries in local development environment.`,
}

// Collection Phase  
var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Multi-platform binary collection",
	Long:  `Commands for collecting binaries across multiple platforms.`,
}

// Release Phase
var releaseCmd = &cobra.Command{
	Use:   "release", 
	Short: "Binary distribution and releases",
	Long:  `Commands for publishing collected binaries to managed releases.`,
}

// ========================
// Local Phase Commands
// ========================

var localInstallCmd = &cobra.Command{
	Use:   "install [binary]",
	Short: "Install/sync binaries to configured versions",
	Long:  `Install all configured binary dependencies or a specific binary to match dep.json versions.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		debug, _ := cmd.Flags().GetBool("debug")
		crossPlatform, _ := cmd.Flags().GetBool("cross-platform")

		if len(args) == 0 {
			// Install all binaries
			if crossPlatform {
				fmt.Println("Installing all configured binaries (cross-platform mode)...")
			} else {
				fmt.Println("Installing all configured binaries...")
			}
			if err := EnsureWithCrossPlatform(debug, crossPlatform); err != nil {
				fmt.Fprintf(os.Stderr, "Error installing binaries: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ All binaries installed successfully")
		} else {
			// Install specific binary
			binaryName := args[0]
			if crossPlatform {
				fmt.Printf("Installing %s (cross-platform mode)...\n", binaryName)
			} else {
				fmt.Printf("Installing %s...\n", binaryName)
			}
			if err := InstallBinaryWithCrossPlatform(binaryName, debug, crossPlatform); err != nil {
				fmt.Fprintf(os.Stderr, "Error installing %s: %v\n", binaryName, err)
				os.Exit(1)
			}
			fmt.Printf("✓ %s installed successfully\n", binaryName)
		}
	},
}

var localRemoveCmd = &cobra.Command{
	Use:   "remove [binary]",
	Short: "Remove an installed binary",
	Long:  `Remove an installed binary dependency from local development environment.`,
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

var localStatusCmd = &cobra.Command{
	Use:   "status [binary]",
	Short: "Show installation status of binaries",
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

var localSyncCmd = &cobra.Command{
	Use:   "sync [binary]",
	Short: "Sync binaries to configured versions",
	Long:  `Sync all binaries or a specific binary to match their configured versions in dep.json.`,
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

var localListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured binaries",
	Long:  `List all configured binary dependencies with their repositories and versions.`,
	Run: func(cmd *cobra.Command, args []string) {
		binaries, err := LoadConfigForTest()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tREPOSITORY\tVERSION")
		fmt.Fprintln(w, "----\t----------\t-------")
		for _, binary := range binaries {
			fmt.Fprintf(w, "%s\t%s\t%s\n", binary.Name, binary.Repo, binary.Version)
		}
		w.Flush()
	},
}

// ================================
// Collection Phase Commands
// ================================

var collectBinaryCmd = &cobra.Command{
	Use:   "binary [binary] [version]",
	Short: "Collect binary for all platforms",
	Long: `Collect a binary from its original source for all configured platforms.
This downloads the binary for all supported platforms and stores them locally
in preparation for creating managed releases.

Examples:
  go run . dep collect:binary flyctl v0.3.162
  go run . dep collect:binary caddy v2.10.0`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runCollectCommand(args[0], args[1])
	},
}

var collectAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Collect all configured binaries for all platforms",
	Long: `Collect all binaries defined in dep.json for all configured platforms.
This is useful for bulk collection before creating managed releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🔄 Collection system not implemented yet\n")
		fmt.Printf("Would collect all configured binaries for all platforms\n")
	},
}

var collectStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show collection status",
	Long: `Show the status of collected binaries.
This displays what has been collected locally and what platforms are available.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📊 Collection status not implemented yet\n")
		fmt.Printf("Would show collection status\n")
	},
}

// ================================
// Release Phase Commands  
// ================================

var releaseBinaryCmd = &cobra.Command{
	Use:   "binary [binary] [version]",
	Short: "Publish collected binary to managed release",
	Long: `Publish a collected binary to your managed GitHub release.
This uploads all platform variants as release assets.

Examples:
  go run . dep release:binary flyctl v0.3.162
  go run . dep release:binary caddy v2.10.0`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📤 Release system not implemented yet\n")
		fmt.Printf("Would publish %s %s to managed release\n", args[0], args[1])
	},
}

var releaseAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Publish all collected binaries to managed releases",
	Long: `Publish all collected binaries to your managed GitHub releases.
This creates releases and uploads assets for all collected binaries.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📤 Release system not implemented yet\n")
		fmt.Printf("Would publish all collected binaries to managed releases\n")
	},
}

var releaseStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show release status",
	Long: `Show the status of published binaries.
This displays what has been published to managed releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📊 Release status not implemented yet\n")
		fmt.Printf("Would show release status\n")
	},
}

// ========================
// Command Registration
// ========================

func init() {
	// Phase group commands
	Cmd.AddCommand(localCmd)
	Cmd.AddCommand(collectCmd)
	Cmd.AddCommand(releaseCmd)
	
	// Local phase commands
	localCmd.AddCommand(localInstallCmd)
	localCmd.AddCommand(localRemoveCmd)
	localCmd.AddCommand(localListCmd)
	localCmd.AddCommand(localStatusCmd)
	localCmd.AddCommand(localSyncCmd)
	
	// Collection phase commands
	collectCmd.AddCommand(collectBinaryCmd)
	collectCmd.AddCommand(collectAllCmd)
	collectCmd.AddCommand(collectStatusCmd)
	
	// Release phase commands
	releaseCmd.AddCommand(releaseBinaryCmd)
	releaseCmd.AddCommand(releaseAllCmd)
	releaseCmd.AddCommand(releaseStatusCmd)

	// Add debug flag to relevant commands
	localInstallCmd.Flags().Bool("debug", false, "Enable debug output")
	localInstallCmd.Flags().Bool("cross-platform", false, "Build for all supported platforms (go-build binaries only)")
	localSyncCmd.Flags().Bool("debug", false, "Enable debug output")
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
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("📊 Collection Results for %s %s\n", result.Binary, result.Version)
	fmt.Printf("%s\n", strings.Repeat("=", 60))
	
	successCount := 0
	failureCount := 0
	
	// Use tabwriter for consistent column alignment
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PLATFORM\tSTATUS\tSIZE\tDURATION\tERROR")
	fmt.Fprintln(w, "--------\t------\t----\t--------\t-----")
	
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
				if len(errorStr) > 35 {
					errorStr = errorStr[:32] + "..."
				}
			}
		}
		
		durationStr := platformResult.Duration.Round(time.Millisecond).String()
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", 
			platform, status, sizeStr, durationStr, errorStr)
	}
	
	w.Flush()
	fmt.Printf("\n📈 Summary: %d success, %d failed, %s total\n", 
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

