package cmd

import "fmt"

import "github.com/spf13/cobra"

var (
	releaseBinaryCmd = &cobra.Command{
		Use:   "binary [binary] [version]",
		Short: "Publish collected binary to managed release",
		Long: `Publish a collected binary to your managed GitHub release.
This uploads all platform variants as release assets.

Examples:
  go run . tools dep release:binary flyctl v0.3.162
  go run . tools dep release:binary caddy v2.10.0`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("📤 Release system not implemented yet")
			fmt.Printf("Would publish %s %s to managed release\n", args[0], args[1])
		},
	}

	releaseAllCmd = &cobra.Command{
		Use:   "all",
		Short: "Publish all collected binaries to managed releases",
		Long: `Publish all collected binaries to your managed GitHub releases.
This creates releases and uploads assets for all collected binaries.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("📤 Release system not implemented yet")
			fmt.Println("Would publish all collected binaries to managed releases")
		},
	}

	releaseStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show release status",
		Long: `Show the status of published binaries.
This displays what has been published to managed releases.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("📊 Release status not implemented yet")
			fmt.Println("Would show release status")
		},
	}
)

func attachReleaseCommands() {
	releaseCmd.AddCommand(releaseBinaryCmd)
	releaseCmd.AddCommand(releaseAllCmd)
	releaseCmd.AddCommand(releaseStatusCmd)
}
