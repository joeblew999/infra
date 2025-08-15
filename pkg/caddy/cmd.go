package caddy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/spf13/cobra"
)

func NewCaddyCmd() *cobra.Command {
	caddyCmd := &cobra.Command{
		Use:   "caddy",
		Short: "Manage Caddy reverse proxy server",
		Long:  `Commands for managing the Caddy reverse proxy server including bento playground integration`,
	}

	caddyCmd.AddCommand(
		newStartCmd(),
		newStopCmd(),
		newStatusCmd(),
		newGenerateCmd(),
	)

	return caddyCmd
}

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start Caddy reverse proxy server",
		Long: `Start Caddy reverse proxy server that provides:
- Main web server proxy (port 80 → 1337)
- Bento playground proxy (port 80/bento-playground → 4195)
- Environment-aware HTTPS/HTTP configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startCaddy()
		},
	}
}

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop Caddy reverse proxy server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return stopCaddy()
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check Caddy reverse proxy status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkCaddyStatus()
		},
	}
}

func newGenerateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate Caddyfile configuration",
		Long: `Generate Caddyfile with bento playground proxy configuration.
The Caddyfile will be saved to .data/caddy/Caddyfile`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateCaddyfile()
		},
	}
}

func startCaddy() error {
	// Ensure caddy directory exists
	caddyDir := config.GetCaddyPath()
	if err := os.MkdirAll(caddyDir, 0755); err != nil {
		return fmt.Errorf("failed to create caddy directory: %w", err)
	}

	// Generate Caddyfile
	caddyfilePath := filepath.Join(caddyDir, "Caddyfile")
	caddyfile := GenerateCaddyfile(80, 1337)
	if err := os.WriteFile(caddyfilePath, []byte(caddyfile), 0644); err != nil {
		return fmt.Errorf("failed to write Caddyfile: %w", err)
	}

	runner := New()
	args := []string{"run", "--config", caddyfilePath}
	
	fmt.Printf("Starting Caddy reverse proxy with bento playground integration...\n")
	fmt.Printf("- Main web server: http://localhost\n")
	fmt.Printf("- Bento playground: http://localhost/bento-playground\n")
	fmt.Printf("- Caddyfile: %s\n", caddyfilePath)
	
	return runner.Run(args...)
}

func stopCaddy() error {
	// Implementation would use caddy stop command
	// For now, this is a placeholder
	fmt.Println("Stopping Caddy reverse proxy...")
	return nil
}

func checkCaddyStatus() error {
	// Implementation would check if caddy is running
	// For now, this is a placeholder
	fmt.Println("Checking Caddy reverse proxy status...")
	return nil
}

func generateCaddyfile() error {
	caddyDir := config.GetCaddyPath()
	if err := os.MkdirAll(caddyDir, 0755); err != nil {
		return fmt.Errorf("failed to create caddy directory: %w", err)
	}

	caddyfilePath := filepath.Join(caddyDir, "Caddyfile")
	caddyfile := GenerateCaddyfile(80, 1337)
	
	if err := os.WriteFile(caddyfilePath, []byte(caddyfile), 0644); err != nil {
		return fmt.Errorf("failed to write Caddyfile: %w", err)
	}

	fmt.Printf("Generated Caddyfile at: %s\n", caddyfilePath)
	fmt.Printf("Configuration:\n%s\n", caddyfile)
	return nil
}