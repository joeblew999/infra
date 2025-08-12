package dep

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/config"
)

type dshlintInstaller struct{}

func (i *dshlintInstaller) Install(binary DepBinary, debug bool) error {
	log.Info("Attempting download and installation", "binary", binary.Name)

	installPath, err := Get(binary.Name)
	if err != nil {
		return fmt.Errorf("failed to get install path for %s: %w", binary.Name, err)
	}

	var release *gitHubRelease

	if debug {
		log.Info("Using gh cli for dshlint release info (debug mode).")
		release, err = getGitHubReleaseDebug(binary.Repo, binary.Version)
	} else {
		release, err = getGitHubRelease(binary.Repo, binary.Version)
	}

	if err != nil {
		return fmt.Errorf("failed to get GitHub release for %s: %w", binary.Name, err)
	}

	asset, err := selectAsset(release, binary.Assets)
	if err != nil {
		return fmt.Errorf("failed to select asset for %s: %w", binary.Name, err)
	}

	tmpDir := filepath.Join(config.GetDepPath(), "tmp", binary.Name)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory %s: %w", tmpDir, err)
	}
	defer os.RemoveAll(tmpDir) // Clean up temporary directory

	assetPath, err := downloadFile(asset.BrowserDownloadURL, tmpDir, asset.Name)
	if err != nil {
		return fmt.Errorf("failed to download asset %s: %w", asset.Name, err)
	}

	// Handle different archive formats
	if strings.HasSuffix(asset.Name, ".zip") {
		if err := unzip(assetPath, tmpDir); err != nil {
			return fmt.Errorf("failed to unzip %s: %w", asset.Name, err)
		}
	} else if strings.HasSuffix(asset.Name, ".tar.gz") {
		if err := untarGz(assetPath, tmpDir); err != nil {
			return fmt.Errorf("failed to untar.gz %s: %w", asset.Name, err)
		}
	} else {
		return fmt.Errorf("unsupported archive format for %s", asset.Name)
	}

	// Find the binary within the extracted files
	binaryName := "dshlint"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	binaryPath := filepath.Join(tmpDir, binaryName)
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Try to find the binary in any subdirectory
		_ = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Base(path) == binaryName {
				binaryPath = path
				return filepath.SkipDir
			}
			return nil
		})
	}

	// Ensure the binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("binary %s not found in extracted files", binaryName)
	}

	// Move the binary to the final location
	if err := os.Rename(binaryPath, installPath); err != nil {
		return fmt.Errorf("failed to move binary to final location: %w", err)
	}

	// Make the binary executable
	if err := os.Chmod(installPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	log.Info("Binary installed successfully", "binary", binary.Name, "path", installPath)
	return nil
}