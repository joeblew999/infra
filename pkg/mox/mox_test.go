package mox

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joeblew999/infra/pkg/config"
)

func TestMoxServerConfiguration(t *testing.T) {
	// Test mox server configuration with test isolation
	domain := "test.example.com"
	adminEmail := "admin@test.example.com"
	
	server := NewServer(domain, adminEmail)
	
	// Verify server configuration
	if server.domain != domain {
		t.Errorf("Domain mismatch: got %s, want %s", server.domain, domain)
	}
	
	if server.adminEmail != adminEmail {
		t.Errorf("Admin email mismatch: got %s, want %s", server.adminEmail, adminEmail)
	}
	
	t.Logf("✅ Mox server configured: domain=%s, admin=%s", server.domain, server.adminEmail)
}

func TestMoxDataDirectory(t *testing.T) {
	// Test mox data directory creation and isolation
	server := NewServer("test.local", "admin@test.local")
	
	// Initialize to create directories
	err := server.Init()
	if err != nil {
		t.Fatalf("Failed to initialize mox server: %v", err)
	}
	
	// Verify data directory was created
	dataPath := config.GetMoxDataPath()
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		t.Fatalf("Mox data directory not created: %s", dataPath)
	}
	
	t.Logf("✅ Mox data directory created: %s", dataPath)
	
	// Verify config directory structure
	configPath := config.GetMoxConfigPath()
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Fatalf("Mox config directory not created: %s", configDir)
	}
	
	t.Logf("✅ Mox config directory ready: %s", configDir)
	t.Logf("✅ Mox config file path: %s", configPath)
}

func TestMoxConfigGeneration(t *testing.T) {
	// Test mox configuration file generation
	domain := "mail.test.local"
	adminEmail := "postmaster@mail.test.local"
	
	server := NewServer(domain, adminEmail)
	
	// Initialize server
	err := server.Init()
	if err != nil {
		t.Fatalf("Failed to initialize mox server: %v", err)
	}
	
	// Generate configuration
	err = server.GenerateConfig()
	if err != nil {
		t.Fatalf("Failed to generate mox config: %v", err)
	}
	
	// Verify config file exists
	configPath := config.GetMoxConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Mox config file not created: %s", configPath)
	}
	
	// Read and verify config content
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read mox config: %v", err)
	}
	
	configStr := string(configContent)
	if !strings.Contains(configStr, domain) {
		t.Errorf("Config does not contain domain: %s", domain)
	}
	
	if !strings.Contains(configStr, adminEmail) {
		t.Errorf("Config does not contain admin email: %s", adminEmail)
	}
	
	t.Logf("✅ Mox config generated: %s (%d bytes)", configPath, len(configContent))
	t.Logf("📁 Test artifacts in: %s", config.GetMoxDataPath())
}

func TestMoxEnvironmentIsolation(t *testing.T) {
	// Test that mox uses test-isolated paths
	moxDataPath := config.GetMoxDataPath()
	
	// Verify test isolation
	if !strings.HasPrefix(moxDataPath, ".data-test") {
		t.Errorf("Mox data path not test-isolated: %s", moxDataPath)
	}
	
	t.Logf("✅ Test-isolated mox path: %s", moxDataPath)
	
	// Create test artifacts to demonstrate isolation
	testFile := filepath.Join(moxDataPath, "test-isolation.txt")
	err := os.MkdirAll(filepath.Dir(testFile), 0755)
	if err != nil {
		t.Fatalf("Failed to create mox test directory: %v", err)
	}
	
	isolationData := "Mox Test Isolation Verification:\n"
	isolationData += "Environment: Test\n"
	isolationData += "Data Path: " + moxDataPath + "\n"
	isolationData += "Config Path: " + config.GetMoxConfigPath() + "\n"
	isolationData += "Binary Path: " + config.GetMoxBinPath() + "\n"
	
	err = os.WriteFile(testFile, []byte(isolationData), 0644)
	if err != nil {
		t.Fatalf("Failed to save isolation test: %v", err)
	}
	
	t.Logf("✅ Isolation test saved: %s", testFile)
}

func TestMoxBinaryPath(t *testing.T) {
	// Test mox binary path configuration
	binaryPath := config.GetMoxBinPath()
	
	if binaryPath == "" {
		t.Fatal("Mox binary path is empty")
	}
	
	// The binary should be in .dep directory
	if !strings.Contains(binaryPath, ".dep") {
		t.Errorf("Mox binary path should contain .dep: %s", binaryPath)
	}
	
	if !strings.Contains(binaryPath, "mox") {
		t.Errorf("Mox binary path should contain mox: %s", binaryPath)
	}
	
	t.Logf("✅ Mox binary path: %s", binaryPath)
	
	// Check if binary exists (it should since we installed it earlier)
	if _, err := os.Stat(binaryPath); err == nil {
		t.Logf("✅ Mox binary exists at: %s", binaryPath)
	} else {
		t.Logf("ℹ️  Mox binary not found (expected in test): %s", binaryPath)
	}
}