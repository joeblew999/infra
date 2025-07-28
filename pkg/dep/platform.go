// Package dep provides cross-platform binary naming utilities.
// This package delegates to pkg/config for centralized platform handling.
package dep

import "github.com/joeblew999/infra/pkg/config"

// GetBinaryName returns the platform-specific binary name with proper extension.
// On Windows, this adds .exe extension; on other platforms, it returns the base name.
func GetBinaryName(baseName string) string {
	return config.GetBinaryName(baseName)
}

// GetBinaryNameFor returns the platform-specific binary name for given OS/ARCH.
// This allows testing with different platforms without changing runtime.GOOS/GOARCH.
func GetBinaryNameFor(baseName, osName, arch string) string {
	return config.GetBinaryNameFor(baseName, osName, arch)
}

// GetBinaryNames returns platform-specific names for a list of binaries.
func GetBinaryNames(baseNames []string) []string {
	return config.GetBinaryNames(baseNames)
}

// GetBinaryNamesFor returns platform-specific names for given OS/ARCH.
func GetBinaryNamesFor(baseNames []string, osName, arch string) []string {
	return config.GetBinaryNamesFor(baseNames, osName, arch)
}

// StripExeSuffix removes .exe suffix from Windows binaries for metadata purposes.
func StripExeSuffix(binaryName string) string {
	return config.StripExeSuffix(binaryName)
}

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return config.IsWindows()
}

// GetPlatform returns the current platform as os/arch string.
func GetPlatform() string {
	return config.GetPlatform()
}

// GetPlatformParts returns the current OS and architecture separately.
func GetPlatformParts() (string, string) {
	return config.GetPlatformParts()
}