package cmd

import "github.com/spf13/cobra"

var (
	// Cmd is the top-level `infra dep` command exposed to the CLI layer.
	Cmd = &cobra.Command{
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
  dep release:status  - Show release status

Maintenance:
  dep clean           - Clean all dependency system data and caches`,
	}

	localCmd = &cobra.Command{
		Use:   "local",
		Short: "Local development binary management",
		Long:  `Commands for managing binaries in the local development environment.`,
	}

	collectCmd = &cobra.Command{
		Use:   "collect",
		Short: "Multi-platform binary collection",
		Long:  `Commands for collecting binaries across multiple platforms.`,
	}

	releaseCmd = &cobra.Command{
		Use:   "release",
		Short: "Binary distribution and releases",
		Long:  `Commands for publishing collected binaries to managed releases.`,
	}

	cleanCmd = &cobra.Command{
		Use:   "clean",
		Short: "Clean all dependency system data and caches",
		Long: `Clean all dependency system data and caches to debug issues.

This command removes:
  • .dep/ directory (local binaries)
  • .dep/.collection/ directory (multi-platform collections)
  • Generated code (pkg/config/binaries_gen.go)

Use this when the dependency system is having unexplained issues.`,
		Run: func(cmd *cobra.Command, args []string) {
			runClean()
		},
	}
)

func init() {
	Cmd.AddCommand(localCmd)
	Cmd.AddCommand(collectCmd)
	Cmd.AddCommand(releaseCmd)
	Cmd.AddCommand(cleanCmd)

	attachLocalCommands()
	attachCollectCommands()
	attachReleaseCommands()
	configureFlags()
}

func configureFlags() {
	localInstallCmd.Flags().Bool("debug", false, "Enable debug output")
	localInstallCmd.Flags().Bool("cross-platform", false, "Build for all supported platforms (go-build binaries only)")
	localSyncCmd.Flags().Bool("debug", false, "Enable debug output")
}
