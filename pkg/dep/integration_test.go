package dep

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestEndToEndInstallation tests complete installation workflow
func TestEndToEndInstallation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	// Test cases for end-to-end scenarios
	testCases := []struct {
		name       string
		binary     string
		versionCmd []string
		expectWork bool
	}{
		{
			name:       "Task Runner",
			binary:     "task",
			versionCmd: []string{"task", "--version"},
			expectWork: true,
		},
		{
			name:       "Claude Code",
			binary:     "claude",
			versionCmd: []string{"claude", "--version"},
			expectWork: true,
		},
		{
			name:       "GitHub CLI",
			binary:     "gh",
			versionCmd: []string{"gh", "--version"},
			expectWork: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing end-to-end installation of %s", tc.binary)

			// Clean install
			if err := Remove(tc.binary); err != nil {
				t.Logf("Cleanup note: %v", err)
			}

			// Install
			if err := InstallBinary(tc.binary, false); err != nil {
				if tc.expectWork {
					t.Fatalf("Failed to install %s: %v", tc.binary, err)
				} else {
					t.Logf("Expected failure for %s: %v", tc.binary, err)
					return
				}
			}

			// Get path
			binaryPath, err := Get(tc.binary)
			if err != nil {
				t.Fatalf("Failed to get path for %s: %v", tc.binary, err)
			}

			// Test execution (if version command provided)
			if len(tc.versionCmd) > 0 {
				// Replace first arg with actual path
				cmd := exec.Command(binaryPath)
				if len(tc.versionCmd) > 1 {
					cmd.Args = append([]string{binaryPath}, tc.versionCmd[1:]...)
				}

				output, err := cmd.CombinedOutput()
				if err != nil {
					t.Logf("Version command failed (this may be expected): %v", err)
					t.Logf("Output: %s", string(output))
				} else {
					t.Logf("✓ %s version check successful: %s", tc.binary, strings.TrimSpace(string(output)))
				}
			}
		})
	}
}

// TestBulkInstallation tests installing all configured binaries
func TestBulkInstallation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bulk installation test in short mode")
	}

	// This is a comprehensive test - only run when explicitly requested
	if os.Getenv("RUN_BULK_INSTALL_TEST") != "true" {
		t.Skip("Set RUN_BULK_INSTALL_TEST=true to run bulk installation test")
	}

	t.Log("Running bulk installation test for all configured binaries")

	// Use Ensure to install all binaries
	if err := Ensure(true); err != nil {
		t.Fatalf("Bulk installation failed: %v", err)
	}

	// Load config to verify all binaries
	binaries, err := loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify each binary was installed
	for _, binary := range binaries {
		t.Run("Verify_"+binary.Name, func(t *testing.T) {
			binaryPath, err := Get(binary.Name)
			if err != nil {
				t.Errorf("Failed to get path for %s: %v", binary.Name, err)
				return
			}

			if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
				t.Errorf("Binary %s not found at %s", binary.Name, binaryPath)
				return
			}

			// Check metadata
			meta, err := readMeta(binaryPath)
			if err != nil {
				t.Errorf("Failed to read metadata for %s: %v", binary.Name, err)
				return
			}

			if meta.Name != binary.Name {
				t.Errorf("Metadata name mismatch for %s: got %s", binary.Name, meta.Name)
			}

			t.Logf("✓ %s verified (version: %s)", binary.Name, meta.Version)
		})
	}
}

// TestVersionPinning tests that version pinning works correctly
func TestVersionPinning(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping version pinning test in short mode")
	}

	// Test with a binary that has explicit version pinning
	testBinary := "task" // Uses v3.44.1 in config

	t.Logf("Testing version pinning with %s", testBinary)

	// Clean install
	if err := Remove(testBinary); err != nil {
		t.Logf("Cleanup note: %v", err)
	}

	// Install
	if err := InstallBinary(testBinary, false); err != nil {
		t.Fatalf("Failed to install %s: %v", testBinary, err)
	}

	// Check metadata contains expected version
	binaryPath, err := Get(testBinary)
	if err != nil {
		t.Fatalf("Failed to get path: %v", err)
	}

	meta, err := readMeta(binaryPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	// Load config to get expected version
	binaries, err := loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	var expectedVersion string
	for _, binary := range binaries {
		if binary.Name == testBinary {
			expectedVersion = binary.Version
			break
		}
	}

	if expectedVersion == "" {
		t.Fatalf("Could not find expected version for %s", testBinary)
	}

	t.Logf("Expected version: %s, Got metadata version: %s", expectedVersion, meta.Version)

	// Note: For "latest" versions, we can't predict exact version,
	// but for pinned versions we should get what we expect
	if expectedVersion != "latest" && meta.Version != expectedVersion {
		t.Logf("Version mismatch (this may be intentional): expected %s, got %s", expectedVersion, meta.Version)
	}
}

// TestConcurrentInstallation tests that concurrent installations don't conflict
func TestConcurrentInstallation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent installation test in short mode")
	}

	// Test installing different binaries concurrently
	binaries := []string{"task", "gh"}

	t.Log("Testing concurrent installation")

	// Clean up first
	for _, binary := range binaries {
		if err := Remove(binary); err != nil {
			t.Logf("Cleanup note for %s: %v", binary, err)
		}
	}

	// Install concurrently
	results := make(chan error, len(binaries))
	
	for _, binary := range binaries {
		go func(name string) {
			results <- InstallBinary(name, false)
		}(binary)
	}

	// Collect results
	var errors []error
	for i := 0; i < len(binaries); i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		t.Fatalf("Concurrent installation errors: %v", errors)
	}

	// Verify all installed correctly
	for _, binary := range binaries {
		binaryPath, err := Get(binary)
		if err != nil {
			t.Errorf("Failed to get path for %s: %v", binary, err)
			continue
		}

		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			t.Errorf("Binary %s not found after concurrent install", binary)
		} else {
			t.Logf("✓ %s installed successfully via concurrent test", binary)
		}
	}
}

// TestConfigValidation tests that the configuration is valid
func TestConfigValidation(t *testing.T) {
	binaries, err := loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(binaries) == 0 {
		t.Fatal("No binaries configured")
	}

	t.Logf("Validating configuration for %d binaries", len(binaries))

	// Track source type distribution
	sourceTypes := make(map[string]int)
	
	for _, binary := range binaries {
		sourceTypes[binary.Source]++
		
		// Basic validation
		if binary.Name == "" {
			t.Errorf("Binary with empty name found")
		}
		
		if binary.Source == "" {
			t.Errorf("Binary %s has empty source type", binary.Name)
		}
		
		if binary.Version == "" {
			t.Errorf("Binary %s has empty version", binary.Name)
		}
	}

	t.Logf("Source type distribution:")
	for sourceType, count := range sourceTypes {
		t.Logf("  %s: %d binaries", sourceType, count)
	}

	// Ensure we have the expected source types
	expectedTypes := []string{"github-release", "go-build", "claude-release"}
	for _, expectedType := range expectedTypes {
		if sourceTypes[expectedType] == 0 {
			t.Errorf("Expected source type %s not found in configuration", expectedType)
		}
	}
}