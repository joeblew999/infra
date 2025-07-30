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

type natsInstaller struct{}

func (i *natsInstaller) Install(binary DepBinary, debug bool) error {
	log.Info("Attempting download and installation", "binary", binary.Name)

	installPath, err := Get(binary.Name)
	if err != nil {
		return fmt.Errorf("failed to get install path for %s: %w", binary.Name, err)
	}

	var release *gitHubRelease

	if debug && binary.Name == "nats" {
		log.Info("Using gh cli for NATS release info (debug mode).")
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

	// Move the extracted binary to its final destination
	// For NATS CLI, the binary is inside a versioned subdirectory
	versionedDir := fmt.Sprintf("nats-%s-%s-%s", strings.TrimPrefix(binary.Version, "v"), runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "darwin" {
		versionedDir = fmt.Sprintf("nats-%s-darwin-%s", strings.TrimPrefix(binary.Version, "v"), runtime.GOARCH)
	}
	if runtime.GOOS == "linux" {
		versionedDir = fmt.Sprintf("nats-%s-linux-%s", strings.TrimPrefix(binary.Version, "v"), runtime.GOARCH)
	}
	if runtime.GOOS == "windows" {
		versionedDir = fmt.Sprintf("nats-%s-windows-%s", strings.TrimPrefix(binary.Version, "v"), runtime.GOARCH)
	}
	
	srcPath := filepath.Join(tmpDir, versionedDir, "nats")
	if runtime.GOOS == "windows" {
		srcPath += ".exe"
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