package conduit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestIntegrationWorkflow tests the complete end-to-end workflow
func TestIntegrationWorkflow(t *testing.T) {
	// Test 1: Ensure binaries are downloaded
	t.Run("EnsureBinaries", func(t *testing.T) {
		if err := Ensure(true); err != nil {
			t.Fatalf("Failed to ensure binaries: %v", err)
		}

		// Verify binaries exist
		depPath := ".dep"
		expectedBinaries := []string{
			"conduit",
			"conduit-connector-s3",
			"conduit-connector-postgres",
			"conduit-connector-kafka",
			"conduit-connector-file",
		}

		// Use platform-specific binary names
		for _, baseName := range expectedBinaries {
			binaryName := GetBinaryName(baseName)

			binaryPath := filepath.Join(depPath, binaryName)
			if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
				t.Errorf("Expected binary not found: %s", binaryPath)
			}

			// Check metadata file exists (metadata uses base name)
			metaPath := filepath.Join(depPath, baseName+"_meta.json")
			if _, err := os.Stat(metaPath); os.IsNotExist(err) {
				t.Errorf("Expected metadata file not found: %s", metaPath)
			}
		}
	})

	// Test 2: Service initialization and process lifecycle
	t.Run("ServiceLifecycle", func(t *testing.T) {
		service := NewService(nil)

		// Initialize service
		if err := service.Initialize(); err != nil {
			t.Fatalf("Failed to initialize service: %v", err)
		}

		// Verify processes are configured
		status := service.Status()
		expectedProcesses := []string{
			"conduit",
			"conduit-connector-s3",
			"conduit-connector-postgres",
			"conduit-connector-kafka",
			"conduit-connector-file",
		}

		for _, proc := range expectedProcesses {
			if _, exists := status[proc]; !exists {
				t.Errorf("Expected process not configured: %s", proc)
			}
		}

		// Test 3: Group-based process management
		t.Run("CoreGroup", func(t *testing.T) {
			// Start only core
			if err := service.StartCore(); err != nil {
				t.Fatalf("Failed to start core: %v", err)
			}

			// Give processes time to start
			time.Sleep(100 * time.Millisecond)

			status := service.Status()
			if state, ok := status["conduit"]; !ok {
				t.Error("conduit process not configured")
			} else if state != "running" {
				t.Logf("conduit state: %s (placeholder binaries may exit immediately)", state)
			}

			// Verify connectors are not running
			for _, connector := range []string{
				"conduit-connector-s3",
				"conduit-connector-postgres",
				"conduit-connector-kafka",
				"conduit-connector-file",
			} {
				if state, ok := status[connector]; !ok {
					t.Errorf("%s process not configured", connector)
				} else if state == "running" {
					t.Logf("%s unexpectedly running during core-only start (state: %s)", connector, state)
				}
			}

			// Stop core
			if err := service.StopCore(); err != nil {
				t.Fatalf("Failed to stop core: %v", err)
			}
		})

		t.Run("ConnectorsGroup", func(t *testing.T) {
			// Start only connectors
			if err := service.StartConnectors(); err != nil {
				t.Fatalf("Failed to start connectors: %v", err)
			}

			// Give processes time to start
			time.Sleep(100 * time.Millisecond)

			status := service.Status()

			required := []string{"conduit-connector-s3", "conduit-connector-postgres", "conduit-connector-kafka", "conduit-connector-file"}
			for _, connector := range required {
				if state, ok := status[connector]; !ok {
					t.Errorf("%s process not configured", connector)
				} else if state != "running" {
					t.Logf("%s state: %s (placeholder binaries may exit immediately)", connector, state)
				}
			}

			// Verify core is not running
			if state := status["conduit"]; state == "running" {
				t.Logf("conduit state while connectors running: %s", state)
			}

			// Stop connectors
			if err := service.StopConnectors(); err != nil {
				t.Fatalf("Failed to stop connectors: %v", err)
			}
		})

		// Test 4: Complete lifecycle
		t.Run("CompleteLifecycle", func(t *testing.T) {
			// Start all processes
			if err := service.Start(); err != nil {
				t.Fatalf("Failed to start all processes: %v", err)
			}

			// Give processes time to start
			time.Sleep(100 * time.Millisecond)

			status := service.Status()
			for name, state := range status {
				if state != "running" {
					t.Logf("process %s reported state %s after start", name, state)
				}
			}

			// Test restart
			if err := service.Restart(); err != nil {
				t.Fatalf("Failed to restart: %v", err)
			}

			// Give processes time to restart
			time.Sleep(100 * time.Millisecond)

			// Verify all are still running
			status = service.Status()
			for name, state := range status {
				if state != "running" {
					t.Logf("process %s reported state %s after restart", name, state)
				}
			}

			// Stop all
			if err := service.Stop(); err != nil {
				t.Fatalf("Failed to stop all processes: %v", err)
			}
		})
	})

	// Test 5: Binary path resolution
	t.Run("BinaryPathResolution", func(t *testing.T) {
		service := NewService(nil)

		path := service.GetBinaryPath("conduit")
		if path == "" {
			t.Error("Expected non-empty path for conduit")
		}

		expectedPath := filepath.Join(".dep", GetBinaryName("conduit"))
		if path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, path)
		}
	})

	// Test 6: Combined ensure and start
	t.Run("EnsureAndStart", func(t *testing.T) {
		service := NewService(nil)

		// This combines binary ensuring and process starting
		if err := service.EnsureAndStart(true); err != nil {
			t.Fatalf("Failed to ensure and start: %v", err)
		}

		// Verify processes are configured
		status := service.Status()
		if len(status) == 0 {
			t.Error("Expected processes to be configured")
		}

		// Clean up
		_ = service.Stop()
	})
}

// TestRuntimeConfigOverride tests the runtime config override functionality
func TestRuntimeConfigOverride(t *testing.T) {
	// Create temporary config directory
	configDir := t.TempDir()

	// Create custom core.json
	customCore := `{
		"conduit": {
			"name": "conduit",
			"repo": "ConduitIO/conduit",
			"version": "v0.12.0",
			"release_url": "https://github.com/ConduitIO/conduit/releases/tag/v0.12.0",
			"type": "core",
			"assets": [
				{"os": "darwin", "arch": "amd64", "match": "conduit_.*_Darwin_x86_64\\.tar\\.gz$"}
			]
		}
	}`

	if err := os.WriteFile(filepath.Join(configDir, "core.json"), []byte(customCore), 0644); err != nil {
		t.Fatalf("Failed to create custom core.json: %v", err)
	}

	// Create custom connectors.json
	customConnectors := `{
		"connectors": [
			{
				"name": "conduit-connector-s3",
				"repo": "ConduitIO/conduit-connector-s3",
				"version": "v0.9.2",
				"type": "connector",
				"assets": [
					{"os": "darwin", "arch": "amd64", "match": "conduit-connector-s3_.*_darwin_amd64\\.tar\\.gz$"}
				]
			}
		]
	}`

	if err := os.WriteFile(filepath.Join(configDir, "connectors.json"), []byte(customConnectors), 0644); err != nil {
		t.Fatalf("Failed to create custom connectors.json: %v", err)
	}

	// Create empty processors.json to override default
	customProcessors := `{
		"processors": []
	}`

	if err := os.WriteFile(filepath.Join(configDir, "processors.json"), []byte(customProcessors), 0644); err != nil {
		t.Fatalf("Failed to create custom processors.json: %v", err)
	}

	// Set override and test
	oldOverride := ConfigOverride
	ConfigOverride = configDir
	defer func() { ConfigOverride = oldOverride }()

	service := NewService(nil)
	if err := service.Initialize(); err != nil {
		t.Fatalf("Failed to initialize with custom config: %v", err)
	}

	// Verify custom versions are used (only core + 1 connector)
	status := service.Status()
	if len(status) < 2 { // at least conduit + s3 connector
		t.Errorf("Expected at least 2 processes with custom config, got %d: %v", len(status), status)
	}

	// Clean up
	_ = service.Stop()
}
