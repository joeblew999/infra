package config

import "testing"

func TestGetBinaryNameFor(t *testing.T) {
	tests := []struct {
		name     string
		baseName string
		os       string
		arch     string
		expected string
	}{
		{"Linux binary", "claude", "linux", "amd64", "claude"},
		{"macOS binary", "task", "darwin", "arm64", "task"},
		{"Windows binary", "claude", "windows", "amd64", "claude.exe"},
		{"Windows ARM64", "task", "windows", "arm64", "task.exe"},
		{"Linux ARM64", "tofu", "linux", "arm64", "tofu"},
		{"Empty base", "", "linux", "amd64", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBinaryNameFor(tt.baseName, tt.os, tt.arch)
			if result != tt.expected {
				t.Errorf("GetBinaryNameFor(%q, %q, %q) = %q, want %q",
					tt.baseName, tt.os, tt.arch, result, tt.expected)
			}
		})
	}
}

func TestGetBinaryNamesFor(t *testing.T) {
	baseNames := []string{"claude", "task", "tofu", "caddy"}

	tests := []struct {
		name     string
		os       string
		arch     string
		expected []string
	}{
		{"Linux batch", "linux", "amd64", []string{"claude", "task", "tofu", "caddy"}},
		{"Windows batch", "windows", "amd64", []string{"claude.exe", "task.exe", "tofu.exe", "caddy.exe"}},
		{"macOS batch", "darwin", "arm64", []string{"claude", "task", "tofu", "caddy"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBinaryNamesFor(baseNames, tt.os, tt.arch)
			if len(result) != len(tt.expected) {
				t.Fatalf("GetBinaryNamesFor() returned %d names, want %d", len(result), len(tt.expected))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("GetBinaryNamesFor()[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
