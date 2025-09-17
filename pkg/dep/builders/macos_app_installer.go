package builders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joeblew999/infra/pkg/dep/util"
	"github.com/joeblew999/infra/pkg/log"
)

// MacOSAppInstaller handles macOS .app installation from DMG files
type MacOSAppInstaller struct{}

// Install downloads and installs a macOS .app from DMG
func (i *MacOSAppInstaller) Install(name, repo, version string, assets []AssetSelector, debug bool) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("%s is only available for macOS", name)
	}

	log.Info("Installing macOS app", "name", name, "version", version)

	// Determine app name and installation path
	appName := fmt.Sprintf("%s.app", strings.Title(name))
	appPath := filepath.Join("/Applications", appName)
	
	// Check if app is already installed
	if _, err := os.Stat(appPath); err == nil {
		log.Info("App already installed", "name", name, "path", appPath)
		return nil
	}

	// Find the appropriate asset
	asset, err := i.findMatchingAsset(name, repo, version, assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		// For macOS apps, try universal architecture if specific arch not found
		if runtime.GOARCH == "arm64" || runtime.GOARCH == "amd64" {
			asset, err = i.findMatchingAsset(name, repo, version, assets, runtime.GOOS, "universal")
		}
		if err != nil {
			return fmt.Errorf("no matching asset found for %s: %w", name, err)
		}
	}

	// Download the DMG file
	dmgPath := filepath.Join(".dep", fmt.Sprintf("%s.dmg", name))
	if err := os.MkdirAll(".dep", 0755); err != nil {
		return fmt.Errorf("failed to create .dep directory: %w", err)
	}

	log.Info("Downloading DMG", "name", name, "url", asset.URL, "output", dmgPath)
	if err := util.DownloadFile(asset.URL, dmgPath, true); err != nil {
		return fmt.Errorf("failed to download %s.dmg: %w", name, err)
	}

	// Mount the DMG
	log.Info("Mounting DMG", "name", name)
	mountPoint, err := i.mountDMG(dmgPath, debug)
	if err != nil {
		return fmt.Errorf("failed to mount DMG: %w", err)
	}
	defer i.unmountDMG(mountPoint, debug)

	// Copy .app to Applications
	sourceApp := filepath.Join(mountPoint, appName)
	if _, err := os.Stat(sourceApp); err != nil {
		return fmt.Errorf("%s not found in DMG at %s: %w", appName, sourceApp, err)
	}

	log.Info("Installing app to /Applications", "name", name, "app", appName)
	if err := i.copyApp(sourceApp, appPath, debug); err != nil {
		return fmt.Errorf("failed to install %s: %w", appName, err)
	}

	// Clean up downloaded DMG
	if err := os.Remove(dmgPath); err != nil {
		log.Warn("Failed to clean up DMG file", "path", dmgPath, "error", err)
	}

	log.Info("macOS app installation completed", "name", name, "path", appPath)
	return nil
}

// Uninstall removes a macOS .app from the Applications directory
func (i *MacOSAppInstaller) Uninstall(name string, debug bool) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("%s uninstall is only available for macOS", name)
	}

	// Determine app name and installation path
	appName := fmt.Sprintf("%s.app", strings.Title(name))
	appPath := filepath.Join("/Applications", appName)
	
	// Check if app exists
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		log.Info("App not installed", "name", name, "path", appPath)
		return nil
	}

	log.Info("Uninstalling macOS app", "name", name, "path", appPath)
	
	// Remove the .app bundle
	if err := os.RemoveAll(appPath); err != nil {
		return fmt.Errorf("failed to remove app bundle: %w", err)
	}

	log.Info("macOS app uninstalled successfully", "name", name)
	return nil
}

// mountDMG mounts a DMG file and returns the mount point
func (i *MacOSAppInstaller) mountDMG(dmgPath string, debug bool) (string, error) {
	cmd := exec.Command("hdiutil", "attach", dmgPath, "-nobrowse", "-quiet")
	if debug {
		log.Debug("Mounting DMG", "command", cmd.String())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("hdiutil attach failed: %w", err)
	}

	// Parse hdiutil output to find mount point
	// Output format: "/dev/disk2s1     Apple_HFS     /Volumes/UTM"
	lines := parseHdiutilOutput(string(output))
	for _, line := range lines {
		if filepath.HasPrefix(line, "/Volumes/") {
			return line, nil
		}
	}

	return "", fmt.Errorf("could not determine mount point from hdiutil output")
}

// unmountDMG unmounts a DMG
func (i *MacOSAppInstaller) unmountDMG(mountPoint string, debug bool) error {
	cmd := exec.Command("hdiutil", "detach", mountPoint, "-quiet")
	if debug {
		log.Debug("Unmounting DMG", "command", cmd.String(), "mountPoint", mountPoint)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		log.Warn("Failed to unmount DMG", "mountPoint", mountPoint, "error", err)
		return err
	}

	return nil
}

// copyApp copies an .app bundle using cp -R
func (i *MacOSAppInstaller) copyApp(source, dest string, debug bool) error {
	// Remove existing app if present
	if _, err := os.Stat(dest); err == nil {
		if err := os.RemoveAll(dest); err != nil {
			return fmt.Errorf("failed to remove existing app: %w", err)
		}
	}

	cmd := exec.Command("cp", "-R", source, dest)
	if debug {
		log.Debug("Copying app", "command", cmd.String())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy app bundle: %w", err)
	}

	return nil
}

// parseHdiutilOutput parses hdiutil attach output to extract mount points
func parseHdiutilOutput(output string) []string {
	var mountPoints []string
	lines := splitLines(output)
	
	for _, line := range lines {
		fields := splitFields(line)
		for _, field := range fields {
			if filepath.HasPrefix(field, "/Volumes/") {
				mountPoints = append(mountPoints, field)
			}
		}
	}
	
	return mountPoints
}

// splitLines splits text into lines
func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	
	var lines []string
	start := 0
	for i, r := range text {
		if r == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	
	return lines
}

// splitFields splits a line into whitespace-separated fields
func splitFields(line string) []string {
	var fields []string
	var current string
	var inField bool
	
	for _, r := range line {
		if r == ' ' || r == '\t' {
			if inField {
				fields = append(fields, current)
				current = ""
				inField = false
			}
		} else {
			current += string(r)
			inField = true
		}
	}
	
	if inField {
		fields = append(fields, current)
	}
	
	return fields
}

// Asset represents a download asset
type Asset struct {
	URL      string
	Name     string
	OS       string
	Arch     string
	Size     int64
}

// findMatchingAsset finds the best matching asset for the given platform
func (i *MacOSAppInstaller) findMatchingAsset(name, repo, version string, assets []AssetSelector, targetOS, targetArch string) (*Asset, error) {
	// For macOS apps, we typically have DMG files
	// This is a simplified implementation - in production you'd want to:
	// 1. Fetch release information from GitHub API
	// 2. Parse asset names and match against patterns
	// 3. Handle different naming conventions
	
	for _, assetSelector := range assets {
		if assetSelector.OS == targetOS && assetSelector.Arch == targetArch {
			// Build the download URL - this is a simplified approach
			// In practice, you'd fetch from GitHub releases API
			baseURL := fmt.Sprintf("https://github.com/%s/releases/download/%s", repo, version)
			
			// For UTM specifically, we know it's UTM.dmg
			var assetName string
			if name == "utm" {
				assetName = "UTM.dmg"
			} else {
				// Generic approach for other apps
				assetName = fmt.Sprintf("%s.dmg", strings.Title(name))
			}
			
			return &Asset{
				URL:  fmt.Sprintf("%s/%s", baseURL, assetName),
				Name: assetName,
				OS:   targetOS,
				Arch: targetArch,
			}, nil
		}
	}
	
	return nil, fmt.Errorf("no matching asset found for %s/%s", targetOS, targetArch)
}