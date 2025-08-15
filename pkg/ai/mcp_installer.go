package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
)

// MCPServer represents an MCP server configuration for installation
// This follows the same pattern as ConduitBinary but adapted for MCP servers
type MCPServer struct {
	Name       string              `json:"name"`
	Repo       string              `json:"repo"`
	Version    string              `json:"version"`
	ReleaseURL string              `json:"release_url"`
	Type       string              `json:"type"` // "go", "node", "binary"
	Assets     []dep.AssetSelector `json:"assets"`
	Command    string              `json:"command"`
	Args       []string            `json:"args"`
	Env        map[string]string   `json:"env"`
}

// MCPInstaller handles the installation of MCP servers
// This adapts the conduit pattern for MCP servers

type MCPInstaller struct{}

// Install installs an MCP server based on its type
func (i *MCPInstaller) Install(server MCPServer, debug bool) error {
	log.Info("Installing MCP server", "name", server.Name, "type", server.Type, "repo", server.Repo)

	// Convert MCPServer to DepBinary for compatibility with pkg/dep
	depBinary := dep.DepBinary{
		Name:       server.Name,
		Repo:       server.Repo,
		Version:    server.Version,
		ReleaseURL: server.ReleaseURL,
		Assets:     server.Assets,
	}

	// Use appropriate installer based on server type
	switch server.Type {
	case "go":
		installer := &goInstaller{}
		return installer.Install(depBinary, debug)
	case "node":
		installer := &nodeInstaller{}
		return installer.Install(depBinary, debug)
	case "binary":
		installer := &binaryInstaller{}
		return installer.Install(depBinary, debug)
	default:
		return fmt.Errorf("unsupported MCP server type: %s", server.Type)
	}
}

// goInstaller handles Go-based MCP server installation

type goInstaller struct{}

func (i *goInstaller) Install(binary dep.DepBinary, debug bool) error {
	log.Info("Installing Go-based MCP server", "name", binary.Name, "repo", binary.Repo)

	// For Go projects, we use go install directly
	// This is different from conduit as we don't need to download binaries
	// Instead we install from source using go install

	// Build go install command
	cmd := fmt.Sprintf("go install %s@%s", binary.Repo, binary.Version)
	if debug {
		log.Info("Running go install", "command", cmd)
	}

	// Execute go install
	return execGoInstall(binary.Repo, binary.Version, "")
}

// nodeInstaller handles Node.js-based MCP server installation

type nodeInstaller struct{}

func (i *nodeInstaller) Install(binary dep.DepBinary, debug bool) error {
	log.Info("Installing Node.js-based MCP server", "name", binary.Name, "repo", binary.Repo)

	// Use the standard dep installer for Node.js projects
	// This will download the appropriate binary/archive
	return dep.InstallBinary(binary.Name, debug)
}

// binaryInstaller handles pre-compiled binary MCP server installation

type binaryInstaller struct{}

func (i *binaryInstaller) Install(binary dep.DepBinary, debug bool) error {
	log.Info("Installing binary MCP server", "name", binary.Name, "repo", binary.Repo)

	// Use the standard dep installer
	return dep.InstallBinary(binary.Name, debug)
}

// execGoInstall executes go install command for Go-based MCP servers
func execGoInstall(repo, version, _ string) error {
	// Ensure GOPATH/bin exists
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = filepath.Join(os.Getenv("HOME"), "go")
	}
	binPath := filepath.Join(goPath, "bin")

	// Create bin directory if it doesn't exist
	if err := os.MkdirAll(binPath, 0755); err != nil {
		return fmt.Errorf("failed to create GOPATH/bin: %w", err)
	}

	// Build go install command
	var installTarget string
	if version == "latest" {
		installTarget = repo
	} else {
		installTarget = fmt.Sprintf("%s@%s", repo, version)
	}
	cmd := exec.Command("go", "install", installTarget)
	cmd.Env = append(os.Environ(), "GO111MODULE=on")

	log.Info("Running go install", "repo", repo, "version", version, "target", binPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go install failed: %w\nOutput: %s", err, string(output))
	}

	// The binary will be installed to GOPATH/bin
	binaryName := filepath.Base(repo)
	installedPath := filepath.Join(binPath, binaryName)
	if runtime.GOOS == "windows" {
		installedPath += ".exe"
	}

	// Verify binary was installed
	if _, err := os.Stat(installedPath); err != nil {
		return fmt.Errorf("binary not found after installation: %w", err)
	}

	// Create symlink in .dep-mcp directory for consistency
	mcpPath := config.GetMCPPath()
	if err := os.MkdirAll(mcpPath, 0755); err != nil {
		return fmt.Errorf("failed to create .dep-mcp directory: %w", err)
	}

	targetPath := filepath.Join(mcpPath, binaryName)
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	// Remove existing symlink/file
	os.Remove(targetPath)

	// Create symlink or copy the binary
	if err := os.Link(installedPath, targetPath); err != nil {
		// Fallback to copy if hard link fails
		if copyErr := copyFile(installedPath, targetPath); copyErr != nil {
			return fmt.Errorf("failed to copy binary: %w", copyErr)
		}
	}

	log.Info("Go-based MCP server installed", "name", binaryName, "path", targetPath)
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = sourceFile.WriteTo(destFile)
	return err
}

// MCPServerManager manages MCP server installations

type MCPServerManager struct {
	servers map[string]MCPServer
}

// NewMCPServerManager creates a new MCP server manager
func NewMCPServerManager() *MCPServerManager {
	manager := &MCPServerManager{
		servers: make(map[string]MCPServer),
	}

	// Load configuration from mcp-servers.json
	if err := manager.loadConfig(); err != nil {
		log.Warn("Failed to load MCP servers config, using defaults", "error", err)
		// Use default servers as fallback
		for _, server := range GetDefaultServers() {
			manager.servers[server.Name] = server
		}
	}

	return manager
}

// loadConfig loads MCP server configuration from mcp-servers.json
func (m *MCPServerManager) loadConfig() error {
	configPath := filepath.Join("pkg", "mcp", "mcp-servers.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read mcp-servers.json: %w", err)
	}

	var servers []MCPServer
	if err := json.Unmarshal(data, &servers); err != nil {
		return fmt.Errorf("failed to parse mcp-servers.json: %w", err)
	}

	for _, server := range servers {
		m.servers[server.Name] = server
	}

	return nil
}

// AddServer adds a new MCP server configuration
func (m *MCPServerManager) AddServer(server MCPServer) {
	m.servers[server.Name] = server
}

// InstallServer installs a specific MCP server
func (m *MCPServerManager) InstallServer(name string, debug bool) error {
	server, exists := m.servers[name]
	if !exists {
		return fmt.Errorf("MCP server not found: %s", name)
	}

	installer := &MCPInstaller{}
	return installer.Install(server, debug)
}

// InstallAll installs all configured MCP servers
func (m *MCPServerManager) InstallAll(debug bool) error {
	var errs []error
	for name := range m.servers {
		if err := m.InstallServer(name, debug); err != nil {
			errs = append(errs, fmt.Errorf("failed to install %s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("install errors: %v", errs)
	}
	return nil
}

// GetDefaultServers returns the default MCP servers configuration
func GetDefaultServers() []MCPServer {
	return []MCPServer{
		{
			Name:       "github-mcp",
			Repo:       "modelcontextprotocol/server-github",
			Version:    "2025.1.24",
			ReleaseURL: "https://github.com/modelcontextprotocol/server-github/releases",
			Type:       "node",
			Command:    "node",
			Args:       []string{"dist/index.js"},
			Env:        map[string]string{"GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}"},
			Assets: []dep.AssetSelector{
				{OS: "darwin", Arch: "arm64", Match: `mcp-server-github.*darwin-arm64\.tar\.gz$`},
				{OS: "darwin", Arch: "amd64", Match: `mcp-server-github.*darwin-amd64\.tar\.gz$`},
				{OS: "linux", Arch: "amd64", Match: `mcp-server-github.*linux-x64\.tar\.gz$`},
				{OS: "linux", Arch: "arm64", Match: `mcp-server-github.*linux-arm64\.tar\.gz$`},
				{OS: "windows", Arch: "amd64", Match: `mcp-server-github.*win32-x64\.zip$`},
			},
		},
		{
			Name:       "filesystem-mcp",
			Repo:       "modelcontextprotocol/server-filesystem",
			Version:    "2025.1.24",
			ReleaseURL: "https://github.com/modelcontextprotocol/server-filesystem/releases",
			Type:       "node",
			Command:    "node",
			Args:       []string{"dist/index.js"},
			Env:        map[string]string{"ALLOWED_PATHS": "${HOME}"},
			Assets: []dep.AssetSelector{
				{OS: "darwin", Arch: "arm64", Match: `mcp-server-filesystem.*darwin-arm64\.tar\.gz$`},
				{OS: "darwin", Arch: "amd64", Match: `mcp-server-filesystem.*darwin-amd64\.tar\.gz$`},
				{OS: "linux", Arch: "amd64", Match: `mcp-server-filesystem.*linux-x64\.tar\.gz$`},
				{OS: "linux", Arch: "arm64", Match: `mcp-server-filesystem.*linux-arm64\.tar\.gz$`},
				{OS: "windows", Arch: "amd64", Match: `mcp-server-filesystem.*win32-x64\.zip$`},
			},
		},
		{
			Name:       "fetch-mcp",
			Repo:       "modelcontextprotocol/server-fetch",
			Version:    "2025.1.24",
			ReleaseURL: "https://github.com/modelcontextprotocol/server-fetch/releases",
			Type:       "node",
			Command:    "node",
			Args:       []string{"dist/index.js"},
			Assets: []dep.AssetSelector{
				{OS: "darwin", Arch: "arm64", Match: `mcp-server-fetch.*darwin-arm64\.tar\.gz$`},
				{OS: "darwin", Arch: "amd64", Match: `mcp-server-fetch.*darwin-amd64\.tar\.gz$`},
				{OS: "linux", Arch: "amd64", Match: `mcp-server-fetch.*linux-x64\.tar\.gz$`},
				{OS: "linux", Arch: "arm64", Match: `mcp-server-fetch.*linux-arm64\.tar\.gz$`},
				{OS: "windows", Arch: "amd64", Match: `mcp-server-fetch.*win32-x64\.zip$`},
			},
		},
	}
}

// ListServers returns a list of all configured MCP server names
func (m *MCPServerManager) ListServers() []string {
	var names []string
	for name := range m.servers {
		names = append(names, name)
	}
	return names
}

// GetServer returns the configuration for a specific MCP server
func (m *MCPServerManager) GetServer(name string) (*MCPServer, error) {
	server, exists := m.servers[name]
	if !exists {
		return nil, fmt.Errorf("MCP server not found: %s", name)
	}
	return &server, nil
}
