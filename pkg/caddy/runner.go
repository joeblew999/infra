package caddy

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
)

// Runner executes caddy commands with environment-aware configuration
type Runner struct {
	binaryPath string
}

// New creates a new caddy runner
func New() *Runner {
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

// GenerateCaddyfile creates a Caddyfile with environment-specific configuration
func GenerateCaddyfile(port int, targetPort int) string {
	if config.ShouldUseHTTPS() {
		// Development: HTTPS with automatic certificates
		return fmt.Sprintf(`localhost:%d {
	reverse_proxy localhost:%d
	tls internal
}`, port, targetPort)
	} else {
		// Production: HTTP only
		return fmt.Sprintf(`:%d {
	reverse_proxy localhost:%d
}`, port, targetPort)
	}
}