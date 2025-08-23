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
}