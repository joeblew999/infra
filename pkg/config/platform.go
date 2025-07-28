package config

import (
	"runtime"
	"strings"
)

// GetBinaryName returns the platform-specific binary name with proper extension.
// On Windows, this adds .exe extension; on other platforms, it returns the base name.
func GetBinaryName(baseName string) string {
	osName, _ := GetCurrentPlatform()
	return GetBinaryNameFor(baseName, osName, "")
}

// GetBinaryNameFor returns the platform-specific binary name for given OS/ARCH.
// This allows testing with different platforms without changing runtime.GOOS/GOARCH.
func GetBinaryNameFor(baseName, osName, arch string) string {
	if osName == "windows" {
		return baseName + ".exe"
	}
	return baseName
}

// PlatformOverride allows overriding the platform for testing purposes.
// This enables CI matrix testing across different platforms.
var PlatformOverride struct {
	GOOS   string
	GOARCH string
}

// GetCurrentPlatform returns the current platform, respecting overrides for testing.
func GetCurrentPlatform() (string, string) {
	if PlatformOverride.GOOS != "" {
		return PlatformOverride.GOOS, PlatformOverride.GOARCH
	}
	return runtime.GOOS, runtime.GOARCH
}

// GetBinaryNameWithOverride returns the binary name respecting platform overrides.
func GetBinaryNameWithOverride(baseName string) string {
	osName, _ := GetCurrentPlatform()
	return GetBinaryNameFor(baseName, osName, "")
}

// IsWindowsWithOverride returns true if running on Windows, respecting overrides.
func IsWindowsWithOverride() bool {
	osName, _ := GetCurrentPlatform()
	return osName == "windows"
}

// GetBinaryNames returns platform-specific names for a list of binaries.
func GetBinaryNames(baseNames []string) []string {
	names := make([]string, len(baseNames))
	for i, name := range baseNames {
		names[i] = GetBinaryName(name)
	}
	return names
}

// GetBinaryNamesFor returns platform-specific names for given OS/ARCH.
func GetBinaryNamesFor(baseNames []string, osName, arch string) []string {
	names := make([]string, len(baseNames))
	for i, name := range baseNames {
		names[i] = GetBinaryNameFor(name, osName, arch)
	}
	return names
}

// StripExeSuffix removes .exe suffix from Windows binaries for metadata purposes.
func StripExeSuffix(binaryName string) string {
	return strings.TrimSuffix(binaryName, ".exe")
}

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return IsWindowsWithOverride()
}

// GetPlatform returns the current platform as os/arch string.
func GetPlatform() string {
	osName, arch := GetCurrentPlatform()
	return osName + "/" + arch
}

// GetPlatformParts returns the current OS and architecture separately.
func GetPlatformParts() (string, string) {
	return GetCurrentPlatform()
}