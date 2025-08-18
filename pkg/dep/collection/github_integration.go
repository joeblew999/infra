package collection

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/joeblew999/infra/pkg/log"
)

// GitHubAPI provides methods for interacting with GitHub releases
type GitHubAPI struct{}

// NewGitHubAPI creates a new GitHub API client
func NewGitHubAPI() *GitHubAPI {
	return &GitHubAPI{}
}

// GetRelease fetches release information from GitHub API
func (g *GitHubAPI) GetRelease(repo, version string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, version)
	
	log.Debug("Fetching GitHub release", "repo", repo, "version", version, "url", url)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub release from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for %s", resp.StatusCode, url)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub release response: %w", err)
	}

	log.Debug("Successfully fetched GitHub release", 
		"repo", repo, 
		"version", version, 
		"assets", len(release.Assets))

	return &release, nil
}

// SelectAssetForPlatform selects the appropriate asset for a target platform
func (g *GitHubAPI) SelectAssetForPlatform(release *GitHubRelease, assets []AssetInfo, platformConfig *PlatformConfig) (*GitHubReleaseAsset, error) {
	targetOS := platformConfig.OS
	targetArch := platformConfig.Arch

	log.Debug("Selecting asset for platform", 
		"target_os", targetOS,
		"target_arch", targetArch,
		"available_assets", len(release.Assets))

	// Find matching asset selector
	var selector *AssetInfo
	for _, asset := range assets {
		if asset.OS == targetOS && asset.Arch == targetArch {
			selector = &asset
			break
		}
	}

	if selector == nil {
		return nil, fmt.Errorf("no asset selector found for %s/%s", targetOS, targetArch)
	}

	log.Debug("Found asset selector", "pattern", selector.Match)

	// Find matching asset in release
	for _, asset := range release.Assets {
		matched, err := regexp.MatchString(selector.Match, asset.Name)
		if err != nil {
			log.Warn("Invalid regex pattern", "pattern", selector.Match, "error", err)
			continue
		}
		
		if matched {
			log.Info("Selected asset for platform", 
				"asset", asset.Name,
				"platform", fmt.Sprintf("%s-%s", targetOS, targetArch),
				"size", asset.Size,
				"pattern", selector.Match)
			return &asset, nil
		}
	}

	// Log available assets for debugging
	var assetNames []string
	for _, asset := range release.Assets {
		assetNames = append(assetNames, asset.Name)
	}
	
	return nil, fmt.Errorf("no matching asset found for %s/%s with pattern '%s', available assets: %v", 
		targetOS, targetArch, selector.Match, assetNames)
}


// Platform mapping for common variations
var platformMappings = map[string]map[string]string{
	"os": {
		"darwin":  "darwin",
		"linux":   "linux", 
		"windows": "windows",
		// Some binaries use different naming
		"macos": "darwin",
		"win32": "windows",
		"win":   "windows",
	},
	"arch": {
		"amd64": "amd64",
		"arm64": "arm64",
		// Some binaries use different naming
		"x86_64": "amd64",
		"x64":    "amd64",
		"aarch64": "arm64",
	},
}

// NormalizePlatform normalizes platform names for asset matching
func NormalizePlatform(os, arch string) (string, string) {
	normalizedOS := os
	normalizedArch := arch
	
	if mapped, exists := platformMappings["os"][strings.ToLower(os)]; exists {
		normalizedOS = mapped
	}
	
	if mapped, exists := platformMappings["arch"][strings.ToLower(arch)]; exists {
		normalizedArch = mapped
	}
	
	return normalizedOS, normalizedArch
}