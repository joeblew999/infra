package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	dep "github.com/joeblew999/infra/pkg/dep"
	"github.com/spf13/cobra"
)

var (
	localInstallCmd = &cobra.Command{
		Use:   "install [binary]",
		Short: "Install/sync binaries to configured versions",
		Long:  `Install all configured binary dependencies or a specific binary to match dep.json versions.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			crossPlatform, _ := cmd.Flags().GetBool("cross-platform")

			if len(args) == 0 {
				if crossPlatform {
					fmt.Println("Installing all configured binaries (cross-platform mode)...")
				} else {
					fmt.Println("Installing all configured binaries...")
				}
				if err := dep.EnsureWithCrossPlatform(debug, crossPlatform); err != nil {
					fmt.Fprintf(os.Stderr, "Error installing binaries: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("✓ All binaries installed successfully")
				return
			}

			binaryName := args[0]
			if crossPlatform {
				fmt.Printf("Installing %s (cross-platform mode)...\n", binaryName)
			} else {
				fmt.Printf("Installing %s...\n", binaryName)
			}
			if err := dep.InstallBinaryWithCrossPlatform(binaryName, debug, crossPlatform); err != nil {
				fmt.Fprintf(os.Stderr, "Error installing %s: %v\n", binaryName, err)
				os.Exit(1)
			}
			fmt.Printf("✓ %s installed successfully\n", binaryName)
		},
	}

	localRemoveCmd = &cobra.Command{
		Use:   "remove [binary]",
		Short: "Remove an installed binary",
		Long:  `Remove an installed binary dependency from the local development environment.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			binaryName := args[0]
			fmt.Printf("Removing %s...\n", binaryName)
			if err := dep.Remove(binaryName); err != nil {
				fmt.Fprintf(os.Stderr, "Error removing %s: %v\n", binaryName, err)
				os.Exit(1)
			}
			fmt.Printf("✓ %s removed successfully\n", binaryName)
		},
	}

	localStatusCmd = &cobra.Command{
		Use:   "status [binary]",
		Short: "Show installation status of binaries",
		Long:  `Show installation status and version information for all binaries or a specific binary.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				if err := dep.ShowStatus(); err != nil {
					fmt.Fprintf(os.Stderr, "Error showing status: %v\n", err)
					os.Exit(1)
				}
				return
			}

			binaryName := args[0]
			if err := dep.ShowBinaryStatus(binaryName); err != nil {
				fmt.Fprintf(os.Stderr, "Error showing status for %s: %v\n", binaryName, err)
				os.Exit(1)
			}
		},
	}

	localSyncCmd = &cobra.Command{
		Use:   "sync [binary]",
		Short: "Sync binaries to configured versions",
		Long:  `Sync all binaries or a specific binary to match their configured versions in dep.json.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			debug, _ := cmd.Flags().GetBool("debug")

			if len(args) == 0 {
				fmt.Println("Upgrading all binaries...")
				if err := dep.UpgradeAll(debug); err != nil {
					fmt.Fprintf(os.Stderr, "Error upgrading binaries: %v\n", err)
					os.Exit(1)
				}
				fmt.Println("✓ All binaries upgraded successfully")
				return
			}

			binaryName := args[0]
			fmt.Printf("Upgrading %s...\n", binaryName)
			if err := dep.UpgradeBinary(binaryName, debug); err != nil {
				fmt.Fprintf(os.Stderr, "Error upgrading %s: %v\n", binaryName, err)
				os.Exit(1)
			}
			fmt.Printf("✓ %s upgraded successfully\n", binaryName)
		},
	}

	localListCmd = &cobra.Command{
		Use:   "list",
		Short: "List configured binaries",
		Long:  `List all configured binary dependencies with their repositories and versions.`,
		Run: func(cmd *cobra.Command, args []string) {
			binaries, err := dep.LoadConfigForTest()
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
)

func attachLocalCommands() {
	localCmd.AddCommand(localInstallCmd)
	localCmd.AddCommand(localRemoveCmd)
	localCmd.AddCommand(localListCmd)
	localCmd.AddCommand(localStatusCmd)
	localCmd.AddCommand(localSyncCmd)
}
