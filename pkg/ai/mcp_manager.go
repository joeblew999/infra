package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joeblew999/infra/pkg/log"
)

// Server represents an our own MCP server configuration
type Server struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Repo    string            `json:"repo"`
	Type    string            `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// ClaudeServerStatus represents the status of a single MCP server from Claude
type ClaudeServerStatus struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

// ClaudeStatus represents the complete MCP status from Claude
type ClaudeStatus struct {
	Servers []ClaudeServerStatus `json:"servers"`
}

// Config represents the MCP configuration structure
type Config struct {
	Servers []Server `json:"servers"`
}

// Manager handles MCP server management
type Manager struct {
	configPath string
	config     Config
}

// NewManager creates a new MCP manager
func NewManager() (*Manager, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ClaudeConfigDir)
	configPath := filepath.Join(configDir, ClaudeConfigFile)

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	manager := &Manager{
		configPath: configPath,
		config:     Config{Servers: []Server{}},
	}

	// Load existing config if it exists
	if err := manager.loadConfig(); err != nil {
		// If file doesn't exist, we'll create it later
		log.Info("No existing MCP config found, will create new one", "path", configPath)
	}

	return manager, nil
}

// loadConfig loads MCP configuration from file
func (m *Manager) loadConfig() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	return json.Unmarshal(data, &m.config)
}

// saveConfig saves MCP configuration to file
func (m *Manager) saveConfig() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// Install adds MCP servers to configuration
func (m *Manager) Install(servers []Server) error {
	// Add new servers or update existing ones
	for _, newServer := range servers {
		found := false
		for i, existing := range m.config.Servers {
			if existing.Name == newServer.Name {
				m.config.Servers[i] = newServer
				found = true
				break
			}
		}
		if !found {
			m.config.Servers = append(m.config.Servers, newServer)
		}
	}

	return m.saveConfig()
}

// Uninstall removes MCP servers from configuration
func (m *Manager) Uninstall(serverNames []string) error {
	var newServers []Server
	for _, server := range m.config.Servers {
		keep := true
		for _, name := range serverNames {
			if server.Name == name {
				keep = false
				break
			}
		}
		if keep {
			newServers = append(newServers, server)
		}
	}

	m.config.Servers = newServers
	return m.saveConfig()
}

// List returns all configured MCP servers
func (m *Manager) List() []Server {
	return m.config.Servers
}

// GetClaudeStatus queries Claude for actual MCP server status
func (m *Manager) GetClaudeStatus() ([]ClaudeServerStatus, error) {
	cmd := exec.Command("claude", "mcp", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to query Claude MCP status: %w", err)
	}

	return parseClaudeStatus(string(output)), nil
}

// GetClaudeConfigLocations returns Claude's MCP configuration file locations
func (m *Manager) GetClaudeConfigLocations() ([]string, error) {
	cmd := exec.Command("claude", "mcp", "config", "locations")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get Claude config locations: %w", err)
	}

	var locations []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "• ") && strings.Contains(line, ".json") {
			parts := strings.Split(line, ": ")
			if len(parts) > 1 {
				location := strings.TrimSpace(parts[1])
				if strings.HasSuffix(location, ".json") {
					locations = append(locations, location)
				}
			}
		}
	}

	// Fallback to common locations if no locations found
	if len(locations) == 0 {
		commonLocations := CommonClaudeConfigLocations()
		for _, loc := range commonLocations {
			if _, err := os.Stat(loc); err == nil {
				locations = append(locations, loc)
			}
		}
	}

	return locations, nil
}

// LoadConfigFromFile loads configuration from a specific file
func (m *Manager) LoadConfigFromFile(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return m.Install(config.Servers)
}

// parseClaudeStatus parses the actual output format from "claude mcp list"
func parseClaudeStatus(output string) []ClaudeServerStatus {
	var servers []ClaudeServerStatus

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Checking") {
			continue
		}

		// Parse format: "server-name: command - ✓ Connected" or "✗ Error"
		parts := strings.Split(line, " - ")
		if len(parts) >= 2 {
			serverPart := parts[0]
			statusPart := parts[1]

			// Extract server name and command
			serverParts := strings.SplitN(serverPart, ": ", 2)
			if len(serverParts) == 2 {
				name := strings.TrimSpace(serverParts[0])
				command := strings.TrimSpace(serverParts[1])

				status := StatusUnknown
				if strings.Contains(statusPart, "✓") || strings.Contains(statusPart, "Connected") {
					status = StatusRunning
				} else if strings.Contains(statusPart, "✗") || strings.Contains(statusPart, "Error") {
					status = StatusError
				}

				servers = append(servers, ClaudeServerStatus{
					Name:    name,
					Status:  status,
					Command: command,
				})
			}
		}
	}

	return servers
}
