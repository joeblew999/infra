package dep

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/joeblew999/infra/pkg/log"
)

// BinaryStatus represents the status of a binary
type BinaryStatus struct {
	Name              string
	Configured        bool
	Installed         bool
	ConfiguredVersion string
	InstalledVersion  string
	UpToDate          bool
	InstallPath       string
	LastModified      *time.Time
	Size              int64
	Source            string
}

// ShowStatus displays the status of all configured binaries
func ShowStatus() error {
	binaries, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("Binary Status Report (%d configured)\n", len(binaries))
	fmt.Println(strings.Repeat("=", 80))

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tINSTALLED\tCONFIGURED\tUP-TO-DATE\tSOURCE")
	fmt.Fprintln(w, strings.Repeat("-", 60))

	installedCount := 0
	upToDateCount := 0

	for _, binary := range binaries {
		status, err := getBinaryStatus(binary)
		if err != nil {
			log.Warn("Failed to get status", "binary", binary.Name, "error", err)
			continue
		}

		if status.Installed {
			installedCount++
		}
		if status.UpToDate {
			upToDateCount++
		}

		// Format status indicators
		statusIcon := "❌"
		if status.Installed {
			if status.UpToDate {
				statusIcon = "✅"
			} else {
				statusIcon = "⚠️"
			}
		}

		installedVersion := "not installed"
		if status.Installed {
			installedVersion = status.InstalledVersion
		}

		upToDateIcon := "❌"
		if status.UpToDate {
			upToDateIcon = "✅"
		} else if !status.Installed {
			upToDateIcon = "➖"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			status.Name,
			statusIcon,
			installedVersion,
			status.ConfiguredVersion,
			upToDateIcon,
			status.Source,
		)
	}

	w.Flush()

	// Summary
	fmt.Println()
	fmt.Printf("Summary: %d/%d installed, %d/%d up-to-date\n",
		installedCount, len(binaries),
		upToDateCount, len(binaries))

	if installedCount < len(binaries) {
		fmt.Printf("\nTo install missing binaries: go run . dep install\n")
	}

	if upToDateCount < installedCount {
		fmt.Printf("To upgrade outdated binaries: go run . dep upgrade\n")
	}

	return nil
}

// ShowBinaryStatus displays detailed status for a specific binary
func ShowBinaryStatus(name string) error {
	binaries, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Find the binary
	var targetBinary *DepBinary
	for _, binary := range binaries {
		if binary.Name == name {
			targetBinary = &binary
			break
		}
	}

	if targetBinary == nil {
		return fmt.Errorf("binary '%s' not found in configuration", name)
	}

	status, err := getBinaryStatus(*targetBinary)
	if err != nil {
		return fmt.Errorf("failed to get status for %s: %w", name, err)
	}

	// Display detailed information
	fmt.Printf("Binary: %s\n", status.Name)
	fmt.Printf("Description: %s\n", targetBinary.Description)
	fmt.Printf("Repository: %s\n", targetBinary.Repo)
	fmt.Printf("Source Type: %s\n", status.Source)
	fmt.Printf("Configured Version: %s\n", status.ConfiguredVersion)

	if status.Installed {
		fmt.Printf("Status: ✅ Installed\n")
		fmt.Printf("Installed Version: %s\n", status.InstalledVersion)
		fmt.Printf("Install Path: %s\n", status.InstallPath)
		
		if status.LastModified != nil {
			fmt.Printf("Last Modified: %s\n", status.LastModified.Format("2006-01-02 15:04:05"))
		}
		
		if status.Size > 0 {
			fmt.Printf("Size: %s\n", formatBytes(status.Size))
		}

		if status.UpToDate {
			fmt.Printf("Up-to-date: ✅ Yes\n")
		} else {
			fmt.Printf("Up-to-date: ⚠️  No (configured: %s, installed: %s)\n", 
				status.ConfiguredVersion, status.InstalledVersion)
			fmt.Printf("\nTo upgrade: go run . dep upgrade %s\n", name)
		}
	} else {
		fmt.Printf("Status: ❌ Not installed\n")
		fmt.Printf("\nTo install: go run . dep install %s\n", name)
	}

	return nil
}

// getBinaryStatus retrieves the status information for a binary
func getBinaryStatus(binary DepBinary) (*BinaryStatus, error) {
	status := &BinaryStatus{
		Name:              binary.Name,
		Configured:        true,
		ConfiguredVersion: binary.Version,
		Source:            binary.Source,
	}

	// Get install path
	installPath, err := Get(binary.Name)
	if err != nil {
		return status, nil // Binary not found in config, but we'll still return basic info
	}
	status.InstallPath = installPath

	// Check if binary is installed
	if fileInfo, err := os.Stat(installPath); err == nil {
		status.Installed = true
		modTime := fileInfo.ModTime()
		status.LastModified = &modTime
		status.Size = fileInfo.Size()

		// Try to read metadata
		if meta, err := readMeta(installPath); err == nil {
			status.InstalledVersion = meta.Version
			
			// For "latest" versions, we assume they're up-to-date since we can't easily compare
			if binary.Version == "latest" {
				status.UpToDate = true
			} else {
				status.UpToDate = (meta.Version == binary.Version)
			}
		} else {
			// No metadata, assume it needs updating
			status.InstalledVersion = "unknown"
			status.UpToDate = false
		}
	}

	return status, nil
}

// UpgradeAll upgrades all installed binaries
func UpgradeAll(debug bool) error {
	binaries, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	upgraded := 0
	skipped := 0
	errors := 0

	for _, binary := range binaries {
		status, err := getBinaryStatus(binary)
		if err != nil {
			log.Warn("Failed to get status", "binary", binary.Name, "error", err)
			errors++
			continue
		}

		if !status.Installed {
			fmt.Printf("⬇️  Installing %s (not currently installed)...\n", binary.Name)
		} else if status.UpToDate {
			fmt.Printf("✅ %s is already up-to-date (%s)\n", binary.Name, status.InstalledVersion)
			skipped++
			continue
		} else {
			fmt.Printf("⬆️  Upgrading %s (%s → %s)...\n", binary.Name, status.InstalledVersion, status.ConfiguredVersion)
		}

		if err := InstallBinary(binary.Name, debug); err != nil {
			fmt.Printf("❌ Failed to upgrade %s: %v\n", binary.Name, err)
			errors++
		} else {
			fmt.Printf("✅ %s upgraded successfully\n", binary.Name)
			upgraded++
		}
	}

	fmt.Printf("\nUpgrade complete: %d upgraded, %d skipped, %d errors\n", upgraded, skipped, errors)
	return nil
}

// UpgradeBinary upgrades a specific binary
func UpgradeBinary(name string, debug bool) error {
	binaries, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Find the binary
	var targetBinary *DepBinary
	for _, binary := range binaries {
		if binary.Name == name {
			targetBinary = &binary
			break
		}
	}

	if targetBinary == nil {
		return fmt.Errorf("binary '%s' not found in configuration", name)
	}

	status, err := getBinaryStatus(*targetBinary)
	if err != nil {
		return fmt.Errorf("failed to get status for %s: %w", name, err)
	}

	if !status.Installed {
		fmt.Printf("Installing %s (not currently installed)...\n", name)
	} else if status.UpToDate {
		fmt.Printf("%s is already up-to-date (%s)\n", name, status.InstalledVersion)
		return nil
	} else {
		fmt.Printf("Upgrading %s (%s → %s)...\n", name, status.InstalledVersion, status.ConfiguredVersion)
	}

	return InstallBinary(name, debug)
}

// formatBytes formats byte counts in human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}