package dep

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
)

// Source declares how a binary should be obtained.
type Source string

const (
	// SourcePlaceholder represents a binary that is expected to already exist in
	// the dependency directory. The installer simply verifies the file is present.
	SourcePlaceholder Source = "placeholder"

	// SourceGithubRelease represents a binary fetched from a GitHub release. The
	// concrete download logic will be implemented later.
	SourceGithubRelease Source = "github-release"

	// SourceGoBuild represents a binary built from a Go module.
	SourceGoBuild Source = "go-build"
)

// BinarySpec describes a binary dependency.
type BinarySpec struct {
	Name    string   `json:"name"`
	Version string   `json:"version,omitempty"`
	Source  Source   `json:"source"`
	Asset   *Asset   `json:"asset,omitempty"`
	Path    string   `json:"path,omitempty"`
	Args    []string `json:"args,omitempty"`
}

// Asset captures the minimal information required to locate a release asset.
type Asset struct {
	Match string `json:"match,omitempty"`
	OS    string `json:"os,omitempty"`
	Arch  string `json:"arch,omitempty"`
	URL   string `json:"url,omitempty"`
}

// Manifest is the JSON representation consumed by services.
type Manifest struct {
	Binaries []BinarySpec `json:"binaries"`
}

// Installer installs a binary and returns the absolute path to the executable.
type Installer interface {
	Ensure(spec BinarySpec) (string, error)
}

// EnsureAll iterates through the manifest and installs each binary using the
// provided installer. Returns a map of binary name to absolute path.
func (m *Manifest) EnsureAll(installer Installer) (map[string]string, error) {
	if installer == nil {
		return nil, errors.New("installer is required")
	}
	results := make(map[string]string, len(m.Binaries))
	for _, spec := range m.Binaries {
		path, err := installer.Ensure(spec)
		if err != nil {
			return nil, fmt.Errorf("ensure binary %q: %w", spec.Name, err)
		}
		results[spec.Name] = path
	}
	return results, nil
}

// LoadManifest decodes a manifest from the given reader.
func LoadManifest(r io.Reader) (*Manifest, error) {
	decoder := json.NewDecoder(r)
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	return &manifest, nil
}

// LoadManifestFile reads a manifest from disk.
func LoadManifestFile(path string) (*Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open manifest %s: %w", path, err)
	}
	defer file.Close()
	return LoadManifest(file)
}

// DefaultInstaller is a simple installer that prepares the dependency directory
// and returns the expected binary path. Real download/build logic will replace
// this as the orchestrator matures.
type DefaultInstaller struct{}

// Ensure implements Installer.
func (DefaultInstaller) Ensure(spec BinarySpec) (string, error) {
	if spec.Name == "" {
		return "", errors.New("binary name is required")
	}
	depDir := sharedcfg.GetDepPath()
	if err := os.MkdirAll(depDir, 0o755); err != nil {
		return "", fmt.Errorf("ensure dep dir: %w", err)
	}
	binaryName := spec.Name
	if runtime.GOOS == "windows" && filepath.Ext(binaryName) != ".exe" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(depDir, binaryName)
	if spec.Source == SourcePlaceholder {
		return ensurePlaceholder(spec, binaryPath)
	}
	if spec.Source == SourceGoBuild {
		return buildGoBinary(spec, binaryPath)
	}
	// For now, stub installers simply return the expected path. Future work will
	// add GitHub release downloads, go-build support, etc.
	return binaryPath, nil
}

func buildGoBinary(spec BinarySpec, dest string) (string, error) {
	if strings.TrimSpace(spec.Path) == "" {
		return "", errors.New("go-build source requires path to package")
	}
	// Remove any existing binary so go build overwrites cleanly.
	_ = os.Remove(dest)
	cmd := exec.Command("go", "build", "-o", dest, spec.Path)
	cmd.Dir = sharedcfg.GetAppRoot()
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("go build %s: %w\n%s", spec.Path, err, string(output))
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dest, 0o755); err != nil {
			return "", fmt.Errorf("chmod built binary: %w", err)
		}
	}
	return dest, nil
}

// ResolveBinaryPath combines the dependency directory with a binary name.
func ResolveBinaryPath(name string) string {
	return filepath.Join(sharedcfg.GetDepPath(), name)
}

func ensurePlaceholder(spec BinarySpec, dest string) (string, error) {
	if spec.Path != "" {
		data, err := os.ReadFile(spec.Path)
		if err != nil {
			return "", fmt.Errorf("read placeholder source %s: %w", spec.Path, err)
		}
		if err := os.WriteFile(dest, data, 0o755); err != nil {
			return "", fmt.Errorf("write placeholder binary %s: %w", dest, err)
		}
		return dest, nil
	}
	content := []byte("#!/usr/bin/env bash\n\necho 'placeholder binary: " + spec.Name + "'\n")
	if err := os.WriteFile(dest, content, 0o755); err != nil {
		return "", fmt.Errorf("write placeholder stub %s: %w", dest, err)
	}
	return dest, nil
}
