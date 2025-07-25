package dep

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/joeblew999/infra/pkg/log"
)

// GitHubReleaseAsset represents a single asset in a GitHub release.
type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GitHubRelease represents a GitHub release.
type GitHubRelease struct {
	Assets []GitHubReleaseAsset `json:"assets"`
}

// getGitHubRelease fetches release information directly from GitHub API using net/http.
func getGitHubRelease(repo, version string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, version)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub release from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for %s", resp.StatusCode, url)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub release response: %w", err)
	}

	return &release, nil
}

// getGitHubReleaseDebug fetches release information using gh cli.
func getGitHubReleaseDebug(repo, version string) (*GitHubRelease, error) {
	cmd := exec.Command("gh", "release", "view", version, "--repo", repo, "--json", "assets")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Info("Running command", "command", cmd.Args)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run gh cli: %w\nStderr: %s", err, stderr.String())
	}

	var release GitHubRelease
	if err := json.NewDecoder(&stdout).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode gh cli output: %w", err)
	}

	return &release, nil
}

// selectAsset selects the appropriate asset from a GitHub release based on OS, Arch, and regex match.
func selectAsset(release *GitHubRelease, selectors []AssetSelector) (*GitHubReleaseAsset, error) {
	for _, selector := range selectors {
		if selector.OS == runtime.GOOS && selector.Arch == runtime.GOARCH {
			for _, asset := range release.Assets {
				if matched, _ := regexp.MatchString(selector.Match, asset.Name); matched {
					return &asset, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no matching asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
}

// downloadFile downloads a file from a URL to a specified directory.
func downloadFile(url, destDir, filename string) (string, error) {
	filepath := filepath.Join(destDir, filename)

	out, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", filepath, err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d for %s", resp.StatusCode, url)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy downloaded content to file: %w", err)
	}

	return filepath, nil
}

// unzip extracts a zip archive to a destination directory.
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", fpath, err)
		}

		out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to open output file %s: %w", fpath, err)
		}
		defer out.Close()

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip archive: %w", err)
		}
		defer rc.Close()

		_, err = io.Copy(out, rc)
		if err != nil {
			return fmt.Errorf("failed to copy content from zip to file: %w", err)
		}
	}
	return nil
}

// untarGz extracts a .tar.gz archive to a destination directory.
func untarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		fpath := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fpath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fpath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(fpath), os.FileMode(0755)); err != nil {
				return fmt.Errorf("failed to create directory for file %s: %w", fpath, err)
			}
			out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to open output file %s: %w", fpath, err)
			}
			defer out.Close()
			if _, err := io.Copy(out, tr); err != nil {
				return fmt.Errorf("failed to copy content from tar to file: %w", err)
			}
		default:
			log.Warn("Skipping unsupported tar entry type", "type", header.Typeflag, "name", header.Name)
		}
	}
	return nil
}
