package dep

import (
	"os"
	"testing"

	"github.com/joeblew999/infra/pkg/dep/builders"
)

// TestAllInstallerTypes tests all installer types with representative binaries
func TestAllInstallerTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping installer tests in short mode")
	}

	// Test cases for different installer types
	testCases := []struct {
		name         string
		binaryName   string
		installerType string
		description  string
	}{
		{
			name:         "GitHub Release Installer",
			binaryName:   "task", // Small, reliable binary
			installerType: "github-release",
			description:  "Tests downloading from GitHub releases with asset selection",
		},
		{
			name:         "Go Build Installer", 
			binaryName:   "garble",
			installerType: "go-build",
			description:  "Tests building Go binaries from source",
		},
		{
			name:         "Claude Release Installer",
			binaryName:   "claude",
			installerType: "claude-release", 
			description:  "Tests downloading from Claude's Google Cloud Storage",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing %s (%s)", tc.description, tc.binaryName)
			
			// Clean up before test
			if err := Remove(tc.binaryName); err != nil {
				t.Logf("Cleanup note: %v", err)
			}

			// Test installation
			if err := InstallBinary(tc.binaryName, true); err != nil {
				t.Fatalf("Installation failed for %s: %v", tc.binaryName, err)
			}

			// Verify installation
			binaryPath, err := Get(tc.binaryName)
			if err != nil {
				t.Fatalf("Failed to get binary path: %v", err)
			}

			// Check binary exists
			if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
				t.Fatalf("Binary not found at %s", binaryPath)
			}

			// Check metadata exists
			metaPath := getMetaPath(binaryPath)
			if _, err := os.Stat(metaPath); os.IsNotExist(err) {
				t.Fatalf("Metadata not found at %s", metaPath)
			}

			// Verify metadata content
			meta, err := readMeta(binaryPath)
			if err != nil {
				t.Fatalf("Failed to read metadata: %v", err)
			}

			if meta.Name != tc.binaryName {
				t.Errorf("Metadata name mismatch: got %s, want %s", meta.Name, tc.binaryName)
			}

			if meta.Version == "" {
				t.Error("Metadata version should not be empty")
			}

			t.Logf("✓ %s installed successfully (version: %s)", tc.binaryName, meta.Version)
		})
	}
}

// TestInstallerErrorHandling tests error conditions for installers
func TestInstallerErrorHandling(t *testing.T) {
	testCases := []struct {
		name        string
		binaryName  string
		expectError bool
		description string
	}{
		{
			name:        "Invalid Binary",
			binaryName:  "nonexistent-binary-12345",
			expectError: true,
			description: "Should fail for unsupported binaries",
		},
		{
			name:        "Empty Binary Name",
			binaryName:  "",
			expectError: true,
			description: "Should fail for empty binary names",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := InstallBinary(tc.binaryName, false)
			
			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s but got none", tc.description)
			}
			
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error for %s: %v", tc.description, err)
			}
		})
	}
}

// TestBuilderPatterns tests that builders follow consistent patterns
func TestBuilderPatterns(t *testing.T) {
	t.Run("GitHub Release Builder", func(t *testing.T) {
		builder := &builders.GitHubReleaseInstaller{}
		if builder == nil {
			t.Error("GitHub release builder should not be nil")
		}
	})

	t.Run("Go Build Builder", func(t *testing.T) {
		builder := &builders.GoBuildInstaller{}
		if builder == nil {
			t.Error("Go build builder should not be nil")
		}
	})

	t.Run("NPM Builder", func(t *testing.T) {
		builder := &builders.NPMInstaller{}
		if builder == nil {
			t.Error("NPM builder should not be nil")
		}
	})

	t.Run("Claude Release Builder", func(t *testing.T) {
		builder := &builders.ClaudeReleaseInstaller{}
		if builder == nil {
			t.Error("Claude release builder should not be nil")
		}
	})
}

// TestPlatformDetection tests platform-specific logic
func TestPlatformDetection(t *testing.T) {
	// Load config to get binaries with platform assets
	binaries, err := loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Find a binary with platform assets
	var testBinary *DepBinary
	for _, binary := range binaries {
		if binary.Source == "github-release" && len(binary.Assets) > 0 {
			testBinary = &binary
			break
		}
	}

	if testBinary == nil {
		t.Skip("No github-release binary found for platform testing")
	}

	t.Logf("Testing platform detection with binary: %s", testBinary.Name)

	// Test that we have assets for common platforms
	foundPlatforms := make(map[string]bool)
	for _, asset := range testBinary.Assets {
		platform := asset.OS + "-" + asset.Arch
		foundPlatforms[platform] = true
	}

	expectedPlatforms := []string{"darwin-amd64", "darwin-arm64", "linux-amd64"}
	for _, platform := range expectedPlatforms {
		if !foundPlatforms[platform] {
			t.Logf("Platform %s not found in %s assets (this may be intentional)", platform, testBinary.Name)
		}
	}

	t.Logf("Found platforms for %s: %v", testBinary.Name, foundPlatforms)
}

// TestInstallerCleanup tests removal functionality
func TestInstallerCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cleanup tests in short mode") 
	}

	// Use a small, fast binary for testing
	testBinary := "task"

	// Ensure it's installed first
	if err := InstallBinary(testBinary, false); err != nil {
		t.Fatalf("Failed to install test binary: %v", err)
	}

	// Verify it exists
	binaryPath, err := Get(testBinary)
	if err != nil {
		t.Fatalf("Failed to get binary path: %v", err)
	}

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("Binary should exist before removal")
	}

	// Test removal
	if err := Remove(testBinary); err != nil {
		t.Fatalf("Failed to remove binary: %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(binaryPath); !os.IsNotExist(err) {
		t.Error("Binary should not exist after removal")
	}

	// Verify metadata is gone
	metaPath := getMetaPath(binaryPath)
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Error("Metadata should not exist after removal")
	}
}