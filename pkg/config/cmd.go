package config

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Print current configuration",
	Long:  `Print the current configuration including paths, environment, and platform information.`,
	Run:   runConfig,
}

func runConfig(cmd *cobra.Command, args []string) {
	cfg := GetConfig()

	output, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		return
	}

	fmt.Println(string(output))
}