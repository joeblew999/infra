package config

import (
	"runtime"
	"testing"
)

func TestGetBinaryName(t *testing.T) {
	// Test current platform
	name := GetBinaryName("conduit")
	if runtime.GOOS == "windows" {
		if name != "conduit.exe" {
			t.Errorf("Expected 'conduit.exe' on Windows, got '%s'", name)
		}
	} else {
		if name != "conduit" {
			t.Errorf("Expected 'conduit' on non-Windows, got '%s'", name)
		}
	}
}

func TestGetBinaryNameFor(t *testing.T) {
	// Test specific platforms
	cases := []struct {
		name     string
		osName   string
		expected string
	}{
		{"conduit", "linux", "conduit"},
		{"conduit", "darwin", "conduit"},
		{"conduit", "windows", "conduit.exe"},
		{"myapp", "windows", "myapp.exe"},
		{"myapp", "linux", "myapp"},
	}

	for _, tc := range cases {
		result := GetBinaryNameFor(tc.name, tc.osName, "amd64")
		if result != tc.expected {
			t.Errorf("GetBinaryNameFor(%q, %q) = %q, want %q",
				tc.name, tc.osName, result, tc.expected)
		}
	}
}

func TestPlatformOverride(t *testing.T) {
	// Save original platform
	originalGOOS := PlatformOverride.GOOS
	originalGOARCH := PlatformOverride.GOARCH
	
	// Test Windows override
	PlatformOverride.GOOS = "windows"
	PlatformOverride.GOARCH = "amd64"
	
	name := GetBinaryName("conduit")
	if name != "conduit.exe" {
		t.Errorf("With Windows override, expected 'conduit.exe', got '%s'", name)
	}
	
	// Test Linux override
	PlatformOverride.GOOS = "linux"
	PlatformOverride.GOARCH = "amd64"
	
	name = GetBinaryName("conduit")
	if name != "conduit" {
		t.Errorf("With Linux override, expected 'conduit', got '%s'", name)
	}
	
	// Restore original platform
	PlatformOverride.GOOS = originalGOOS
	PlatformOverride.GOARCH = originalGOARCH
}

func TestStripExeSuffix(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"conduit.exe", "conduit"},
		{"conduit", "conduit"},
		{"myapp.exe", "myapp"},
		{"myapp", "myapp"},
		{"app.exe.exe", "app.exe"},
	}

	for _, tc := range cases {
		result := StripExeSuffix(tc.input)
		if result != tc.expected {
			t.Errorf("StripExeSuffix(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestGetBinaryNames(t *testing.T) {
	// Test with platform override
	PlatformOverride.GOOS = "windows"
	PlatformOverride.GOARCH = "amd64"
	defer func() {
		PlatformOverride.GOOS = ""
		PlatformOverride.GOARCH = ""
	}()

	names := []string{"conduit", "bento", "task"}
	result := GetBinaryNames(names)
	expected := []string{"conduit.exe", "bento.exe", "task.exe"}

	for i, name := range result {
		if name != expected[i] {
			t.Errorf("GetBinaryNames[%d] = %q, want %q", i, name, expected[i])
		}
	}
}