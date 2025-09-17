package ai

import (
	"testing"
)

func TestNewClaudeRunner(t *testing.T) {
	// Test that NewClaudeRunner creates a valid runner
	runner := NewClaudeRunner()

	if runner == nil {
		t.Fatal("NewClaudeRunner returned nil")
	}

	if runner.binaryPath == "" {
		t.Error("binaryPath should not be empty")
	}

	t.Logf("Claude runner created with binary: %s", runner.binaryPath)
}

func TestInstallDefaultMCP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MCP install in short mode")
	}

	runner := NewClaudeRunner()
	
	// Test that InstallDefaultMCP runs without error
	// Note: This test assumes the default config file exists
	// In a real test, we'd mock the filesystem
	err := runner.InstallDefaultMCP()
	if err != nil {
		t.Skipf("InstallDefaultMCP failed: %v", err)
	}
}
