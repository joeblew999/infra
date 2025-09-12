// Package process-compose provides a Go interface to the process-compose binary
// for YAML-based process orchestration and management.
//
// Process Compose is a process orchestrator that can start, stop, and manage
// multiple processes based on YAML configuration files. It provides features like
// dependency management, health checks, auto-restart, and log aggregation.
package processcompose

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// Client provides an interface to the process-compose binary
type Client struct {
	binaryPath string
	configPath string
}

// New creates a new process-compose client
func New() (*Client, error) {
	binaryPath := config.Get(config.BinaryProcessCompose)
	if _, err := os.Stat(binaryPath); err != nil {
		return nil, fmt.Errorf("process-compose binary not found at %s: %w", binaryPath, err)
	}

	return &Client{
		binaryPath: binaryPath,
		configPath: filepath.Join(config.GetDataPath(), "process-compose"),
	}, nil
}

// ProcessConfig represents a process configuration for process-compose
type ProcessConfig struct {
	Command     string            `yaml:"command"`
	WorkingDir  string            `yaml:"working_dir,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	RestartMode string            `yaml:"restart,omitempty"`
	HealthCheck *HealthCheck      `yaml:"availability,omitempty"`
}

// HealthCheck represents a health check configuration
type HealthCheck struct {
	RestartOnFailure bool   `yaml:"restart_on_failure,omitempty"`
	HTTPGet          string `yaml:"http_get,omitempty"`
	Timeout          string `yaml:"timeout,omitempty"`
}

// ComposeConfig represents the top-level process-compose configuration
type ComposeConfig struct {
	Version   string                    `yaml:"version"`
	Processes map[string]ProcessConfig  `yaml:"processes"`
}

// ProcessInfo represents information about a running process
type ProcessInfo struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	PID       int    `json:"pid"`
	Restarts  int    `json:"restarts"`
	SystemCPU string `json:"system_cpu"`
	SystemMem string `json:"system_mem"`
}

// Start starts the process-compose daemon with the given configuration
func (c *Client) Start(ctx context.Context, configFile string) error {
	if configFile == "" {
		configFile = filepath.Join(c.configPath, "process-compose.yaml")
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, c.binaryPath, "-f", configFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Info("Starting process-compose", "config", configFile)
	return cmd.Run()
}

// Stop stops the process-compose daemon
func (c *Client) Stop() error {
	cmd := exec.Command(c.binaryPath, "down")
	log.Info("Stopping process-compose")
	return cmd.Run()
}

// Status returns the status of all processes
func (c *Client) Status() ([]ProcessInfo, error) {
	cmd := exec.Command(c.binaryPath, "process", "list", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get process status: %w", err)
	}

	var processes []ProcessInfo
	if err := json.Unmarshal(output, &processes); err != nil {
		return nil, fmt.Errorf("failed to parse process status: %w", err)
	}

	return processes, nil
}

// Scale scales a process to the specified number of replicas
func (c *Client) Scale(processName string, replicas int) error {
	cmd := exec.Command(c.binaryPath, "process", "scale", processName, fmt.Sprintf("%d", replicas))
	log.Info("Scaling process", "name", processName, "replicas", replicas)
	return cmd.Run()
}

// Restart restarts a specific process
func (c *Client) Restart(processName string) error {
	cmd := exec.Command(c.binaryPath, "process", "restart", processName)
	log.Info("Restarting process", "name", processName)
	return cmd.Run()
}

// Logs returns the logs for a specific process
func (c *Client) Logs(processName string, follow bool) (*exec.Cmd, error) {
	args := []string{"logs", processName}
	if follow {
		args = append(args, "--follow")
	}

	cmd := exec.Command(c.binaryPath, args...)
	return cmd, nil
}

// WriteConfig writes a process-compose configuration to a file
func (c *Client) WriteConfig(config ComposeConfig, configFile string) error {
	if configFile == "" {
		configFile = filepath.Join(c.configPath, "process-compose.yaml")
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// For now, we'll use a simple YAML format string
	// In the future, we could use a proper YAML library
	yamlContent := fmt.Sprintf(`version: "%s"

processes:
`, config.Version)

	for name, proc := range config.Processes {
		yamlContent += fmt.Sprintf("  %s:\n", name)
		yamlContent += fmt.Sprintf("    command: %s\n", proc.Command)
		
		if proc.WorkingDir != "" {
			yamlContent += fmt.Sprintf("    working_dir: %s\n", proc.WorkingDir)
		}
		
		if len(proc.Environment) > 0 {
			yamlContent += "    environment:\n"
			for key, value := range proc.Environment {
				yamlContent += fmt.Sprintf("      %s: %s\n", key, value)
			}
		}
		
		if len(proc.DependsOn) > 0 {
			yamlContent += "    depends_on:\n"
			for _, dep := range proc.DependsOn {
				yamlContent += fmt.Sprintf("      - %s\n", dep)
			}
		}
		
		if proc.RestartMode != "" {
			yamlContent += fmt.Sprintf("    restart: %s\n", proc.RestartMode)
		}
		
		yamlContent += "\n"
	}

	return os.WriteFile(configFile, []byte(yamlContent), 0644)
}

// StartSupervised starts process-compose in supervised mode with goreman integration
func StartSupervised(configFile string) error {
	client, err := New()
	if err != nil {
		return fmt.Errorf("failed to create process-compose client: %w", err)
	}

	// For now, we'll start process-compose directly
	// In the future, this could be integrated with goreman
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("Starting process-compose in supervised mode", "config", configFile)
	return client.Start(ctx, configFile)
}

// CreateExampleConfig creates an example process-compose configuration
func CreateExampleConfig() ComposeConfig {
	return ComposeConfig{
		Version: "0.5",
		Processes: map[string]ProcessConfig{
			"web": {
				Command:     "python -m http.server 8000",
				WorkingDir:  ".",
				RestartMode: "always",
				HealthCheck: &HealthCheck{
					HTTPGet:          "http://localhost:8000",
					RestartOnFailure: true,
					Timeout:          "30s",
				},
			},
			"worker": {
				Command:     "python worker.py",
				WorkingDir:  ".",
				DependsOn:   []string{"web"},
				RestartMode: "on_failure",
				Environment: map[string]string{
					"WORKER_ID": "1",
					"LOG_LEVEL": "info",
				},
			},
		},
	}
}