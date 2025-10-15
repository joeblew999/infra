package profiles

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
)

// ResolveProfile returns the tooling profile matching the override or the active default.
func ResolveProfile(override string) (sharedcfg.ToolingProfile, string) {
	settings := sharedcfg.Tooling()
	trimmed := strings.TrimSpace(override)
	if trimmed != "" {
		if profile, ok := settings.Profiles[trimmed]; ok {
			return profile, profile.Name
		}
	}
	active := settings.Active
	if strings.TrimSpace(active.Name) == "" {
		active.Name = "local"
	}
	return active, active.Name
}

// FindRepoRoot walks parent directories looking for go.work or .git.
func FindRepoRoot(start string) (string, error) {
	dir := strings.TrimSpace(start)
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir = wd
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", errors.New("profiles: unable to locate repository root (missing go.work or .git)")
}

// ResolveCoreDir returns the core directory for the tooling workflow.
func ResolveCoreDir(repoRoot string) string {
	root := strings.TrimSpace(repoRoot)
	if root == "" {
		return "core"
	}
	return filepath.Join(root, "core")
}

// FirstNonEmpty returns the first non-empty trimmed value.
func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
