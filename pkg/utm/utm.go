package utm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	defaultAppPath  = "/Applications/UTM.app"
	defaultDocsPath = "Library/Containers/com.utmapp.UTM/Data/Documents"
)

// Manager provides the tiny bit of functionality we need from UTM:
// find existing VMs and open them from the CLI.
type Manager struct {
	AppPath  string
	DocsPath string
}

// VMInfo describes a discovered .utm bundle.
type VMInfo struct {
	Name string
	Path string
}

// NewManager creates a manager for the current host. We only support macOS
// because UTM itself is macOS-only.
func NewManager() (*Manager, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("utm helpers are only available on macOS")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	docsPath := filepath.Join(home, defaultDocsPath)

	return &Manager{
		AppPath:  defaultAppPath,
		DocsPath: docsPath,
	}, nil
}

// LaunchApp opens UTM.app using the default macOS handler.
func (m *Manager) LaunchApp() error {
	if err := exec.Command("open", m.AppPath).Run(); err != nil {
		return fmt.Errorf("open UTM.app: %w", err)
	}
	return nil
}

// ListVMs returns every .utm bundle under the standard UTM documents folder.
func (m *Manager) ListVMs() ([]VMInfo, error) {
	entries, err := os.ReadDir(m.DocsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []VMInfo{}, nil
		}
		return nil, fmt.Errorf("read UTM documents directory: %w", err)
	}

	var vms []VMInfo
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".utm") {
			continue
		}
		vms = append(vms, VMInfo{
			Name: strings.TrimSuffix(entry.Name(), ".utm"),
			Path: filepath.Join(m.DocsPath, entry.Name()),
		})
	}
	return vms, nil
}

// OpenVM locates the requested VM bundle and uses `open` to launch it in UTM.
// The identifier can be a bare VM name (without the .utm suffix) or any path
// pointing to a .utm bundle.
func (m *Manager) OpenVM(identifier string) error {
	bundlePath, err := m.resolveVMPath(identifier)
	if err != nil {
		return err
	}
	if err := exec.Command("open", bundlePath).Run(); err != nil {
		return fmt.Errorf("open VM %s: %w", bundlePath, err)
	}
	return nil
}

func (m *Manager) resolveVMPath(identifier string) (string, error) {
	candidate := identifier

	if !strings.Contains(identifier, string(os.PathSeparator)) && !strings.HasSuffix(identifier, ".utm") {
		candidate = filepath.Join(m.DocsPath, identifier+".utm")
	}

	if !strings.HasSuffix(candidate, ".utm") {
		candidate = candidate + ".utm"
	}

	absPath, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve VM path: %w", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("vm bundle not found at %s", absPath)
		}
		return "", fmt.Errorf("stat VM bundle: %w", err)
	}

	return absPath, nil
}
