package dep

import (
	"errors"
	"testing"
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