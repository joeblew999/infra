package caddy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/goreman"
)

func init() {
	// Register caddy service factory for decoupled access
	goreman.RegisterService("caddy", func() error {
		return StartSupervised(nil)
	})
}

// Runner executes caddy commands with environment-aware configuration
type Runner struct {
	binaryPath string
}

// New creates a new caddy runner and ensures caddy is installed
func New() *Runner {
	// Ensure caddy binary is installed
	if err := dep.InstallBinary("caddy", false); err != nil {
		// Log warning but continue - binary path will be set regardless
		fmt.Printf("Warning: failed to install caddy binary: %v\n", err)
	}

	binaryPath, _ := filepath.Abs(config.GetCaddyBinPath())
	return &Runner{
		binaryPath: binaryPath,
	}
}

// Run executes a caddy command with the given arguments
func (r *Runner) Run(args ...string) error {
	cmd := exec.Command(r.binaryPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("caddy command failed: %w", err)
	}
	return nil
}

// RunWithOutput executes a caddy command and returns the output
func (r *Runner) RunWithOutput(args ...string) ([]byte, error) {
	cmd := exec.Command(r.binaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("caddy command failed: %w", err)
	}
	return output, nil
}

// StartInBackground starts Caddy in a goroutine with the specified config path
func (r *Runner) StartInBackground(configPath string) {
	go func() {
		if err := r.Run("run", "--config", configPath); err != nil {
			fmt.Printf("Caddy failed: %v\n", err)
		}
	}()
}

// StartWithConfig generates a Caddyfile and starts Caddy in the background
func StartWithConfig(config *CaddyConfig) *Runner {
	// Generate and save Caddyfile
	if err := config.GenerateAndSave("Caddyfile"); err != nil {
		fmt.Printf("Failed to generate Caddyfile: %v\n", err)
		return nil
	}
	
	// Create runner and start in background
	runner := New()
	runner.StartInBackground(".data/caddy/Caddyfile")
	return runner
}

// StartSupervised starts Caddy under goreman supervision (idempotent)
// This is the recommended way to start Caddy in service mode
func StartSupervised(caddyConfig *CaddyConfig) error {
	// Generate Caddyfile if config provided
	var configPath string
	if caddyConfig != nil {
		if err := caddyConfig.GenerateAndSave("Caddyfile"); err != nil {
			return fmt.Errorf("failed to generate Caddyfile: %w", err)
		}
		configPath = filepath.Join(config.GetCaddyPath(), "Caddyfile")
	} else {
		// Default config path
		configPath = ".data/caddy/Caddyfile"
	}
	
	// Ensure config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default development config
		defaultConfig := NewPresetConfig(PresetDevelopment, 80)
		if err := defaultConfig.GenerateAndSave("Caddyfile"); err != nil {
			return fmt.Errorf("failed to create default Caddyfile: %w", err)
		}
		configPath = filepath.Join(config.GetCaddyPath(), "Caddyfile")
	}
	
	// Register and start with goreman supervision
	return goreman.RegisterAndStart("caddy", &goreman.ProcessConfig{
		Command:    config.GetCaddyBinPath(),
		Args:       []string{"run", "--config", configPath},
		WorkingDir: ".",
		Env:        os.Environ(),
	})
}

// FileServer starts a file server with environment-aware HTTPS configuration
// In development: enables HTTPS for localhost
// In production: uses HTTP only (assumes SSL termination by proxy)
func (r *Runner) FileServer(root string, port int) error {
	args := []string{"file-server", "--root", root}

	if config.ShouldUseHTTPS() {
		// Development mode: enable HTTPS
		args = append(args, "--domain", "localhost")
		args = append(args, "--listen", fmt.Sprintf(":%d", port))
	} else {
		// Production mode: HTTP only (SSL terminated by Cloudflare/proxy)
		args = append(args, "--listen", fmt.Sprintf(":%d", port))
	}

	return r.Run(args...)
}

// ReverseProxy starts a reverse proxy with environment-aware HTTPS configuration
func (r *Runner) ReverseProxy(from string, to string) error {
	args := []string{"reverse-proxy"}

	if config.ShouldUseHTTPS() {
		// Development mode: enable HTTPS
		args = append(args, "--from", from, "--to", to)
	} else {
		// Production mode: HTTP only
		args = append(args, "--from", from, "--to", to)
	}

	return r.Run(args...)
}

// ProxyRoute represents a single proxy route configuration
type ProxyRoute struct {
	Path   string // URL path pattern (e.g., "/bento-playground/*")
	Target string // Target URL (e.g., "localhost:4195")
}

// CaddyConfig represents complete Caddy server configuration
type CaddyConfig struct {
	Port   int          // Main listening port
	Target string       // Default reverse proxy target
	Routes []ProxyRoute // Additional proxy routes
}

// DefaultConfig returns a default Caddy configuration
func DefaultConfig() CaddyConfig {
	return CaddyConfig{
		Port:   80,
		Target: "localhost:1337",
		Routes: []ProxyRoute{
			{Path: "/bento-playground/*", Target: "localhost:4195"},
		},
	}
}

// GenerateCaddyfile creates a Caddyfile with environment-specific configuration
func GenerateCaddyfile(cfg CaddyConfig) string {
	var content string

	// Add generation header
	content += "# This Caddyfile is auto-generated by pkg/caddy\n"
	content += "# DO NOT EDIT MANUALLY - changes will be overwritten\n"
	content += "# Generated at: " + time.Now().Format("2006-01-02 15:04:05") + "\n"
	content += "#\n"
	content += "# Configuration:\n"
	content += fmt.Sprintf("# - Port: %d\n", cfg.Port)
	content += fmt.Sprintf("# - Target: %s\n", cfg.Target)
	content += fmt.Sprintf("# - Routes: %d\n", len(cfg.Routes))
	content += "#\n\n"

	if config.ShouldUseHTTPS() {
		// Development: HTTPS with automatic certificates
		content += fmt.Sprintf("localhost:%d {\n", cfg.Port)
	} else {
		// Production: HTTP only
		content += fmt.Sprintf(":%d {\n", cfg.Port)
	}

	// Add specific routes first
	for _, route := range cfg.Routes {
		content += fmt.Sprintf("\thandle %s {\n", route.Path)
		content += fmt.Sprintf("\t\treverse_proxy %s\n", route.Target)
		content += "\t}\n"
	}

	// Add default reverse proxy
	content += fmt.Sprintf("\treverse_proxy %s\n", cfg.Target)

	// Add development-specific headers for cache busting
	if config.ShouldUseHTTPS() {
		content += "\ttls internal\n"
		// Development mode: disable caching to avoid stale content issues
		content += "\theader {\n"
		content += "\t\tCache-Control \"no-cache, no-store, must-revalidate\"\n"
		content += "\t\tPragma \"no-cache\"\n"
		content += "\t\tExpires \"0\"\n"
		content += "\t}\n"
	}

	content += "}"
	return content
}

// GenerateCaddyfileSimple creates a Caddyfile with legacy signature for backward compatibility
func GenerateCaddyfileSimple(port int, targetPort int) string {
	cfg := CaddyConfig{
		Port:   port,
		Target: fmt.Sprintf("localhost:%d", targetPort),
		Routes: []ProxyRoute{
			{Path: "/bento-playground/*", Target: "localhost:4195"},
		},
	}
	return GenerateCaddyfile(cfg)
}
