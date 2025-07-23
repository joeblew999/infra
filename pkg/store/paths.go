package store

import (
	"path/filepath"
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
