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

type deckToolsInstaller struct{}

func (i *deckToolsInstaller) Install(binary DepBinary, debug bool) error {
	log.Info("Installing deck tools package", "version", binary.Version)

	installDir := config.GetDepPath()
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory %s: %w", installDir, err)
	}

	var release *gitHubRelease
	var err error

	if debug {
		log.Info("Using gh cli for deck-tools release info (debug mode)")
		release, err = getGitHubReleaseDebug(binary.Repo, binary.Version)
	} else {
		release, err = getGitHubRelease(binary.Repo, binary.Version)
	}

	if err != nil {
		return fmt.Errorf("failed to get GitHub release for deck-tools: %w", err)
	}

	asset, err := selectAsset(release, binary.Assets)
	if err != nil {
		return fmt.Errorf("failed to select asset for deck-tools: %w", err)
	}

	tmpDir := filepath.Join(config.GetDepPath(), "tmp", "deck-tools")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory %s: %w", tmpDir, err)
	}
	defer os.RemoveAll(tmpDir)

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

	// List of tools to install from the package
	tools := []string{"decksh", "dshfmt", "dshlint", "svgdeck", "pngdeck", "pdfdeck"}

	// Install each tool
	for _, tool := range tools {
		binaryName := tool
		if runtime.GOOS == "windows" {
			binaryName += ".exe"
		}

		// Look for the binary in the extracted files
		var binaryPath string
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

		if binaryPath == "" {
			return fmt.Errorf("binary %s not found in extracted deck-tools package", binaryName)
		}

		// Move the binary to the final location
		finalPath := config.Get(tool)
		if err := os.Rename(binaryPath, finalPath); err != nil {
			return fmt.Errorf("failed to move %s to final location: %w", tool, err)
		}

		// Make the binary executable
		if err := os.Chmod(finalPath, 0755); err != nil {
			return fmt.Errorf("failed to make %s executable: %w", tool, err)
		}

		// Write metadata for this tool
		if err := writeMeta(finalPath, &BinaryMeta{Name: tool, Version: binary.Version}); err != nil {
			return fmt.Errorf("failed to write metadata for %s: %w", tool, err)
		}

		log.Info("Tool installed", "tool", tool, "path", finalPath)
	}

	log.Info("All deck tools installed successfully")
	return nil
}