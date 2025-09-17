package bento

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/spf13/cobra"
)

func NewBentoCmd() *cobra.Command {
	bentoCmd := &cobra.Command{
		Use:   "bento",
		Short: "Manage bento stream processing service",
		Long:  `Commands for managing the bento stream processing service and configuration`,
	}

	bentoCmd.AddCommand(
		newStartCmd(),
		newStopCmd(),
		newConfigCmd(),
		newStatusCmd(),
		newRunCmd(), // Add the new run command
	)

	return bentoCmd
}

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a bento data pipeline from a file",
		Long:  `Executes a bento data pipeline defined in a YAML or JSON file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pipelineFile, _ := cmd.Flags().GetString("pipeline-file")
			if pipelineFile == "" {
				return fmt.Errorf("missing --pipeline-file flag")
			}

			pipelineConfig, err := os.ReadFile(pipelineFile)
			if err != nil {
				return fmt.Errorf("failed to read pipeline file: %w", err)
			}

			port, err := strconv.Atoi(config.GetBentoPort())
			if err != nil {
				return fmt.Errorf("invalid bento port: %w", err)
			}

			service, err := NewService(port) // Use config for port
			if err != nil {
				return fmt.Errorf("failed to create bento service: %w", err)
			}

			if err := service.RunPipeline(pipelineConfig); err != nil {
				return fmt.Errorf("failed to run pipeline: %w", err)
			}

			fmt.Println("Bento pipeline executed successfully")
			return nil
		},
	}
	cmd.Flags().String("pipeline-file", "", "Path to the pipeline definition file (YAML or JSON)")
	return cmd
}

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the bento service",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Ensure config directory and default config exist
			if err := EnsureConfigDir(); err != nil {
				return fmt.Errorf("failed to ensure config directory: %w", err)
			}
			if err := CreateDefaultConfig(); err != nil {
				return fmt.Errorf("failed to ensure default config: %w", err)
			}

			service, err := NewService(4195) // Default bento port
			if err != nil {
				return fmt.Errorf("failed to create bento service: %w", err)
			}

			if err := service.Start(); err != nil {
				return fmt.Errorf("failed to start bento service: %w", err)
			}

			fmt.Println("Bento service started successfully")
			return service.Wait()
		},
	}
}

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the bento service",
		RunE: func(cmd *cobra.Command, args []string) error {
			// This is a placeholder - actual stop would need process management
			fmt.Println("Stop command - implement process management as needed")
			return nil
		},
	}
}

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Manage bento configuration",
		Long:  `Commands for managing bento configuration files`,
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check bento service status",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := config.GetBentoPath()
			configFile := filepath.Join(configDir, "bento.yaml")

			// Ensure config exists (idempotent)
			if err := EnsureConfigDir(); err != nil {
				return fmt.Errorf("failed to ensure config directory: %w", err)
			}
			if err := CreateDefaultConfig(); err != nil {
				return fmt.Errorf("failed to create default config: %w", err)
			}

			fmt.Printf("✅ Bento config: %s\n", configFile)
			fmt.Printf("📁 Config directory: %s\n", configDir)
			fmt.Printf("🌐 HTTP endpoint: http://localhost:4195\n")
			return nil
		},
	}
}
