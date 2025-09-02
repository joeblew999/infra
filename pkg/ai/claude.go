package ai

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
)

//go:embed claude-mcp-default.json
var defaultMCPConfig []byte

// ClaudeConfig represents the Claude configuration structure

// ClaudeMCPConfig represents the MCP server configuration structure
type ClaudeMCPConfig struct {
	Servers []ClaudeMCPServer `json:"servers"`
}

type ClaudeMCPServer struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Repo    string            `json:"repo"`
	Type    string            `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// ClaudeSettings represents the Claude settings structure
type ClaudeSettings struct {
	MCPConfigPath string `json:"mcp_config_path"`
	DefaultModel  string `json:"default_model,omitempty"`
	APIKey        string `json:"api_key,omitempty"`
}

// GetClaudeConfigDir returns the actual path to the Claude config directory
func GetClaudeConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".claude"), nil
}

// GetMCPConfigPath returns the actual path to the Claude MCP config file
func GetMCPConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "claude", "mcp.json"), nil
}

// GetClaudeSettingsPath returns the actual path to the Claude settings file
func GetClaudeSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

// EnsureClaudeDirectories ensures the Claude configuration directories exist
func EnsureClaudeDirectories() error {
	// Ensure .config/claude directory exists
	mcpPath, err := GetMCPConfigPath()
	if err != nil {
		return err
	}
	mcpDir := filepath.Dir(mcpPath)
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		return fmt.Errorf("failed to create Claude MCP config directory: %w", err)
	}

	// Ensure .claude directory exists
	configDir, err := GetClaudeConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create Claude config directory: %w", err)
	}

	return nil
}

// LoadMCPConfig loads the MCP configuration from the standard location
func LoadMCPConfig() (*ClaudeMCPConfig, error) {
	configPath, err := GetMCPConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &ClaudeMCPConfig{Servers: []ClaudeMCPServer{}}, nil
		}
		return nil, fmt.Errorf("failed to read MCP config: %w", err)
	}

	var config ClaudeMCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse MCP config: %w", err)
	}

	return &config, nil
}

// SaveMCPConfig saves the MCP configuration to the standard location
func SaveMCPConfig(config *ClaudeMCPConfig) error {
	configPath, err := GetMCPConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal MCP config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write MCP config: %w", err)
	}

	return nil
}

// LoadClaudeSettings loads the Claude settings from the standard location
func LoadClaudeSettings() (*ClaudeSettings, error) {
	settingsPath, err := GetClaudeSettingsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default settings if file doesn't exist
			return &ClaudeSettings{
				MCPConfigPath: "~/.config/claude/mcp.json",
				DefaultModel:  "claude-3-5-sonnet-20241022",
			}, nil
		}
		return nil, fmt.Errorf("failed to read Claude settings: %w", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse Claude settings: %w", err)
	}

	return &settings, nil
}

// SaveClaudeSettings saves the Claude settings to the standard location
func SaveClaudeSettings(settings *ClaudeSettings) error {
	settingsPath, err := GetClaudeSettingsPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Claude settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write Claude settings: %w", err)
	}

	return nil
}

// RunClaudeConfigure runs the Claude configuration wizard
func RunClaudeConfigure() error {
	fmt.Println("🤖 Configuring Claude AI...")

	// Ensure directories exist
	if err := EnsureClaudeDirectories(); err != nil {
		return err
	}

	// Use ClaudeRunner for consistent binary resolution
	runner := NewClaudeRunner()
	fmt.Printf("✅ Claude CLI: %s\n", runner.binaryPath)

	// Run Claude configuration
	fmt.Println("\n🔧 Running Claude configuration...")
	if err := runner.RunInteractive("auth", "login"); err != nil {
		fmt.Printf("⚠️  Failed to configure Claude: %v\n", err)
		return nil
	}

	// Copy default MCP configuration
	if err := CopyDefaultMCPConfig(); err != nil {
		fmt.Printf("⚠️  Failed to copy MCP configuration: %v\n", err)
	} else {
		fmt.Println("✅ MCP configuration updated")
	}

	fmt.Println("\n✅ Claude configuration complete!")
	fmt.Println("   You can now use Claude with:")
	fmt.Println("   - go run . ai claude session")
	fmt.Println("   - go run . ai claude mcp list")

	return nil
}

// CopyDefaultMCPConfig copies the default MCP configuration to Claude's config
func CopyDefaultMCPConfig() error {
	// Load default configuration
	defaultConfig, err := loadDefaultMCPConfig()
	if err != nil {
		return fmt.Errorf("failed to load default MCP config: %w", err)
	}

	// Save to Claude's configuration
	if err := SaveMCPConfig(defaultConfig); err != nil {
		return fmt.Errorf("failed to save MCP config: %w", err)
	}

	return nil
}

// loadDefaultMCPConfig loads the default MCP configuration from our package
func loadDefaultMCPConfig() (*ClaudeMCPConfig, error) {
	configFile := "pkg/ai/claude-mcp-default.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read default MCP config: %w", err)
	}

	var config ClaudeMCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse default MCP config: %w", err)
	}

	return &config, nil
}

// GetClaudeConfigLocations returns all Claude configuration locations
func GetClaudeConfigLocations() ([]string, error) {
	locations := []string{}

	// MCP config location
	mcpPath, err := GetMCPConfigPath()
	if err != nil {
		return nil, err
	}
	locations = append(locations, mcpPath)

	// Settings location
	settingsPath, err := GetClaudeSettingsPath()
	if err != nil {
		return nil, err
	}
	locations = append(locations, settingsPath)

	// Config directory
	configDir, err := GetClaudeConfigDir()
	if err != nil {
		return nil, err
	}
	locations = append(locations, configDir)

	return locations, nil
}

// ClaudeRunner executes claude commands with proper binary path resolution
type ClaudeRunner struct {
	binaryPath string
}

// NewClaudeRunner creates a new claude runner
func NewClaudeRunner() *ClaudeRunner {
	// Get the claude binary path from dep system
	binaryPath, err := dep.Get("claude")
	if err != nil {
		log.Info("Claude binary not found, attempting to install", "error", err)
		// Try to install claude automatically for idempotent behavior
		if installErr := dep.InstallBinary("claude", false); installErr != nil {
			log.Warn("Could not auto-install claude", "install_error", installErr)
			// Fallback to system claude if available
			binaryPath = "claude"
		} else {
			// Try to get the path again after installation
			binaryPath, err = dep.Get("claude")
			if err != nil {
				log.Warn("Could not get claude path after installation", "error", err)
				binaryPath = "claude"
			}
		}
	}
	
	return &ClaudeRunner{
		binaryPath: binaryPath,
	}
}

// Run executes a claude command with the given arguments
func (r *ClaudeRunner) Run(args ...string) error {
	cmd := exec.Command(r.binaryPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude command failed: %w", err)
	}
	return nil
}

// RunWithOutput executes a claude command and returns the output
func (r *ClaudeRunner) RunWithOutput(args ...string) ([]byte, error) {
	cmd := exec.Command(r.binaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("claude command failed: %w", err)
	}
	return output, nil
}

// RunInteractive executes a claude command with interactive input/output
func (r *ClaudeRunner) RunInteractive(args ...string) error {
	cmd := exec.Command(r.binaryPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude interactive command failed: %w", err)
	}
	return nil
}

// Session starts or resumes a Claude session
func (r *ClaudeRunner) Session(sessionName string) error {
	args := []string{}
	if sessionName != "" {
		args = append(args, "--session", sessionName)
	}
	
	log.Info("Starting Claude session", "session", sessionName)
	return r.RunInteractive(args...)
}

// RunFile executes Claude commands from a file
func (r *ClaudeRunner) RunFile(filename string) error {
	args := []string{filename}
	
	log.Info("Running Claude from file", "file", filename)
	return r.RunInteractive(args...)
}

// RunStdin executes Claude commands from stdin
func (r *ClaudeRunner) RunStdin() error {
	log.Info("Running Claude from stdin")
	return r.RunInteractive()
}

// Configure runs Claude configuration setup
func (r *ClaudeRunner) Configure() error {
	log.Info("Configuring Claude")
	return r.RunInteractive()
}

// Info displays Claude information
func (r *ClaudeRunner) Info() error {
	return r.RunInteractive("--version")
}

// MCPList lists MCP servers for Claude
func (r *ClaudeRunner) MCPList() error {
	return r.RunInteractive("mcp", "list")
}

// MCPAdd adds an MCP server to Claude
func (r *ClaudeRunner) MCPAdd(name, command string) error {
	return r.RunInteractive("mcp", "add", name, command)
}

// MCPRemove removes an MCP server from Claude
func (r *ClaudeRunner) MCPRemove(name string) error {
	return r.RunInteractive("mcp", "remove", name)
}

// InstallDefaultMCP installs the default MCP servers from config
func (r *ClaudeRunner) InstallDefaultMCP() error {
	// Use embedded default config
	var config ClaudeMCPConfig
	if err := json.Unmarshal(defaultMCPConfig, &config); err != nil {
		return fmt.Errorf("failed to parse embedded config: %w", err)
	}

	for _, server := range config.Servers {
		fullCommand := server.Command + " " + strings.Join(server.Args, " ")
		
		if err := r.MCPAdd(server.Name, fullCommand); err != nil {
			return fmt.Errorf("failed to install %s: %w", server.Name, err)
		}
		fmt.Printf("✅ Installed %s: %s\n", server.Name, fullCommand)
	}

	fmt.Println("🎉 Default MCP servers installed successfully!")
	return nil
}

// PresetList lists all available preset MCP servers from the default config
func (r *ClaudeRunner) PresetList() error {
	// Use embedded default config
	var config ClaudeMCPConfig
	if err := json.Unmarshal(defaultMCPConfig, &config); err != nil {
		return fmt.Errorf("failed to parse embedded config: %w", err)
	}

	fmt.Println("📋 Available Preset MCP Servers")
	fmt.Println(strings.Repeat("=", 35))
	
	if len(config.Servers) == 0 {
		fmt.Println("No preset servers found.")
		return nil
	}

	for _, server := range config.Servers {
		fmt.Printf("\n🔌 %s\n", server.Name)
		fmt.Printf("   Version: %s\n", server.Version)
		fmt.Printf("   Repo: %s\n", server.Repo)
		fmt.Printf("   Command: %s %s\n", server.Command, strings.Join(server.Args, " "))
		if len(server.Env) > 0 {
			fmt.Println("   Environment:")
			for key, value := range server.Env {
				fmt.Printf("     %s: %s\n", key, value)
			}
		}
	}

	fmt.Printf("\n💡 Install with: go run . cli ai claude mcp preset-install [server-name]\n")
	return nil
}

// InstallMCPByName installs a specific MCP server by name from the default config
func (r *ClaudeRunner) InstallMCPByName(serverName string) error {
	// Use embedded default config
	var config ClaudeMCPConfig
	if err := json.Unmarshal(defaultMCPConfig, &config); err != nil {
		return fmt.Errorf("failed to parse embedded config: %w", err)
	}

	// Find the requested server
	var targetServer *ClaudeMCPServer
	for _, server := range config.Servers {
		if server.Name == serverName {
			targetServer = &server
			break
		}
	}

	if targetServer == nil {
		availableServers := make([]string, len(config.Servers))
		for i, server := range config.Servers {
			availableServers[i] = server.Name
		}
		return fmt.Errorf("server '%s' not found. Available servers: %s", 
			serverName, strings.Join(availableServers, ", "))
	}

	// Install the specific server
	fullCommand := targetServer.Command + " " + strings.Join(targetServer.Args, " ")
	
	if err := r.MCPAdd(targetServer.Name, fullCommand); err != nil {
		return fmt.Errorf("failed to install %s: %w", targetServer.Name, err)
	}
	
	fmt.Printf("✅ Installed %s: %s\n", targetServer.Name, fullCommand)
	return nil
}

// DisplayClaudeInfo displays Claude configuration and system information
func DisplayClaudeInfo() error {
	fmt.Println("🤖 Claude AI Configuration")
	fmt.Println(strings.Repeat("=", 30))

	// Check Claude availability using dep system
	runner := NewClaudeRunner()
	fmt.Printf("✅ Claude CLI: %s\n", runner.binaryPath)

	fmt.Println("\n📁 Configuration Files:")

	// MCP config location
	mcpPath, err := GetMCPConfigPath()
	if err != nil {
		fmt.Printf("❌ MCP Config: %v\n", err)
	} else {
		fmt.Printf("✅ MCP Config: %s\n", mcpPath)
		if _, err := os.Stat(mcpPath); err == nil {
			config, err := LoadMCPConfig()
			if err != nil {
				fmt.Printf("   ⚠️  Error loading: %v\n", err)
			} else {
				fmt.Printf("   MCP Servers: %d configured\n", len(config.Servers))
				for _, server := range config.Servers {
					fmt.Printf("   - %s (%s)\n", server.Name, server.Version)
				}
			}
		} else {
			fmt.Println("   ⚠️  File not found")
		}
	}

	// Settings location
	settingsPath, err := GetClaudeSettingsPath()
	if err != nil {
		fmt.Printf("❌ Settings: %v\n", err)
	} else {
		fmt.Printf("✅ Settings: %s\n", settingsPath)
		if _, err := os.Stat(settingsPath); err == nil {
			settings, err := LoadClaudeSettings()
			if err != nil {
				fmt.Printf("   ⚠️  Error loading: %v\n", err)
			} else {
				fmt.Printf("   Config Path: %s\n", settings.MCPConfigPath)
				fmt.Printf("   Default Model: %s\n", settings.DefaultModel)
			}
		} else {
			fmt.Println("   ⚠️  File not found")
		}
	}

	// Check environment variables
	fmt.Println("\n🔑 Environment Variables:")
	if val := os.Getenv("ANTHROPIC_API_KEY"); val != "" {
		fmt.Println("✅ ANTHROPIC_API_KEY: Set")
	} else {
		fmt.Println("❌ ANTHROPIC_API_KEY: Not set")
	}

	return nil
}
