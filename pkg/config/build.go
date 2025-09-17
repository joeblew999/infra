package config

import (
	"fmt"
	"os/exec"
	"strings"
)

// Build info set by main - these are the canonical build variables
var (
	GitHash   = "dev"
	BuildTime = "unknown"
)

// SetBuildInfo sets build information for display
func SetBuildInfo(gitHash, buildTime string) {
	GitHash = gitHash
	BuildTime = buildTime
}

// GetShortHash returns first 7 characters of git hash
func GetShortHash() string {
	if len(GitHash) >= 7 {
		return GitHash[:7]
	}
	return GitHash
}

// GetVersion returns semantic version based on git hash
func GetVersion() string {
	if GitHash == "dev" {
		return "dev"
	}
	return GetShortHash()
}

// GetFullVersionString returns complete version info
func GetFullVersionString() string {
	version := GetVersion()
	shortHash := GetShortHash()
	return fmt.Sprintf("%s (build: %s, time: %s)", version, shortHash, BuildTime)
}

// GetRuntimeGitHash gets git hash at runtime (centralized implementation)
func GetRuntimeGitHash() string {
	if cmd := exec.Command("git", "rev-parse", "HEAD"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	return ""
}

// GetRuntimeGitBranch gets git branch at runtime
func GetRuntimeGitBranch() string {
	if cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	return ""
}

