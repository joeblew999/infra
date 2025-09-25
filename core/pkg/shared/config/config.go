package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// Environment indicators.
	EnvVarEnvironment       = "CORE_ENVIRONMENT"
	EnvVarAppRoot           = "CORE_APP_ROOT"
	EnvVarBusClusterEnabled = "CORE_BUS_CLUSTER_ENABLED"

	EnvProduction  = "production"
	EnvDevelopment = "development"

	// Filesystem layout.
	DepDir      = ".dep"
	BinDir      = ".bin"
	DataDir     = ".data"
	TestDataDir = ".data-test"
	LogsDir     = ".logs"
)

// Environment returns the effective runtime environment string.
func Environment() string {
	value := strings.TrimSpace(os.Getenv(EnvVarEnvironment))
	if value == "" {
		return EnvDevelopment
	}
	return strings.ToLower(value)
}

// IsProduction reports if the orchestrator is running in production.
func IsProduction() bool {
	return Environment() == EnvProduction
}

// GetAppRoot returns the base directory for runtime assets.
func GetAppRoot() string {
	if override := strings.TrimSpace(os.Getenv(EnvVarAppRoot)); override != "" {
		return filepath.Clean(override)
	}
	if IsProduction() {
		if isContainerEnvironment() {
			return "/app"
		}
	}
	return "."
}

// GetDepPath returns the absolute path to the dependency directory.
func GetDepPath() string {
	return filepath.Join(GetAppRoot(), DepDir)
}

// GetBinPath returns the absolute path to the compiled binaries directory.
func GetBinPath() string {
	return filepath.Join(GetAppRoot(), BinDir)
}

// GetDataPath returns the absolute path to the persistent data directory.
func GetDataPath() string {
	if IsTestEnvironment() {
		return GetTestDataPath()
	}
	return filepath.Join(GetAppRoot(), DataDir)
}

// GetTestDataPath returns the absolute path to the test data directory.
func GetTestDataPath() string {
	return filepath.Join(GetAppRoot(), TestDataDir)
}

// GetLogsPath returns the absolute path to the log directory.
func GetLogsPath() string {
	return filepath.Join(GetAppRoot(), LogsDir)
}

// IsTestEnvironment detects go test execution.
func IsTestEnvironment() bool {
	for _, arg := range os.Args {
		if strings.HasSuffix(arg, ".test") {
			return true
		}
	}
	return false
}

// ShouldEnsureBusCluster reports whether the orchestrator should attempt to join
// or provision an external bus cluster in addition to the embedded instance.
func ShouldEnsureBusCluster() bool {
	value := strings.TrimSpace(os.Getenv(EnvVarBusClusterEnabled))
	if value == "" {
		return false
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func isContainerEnvironment() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/run/.containerenv"); err == nil {
			return true
		}
	}
	return false
}
