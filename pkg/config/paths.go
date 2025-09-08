//go:generate go run generate_binaries.go

package config

import (
	"os"
	"path/filepath"
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
	LoggingConfigFileName = "infra.log.json"
	ChangelogFileName = "CHANGELOG.md"
	ClaudeConfigFileName = "claude.json"

	// DepDir is the designated location for all downloaded and managed external binary dependencies.
	DepDir = ".dep"

	// DepMCPDir is the designated location for all downloaded and managed MCP server binaries.
	DepMCPDir = ".dep-mcp"

	// BinDir is the location for the project's own compiled binaries.
	BinDir = ".bin"

	// TaskfilesDir is the directory containing Taskfiles for various project automation tasks.
	TaskfilesDir = "taskfiles"

	// DataDir is the root directory for all application data (e.g., databases, NATS stores).
	DataDir = ".data"

	// LogsDir is the directory for application log files.
	LogsDir = ".logs"

	// NATS stream constants
	NATSLogStreamName   = "LOGS"
	NATSLogStreamSubject = "logs.app"

	// DocsDir is the directory containing Markdown documentation files.
	DocsDir = "docs"

	// DocsHTTPPath is the HTTP path prefix for serving documentation.
	DocsHTTPPath = "/docs/"

	// BuildDir is the directory for storing built container images and artifacts.
	BuildDir = ".oci"

	// FontDir is the directory for cached fonts
	FontDir = "font"

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
	
	// Binary constants are now auto-generated in binaries_gen.go
	// Run `go generate` to regenerate from dep.json
)

// GetDepPath returns the absolute path to the .dep directory.
func GetDepPath() string {
	return filepath.Join(".", DepDir)
}

// GetMCPPath returns the absolute path to the .dep-mcp directory.
func GetMCPPath() string {
	return filepath.Join(".", DepMCPDir)
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
// In Fly.io production, this points to the mounted volume at /app/.data
func GetDataPath() string {
	if IsProduction() {
		// In Fly.io production, use the mounted volume at /app/.data
		return "/app/.data"
	}
	return filepath.Join(".", DataDir)
}

// GetLogsPath returns the absolute path to the .logs directory.
// In Fly.io production, this uses the data directory for logs
func GetLogsPath() string {
	if IsProduction() {
		// In Fly.io production, use the data directory for logs
		return filepath.Join(GetDataPath(), "logs")
	}
	return filepath.Join(".", LogsDir)
}

// Get returns the absolute path to a binary dependency.
func Get(name string) string {
	return filepath.Join(GetDepPath(), GetBinaryName(name))
}

// GetTofuBinPath returns the absolute path to the tofu binary.
func GetTofuBinPath() string {
	return Get(BinaryTofu)
}

// GetTaskBinPath returns the absolute path to the task binary.
func GetTaskBinPath() string {
	return Get(BinaryTask)
}

// GetCaddyBinPath returns the absolute path to the caddy binary.
func GetCaddyBinPath() string {
	return Get(BinaryCaddy)
}

// GetKoBinPath returns the absolute path to the ko binary.
func GetKoBinPath() string {
	return Get(BinaryKo)
}

// GetFlyctlBinPath returns the absolute path to the flyctl binary.
func GetFlyctlBinPath() string {
	return Get(BinaryFlyctl)
}

// GetClaudeBinPath returns the absolute path to the claude binary.
func GetClaudeBinPath() string {
	return Get(BinaryClaude)
}

// GetTerraformPath returns the absolute path to the terraform directory.
func GetTerraformPath() string {
	return filepath.Join(".", TerraformDir)
}

// GetBuildPath returns the absolute path to the build directory.
func GetBuildPath() string {
	return filepath.Join(".", BuildDir)
}

// IsProduction returns true if running in production environment
func IsProduction() bool {
	env := os.Getenv(EnvVarEnvironment)
	if env != "" {
		return env == EnvProduction
	}
	// Don't consider Fly.io unless ENVIRONMENT is explicitly set to production
	// This prevents false positives when FLY_APP_NAME is set for other reasons
	return false
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

// GetLoggingConfigFile returns the path to the logging configuration file
func GetLoggingConfigFile() string {
	return filepath.Join(".", LoggingConfigFileName)
}

// GetChangelogFile returns the path to the changelog file
func GetChangelogFile() string {
	return filepath.Join(DocsDir, ChangelogFileName)
}

// GetLoggingLevel returns the default logging level
func GetLoggingLevel() string {
	if IsProduction() {
		return "warn"
	}
	return "info"
}

// GetClaudeConfigPath returns the absolute path to the Claude configuration file
func GetClaudeConfigPath() string {
	return filepath.Join(".", ClaudeConfigFileName)
}

// GetLoggingFormat returns the default logging format
func GetLoggingFormat() string {
	if IsProduction() {
		return "json"
	}
	return "json" // Always JSON for structured logs
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

// GetPocketBaseDataPath returns the absolute path to the PocketBase data directory.
func GetPocketBaseDataPath() string {
	return filepath.Join(GetDataPath(), "pocketbase")
}

// GetPocketBasePort returns the default port for PocketBase server.
func GetPocketBasePort() string {
	return "8090"
}

// GetBentoPath returns the absolute path to the bento configuration directory.
func GetBentoPath() string {
	return filepath.Join(GetDataPath(), "bento")
}

// GetBentoPort returns the default port for bento service.
func GetBentoPort() string {
	return "4195"
}

// GetFontPath returns the absolute path to the font cache directory.
func GetFontPath() string {
	return filepath.Join(GetDataPath(), FontDir)
}

// GetFontPathForFamily returns the absolute path for a specific font family.
func GetFontPathForFamily(family string) string {
	return filepath.Join(GetFontPath(), family)
}

// GetCaddyPath returns the absolute path to the caddy configuration directory.
func GetCaddyPath() string {
	return filepath.Join(GetDataPath(), "caddy")
}

// GetCaddyPort returns the default port for caddy reverse proxy.
func GetCaddyPort() string {
	return "80"
}

// GetTransPath returns the absolute path to the translation cache directory.
func GetTransPath() string {
	return filepath.Join(GetDataPath(), "trans")
}

// GetWebServerPort returns the default port for the web server.
func GetWebServerPort() string {
	return "1337"
}

// GetNATSPort returns the default port for NATS server.
func GetNATSPort() string {
	return "4222"
}

// GetMCPPort returns the default port for MCP server.
func GetMCPPort() string {
	return "8080"
}

// GetDeckAPIPort returns the default port for Deck API server.
func GetDeckAPIPort() string {
	return "8888"
}

// GetMetricsPort returns the default port for metrics server.
func GetMetricsPort() string {
	return "9091"
}

// GetDockerImageName returns the default Docker image name for local builds
func GetDockerImageName() string {
	return "infra-local"
}

// GetDockerImageTag returns the default Docker image tag 
func GetDockerImageTag() string {
	return "latest"
}

// GetDockerImageFullName returns the full Docker image name with tag
func GetDockerImageFullName() string {
	return GetDockerImageName() + ":" + GetDockerImageTag()
}
