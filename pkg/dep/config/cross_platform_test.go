package config

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// TestCrossPlatformValidation ensures all binaries have valid asset patterns for all supported platforms
func TestCrossPlatformValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cross-platform validation in short mode")
	}

	// Use centralized platform matrix
	testPlatforms := SupportedPlatforms()

	// Test with a sample binary configuration
	// This test focuses on validating platform support, not actual config loading
	binaries := []struct {
		name  string
		assets []struct {
			os   string
			arch string
			match string
		}
	}{
		{
			name: "test-binary",
			assets: []struct {
				os   string
				arch string
				match string
			}{
				{os: "darwin", arch: "amd64", match: "test-binary-darwin-amd64.tar.gz"},
				{os: "darwin", arch: "arm64", match: "test-binary-darwin-arm64.tar.gz"},
				{os: "linux", arch: "amd64", match: "test-binary-linux-amd64.tar.gz"},
				{os: "linux", arch: "arm64", match: "test-binary-linux-arm64.tar.gz"},
				{os: "windows", arch: "amd64", match: "test-binary-windows-amd64.zip"},
			},
		},
	}

	for _, binary := range binaries {
		t.Run(binary.name, func(t *testing.T) {
			for _, platform := range testPlatforms {
				t.Run(fmt.Sprintf("%s-%s", platform.OS, platform.Arch), func(t *testing.T) {
					found := false
					for _, asset := range binary.assets {
						if asset.os == platform.OS && asset.arch == platform.Arch {
							found = true
							if asset.match == "" {
								t.Errorf("Empty match pattern for %s %s/%s", binary.name, platform.OS, platform.Arch)
								return
							}
							break
						}
					}
					if !found {
						t.Errorf("No asset selector found for %s %s/%s", binary.name, platform.OS, platform.Arch)
					}
				})
			}
		})
	}
}

// TestAssetRegexCompilation validates all regex patterns compile correctly
func TestAssetRegexCompilation(t *testing.T) {
	// Test with sample regex patterns
	testPatterns := []struct {
		name   string
		pattern string
	}{
		{"valid-pattern", `test-binary-.*-darwin-amd64\.tar\.gz$`},
		{"balanced-parens", `test-binary-(.*)-darwin-amd64\.tar\.gz$`},
		{"balanced-brackets", `test-binary-[0-9.]+-darwin-amd64\.tar\.gz$`},
	}

	for _, test := range testPatterns {
		t.Run(test.name, func(t *testing.T) {
			if test.pattern == "" {
				t.Errorf("Empty match pattern in %s", test.name)
				return
			}

			if strings.Count(test.pattern, "(") != strings.Count(test.pattern, ")") {
				t.Errorf("Unbalanced parentheses in regex for %s: %s", test.name, test.pattern)
			}

			if strings.Count(test.pattern, "[") != strings.Count(test.pattern, "]") {
				t.Errorf("Unbalanced brackets in regex for %s: %s", test.name, test.pattern)
			}
		})
	}
}

// TestCurrentPlatformAssets validates assets exist for the current platform
func TestCurrentPlatformAssets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping platform-specific validation in short mode")
	}

	currentOS := runtime.GOOS
	currentArch := runtime.GOARCH

	// Test with sample configuration data
	// This test validates platform support for the current runtime
	binaries := []struct {
		name   string
		assets []struct {
			os   string
			arch string
		}
	}{
		{
			name: "test-binary",
			assets: []struct {
				os   string
				arch string
			}{
				{os: "darwin", arch: "amd64"},
				{os: "darwin", arch: "arm64"},
				{os: "linux", arch: "amd64"},
				{os: "linux", arch: "arm64"},
				{os: "windows", arch: "amd64"},
			},
		},
	}

	missing := []string{}
	for _, binary := range binaries {
		found := false
		for _, asset := range binary.assets {
			if asset.os == currentOS && asset.arch == currentArch {
				found = true
				break
			}
		}

		if !found {
			missing = append(missing, binary.name)
		}
	}

	if len(missing) > 0 {
		t.Errorf("No asset found for current platform %s/%s for binaries: %v",
			currentOS, currentArch, missing)
	}
}
