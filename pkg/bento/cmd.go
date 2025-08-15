package bento

import (
	"fmt"
	"path/filepath"

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
	)

	return bentoCmd
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