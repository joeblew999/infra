package dep

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

type tokiInstaller struct{}

func (i *tokiInstaller) Install(binary DepBinary, debug bool) error {
	log.Info("Attempting download and installation", "binary", binary.Name)

	installPath, err := Get(binary.Name)
	if err != nil {
		return fmt.Errorf("failed to get install path for %s: %w", binary.Name, err)
	}

	var release *gitHubRelease

	if debug {
		log.Info("Using gh cli for Toki release info (debug mode).")
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

	log.Info("Downloaded asset", "asset_name", asset.Name, "path", assetPath)

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

	// Look for the toki binary in the extracted directory
	// Toki binaries are typically named toki with the version in the directory structure
	srcPath := filepath.Join(tmpDir, "toki")
	if runtime.GOOS == "windows" {
		srcPath += ".exe"
	}

	// If direct path doesn't work, try searching in the extracted directory
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		// Search for toki binary in the extracted directory
		possiblePaths := []string{
			filepath.Join(tmpDir, "toki"),
			filepath.Join(tmpDir, "toki", "toki"),
			filepath.Join(tmpDir, fmt.Sprintf("toki_%s_%s_%s", binary.Version[1:], runtime.GOOS, runtime.GOARCH), "toki"),
		}
		
		for _, path := range possiblePaths {
			if runtime.GOOS == "windows" {
				path += ".exe"
			}
			if _, err := os.Stat(path); err == nil {
				srcPath = path
				break
			}
		}
	}

	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("toki binary not found in extracted archive")
	}

	if err := os.Rename(srcPath, installPath); err != nil {
		return fmt.Errorf("failed to move binary from %s to %s: %w", srcPath, installPath, err)
	}

	if err := os.Chmod(installPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions for %s: %w", installPath, err)
	}

	log.Info("Successfully installed binary", "binary", binary.Name, "path", installPath)
	return nil
}