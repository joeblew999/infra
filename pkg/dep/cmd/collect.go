package cmd

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

var (
	collectBinaryCmd = &cobra.Command{
		Use:   "binary [binary] [version]",
		Short: "Collect binary for all platforms",
		Long: `Collect a binary from its original source for all configured platforms.
This downloads the binary for all supported platforms and stores them locally
in preparation for creating managed releases.

Examples:
  go run . tools dep collect:binary flyctl v0.3.162
  go run . tools dep collect:binary caddy v2.10.0`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			runCollect(args[0], args[1])
		},
	}

	collectAllCmd = &cobra.Command{
		Use:   "all",
		Short: "Collect all configured binaries for all platforms",
		Long: `Collect all binaries defined in dep.json for all configured platforms.
This is useful for bulk collection before creating managed releases.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🔄 Collection system not implemented yet")
			fmt.Println("Would collect all configured binaries for all platforms")
		},
	}

	collectStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show collection status",
		Long: `Show the status of collected binaries.
This displays what has been collected locally and what platforms are available.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("📊 Collection status not implemented yet")
			fmt.Println("Would show collection status")
		},
	}
)

func attachCollectCommands() {
	collectCmd.AddCommand(collectBinaryCmd)
	collectCmd.AddCommand(collectAllCmd)
	collectCmd.AddCommand(collectStatusCmd)
}

func runCollect(binaryName, version string) {
	ctx := context.Background()

	fmt.Printf("🔄 Starting cross-platform collection of %s %s...\n", binaryName, version)
	fmt.Printf("📁 Collection directory: %s\n", ".dep/.collection")
	fmt.Printf("🔧 Target platforms: darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64, windows-arm64\n\n")

	config := collection.DefaultConfig()
	collector, err := collection.NewCrossPlatformCollector(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error creating collector: %v\n", err)
		os.Exit(1)
	}

	start := time.Now()
	result, err := collector.CollectBinary(ctx, binaryName, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error collecting binary: %v\n", err)
		os.Exit(1)
	}

	displayCollectionResult(result, time.Since(start))
}

func displayCollectionResult(result *collection.CollectionResult, duration time.Duration) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("📊 Collection Results for %s %s\n", result.Binary, result.Version)
	fmt.Printf("%s\n", strings.Repeat("=", 60))

	successCount := 0
	failureCount := 0

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PLATFORM\tSTATUS\tSIZE\tDURATION\tERROR")
	fmt.Fprintln(w, "--------\t------\t----\t--------\t-----")

	for platform, platformResult := range result.Platforms {
		status := "❌ FAILED"
		sizeStr := "-"
		errorStr := ""

		if platformResult.Success {
			status = "✅ SUCCESS"
			sizeStr = formatBytes(platformResult.Size)
			successCount++
		} else {
			failureCount++
			if platformResult.Error != "" {
				errorStr = platformResult.Error
			}
		}

		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%s\n",
			platform,
			status,
			sizeStr,
			platformResult.Duration.Round(time.Millisecond),
			errorStr,
		)
	}

	w.Flush()

	fmt.Println("\nSummary:")
	fmt.Printf("  ✅ Success: %d\n", successCount)
	fmt.Printf("  ❌ Failed:  %d\n", failureCount)
	fmt.Printf("  ⏱  Duration: %s\n", duration.Round(time.Second))

	fmt.Printf("\nArtifacts stored in .dep/.collection/%s/%s\n", result.Binary, result.Version)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
