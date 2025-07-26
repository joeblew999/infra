package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// Environment variables
	// NOTE: ALL environment variables MUST be declared as constants here and used throughout the code.
	// This ensures centralized management and prevents hardcoded strings scattered across functions.
	EnvVarEnvironment   = "ENVIRONMENT"
	EnvVarFlyAppName    = "FLY_APP_NAME"
	EnvVarKoDockerRepo  = "KO_DOCKER_REPO"

	// Registry and image constants
	// NOTE: All registry URLs and image names are constants to prevent obfuscation
	FlyRegistryURL           = "registry.fly.io/"
	FlyRegistryFallback      = "registry.fly.io/infra"
	KoLocalRegistry          = "ko.local"
	ChainguardStaticImage    = "cgr.dev/chainguard/static:latest"
	ChainguardGoImage        = "cgr.dev/chainguard/go:latest"

	// Platform constants
	PlatformLinuxAmd64 = "linux/amd64"
	PlatformLinuxArm64 = "linux/arm64"

	// Configuration file names
	KoConfigFileName = ".ko.yaml"

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
	env := os.Getenv(EnvVarEnvironment)
	if env == "" {
		env = os.Getenv(EnvVarFlyAppName) // Fly.io sets this
	}
	return env == EnvProduction || os.Getenv(EnvVarFlyAppName) != ""
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
	return filepath.Join(".", KoConfigFileName)
}

// GetKoDefaultBaseImage returns the appropriate base image for the environment
func GetKoDefaultBaseImage() string {
	if IsProduction() {
		return ChainguardStaticImage // Minimal for production
	}
	return ChainguardGoImage // Debug-friendly for development
}

// GetKoDockerRepo returns the appropriate Docker repository for the environment
func GetKoDockerRepo() string {
	// Check if explicitly set via environment variable
	if repo := os.Getenv(EnvVarKoDockerRepo); repo != "" {
		return repo
	}
	
	if IsProduction() {
		// Use Fly.io registry in production (assuming FLY_APP_NAME is set)
		if appName := os.Getenv(EnvVarFlyAppName); appName != "" {
			return FlyRegistryURL + appName
		}
		return FlyRegistryFallback // fallback
	}
	
	// Local development - use local registry or ko.local
	return KoLocalRegistry
}

// GetKoDefaultPlatforms returns the platforms to build for
func GetKoDefaultPlatforms() []string {
	if IsProduction() {
		// Multi-platform for production
		return []string{PlatformLinuxAmd64, PlatformLinuxArm64}
	}
	// Single platform for development (faster builds)
	return []string{PlatformLinuxAmd64}
}
