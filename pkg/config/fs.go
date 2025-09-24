package config

import (
	"fmt"
	"os"
)

// EnsureAppDirectories creates the runtime directories that services expect.
// It mirrors the Fly.io /app layout locally so commands can run without root access.
func EnsureAppDirectories() error {
	root := GetAppRoot()
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("failed to create app root directory: %w", err)
	}

	if IsTestEnvironment() {
		if err := os.MkdirAll(GetTestDataPath(), 0o755); err != nil {
			return fmt.Errorf("failed to create test data directory: %w", err)
		}
		if err := os.MkdirAll(GetLogsPath(), 0o755); err != nil {
			return fmt.Errorf("failed to create logs directory: %w", err)
		}
		return nil
	}

	if err := os.MkdirAll(GetDataPath(), 0o755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if err := os.MkdirAll(GetLogsPath(), 0o755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	return nil
}
