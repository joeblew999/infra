package conduit

import "github.com/joeblew999/infra/pkg/config"

// GetBinaryName returns the platform-specific binary name
func GetBinaryName(baseName string) string {
	return config.GetBinaryName(baseName)
}

// GetBinaryNames returns platform-specific names for a list of binaries
func GetBinaryNames(baseNames []string) []string {
	return config.GetBinaryNames(baseNames)
}

// StripExeSuffix removes .exe suffix from Windows binaries for metadata purposes
func StripExeSuffix(binaryName string) string {
	return config.StripExeSuffix(binaryName)
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return config.IsWindows()
}

// GetPlatformString returns a human-readable platform string
func GetPlatformString() string {
	return config.GetPlatform()
}