package dep

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// zigInstaller handles zig compiler installation
// Zig is a general-purpose programming language and toolchain for maintaining robust, optimal, and reusable software.
type zigInstaller struct{}

// Install installs the zig compiler
func (i *zigInstaller) Install(binary DepBinary, debug bool) error {
	assetURL, err := i.getAssetURL(binary)
	if err != nil {
		return fmt.Errorf("failed to get zig asset URL: %w", err)
	}

	// Download the asset
	archivePath, err := i.downloadAsset(assetURL, debug)
	if err != nil {
		return fmt.Errorf("failed to download zig: %w", err)
	}
	defer os.Remove(archivePath)

	// Extract based on the expected format from URL
	extractDir := filepath.Join(config.GetDepPath(), "zig-temp")
	
	// Determine format based on the original URL
	if strings.Contains(assetURL, ".tar.xz") {
		if err := i.untarXz(archivePath, extractDir, debug); err != nil {
			return fmt.Errorf("failed to extract tar.xz: %w", err)
		}
	} else if strings.Contains(assetURL, ".zip") {
		if err := i.unzip(archivePath, extractDir, debug); err != nil {
			return fmt.Errorf("failed to extract zip: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported archive format for zig")
	}

	// Find the zig binary
	var zigBinaryPath string
	err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "zig" {
			zigBinaryPath = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to find zig binary: %w", err)
	}

	if zigBinaryPath == "" {
		return fmt.Errorf("zig binary not found in archive")
	}

	// Move binary to final location
	finalPath := config.Get("zig")
	if err := os.Rename(zigBinaryPath, finalPath); err != nil {
		return fmt.Errorf("failed to move zig binary: %w", err)
	}

	// Make binary executable
	if err := os.Chmod(finalPath, 0755); err != nil {
		return fmt.Errorf("failed to make zig executable: %w", err)
	}

	// Cleanup temp directory
	os.RemoveAll(extractDir)

	log.Info("zig installed successfully", "path", finalPath, "version", binary.Version)
	return nil
}

// getAssetURL returns the correct asset URL for the current platform
func (i *zigInstaller) getAssetURL(binary DepBinary) (string, error) {
	os, arch := runtime.GOOS, runtime.GOARCH
	
	// Map to zig naming conventions
	zigOS := map[string]string{
		"darwin":  "macos",
		"linux":   "linux",
		"windows": "windows",
	}[os]
	
	zigArch := map[string]string{
		"amd64": "x86_64",
		"arm64": "aarch64",
	}[arch]
	
	if zigOS == "" || zigArch == "" {
		return "", fmt.Errorf("unsupported platform: %s/%s", os, arch)
	}
	
	// Construct the asset filename based on actual zig naming
	var assetName string
	if os == "windows" {
		assetName = fmt.Sprintf("zig-%s-%s-%s.zip", zigArch, zigOS, binary.Version)
	} else {
		assetName = fmt.Sprintf("zig-%s-%s-%s.tar.xz", zigArch, zigOS, binary.Version)
	}
	
	// Use ziglang.org direct download URLs
	return fmt.Sprintf("https://ziglang.org/download/%s/%s", binary.Version, assetName), nil
}

// downloadAsset downloads the asset from the given URL
func (i *zigInstaller) downloadAsset(url string, debug bool) (string, error) {
	if debug {
		log.Info("Downloading asset", "url", url)
	}

	// Create a temporary file
	tempFile, err := os.CreateTemp(config.GetDepPath(), "zig-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download asset: HTTP %d", resp.StatusCode)
	}

	// Copy the response body to the temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save asset: %w", err)
	}

	return tempFile.Name(), nil
}

// untarXz extracts .tar.xz files
func (i *zigInstaller) untarXz(src, dest string, debug bool) error {
	// Note: This is a simplified extraction - in production, you might want to use
	// a proper xz decoder, but for now we'll use system tar command
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	cmd := exec.Command("tar", "-xf", src, "-C", dest)
	if debug {
		log.Info("Extracting tar.xz", "command", cmd.String())
	}
	return cmd.Run()
}

// unzip extracts .zip files
func (i *zigInstaller) unzip(src, dest string, debug bool) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, filepath.FromSlash(f.Name))
		
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		inFile, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, inFile)
		inFile.Close()
		outFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}