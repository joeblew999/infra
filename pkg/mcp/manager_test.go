package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMCPManager(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)

	// Create test manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test initial state
	servers := manager.List()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers initially, got %d", len(servers))
	}

	// Test server installation
	testServers := []Server{
		{
			Name:    "github",
			Version: "v1.0.0",
			Repo:    "test/repo",
			Command: "node",
			Args:    []string{"index.js"},
		},
		{
			Name:    "filesystem",
			Version: "v2.0.0",
			Repo:    "test/repo2",
			Command: "python",
			Args:    []string{"main.py"},
		},
	}

	// Install test servers
	if err := manager.Install(testServers); err != nil {
		t.Fatalf("Failed to install servers: %v", err)
	}

	// Verify installation
	servers = manager.List()
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers after install, got %d", len(servers))
	}

	// Test server update
	updatedServers := []Server{
		{
			Name:    "github",
			Version: "v1.1.0",
			Repo:    "test/repo",
			Command: "node",
			Args:    []string{"index.js"},
		},
	}
	if err := manager.Install(updatedServers); err != nil {
		t.Fatalf("Failed to update server: %v", err)
	}

	// Verify update
	servers = manager.List()
	if servers[0].Version != "v1.1.0" {
		t.Errorf("Expected version v1.1.0, got %s", servers[0].Version)
	}

	// Test server uninstall
	if err := manager.Uninstall([]string{"filesystem"}); err != nil {
		t.Fatalf("Failed to uninstall server: %v", err)
	}

	// Verify uninstall
	servers = manager.List()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server after uninstall, got %d", len(servers))
	}
	if servers[0].Name != "github" {
		t.Errorf("Expected remaining server to be 'github', got %s", servers[0].Name)
	}

	// Verify config file exists
	configPath := filepath.Join(tempDir, ".config", "claude", "mcp.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected config file to exist at %s", configPath)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)

	// Create test config file
	configFile := filepath.Join(tempDir, "test-mcp.json")
	configContent := `{
		"servers": [
			{
				"name": "test-server",
				"version": "v1.0.0",
				"repo": "test/test",
				"command": "echo",
				"args": ["hello"]
			}
		]
	}`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create manager and load config
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if err := manager.LoadConfigFromFile(configFile); err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	// Verify loaded servers
	servers := manager.List()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server after loading from file, got %d", len(servers))
	}
	if servers[0].Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got %s", servers[0].Name)
	}
}

func TestEmptyConfig(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)

	// Create manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Should handle empty config gracefully
	servers := manager.List()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers for empty config, got %d", len(servers))
	}
}

func TestConfigFileCreation(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)

	// Create manager - should create directories
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Verify config directory exists
	configDir := filepath.Join(tempDir, ".config", "claude")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Expected config directory to be created at %s", configDir)
	}

	// Should not create config file until needed
	configPath := filepath.Join(configDir, "mcp.json")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		// File might exist if we have default servers, which is fine
		t.Logf("Config file exists at %s", configPath)
	}
	
	_ = manager // Use the variable
}

func TestUninstallNonExistent(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)

	// Create manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Uninstalling non-existent server should not fail
	if err := manager.Uninstall([]string{"nonexistent"}); err != nil {
		t.Fatalf("Uninstalling non-existent server should not fail: %v", err)
	}

	// Config should remain unchanged
	serversAfter := manager.List()
	if len(serversAfter) != 0 {
		t.Errorf("Expected 0 servers after uninstalling non-existent, got %d", len(serversAfter))
	}
}

func TestUninstallAll(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)

	// Create manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Install some test servers
	testServers := []Server{
		{Name: "server1", Version: "v1.0.0", Command: "echo"},
		{Name: "server2", Version: "v2.0.0", Command: "echo"},
		{Name: "server3", Version: "v3.0.0", Command: "echo"},
	}
	
	if err := manager.Install(testServers); err != nil {
		t.Fatalf("Failed to install test servers: %v", err)
	}

	// Verify servers are installed
	servers := manager.List()
	if len(servers) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(servers))
	}

	// Collect all server names and uninstall
	serverNames := make([]string, len(servers))
	for i, server := range servers {
		serverNames[i] = server.Name
	}
	
	if err := manager.Uninstall(serverNames); err != nil {
		t.Fatalf("Failed to uninstall all servers: %v", err)
	}

	// Verify all servers are uninstalled
	servers = manager.List()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers after uninstalling all, got %d", len(servers))
	}
}

func TestUninstallAllEmpty(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)

	// Create manager
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Uninstall all from empty config should not fail
	servers := manager.List()
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers initially, got %d", len(servers))
	}

	// Collect server names (empty slice)
	serverNames := make([]string, 0)
	
	if err := manager.Uninstall(serverNames); err != nil {
		t.Fatalf("Uninstalling from empty config should not fail: %v", err)
	}
}

func TestClaudeJSONParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []ClaudeServerStatus
	}{
		{
			name: "json format",
			input: `{"servers": [
				{"name": "github", "command": "npx @modelcontextprotocol/server-github", "status": "running"},
				{"name": "playwright", "command": "npx @playwright/mcp@latest", "status": "running"}
			]}`,
			expected: []ClaudeServerStatus{
				{
					Name:    "github",
					Status:  "running",
					Command: "npx @modelcontextprotocol/server-github",
				},
				{
					Name:    "playwright",
					Status:  "running",
					Command: "npx @playwright/mcp@latest",
				},
			},
		},
		{
			name: "json with error",
			input: `{"servers": [
				{"name": "github", "command": "npx @modelcontextprotocol/server-github", "status": "error", "error": "timeout"},
				{"name": "playwright", "command": "npx @playwright/mcp@latest", "status": "running"}
			]}`,
			expected: []ClaudeServerStatus{
				{
					Name:    "github",
					Status:  "error",
					Command: "npx @modelcontextprotocol/server-github",
					Error:   "timeout",
				},
				{
					Name:    "playwright",
					Status:  "running",
					Command: "npx @playwright/mcp@latest",
				},
			},
		},
		{
			name: "empty servers",
			input: `{"servers": []}`,
			expected: []ClaudeServerStatus{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status ClaudeStatus
			err := json.Unmarshal([]byte(tt.input), &status)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}
			
			result := status.Servers
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d servers, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i].Name != expected.Name {
					t.Errorf("Server %d name mismatch: expected %s, got %s", i, expected.Name, result[i].Name)
				}
				if result[i].Status != expected.Status {
					t.Errorf("Server %d status mismatch: expected %s, got %s", i, expected.Status, result[i].Status)
				}
				if result[i].Command != expected.Command {
					t.Errorf("Server %d command mismatch: expected %s, got %s", i, expected.Command, result[i].Command)
				}
				if result[i].Error != expected.Error {
					t.Errorf("Server %d error mismatch: expected %s, got %s", i, expected.Error, result[i].Error)
				}
			}
		})
	}
}