package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print current configuration",
	Long:  `Print the current configuration including paths, environment, and platform information.`,
	Run:   runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) {
	cfg := config.GetConfig()

	output, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		return
	}

	fmt.Println(string(output))
}