package collection

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
)

// Config holds configuration for the managed binary distribution system
type Config struct {
	// Collection settings
	CollectionDir    string   `json:"collection_dir"`
	PlatformMatrix   []string `json:"platform_matrix"`
	ConcurrentLimit  int      `json:"concurrent_limit"`
	
	// Managed release settings
	ManagedRepo      string `json:"managed_repo"`
	ManagedOwner     string `json:"managed_owner"`
	ReleasePrefix    string `json:"release_prefix"`
	
	// Download settings
	PreferManaged    bool     `json:"prefer_managed"`
	FallbackChain    []string `json:"fallback_chain"`
	CacheEnabled     bool     `json:"cache_enabled"`
	CacheTTL         int      `json:"cache_ttl_hours"`
	
	// Validation settings
	RequireChecksums bool `json:"require_checksums"`
	VerifySignatures bool `json:"verify_signatures"`
}

// getGitOwner attempts to get the git owner from git config or environment
func getGitOwner() string {
	// Try environment variable first
	if owner := os.Getenv("GIT_OWNER"); owner != "" {
		return owner
	}

	// Try git config user.name
	if cmd := exec.Command("git", "config", "user.name"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			if owner := strings.TrimSpace(string(output)); owner != "" {
				return owner
			}
		}
	}

	// Try remote.origin.url to extract owner
	if cmd := exec.Command("git", "config", "--get", "remote.origin.url"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			url := strings.TrimSpace(string(output))
			// Parse GitHub URL to extract owner
			// https://github.com/owner/repo.git or git@github.com:owner/repo.git
			if strings.Contains(url, "github.com") {
				if strings.HasPrefix(url, "git@") {
					// git@github.com:owner/repo.git
					parts := strings.Split(url, ":")
					if len(parts) > 1 {
						repoPath := strings.TrimSuffix(parts[1], ".git")
						if pathParts := strings.Split(repoPath, "/"); len(pathParts) > 0 {
							return pathParts[0]
						}
					}
				} else if strings.HasPrefix(url, "https://") {
					// https://github.com/owner/repo.git
					parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
					if len(parts) > 0 {
						return parts[0]
					}
				}
			}
		}
	}

	// Fallback to hardcoded value
	return "joeblew999"
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		CollectionDir:   filepath.Join(config.GetDepPath(), ".collection"),
		PlatformMatrix: []string{
			"darwin-amd64",
			"darwin-arm64",
			"linux-amd64",
			"linux-arm64",
			"windows-amd64",
			"windows-arm64",
		},
		ConcurrentLimit:  4,
		ManagedRepo:      "infra-binaries",
		ManagedOwner:     getGitOwner(),
		ReleasePrefix:    "",
		PreferManaged:    true,
		FallbackChain:    []string{"managed", "original", "cache"},
		CacheEnabled:     true,
		CacheTTL:         72, // 3 days
		RequireChecksums: true,
		VerifySignatures: false,
	}
}

// GetCollectionPath returns the collection path for a binary
func (c *Config) GetCollectionPath(name, version string) string {
	return filepath.Join(c.CollectionDir, "binaries", name, version)
}

// GetManifestPath returns the manifest path for a binary
func (c *Config) GetManifestPath(name, version string) string {
	return filepath.Join(c.GetCollectionPath(name, version), "manifest.json")
}

// GetBinaryPath returns the path for a platform-specific binary
func (c *Config) GetBinaryPath(name, version, platform string) string {
	filename := name
	if platform == "windows-amd64" || platform == "windows-arm64" {
		filename += ".exe"
	}
	return filepath.Join(c.GetCollectionPath(name, version), platform, filename)
}

// GetReleaseTag returns the release tag for a binary version
func (c *Config) GetReleaseTag(name, version string) string {
	if c.ReleasePrefix != "" {
		return fmt.Sprintf("%s-%s-%s", c.ReleasePrefix, name, version)
	}
	return fmt.Sprintf("%s-%s", name, version)
}

// GetManagedRepoURL returns the full URL for the managed repository
func (c *Config) GetManagedRepoURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", c.ManagedOwner, c.ManagedRepo)
}

// GetManagedReleaseURL returns the URL for a specific managed release
func (c *Config) GetManagedReleaseURL(name, version string) string {
	tag := c.GetReleaseTag(name, version)
	return fmt.Sprintf("%s/releases/tag/%s", c.GetManagedRepoURL(), tag)
}

// GetCurrentPlatform returns the current platform string
func (c *Config) GetCurrentPlatform() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}

// IsPlatformSupported checks if a platform is in the matrix
func (c *Config) IsPlatformSupported(platform string) bool {
	for _, p := range c.PlatformMatrix {
		if p == platform {
			return true
		}
	}
	return false
}

// GetSupportedPlatforms returns the list of supported platforms
func (c *Config) GetSupportedPlatforms() []string {
	return append([]string{}, c.PlatformMatrix...)
}

// GetMetadataDir returns the metadata directory path
func (c *Config) GetMetadataDir() string {
	return filepath.Join(c.CollectionDir, "metadata")
}

// GetCollectionReportPath returns the path for collection reports
func (c *Config) GetCollectionReportPath() string {
	return filepath.Join(c.GetMetadataDir(), "collection-report.json")
}

// GetReleaseStatusPath returns the path for release status tracking
func (c *Config) GetReleaseStatusPath() string {
	return filepath.Join(c.GetMetadataDir(), "release-status.json")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.CollectionDir == "" {
		return fmt.Errorf("collection_dir cannot be empty")
	}
	
	if len(c.PlatformMatrix) == 0 {
		return fmt.Errorf("platform_matrix cannot be empty")
	}
	
	if c.ManagedOwner == "" {
		return fmt.Errorf("managed_owner cannot be empty")
	}
	
	if c.ManagedRepo == "" {
		return fmt.Errorf("managed_repo cannot be empty")
	}
	
	if c.ConcurrentLimit <= 0 {
		return fmt.Errorf("concurrent_limit must be positive")
	}
	
	if c.CacheTTL < 0 {
		return fmt.Errorf("cache_ttl_hours cannot be negative")
	}
	
	// Validate platform matrix format
	for _, platform := range c.PlatformMatrix {
		if !isValidPlatform(platform) {
			return fmt.Errorf("invalid platform format: %s (expected format: os-arch)", platform)
		}
	}
	
	return nil
}

// isValidPlatform checks if a platform string is valid
func isValidPlatform(platform string) bool {
	validPlatforms := map[string]bool{
		"darwin-amd64":   true,
		"darwin-arm64":   true,
		"linux-amd64":    true,
		"linux-arm64":    true,
		"windows-amd64":  true,
		"windows-arm64":  true,
		"freebsd-amd64":  true,
		"freebsd-arm64":  true,
		"openbsd-amd64":  true,
		"openbsd-arm64":  true,
		"netbsd-amd64":   true,
		"netbsd-arm64":   true,
	}
	
	return validPlatforms[platform]
}

// PlatformConfig represents configuration for a specific platform
type PlatformConfig struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	FileExt      string `json:"file_ext"`
	Executable   bool   `json:"executable"`
	CrossCompile bool   `json:"cross_compile"`
}

// GetPlatformConfig returns configuration for a platform
func (c *Config) GetPlatformConfig(platform string) (*PlatformConfig, error) {
	switch platform {
	case "darwin-amd64":
		return &PlatformConfig{
			OS: "darwin", Arch: "amd64", FileExt: "", Executable: true, CrossCompile: true,
		}, nil
	case "darwin-arm64":
		return &PlatformConfig{
			OS: "darwin", Arch: "arm64", FileExt: "", Executable: true, CrossCompile: true,
		}, nil
	case "linux-amd64":
		return &PlatformConfig{
			OS: "linux", Arch: "amd64", FileExt: "", Executable: true, CrossCompile: true,
		}, nil
	case "linux-arm64":
		return &PlatformConfig{
			OS: "linux", Arch: "arm64", FileExt: "", Executable: true, CrossCompile: true,
		}, nil
	case "windows-amd64":
		return &PlatformConfig{
			OS: "windows", Arch: "amd64", FileExt: ".exe", Executable: true, CrossCompile: true,
		}, nil
	case "windows-arm64":
		return &PlatformConfig{
			OS: "windows", Arch: "arm64", FileExt: ".exe", Executable: true, CrossCompile: true,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}