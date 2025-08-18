package dep

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/joeblew999/infra/pkg/log"
)

// GitHubRelease represents GitHub API response for latest release
// This is a minimal struct - we only need the tag_name field
// We use interface{} for unused fields to reduce memory usage
// https://docs.github.com/en/rest/releases/releases#get-the-latest-release
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// CheckGitHubRelease checks the latest GitHub release for a repository
func CheckGitHubRelease(owner, repo string) (GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return GitHubRelease{}, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return GitHubRelease{}, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GitHubRelease{}, fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, &release); err != nil {
		return GitHubRelease{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return release, nil
}

// CheckAllReleases checks all configured binaries for their latest versions
func CheckAllReleases() error {
	binaries, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	for _, binary := range binaries {
		if binary.Repo == "" {
			continue
		}

		// Skip npm registry sources (like claude)
		if binary.ReleaseURL != "" && binary.ReleaseURL != fmt.Sprintf("https://github.com/%s/releases", binary.Repo) {
			log.Info("Skipping non-GitHub source", "binary", binary.Name, "source", binary.ReleaseURL)
			continue
		}

		// Parse owner/repo from repo string
		parts := []string{}
		for i, c := range binary.Repo {
			if c == '/' {
				parts = append(parts, binary.Repo[:i], binary.Repo[i+1:])
				break
			}
		}

		if len(parts) != 2 {
			log.Warn("Invalid repo format", "repo", binary.Repo)
			continue
		}

		owner, repo := parts[0], parts[1]

		release, err := CheckGitHubRelease(owner, repo)
		if err != nil {
			log.Error("Failed to check release", "binary", binary.Name, "error", err)
			continue
		}

		status := "✓"
		if release.TagName != binary.Version {
			status = "↑"
		}

		log.Info("Release check",
			"binary", binary.Name,
			"current", binary.Version,
			"latest", release.TagName,
			"status", status)
	}

	return nil
}
