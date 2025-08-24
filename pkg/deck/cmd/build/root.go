package build

import (
	"github.com/spf13/cobra"
)

// BuildCmd is the root command for build operations
var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build management commands",
	Long:  "Commands for managing deck tool builds and releases",
}

func init() {
	BuildCmd.AddCommand(installCmd)
	BuildCmd.AddCommand(statusCmd)
	BuildCmd.AddCommand(testCmd)
	BuildCmd.AddCommand(updateSourceCmd)
	
	// Set up flags for update-source command
	updateSourceCmd.Flags().BoolP("force", "f", false, "Force update by removing existing .source directory")
	updateSourceCmd.Flags().BoolP("dry-run", "n", false, "Show what would be done without making changes")
}