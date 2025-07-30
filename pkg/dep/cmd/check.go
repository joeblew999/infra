package cmd

import (
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/dep"
)

// CheckReleasesCmd checks for latest versions of all configured binaries
func CheckReleasesCmd(args []string) error {
	fmt.Println("Checking latest releases for configured binaries...")
	fmt.Println()
	
	if err := dep.CheckAllReleases(); err != nil {
		fmt.Fprintf(os.Stderr, "Error checking releases: %v\n", err)
		return err
	}
	
	return nil
}

// CheckReleaseCmd checks a specific binary's latest release
func CheckReleaseCmd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: check-release <binary-name>")
	}

	binaryName := args[0]
	
	binaries, err := dep.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	var targetBinary *dep.DepBinary
	for _, b := range binaries {
		if b.Name == binaryName {
			targetBinary = &b
			break
		}
	}

	if targetBinary == nil {
		return fmt.Errorf("binary '%s' not found in configuration", binaryName)
	}

	// Skip npm registry sources
	if targetBinary.ReleaseURL != "" && targetBinary.ReleaseURL != fmt.Sprintf("https://github.com/%s/releases", targetBinary.Repo) {
		fmt.Printf("Binary '%s' uses non-GitHub source: %s\n", binaryName, targetBinary.ReleaseURL)
		return nil
	}

	// Parse owner/repo
	parts := []string{}
	for i, c := range targetBinary.Repo {
		if c == '/' {
			parts = append(parts, targetBinary.Repo[:i], targetBinary.Repo[i+1:])
			break
		}
	}
	
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: %s", targetBinary.Repo)
	}

	owner, repo := parts[0], parts[1]
	
	release, err := dep.CheckGitHubRelease(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to check release: %w", err)
	}

	fmt.Printf("Binary: %s\n", binaryName)
	fmt.Printf("Current: %s\n", targetBinary.Version)
	fmt.Printf("Latest: %s\n", release.TagName)
	fmt.Printf("Status: ")
	
	if release.TagName == targetBinary.Version {
		fmt.Println("✓ Up to date")
	} else {
		fmt.Printf("↑ Update available (%s → %s)\n", targetBinary.Version, release.TagName)
	}

	if release.Prerelease {
		fmt.Println("⚠ Latest is a prerelease")
	}
	
	return nil
}