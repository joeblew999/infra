package store

import (
	"fmt"
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

// GetTerraformPath returns the absolute path to the terraform directory.
func GetTerraformPath() string {
	return filepath.Join(".", TerraformDir)
}
