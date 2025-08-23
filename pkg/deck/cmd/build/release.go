package build

import (
	"fmt"
	"runtime"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create release packages for GitHub",
	Long: `Build and package deck tools for cross-platform releases.

This creates tar.gz packages for GitHub releases, matching the 
exact same process used in GitHub Actions workflows.

Examples:
  deck release --version v1.0.0
  deck release --version v1.0.0 --os linux --arch amd64
  deck release --version v1.0.0 --all`,
	Run: func(cmd *cobra.Command, args []string) {
		version, _ := cmd.Flags().GetString("version")
		os, _ := cmd.Flags().GetString("os")
		arch, _ := cmd.Flags().GetString("arch")
		all, _ := cmd.Flags().GetBool("all")
		
		if version == "" {
			version = "v1.0.0" // Default version
		}
		
		release := deck.NewRelease(version)
		
		fmt.Printf("Creating deck release %s...\n", version)
		
		if all {
			if err := release.BuildAllTargets(); err != nil {
				fmt.Printf("Error building all targets: %v\n", err)
				return
			}
			fmt.Println("All release packages created successfully!")
		} else if os != "" && arch != "" {
			targetRelease := deck.NewTargetRelease(version, os, arch)
			if err := targetRelease.Build(); err != nil {
				fmt.Printf("Error building %s/%s: %v\n", os, arch, err)
				return
			}
			fmt.Printf("Release package created: %s\n", deck.GetPackageName(version, os, arch))
		} else {
			// Build for current platform
			if err := release.Build(); err != nil {
				fmt.Printf("Error building release: %v\n", err)
				return
			}
			fmt.Printf("Release package created: %s\n", deck.GetPackageName(version, runtime.GOOS, runtime.GOARCH))
		}
	},
}

func init() {
	releaseCmd.Flags().String("version", "v1.0.0", "Release version (e.g., v1.0.0)")
	releaseCmd.Flags().String("os", "", "Target OS (darwin, linux, windows)")
	releaseCmd.Flags().String("arch", "", "Target architecture (amd64, arm64)")
	releaseCmd.Flags().Bool("all", false, "Build for all supported platforms")
	
	BuildCmd.AddCommand(releaseCmd)
}