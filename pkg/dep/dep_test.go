package dep

import (
	"errors"
	"testing"
)

func TestEmbeddedDepBinaries(t *testing.T) {
	// Load actual configuration and validate its structure
	binaries, err := loadConfig()
	if err != nil {
		t.Fatalf("Failed to load embedded configuration: %v", err)
	}

	if len(binaries) == 0 {
		t.Fatal("Expected at least one binary in configuration")
	}

	// Validate each binary has required fields
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
		
		// Only require release_url for github-release source types
		if binary.Source == "github-release" && binary.ReleaseURL == "" {
			t.Errorf("Binary %s with github-release source must have release URL", binary.Name)
		}
		
		// Only require assets for github-release source types
		if binary.Source == "github-release" && len(binary.Assets) == 0 {
			t.Errorf("Binary %s with github-release source must have at least one asset selector", binary.Name)
		}
		
		// go-build source types need package field
		if binary.Source == "go-build" && binary.Package == "" {
			t.Errorf("Binary %s with go-build source must have package field", binary.Name)
		}
		
		// claude-release source type is valid (no additional validation needed)
		if binary.Source == "claude-release" {
			// Claude release installer handles its own validation
		}
	}

	// Ensure all binaries have unique names
	names := make(map[string]bool)
	for _, binary := range binaries {
		if names[binary.Name] {
			t.Errorf("Duplicate binary name found: %s", binary.Name)
		}
		names[binary.Name] = true
	}

	// Validate asset selectors for each binary
	for _, binary := range binaries {
		if len(binary.Assets) > 0 {
			for _, asset := range binary.Assets {
				if asset.OS == "" {
					t.Errorf("Binary %s has asset with empty OS", binary.Name)
				}
				if asset.Arch == "" {
					t.Errorf("Binary %s has asset with empty Arch", binary.Name)
				}
				if asset.Match == "" {
					t.Errorf("Binary %s has asset with empty Match pattern", binary.Name)
				}
			}
		}
	}
}

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