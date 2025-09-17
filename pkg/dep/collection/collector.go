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

	"github.com/joeblew999/infra/pkg/dep/util"
	"github.com/joeblew999/infra/pkg/log"
)

// DefaultCollector implements the Collector interface
type DefaultCollector struct {
	config   *Config
	binaries []DepBinary
}

// NewCollector creates a new collector with the given configuration
func NewCollector(config *Config) (*DefaultCollector, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Load binaries from dep.json
	binaries, err := loadDepConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load dep.json: %w", err)
	}

	return &DefaultCollector{
		config:   config,
		binaries: binaries,
	}, nil
}

// CollectBinary downloads a binary for all configured platforms
func (c *DefaultCollector) CollectBinary(ctx context.Context, name, version string) (*CollectionResult, error) {
	log.Info("Starting binary collection", "name", name, "version", version)

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

	// Collect for each platform in parallel
	platformResults := make(chan *PlatformResult, len(c.config.PlatformMatrix))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, c.config.ConcurrentLimit)

	for _, platform := range c.config.PlatformMatrix {
		wg.Add(1)
		go func(platform string) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			platformResult := c.collectForPlatform(ctx, binary, platform)
			platformResults <- platformResult
		}(platform)
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

	log.Info("Binary collection completed", 
		"name", name, 
		"version", binary.Version,
		"platforms", len(result.Platforms),
		"success", successCount,
		"errors", len(errors))

	return result, nil
}

// CollectAll downloads all configured binaries for all platforms
func (c *DefaultCollector) CollectAll(ctx context.Context) (*BatchCollectionResult, error) {
	log.Info("Starting batch collection for all binaries", "count", len(c.binaries))

	result := &BatchCollectionResult{
		Results:     make(map[string]*CollectionResult),
		TotalCount:  len(c.binaries),
		StartedAt:   time.Now(),
	}

	successCount := 0
	for _, binary := range c.binaries {
		log.Info("Collecting binary", "name", binary.Name, "version", binary.Version)
		
		collectionResult, err := c.CollectBinary(ctx, binary.Name, binary.Version)
		if err != nil {
			log.Error("Failed to collect binary", "name", binary.Name, "error", err)
			// Create a failed result
			collectionResult = &CollectionResult{
				Binary:      binary.Name,
				Version:     binary.Version,
				Success:     false,
				Errors:      []string{err.Error()},
				CollectedAt: time.Now(),
				Platforms:   make(map[string]*PlatformResult),
			}
		}

		result.Results[binary.Name] = collectionResult
		if collectionResult.Success {
			successCount++
		}
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)
	result.SuccessCount = successCount
	result.FailureCount = result.TotalCount - successCount

	log.Info("Batch collection completed", 
		"total", result.TotalCount,
		"success", result.SuccessCount,
		"failed", result.FailureCount,
		"duration", result.Duration.Round(time.Second))

	return result, nil
}

// GetCollectionStatus returns the current collection status
func (c *DefaultCollector) GetCollectionStatus(name, version string) (*CollectionStatus, error) {
	manifestPath := c.config.GetManifestPath(name, version)

	status := &CollectionStatus{
		Binary:    name,
		Version:   version,
		Collected: false,
		Platforms: make(map[string]bool),
	}

	// Check if manifest exists
	if _, err := os.Stat(manifestPath); err == nil {
		status.Collected = true
		status.ManifestPath = manifestPath

		// Load manifest to get detailed status
		if manifest, err := c.loadManifest(manifestPath); err == nil {
			status.PlatformCount = len(manifest.Platforms)
			collectedAt := manifest.CollectionDate
			status.CollectedAt = &collectedAt

			for platform := range manifest.Platforms {
				binaryPath := c.config.GetBinaryPath(name, version, platform)
				_, err := os.Stat(binaryPath)
				status.Platforms[platform] = err == nil
			}
		}
	}

	return status, nil
}

// ListCollected returns all collected binaries
func (c *DefaultCollector) ListCollected() ([]CollectedBinary, error) {
	var collected []CollectedBinary

	binariesDir := filepath.Join(c.config.CollectionDir, "binaries")
	if _, err := os.Stat(binariesDir); os.IsNotExist(err) {
		return collected, nil
	}

	err := filepath.Walk(binariesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "manifest.json" {
			manifest, err := c.loadManifest(path)
			if err != nil {
				log.Warn("Failed to load manifest", "path", path, "error", err)
				return nil
			}

			var totalSize int64
			var platforms []string
			for platform, platformInfo := range manifest.Platforms {
				platforms = append(platforms, platform)
				totalSize += platformInfo.Size
			}

			collected = append(collected, CollectedBinary{
				Name:         manifest.Binary,
				Version:      manifest.Version,
				Platforms:    platforms,
				Size:         totalSize,
				CollectedAt:  manifest.CollectionDate,
				ManifestPath: path,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk collection directory: %w", err)
	}

	return collected, nil
}

// collectForPlatform collects a binary for a specific platform
func (c *DefaultCollector) collectForPlatform(ctx context.Context, binary *DepBinary, platform string) *PlatformResult {
	startTime := time.Now()
	
	result := &PlatformResult{
		Platform: platform,
		Success:  false,
	}

	defer func() {
		result.Duration = time.Since(startTime)
	}()

	// Parse platform
	platformConfig, err := c.config.GetPlatformConfig(platform)
	if err != nil {
		result.Error = fmt.Sprintf("invalid platform: %v", err)
		return result
	}

	// Create temporary directory for download
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("collect-%s-%s-%s", binary.Name, binary.Version, platform))
	if err != nil {
		result.Error = fmt.Sprintf("failed to create temp dir: %v", err)
		return result
	}
	defer os.RemoveAll(tempDir)

	// Create installer for this platform
	installer, err := c.createPlatformInstaller(binary, platformConfig)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create installer: %v", err)
		return result
	}

	// Download binary
	tempBinaryPath, err := c.downloadBinaryForPlatform(installer, binary, platformConfig, tempDir)
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
	finalPath := c.config.GetBinaryPath(binary.Name, binary.Version, platform)
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

	log.Info("Platform collection successful", 
		"binary", binary.Name,
		"version", binary.Version,
		"platform", platform,
		"size", util.FormatBytes(result.Size),
		"duration", result.Duration.Round(time.Millisecond))

	return result
}

// createPlatformInstaller creates an installer configured for a specific platform
func (c *DefaultCollector) createPlatformInstaller(binary *DepBinary, platformConfig *PlatformConfig) (any, error) {
	// For now, we'll use a simplified approach
	// In a full implementation, we'd need to mock the platform detection
	// to make installers think they're running on the target platform
	
	// This is a placeholder - full implementation would need platform simulation
	return nil, fmt.Errorf("platform installer creation not implemented yet")
}

// downloadBinaryForPlatform downloads a binary for a specific platform
func (c *DefaultCollector) downloadBinaryForPlatform(installer any, binary *DepBinary, platformConfig *PlatformConfig, tempDir string) (string, error) {
	// This is a simplified implementation
	// Full implementation would need to:
	// 1. Mock runtime.GOOS and runtime.GOARCH for the target platform
	// 2. Use the appropriate installer with platform-specific logic
	// 3. Handle different binary types (GitHub releases, Claude releases, go build, etc.)
	
	// For now, return an error indicating this needs implementation
	return "", fmt.Errorf("cross-platform collection not yet implemented - would need to simulate %s/%s environment", platformConfig.OS, platformConfig.Arch)
}

// calculateSHA256 calculates the SHA256 hash of a file
func (c *DefaultCollector) calculateSHA256(filePath string) (string, error) {
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

// moveBinary moves a binary file to its final location
func (c *DefaultCollector) moveBinary(src, dst string) error {
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

// generateManifest creates a manifest from collection results
func (c *DefaultCollector) generateManifest(result *CollectionResult) (*BinaryManifest, error) {
	manifest := &BinaryManifest{
		Binary:         result.Binary,
		Version:        result.Version,
		CollectionDate: result.CollectedAt,
		Platforms:      make(map[string]*PlatformInfo),
	}

	// Add source information (simplified)
	manifest.Source = &SourceInfo{
		Type:    "github-release", // This should be determined from the actual installer type
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

// saveManifest saves a manifest to disk
func (c *DefaultCollector) saveManifest(manifest *BinaryManifest, path string) error {
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

// loadManifest loads a manifest from disk
func (c *DefaultCollector) loadManifest(path string) (*BinaryManifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var manifest BinaryManifest
	if err := json.NewDecoder(file).Decode(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}