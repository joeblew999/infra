package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/joeblew999/infra/pkg/config"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug utilities for container and environment",
	Long:  "Provides debugging utilities for container environments and configuration",
}

var debugEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Show environment variables and configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Environment Debug Report\n")
		fmt.Printf("========================\n\n")
		
		fmt.Printf("Runtime:\n")
		fmt.Printf("  OS: %s\n", runtime.GOOS)
		fmt.Printf("  Arch: %s\n", runtime.GOARCH)
		fmt.Printf("  Go Version: %s\n", runtime.Version())
		
		fmt.Printf("\nEnvironment Variables:\n")
		fmt.Printf("  ENVIRONMENT: %s\n", os.Getenv("ENVIRONMENT"))
		fmt.Printf("  FLY_APP_NAME: %s\n", os.Getenv("FLY_APP_NAME"))
		fmt.Printf("  KO_DOCKER_REPO: %s\n", os.Getenv("KO_DOCKER_REPO"))
		fmt.Printf("  PORT: %s\n", os.Getenv("PORT"))
		
		fmt.Printf("\nConfiguration:\n")
		fmt.Printf("  Production: %v\n", config.IsProduction())
		fmt.Printf("  Data Path: %s\n", config.GetDataPath())
		fmt.Printf("  Logs Path: %s\n", config.GetLogsPath())
		fmt.Printf("  PocketBase Path: %s\n", config.GetPocketBaseDataPath())
	},
}

var debugPathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Show all configured paths",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Path Configuration\n")
		fmt.Printf("==================\n\n")
		
		fmt.Printf("Directories:\n")
		fmt.Printf("  .dep: %s\n", config.GetDepPath())
		fmt.Printf("  .bin: %s\n", config.GetBinPath())
		fmt.Printf("  .data: %s\n", config.GetDataPath())
		fmt.Printf("  .logs: %s\n", config.GetLogsPath())
		fmt.Printf("  docs: %s\n", config.DocsDir)
		fmt.Printf("  build: %s\n", config.GetBuildPath())
		fmt.Printf("  terraform: %s\n", config.GetTerraformPath())
		
		fmt.Printf("\nBinaries:\n")
		fmt.Printf("  tofu: %s\n", config.GetTofuBinPath())
		fmt.Printf("  task: %s\n", config.GetTaskBinPath())
		fmt.Printf("  caddy: %s\n", config.GetCaddyBinPath())
		fmt.Printf("  ko: %s\n", config.GetKoBinPath())
		fmt.Printf("  flyctl: %s\n", config.GetFlyctlBinPath())
		fmt.Printf("  claude: %s\n", config.GetClaudeBinPath())
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.AddCommand(debugEnvCmd)
	debugCmd.AddCommand(debugPathsCmd)
}