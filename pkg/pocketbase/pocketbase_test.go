package pocketbase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joeblew999/infra/pkg/config"
)

func TestPocketBaseDataDirectory(t *testing.T) {
	// Test data directory creation and isolation
	server := NewServer("test")

	// Get test-isolated data path
	dataPath := config.GetPocketBaseDataPath()

	// Create data directory
	err := os.MkdirAll(dataPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create PocketBase data directory: %v", err)
	}

	t.Logf("✅ PocketBase data directory created: %s", dataPath)

	// Create some test artifacts to simulate database files
	testFiles := []string{
		"data.db",
		"data.db-shm",
		"data.db-wal",
		"logs.db",
	}

	for _, filename := range testFiles {
		testFile := filepath.Join(dataPath, filename)
		err := os.WriteFile(testFile, []byte("test data"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		t.Logf("✅ Test database file: %s", testFile)
	}

	t.Logf("📁 Test artifacts in: %s", dataPath)

	// Verify server configuration
	if server.dataDir != dataPath {
		t.Errorf("Server data dir mismatch: got %s, want %s", server.dataDir, dataPath)
	}

	// Test custom data directory
	customPath := filepath.Join(dataPath, "custom")
	server.SetDataDir(customPath)

	if server.dataDir != customPath {
		t.Errorf("Custom data dir not set: got %s, want %s", server.dataDir, customPath)
	}

	t.Logf("✅ Custom data directory: %s", customPath)
}

func TestPocketBaseConfiguration(t *testing.T) {
	// Test server configuration with test isolation
	server := NewServer("test")

	// Verify environment settings
	if server.env != "test" {
		t.Errorf("Environment not set: got %s, want test", server.env)
	}

	// Verify port configuration
	expectedPort := config.GetPocketBasePort()
	if server.port != expectedPort {
		t.Errorf("Port mismatch: got %s, want %s", server.port, expectedPort)
	}

	// Test URL generation
	appURL := GetAppURL(server.port)
	expectedURL := config.FormatLocalHTTP(server.port)
	if appURL != expectedURL {
		t.Errorf("App URL mismatch: got %s, want %s", appURL, expectedURL)
	}

	apiURL := GetAPIURL(server.port)
	expectedAPIURL := expectedURL + "/api"
	if apiURL != expectedAPIURL {
		t.Errorf("API URL mismatch: got %s, want %s", apiURL, expectedAPIURL)
	}

	t.Logf("✅ App URL: %s", appURL)
	t.Logf("✅ API URL: %s", apiURL)

	// Verify data directory is test-isolated
	dataDir := GetDataDir()
	if !filepath.HasPrefix(dataDir, ".data-test") {
		t.Errorf("Data directory not test-isolated: %s", dataDir)
	}

	t.Logf("✅ Test-isolated data directory: %s", dataDir)
}
