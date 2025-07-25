package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// DepDir is the designated location for all downloaded and managed external binary dependencies.
	DepDir = ".dep"

	// BinDir is the location for the project's own compiled binaries.
	BinDir = ".bin"

	// TaskfilesDir is the directory containing Taskfiles for various project automation tasks.
	TaskfilesDir = "taskfiles"

	// DataDir is the root directory for all application data (e.g., databases, NATS stores).
	DataDir = ".data"

	// DocsDir is the directory containing Markdown documentation files.
	DocsDir = "docs"

	// DocsHTTPPath is the HTTP path prefix for serving documentation.
	DocsHTTPPath = "/docs/"

	// TerraformDir is the directory containing Terraform/OpenTofu configuration files.
	TerraformDir = "terraform"

	// BinaryDepNameFormat is the format string for naming downloaded binary dependencies.
	// It uses placeholders for name, OS, and architecture.
	BinaryDepNameFormat = "%s_%s_%s"

	// Navigation paths for web interface
	HomeHTTPPath    = "/"
	MetricsHTTPPath = "/metrics"
	LogsHTTPPath    = "/logs"
	StatusHTTPPath  = "/status"

	// Environment detection
	EnvProduction = "production"
	EnvDevelopment = "development"
)

// GetDepPath returns the absolute path to the .dep directory.
func GetDepPath() string {
	return filepath.Join(".", DepDir)
}

// GetBinPath returns the absolute path to the .bin directory.
func GetBinPath() string {
	return filepath.Join(".", BinDir)
}

// GetTaskfilesPath returns the absolute path to the taskfiles directory.
func GetTaskfilesPath() string {
	return filepath.Join(".", TaskfilesDir)
}

// GetDataPath returns the absolute path to the .data directory.
func GetDataPath() string {
	return filepath.Join(".", DataDir)
}

func Get(name string) string {
	return filepath.Join(GetDepPath(), fmt.Sprintf(BinaryDepNameFormat, name, runtime.GOOS, runtime.GOARCH))
}

// GetTofuBinPath returns the absolute path to the tofu binary.
func GetTofuBinPath() string {
	return Get("tofu")
}

// GetTaskBinPath returns the absolute path to the task binary.
func GetTaskBinPath() string {
	return Get("task")
}

// GetCaddyBinPath returns the absolute path to the caddy binary.
func GetCaddyBinPath() string {
	return Get("caddy")
}

// GetKoBinPath returns the absolute path to the ko binary.
func GetKoBinPath() string {
	return Get("ko")
}

// GetFlyctlBinPath returns the absolute path to the flyctl binary.
func GetFlyctlBinPath() string {
	return Get("flyctl")
}

// GetTerraformPath returns the absolute path to the terraform directory.
func GetTerraformPath() string {
	return filepath.Join(".", TerraformDir)
}

// IsProduction returns true if running in production environment
func IsProduction() bool {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("FLY_APP_NAME") // Fly.io sets this
	}
	return env == EnvProduction || os.Getenv("FLY_APP_NAME") != ""
}

// IsDevelopment returns true if running in development environment
func IsDevelopment() bool {
	return !IsProduction()
}

// ShouldUseHTTPS returns true if HTTPS should be enabled
// Local dev: use HTTPS, Production (Fly.io): no HTTPS (Cloudflare terminates SSL)
func ShouldUseHTTPS() bool {
	return IsDevelopment()
}

// GetKoConfigPath returns the path to the ko configuration file
func GetKoConfigPath() string {
	return filepath.Join(".", ".ko.yaml")
}

// GetKoDefaultBaseImage returns the appropriate base image for the environment
func GetKoDefaultBaseImage() string {
	if IsProduction() {
		return "cgr.dev/chainguard/static:latest" // Minimal for production
	}
	return "cgr.dev/chainguard/go:latest" // Debug-friendly for development
}

// GetKoDockerRepo returns the appropriate Docker repository for the environment
func GetKoDockerRepo() string {
	// Check if explicitly set via environment variable
	if repo := os.Getenv("KO_DOCKER_REPO"); repo != "" {
		return repo
	}
	
	if IsProduction() {
		// Use Fly.io registry in production (assuming FLY_APP_NAME is set)
		if appName := os.Getenv("FLY_APP_NAME"); appName != "" {
			return "registry.fly.io/" + appName
		}
		return "registry.fly.io/infra" // fallback
	}
	
	// Local development - use local registry or ko.local
	return "ko.local"
}

// GetKoDefaultPlatforms returns the platforms to build for
func GetKoDefaultPlatforms() []string {
	if IsProduction() {
		// Multi-platform for production
		return []string{"linux/amd64", "linux/arm64"}
	}
	// Single platform for development (faster builds)
	return []string{"linux/amd64"}
}
