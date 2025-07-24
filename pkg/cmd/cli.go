package cmd

import (
	"github.com/joeblew999/infra/pkg/store"
	"github.com/spf13/cobra"
)

var caddyCmd = &cobra.Command{
	Use:   "caddy",
	Short: "Run caddy commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteBinary(store.GetCaddyBinPath(), args...)
	},
}

var tofuCmd = &cobra.Command{
	Use:   "tofu",
	Short: "Run tofu commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteBinary(store.GetTofuBinPath(), args...)
	},
}

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Run task commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteBinary(store.GetTaskBinPath(), args...)
	},
}

// RunCLI adds all CLI-specific commands to the root command.
func RunCLI() {
	rootCmd.AddCommand(tofuCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(caddyCmd)
}
