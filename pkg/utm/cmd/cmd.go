package cmd

import (
	"fmt"
	"github.com/joeblew999/infra/pkg/utm"
	"github.com/spf13/cobra"
)

func GetUTMCmd() *cobra.Command {
	utmCmd := &cobra.Command{
		Use:   "utm",
		Short: "Launch local UTM VMs (macOS only)",
		Long:  "Minimal helpers for opening UTM and existing .utm bundles on macOS.",
	}

	utmCmd.AddCommand(newUTMListCmd())
	utmCmd.AddCommand(newUTMLaunchCmd())
	utmCmd.AddCommand(newUTMOpenCmd())
	return utmCmd
}

func newUTMListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List discovered UTM virtual machines",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := utm.NewManager()
			if err != nil {
				return err
			}

			vms, err := manager.ListVMs()
			if err != nil {
				return err
			}

			if len(vms) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no .utm bundles found)")
				return nil
			}

			for _, vm := range vms {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", vm.Name, vm.Path)
			}
			return nil
		},
	}
}

func newUTMLaunchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "launch",
		Short: "Open the UTM application",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := utm.NewManager()
			if err != nil {
				return err
			}
			return manager.LaunchApp()
		},
	}
}

func newUTMOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open [name-or-path]",
		Short: "Open a specific VM bundle in UTM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := utm.NewManager()
			if err != nil {
				return err
			}
			return manager.OpenVM(args[0])
		},
	}
}
