package dep

import (
	"errors"
	"testing"

	"github.com/joeblew999/infra/pkg/config"
)

func TestEmbeddedDepBinaries(t *testing.T) {
	expectedBinaries := []string{
		"bento",
		"task",
		"tofu",
		"caddy",
		"ko",
		"flyctl",
		"garble",
		"bun",
		"claude",
		"nats",
	}

	// Load configuration using the new loadConfig function
	binaries, err := loadConfig()
	if err != nil {
		t.Fatalf("Failed to load embedded configuration: %v", err)
	}

	if len(binaries) != len(expectedBinaries) {
		t.Errorf("Expected %d binaries, got %d", len(expectedBinaries), len(binaries))
	}

	binaryMap := make(map[string]bool)
	for _, binary := range binaries {
		binaryMap[binary.Name] = true
	}

	for _, expected := range expectedBinaries {
		if !binaryMap[expected] {
			t.Errorf("Expected binary %s not found in embedded configuration", expected)
		}
	}

	for _, binary := range binaries {
		if binary.Name == "" {
			t.Error("Binary name cannot be empty")
		}
		if binary.Repo == "" {
			t.Errorf("Binary %s repo cannot be empty", binary.Name)
		}
		if binary.Version == "" {
			t.Errorf("Binary %s version cannot be empty", binary.Name)
		}
		if binary.ReleaseURL == "" {
			t.Errorf("Binary %s release URL cannot be empty", binary.Name)
		}
		// Garble uses go install, so it intentionally has no assets
		if len(binary.Assets) == 0 && binary.Name != "garble" {
			t.Errorf("Binary %s must have at least one asset", binary.Name)
		}
	}
}

// TestGet tests the Get function with various inputs
func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		binaryName  string
		expectError bool
		errorType   error
	}{
		{
			name:        "Valid binary name",
			binaryName:  "bento",
			expectError: false,
		},
		{
			name:        "Another valid binary name",
			binaryName:  "garble",
			expectError: false,
		},
		{
			name:        "Empty binary name",
			binaryName:  "",
			expectError: true,
			errorType:   ErrInvalidInput,
		},
		{
			name:        "Whitespace only binary name",
			binaryName:  "   ",
			expectError: true,
			errorType:   ErrInvalidInput,
		},
		{
			name:        "Unsupported binary name",
			binaryName:  "nonexistent",
			expectError: true,
			errorType:   ErrBinaryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := Get(tt.binaryName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.errorType != nil && !errors.Is(err, tt.errorType) {
					t.Errorf("Expected error type %v, got %v", tt.errorType, err)
				}
				if path != "" {
					t.Errorf("Expected empty path on error, got %s", path)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if path == "" {
					t.Errorf("Expected non-empty path for valid binary %s", tt.binaryName)
				}
			}
		})
	}
}

// TestBinaryMetaStruct tests the BinaryMeta struct
func TestBinaryMetaStruct(t *testing.T) {
	meta := BinaryMeta{
		Name:    "test-binary",
		Version: "v1.0.0",
	}

	if meta.Name != "test-binary" {
		t.Errorf("Expected name 'test-binary', got %s", meta.Name)
	}
	if meta.Version != "v1.0.0" {
		t.Errorf("Expected version 'v1.0.0', got %s", meta.Version)
	}
}

// TestGetBinaryName tests the platform-specific binary naming from platform.go
func TestGetBinaryName(t *testing.T) {
	tests := []struct {
		name     string
		baseName string
		osName   string
		expected string
	}{
		{
			name:     "Linux binary",
			baseName: "claude",
			osName:   "linux",
			expected: "claude",
		},
		{
			name:     "macOS binary",
			baseName: "task",
			osName:   "darwin",
			expected: "task",
		},
		{
			name:     "Windows binary",
			baseName: "claude",
			osName:   "windows",
			expected: "claude.exe",
		},
		{
			name:     "Windows task binary",
			baseName: "task",
			osName:   "windows",
			expected: "task.exe",
		},
		{
			name:     "Empty base name",
			baseName: "",
			osName:   "linux",
			expected: "",
		},
		{
			name:     "Windows with dots in name",
			baseName: "my.tool",
			osName:   "windows",
			expected: "my.tool.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBinaryNameFor(tt.baseName, tt.osName, "amd64")
			if result != tt.expected {
				t.Errorf("GetBinaryNameFor(%q, %q, \"amd64\") = %q, want %q", 
					tt.baseName, tt.osName, result, tt.expected)
			}
		})
	}
}

// TestGetBinaryNames tests batch binary naming
func TestGetBinaryNames(t *testing.T) {
	baseNames := []string{"claude", "task", "tofu", "caddy"}
	
	tests := []struct {
		name     string
		osName   string
		expected []string
	}{
		{
			name:     "Linux batch",
			osName:   "linux",
			expected: []string{"claude", "task", "tofu", "caddy"},
		},
		{
			name:     "Windows batch",
			osName:   "windows",
			expected: []string{"claude.exe", "task.exe", "tofu.exe", "caddy.exe"},
		},
		{
			name:     "macOS batch",
			osName:   "darwin",
			expected: []string{"claude", "task", "tofu", "caddy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBinaryNamesFor(baseNames, tt.osName, "amd64")
			if len(result) != len(tt.expected) {
				t.Errorf("GetBinaryNamesFor() returned %d names, want %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("GetBinaryNamesFor()[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestStripExeSuffix tests the exe suffix removal functionality
func TestStripExeSuffix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows binary",
			input:    "claude.exe",
			expected: "claude",
		},
		{
			name:     "Linux binary",
			input:    "claude",
			expected: "claude",
		},
		{
			name:     "Multiple extensions",
			input:    "tool.v1.exe",
			expected: "tool.v1",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripExeSuffix(tt.input)
			if result != tt.expected {
				t.Errorf("StripExeSuffix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPlatformFunctions tests platform detection utilities
func TestPlatformFunctions(t *testing.T) {
	// Test that platform functions return reasonable values
	platform := GetPlatform()
	if platform == "" {
		t.Error("GetPlatform() should not return empty string")
	}

	osName, arch := GetPlatformParts()
	if osName == "" {
		t.Error("GetPlatformParts() osName should not be empty")
	}
	if arch == "" {
		t.Error("GetPlatformParts() arch should not be empty")
	}

	// Test IsWindows with override
	originalGOOS := config.PlatformOverride.GOOS
	defer func() { config.PlatformOverride.GOOS = originalGOOS }()

	config.PlatformOverride.GOOS = "windows"
	if !IsWindows() {
		t.Error("IsWindows() should return true when GOOS=windows")
	}

	config.PlatformOverride.GOOS = "linux"
	if IsWindows() {
		t.Error("IsWindows() should return false when GOOS=linux")
	}

	config.PlatformOverride.GOOS = "darwin"
	if IsWindows() {
		t.Error("IsWindows() should return false when GOOS=darwin")
	}
}

// TestBinaryPathConstruction tests the full binary path construction using platform naming
func TestBinaryPathConstruction(t *testing.T) {
	// This tests the integration between dep package and platform naming
	// We'll use platform overrides to test different scenarios
	
	originalGOOS := config.PlatformOverride.GOOS
	defer func() { config.PlatformOverride.GOOS = originalGOOS }()

	tests := []struct {
		name     string
		binary   string
		osName   string
		expected string
	}{
		{
			name:     "claude on Linux",
			binary:   "claude",
			osName:   "linux",
			expected: ".dep/claude",
		},
		{
			name:     "claude on Windows",
			binary:   "claude",
			osName:   "windows",
			expected: ".dep/claude.exe",
		},
		{
			name:     "task on macOS",
			binary:   "task",
			osName:   "darwin",
			expected: ".dep/task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.PlatformOverride.GOOS = tt.osName
			
			// Test that GetBinaryName produces correct name
			binaryName := GetBinaryName(tt.binary)
			
			// Test the full path construction (relative to current dir)
			fullPath := ".dep/" + binaryName
			
			if fullPath != tt.expected {
				t.Errorf("Expected path %q, got %q", tt.expected, fullPath)
			}
		})
	}
}