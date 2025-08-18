package collection

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/joeblew999/infra/pkg/dep/internal"
	"github.com/joeblew999/infra/pkg/dep/platform"
	"github.com/joeblew999/infra/pkg/dep/util"
	"github.com/joeblew999/infra/pkg/log"
)

// CrossPlatformCollector implements cross-platform binary collection
type CrossPlatformCollector struct {
	config   *Config
	binaries []DepBinary
}

// NewCrossPlatformCollector creates a new cross-platform collector
func NewCrossPlatformCollector(config *Config) (*CrossPlatformCollector, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Load binaries from dep.json
	binaries, err := loadDepConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load dep.json: %w", err)
	}

	return &CrossPlatformCollector{
		config:   config,
		binaries: binaries,
	}, nil
}

// CollectBinary downloads a binary for all configured platforms using cross-platform simulation
func (c *CrossPlatformCollector) CollectBinary(ctx context.Context, name, version string) (*CollectionResult, error) {
	log.Info("Starting cross-platform binary collection", "name", name, "version", version)

	// Find the binary in configuration
	var binary *DepBinary
	for i := range c.binaries {
		if c.binaries[i].Name == name {
			binary = &c.binaries[i]
			break
		}
	}

	if binary == nil {
		return nil, fmt.Errorf("binary %s not found in configuration", name)
	}

	// Override version if specified
	if version != "" {
		binary = &DepBinary{
			Name:    binary.Name,
			Repo:    binary.Repo,
			Version: version,
			Assets:  binary.Assets,
		}
	}

	result := &CollectionResult{
		Binary:      name,
		Version:     binary.Version,
		Platforms:   make(map[string]*PlatformResult),
		CollectedAt: time.Now(),
	}

	// Create collection directory
	collectionPath := c.config.GetCollectionPath(name, binary.Version)
	if err := os.MkdirAll(collectionPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create collection directory: %w", err)
	}
	log.Debug("Created collection directory", "path", collectionPath)

	// Collect for each platform in parallel using cross-platform simulation
	platformResults := make(chan *PlatformResult, len(c.config.PlatformMatrix))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, c.config.ConcurrentLimit)

	for _, targetPlatform := range c.config.PlatformMatrix {
		wg.Add(1)
		go func(targetPlatform string) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			platformResult := c.collectForPlatformCrossPlatform(ctx, binary, targetPlatform)
			platformResults <- platformResult
		}(targetPlatform)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(platformResults)
	}()

	// Collect results
	successCount := 0
	var errors []string

	for platformResult := range platformResults {
		result.Platforms[platformResult.Platform] = platformResult
		if platformResult.Success {
			successCount++
		} else if platformResult.Error != "" {
			errors = append(errors, fmt.Sprintf("%s: %s", platformResult.Platform, platformResult.Error))
		}
	}

	result.Success = successCount > 0
	result.Errors = errors

	// Generate manifest
	manifest, err := c.generateManifest(result)
	if err != nil {
		log.Warn("Failed to generate manifest", "error", err)
		errors = append(errors, fmt.Sprintf("manifest generation failed: %v", err))
	} else {
		result.Manifest = manifest
		
		// Save manifest to disk
		manifestPath := c.config.GetManifestPath(name, binary.Version)
		if err := c.saveManifest(manifest, manifestPath); err != nil {
			log.Warn("Failed to save manifest", "error", err)
		}
	}

	log.Info("Cross-platform binary collection completed", 
		"name", name, 
		"version", binary.Version,
		"platforms", len(result.Platforms),
		"success", successCount,
		"errors", len(errors))

	return result, nil
}

// collectForPlatformCrossPlatform collects a binary for a specific platform using simulation
func (c *CrossPlatformCollector) collectForPlatformCrossPlatform(ctx context.Context, binary *DepBinary, targetPlatform string) *PlatformResult {
	startTime := time.Now()
	
	result := &PlatformResult{
		Platform: targetPlatform,
		Success:  false,
	}

	defer func() {
		result.Duration = time.Since(startTime)
	}()

	// Parse target platform
	platformConfig, err := c.config.GetPlatformConfig(targetPlatform)
	if err != nil {
		result.Error = fmt.Sprintf("invalid platform: %v", err)
		return result
	}

	log.Info("Starting collection for platform", 
		"binary", binary.Name,
		"version", binary.Version,
		"platform", targetPlatform,
		"real_platform", fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH))

	// Create platform-specific context
	platformCtx := platform.WithPlatformContext(ctx, platformConfig.OS, platformConfig.Arch)

	// Create temporary directory for download
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("collect-%s-%s-%s", binary.Name, binary.Version, targetPlatform))
	if err != nil {
		result.Error = fmt.Sprintf("failed to create temp dir: %v", err)
		return result
	}
	defer os.RemoveAll(tempDir)

	// Download binary using platform simulation
	tempBinaryPath, err := c.downloadBinaryForPlatformCrossPlatform(platformCtx, binary, platformConfig, tempDir)
	if err != nil {
		result.Error = fmt.Sprintf("download failed: %v", err)
		return result
	}

	// Get file info
	fileInfo, err := os.Stat(tempBinaryPath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to stat binary: %v", err)
		return result
	}

	result.Size = fileInfo.Size()
	result.Filename = filepath.Base(tempBinaryPath)

	// Calculate SHA256
	hash, err := c.calculateSHA256(tempBinaryPath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to calculate checksum: %v", err)
		return result
	}
	result.SHA256 = hash

	// Move to final location
	finalPath := c.config.GetBinaryPath(binary.Name, binary.Version, targetPlatform)
	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		result.Error = fmt.Sprintf("failed to create directory: %v", err)
		return result
	}

	if err := c.moveBinary(tempBinaryPath, finalPath); err != nil {
		result.Error = fmt.Sprintf("failed to move binary: %v", err)
		return result
	}

	result.LocalPath = finalPath
	result.Success = true

	log.Info("Cross-platform collection successful", 
		"binary", binary.Name,
		"version", binary.Version,
		"platform", targetPlatform,
		"size", util.FormatBytes(result.Size),
		"duration", result.Duration.Round(time.Millisecond))

	return result
}

// downloadBinaryForPlatformCrossPlatform downloads a binary for a specific platform using simulation
func (c *CrossPlatformCollector) downloadBinaryForPlatformCrossPlatform(ctx context.Context, binary *DepBinary, platformConfig *PlatformConfig, tempDir string) (string, error) {
	// This is where the cross-platform magic happens
	// We simulate the target platform and use the existing installers
	
	log.Debug("Starting cross-platform download", 
		"target_os", platformConfig.OS,
		"target_arch", platformConfig.Arch,
		"real_os", runtime.GOOS,
		"real_arch", runtime.GOARCH)

	// For GitHub releases, we can download the correct asset directly
	// without needing to run the binary on the target platform
	if binary.Repo != "" {
		return c.downloadFromGitHubRelease(ctx, binary, platformConfig, tempDir)
	}

	// For other types, we need different strategies
	// For now, return error indicating what we need to implement
	return "", fmt.Errorf("cross-platform collection for non-GitHub releases not yet implemented (binary: %s, type: unknown)", binary.Name)
}

// downloadFromGitHubRelease downloads a binary from GitHub releases for a specific platform
func (c *CrossPlatformCollector) downloadFromGitHubRelease(ctx context.Context, binary *DepBinary, platformConfig *PlatformConfig, tempDir string) (string, error) {
	// Use the GitHub API to get release information
	release, err := c.getGitHubRelease(binary.Repo, binary.Version)
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub release: %w", err)
	}

	// Find the asset for the target platform
	asset, err := c.selectAssetForPlatform(release, binary.Assets, platformConfig)
	if err != nil {
		return "", fmt.Errorf("failed to select asset: %w", err)
	}

	log.Info("Found matching asset", 
		"asset", asset.Name,
		"platform", fmt.Sprintf("%s-%s", platformConfig.OS, platformConfig.Arch),
		"url", asset.BrowserDownloadURL)

	// Download the asset
	tempArchivePath := filepath.Join(tempDir, asset.Name)
	if err := util.DownloadFile(asset.BrowserDownloadURL, tempArchivePath, true); err != nil {
		return "", fmt.Errorf("failed to download asset: %w", err)
	}

	// Extract the archive if needed
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create extraction directory: %w", err)
	}

	// Determine if we need to extract
	if c.isArchive(asset.Name) {
		if err := c.extractArchive(tempArchivePath, extractDir); err != nil {
			return "", fmt.Errorf("failed to extract archive: %w", err)
		}
	} else {
		// Direct binary, just copy it
		binaryName := binary.Name
		if platformConfig.OS == "windows" {
			binaryName += ".exe"
		}
		directPath := filepath.Join(extractDir, binaryName)
		if err := c.copyFile(tempArchivePath, directPath); err != nil {
			return "", fmt.Errorf("failed to copy binary: %w", err)
		}
	}

	// Find the binary in the extracted files
	binaryPath, err := c.findBinaryInExtraction(extractDir, binary.Name, platformConfig)
	if err != nil {
		return "", fmt.Errorf("failed to find binary: %w", err)
	}

	return binaryPath, nil
}

// Helper methods that will be used by the cross-platform collector

func (c *CrossPlatformCollector) getGitHubRelease(repo, version string) (*GitHubRelease, error) {
	api := NewGitHubAPI()
	return api.GetRelease(repo, version)
}

func (c *CrossPlatformCollector) selectAssetForPlatform(release *GitHubRelease, assets []AssetInfo, platformConfig *PlatformConfig) (*GitHubReleaseAsset, error) {
	api := NewGitHubAPI()
	return api.SelectAssetForPlatform(release, assets, platformConfig)
}

func (c *CrossPlatformCollector) isArchive(filename string) bool {
	extensions := []string{".zip", ".tar.gz", ".tgz", ".tar.bz2", ".tbz2"}
	for _, ext := range extensions {
		if len(filename) >= len(ext) && filename[len(filename)-len(ext):] == ext {
			return true
		}
	}
	return false
}

func (c *CrossPlatformCollector) extractArchive(archivePath, destDir string) error {
	return internal.ExtractArchive(archivePath, destDir)
}

func (c *CrossPlatformCollector) copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (c *CrossPlatformCollector) findBinaryInExtraction(extractDir, binaryName string, platformConfig *PlatformConfig) (string, error) {
	expectedName := binaryName
	if platformConfig.OS == "windows" {
		expectedName += ".exe"
	}

	// Common locations to check
	possiblePaths := []string{
		filepath.Join(extractDir, expectedName),
		filepath.Join(extractDir, "bin", expectedName),
		filepath.Join(extractDir, binaryName, expectedName),
	}

	// Walk through extracted directory to find the binary
	var foundPath string
	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}
		
		if info.IsDir() {
			return nil
		}
		
		filename := info.Name()
		if filename == expectedName || filename == binaryName {
			// Check if it's executable (or on Windows)
			if info.Mode().Perm()&0111 != 0 || platformConfig.OS == "windows" {
				foundPath = path
				return filepath.SkipDir // Stop walking
			}
		}
		
		return nil
	})

	if err != nil {
		return "", err
	}

	// Try direct paths first
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Use the walked path if found
	if foundPath != "" {
		return foundPath, nil
	}

	return "", fmt.Errorf("binary %s not found in extracted directory %s", expectedName, extractDir)
}

func (c *CrossPlatformCollector) calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (c *CrossPlatformCollector) moveBinary(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Make executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dst, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (c *CrossPlatformCollector) generateManifest(result *CollectionResult) (*BinaryManifest, error) {
	manifest := &BinaryManifest{
		Binary:         result.Binary,
		Version:        result.Version,
		CollectionDate: result.CollectedAt,
		Platforms:      make(map[string]*PlatformInfo),
	}

	// Add source information (simplified for now)
	manifest.Source = &SourceInfo{
		Type:    "github-release", 
		Version: result.Version,
	}

	// Add platform information
	for platform, platformResult := range result.Platforms {
		if platformResult.Success {
			manifest.Platforms[platform] = &PlatformInfo{
				Filename:   platformResult.Filename,
				Size:       platformResult.Size,
				SHA256:     platformResult.SHA256,
				Executable: true,
				LocalPath:  platformResult.LocalPath,
			}
		}
	}

	return manifest, nil
}

func (c *CrossPlatformCollector) saveManifest(manifest *BinaryManifest, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}

// GitHub API types for cross-platform collection
type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type GitHubRelease struct {
	TagName string               `json:"tag_name"`
	Assets  []GitHubReleaseAsset `json:"assets"`
}