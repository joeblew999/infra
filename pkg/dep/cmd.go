package dep

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dep",
	Short: "Manage binary dependencies",
	Long:  `Manage binary dependencies including installation, updates, and removal.`,
}

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
			
			// Use the new single binary installation
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
	Long:  `Remove a specific binary and its metadata.`,
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

var depCheckCmd = &cobra.Command{
	Use:   "check [binary]",
	Short: "Check for updates",
	Long:  `Check for latest versions of all binaries or a specific binary.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Check all binaries
			if err := CheckAllReleases(); err != nil {
				fmt.Fprintf(os.Stderr, "Error checking releases: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Check specific binary
			fmt.Printf("Checking %s...\n", args[0])
			fmt.Printf("Note: Individual binary checking requires enhancement\n")
		}
	},
}

var depListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured binaries",
	Long:  `List all configured binary dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Configured binaries:")
		fmt.Println("  bento - warpstreamlabs/bento")
		fmt.Println("  task - go-task/task")
		fmt.Println("  tofu - opentofu/opentofu")
		fmt.Println("  caddy - caddyserver/caddy")
		fmt.Println("  ko - ko-build/ko")
		fmt.Println("  flyctl - superfly/flyctl")
		fmt.Println("  garble - burrowers/garble")
		fmt.Println("  bun - oven-sh/bun")
		fmt.Println("  claude - anthropics/claude-code")
		fmt.Println("  nats - nats-io/natscli")
		fmt.Println("  litestream - benbjohnson/litestream")
	},
}

func init() {
	// Add subcommands to dep
	Cmd.AddCommand(depInstallCmd)
	Cmd.AddCommand(depRemoveCmd)
	Cmd.AddCommand(depListCmd)
	Cmd.AddCommand(depCheckCmd)
	
	// Add debug flag to install command
	depInstallCmd.Flags().Bool("debug", false, "Enable debug output")
	depCheckCmd.Flags().Bool("debug", false, "Enable debug output")
}