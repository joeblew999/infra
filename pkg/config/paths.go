//go:generate go run generate_binaries.go

package config

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	// Environment variables
	// NOTE: ALL environment variables MUST be declared as constants here and used throughout the code.
	// This ensures centralized management and prevents hardcoded strings scattered across functions.
	EnvVarEnvironment  = "ENVIRONMENT"
	EnvVarFlyAppName   = "FLY_APP_NAME"
	EnvVarKoDockerRepo = "KO_DOCKER_REPO"
	EnvVarAppRoot      = "APP_ROOT"
	EnvVarNATSCluster  = "NATS_CLUSTER_ENABLED"

	// Registry and image constants
	// NOTE: All registry URLs and image names are constants to prevent obfuscation
	FlyRegistryURL        = "registry.fly.io/"
	FlyRegistryFallback   = "registry.fly.io/infra"
	KoLocalRegistry       = "ko.local"
	ChainguardStaticImage = "cgr.dev/chainguard/static:latest"
	ChainguardGoImage     = "cgr.dev/chainguard/go:latest"

	// Platform constants
	PlatformLinuxAmd64 = "linux/amd64"
	PlatformLinuxArm64 = "linux/arm64"

	// Configuration file names
	KoConfigFileName      = ".ko.yaml"
	LoggingConfigFileName = "infra.log.json"
	ChangelogFileName     = "CHANGELOG.md"
	ClaudeConfigFileName  = "claude.json"

	// DepDir is the designated location for all downloaded and managed external binary dependencies.
	DepDir = ".dep"

	// AppRootDir is the development root directory that mirrors the container's /app layout.
	AppRootDir = "app"

	// DepMCPDir is the designated location for all downloaded and managed MCP server binaries.
	DepMCPDir = ".dep-mcp"

	// BinDir is the location for the project's own compiled binaries.
	BinDir = ".bin"

	// TaskfilesDir is the directory containing Taskfiles for various project automation tasks.
	TaskfilesDir = "taskfiles"

	// DataDir is the root directory for all application data (e.g., databases, NATS stores).
	DataDir = ".data"

	// TestDataDir is the root directory for test data (isolated from production data).
	TestDataDir = ".data-test"

	// LogsDir is the directory for application log files.
	LogsDir = ".logs"

	// DocsDir is the directory containing Markdown documentation files.
	DocsDir = "docs"

	// DocsHTTPPath is the HTTP path prefix for serving documentation.
	DocsHTTPPath = "/docs/"

	// BuildDir is the directory for storing built container images and artifacts.
	BuildDir = ".oci"

	// FontDir is the directory for cached fonts
	FontDir = "font"

	// DeckDir is the directory for deck build artifacts
	DeckDir = "deck"

	// MjmlDir is the directory for MJML templates and cache
	MjmlDir = "mjml"

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
	ConfigHTTPPath  = "/config"
	RuntimeHTTPPath = "/runtime"

	// Environment detection
	EnvProduction  = "production"
	EnvDevelopment = "development"

	// Binary constants are now auto-generated in binaries_gen.go
	// Run `go generate` to regenerate from dep.json
)

// GetAppRoot returns the base directory for runtime assets.
// In production containers this resolves to /app, while locally it maps to ./app
// (unless overridden via APP_ROOT). This keeps all runtime artifacts scoped to a
// single folder that mirrors the Fly.io layout.
func GetAppRoot() string {
	if override := strings.TrimSpace(os.Getenv(EnvVarAppRoot)); override != "" {
		return filepath.Clean(override)
	}
	if IsProduction() {
		if isContainerEnvironment() {
			return "/app"
		}
		return filepath.Join(".", AppRootDir)
	}
	return filepath.Join(".", AppRootDir)
}

func isContainerEnvironment() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

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
// In Fly.io production, this points to the mounted volume at /app/.data.
func GetDataPath() string {
	if IsTestEnvironment() {
		return GetTestDataPath()
	}
	return filepath.Join(GetAppRoot(), DataDir)
}

// GetTestDataPath returns the absolute path to the .data-test directory.
// This is used for test isolation, keeping test artifacts separate from production data.
func GetTestDataPath() string {
	return filepath.Join(".", TestDataDir)
}

// IsTestEnvironment returns true if running in a test environment.
// This detects go test execution by checking for the testing package.
func IsTestEnvironment() bool {
	// Check if we're running under go test by looking for .test suffix in args
	for _, arg := range os.Args {
		if strings.HasSuffix(arg, ".test") {
			return true
		}
	}
	return false
}

// GetLogsPath returns the absolute path to the .logs directory.
// In Fly.io production, this uses the data directory for logs
func GetLogsPath() string {
	if IsProduction() {
		// In Fly.io production, use the data directory for logs
		return filepath.Join(GetDataPath(), "logs")
	}
	return filepath.Join(GetAppRoot(), LogsDir)
}

// ShouldEnsureNATSCluster controls whether the orchestrator should attempt to boot
// the full NATS cluster alongside the embedded leaf node. Defaults to true in
// production and false in development unless explicitly overridden via
// NATS_CLUSTER_ENABLED.
func ShouldEnsureNATSCluster() bool {
	if value := strings.TrimSpace(os.Getenv(EnvVarNATSCluster)); value != "" {
		switch strings.ToLower(value) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	}
	return false
}

// Get returns the relative path to a binary dependency.
func Get(name string) string {
	return filepath.Join(GetDepPath(), GetBinaryName(name))
}

// GetAbsoluteDepPath returns the absolute path to a binary dependency.
func GetAbsoluteDepPath(name string) (string, error) {
	relPath := Get(name)
	return filepath.Abs(relPath)
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

// GetNscBinPath returns the absolute path to the nsc binary.
func GetNscBinPath() string {
	return Get(BinaryNsc)
}

// GetClaudeBinPath returns the absolute path to the claude binary.
func GetClaudeBinPath() string {
	return Get(BinaryClaude)
}

// GetGhBinPath returns the absolute path to the gh binary.
func GetGhBinPath() string {
	return Get(BinaryGh)
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

// GetFontPath returns the absolute path to the font cache directory.
// In test environments, uses .data-test/font for isolation.
func GetFontPath() string {
	if IsTestEnvironment() {
		return filepath.Join(GetTestDataPath(), FontDir)
	}
	return filepath.Join(GetDataPath(), FontDir)
}

// GetFontPathForFamily returns the absolute path for a specific font family.
func GetFontPathForFamily(family string) string {
	return filepath.Join(GetFontPath(), family)
}

// GetTransPath returns the absolute path to the translation cache directory.
func GetTransPath() string {
	return filepath.Join(GetDataPath(), "trans")
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

// GetXTemplateBinPath returns the absolute path to the xtemplate binary
func GetXTemplateBinPath() string {
	return Get(BinaryXtemplate)
}

// GetMjmlTemplatePath returns the absolute path to MJML templates directory.
func GetMjmlTemplatePath() string {
	return filepath.Join(GetMjmlPath(), "templates")
}

// GetMjmlCachePath returns the absolute path to MJML cache directory.
func GetMjmlCachePath() string {
	return filepath.Join(GetMjmlPath(), "cache")
}

// GetAuthPath returns the absolute path to the auth data directory.
// In test environments, uses .data-test/auth for isolation.
func GetAuthPath() string {
	if IsTestEnvironment() {
		return filepath.Join(GetTestDataPath(), "auth")
	}
	return filepath.Join(GetDataPath(), "auth")
}

// GetMoxDataPath returns the absolute path to the mox data directory.
// In test environments, uses .data-test/mox for isolation.
func GetMoxDataPath() string {
	if IsTestEnvironment() {
		return filepath.Join(GetTestDataPath(), "mox")
	}
	return filepath.Join(GetDataPath(), "mox")
}

// GetMoxConfigPath returns the absolute path to the mox configuration file.
func GetMoxConfigPath() string {
	return filepath.Join(GetMoxDataPath(), "moxmail.conf")
}

// GetMoxBinPath returns the absolute path to the mox binary.
func GetMoxBinPath() string {
	return Get(BinaryMox)
}

// GetAPIPath returns the absolute path to the API services directory.
func GetAPIPath() string {
	return filepath.Join(".", "api")
}

// GetAPIServices returns a list of API service directories that should be code generated.
// This can be overridden by environment variables or configuration files.
func GetAPIServices() []string {
	// Allow override via environment variable (comma-separated)
	if envServices := os.Getenv("API_SERVICES"); envServices != "" {
		services := strings.Split(envServices, ",")
		var trimmed []string
		for _, service := range services {
			if s := strings.TrimSpace(service); s != "" {
				trimmed = append(trimmed, s)
			}
		}
		return trimmed
	}

	// Default services - these can be discovered or configured
	return []string{
		"api/deck",
		"api/fast",
		"api/testservice",
	}
}
