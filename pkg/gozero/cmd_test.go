package gozero

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/joeblew999/infra/pkg/config"
)

func TestGoZeroRunner_runGoctl(t *testing.T) {
	// Create test runner
	runner := NewGoZeroRunner(false)
	
	// Create a temp directory for testing
	tempDir, err := os.MkdirTemp("", "gozero-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	runner.SetWorkDir(tempDir)
	
	// Test that goctl binary exists and can run --version
	// Find the absolute path by looking for the binary relative to the repo root
	goctlPath := config.Get(config.BinaryGoctl)
	if !filepath.IsAbs(goctlPath) {
		// Try to find the repo root by looking for go.work file
		wd, _ := os.Getwd()
		for dir := wd; dir != "/" && dir != "."; {
			if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
				goctlPath = filepath.Join(dir, goctlPath)
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	
	if _, err := os.Stat(goctlPath); os.IsNotExist(err) {
		t.Skip("goctl binary not found - run 'infra dep local install goctl' first")
	}
	
	// Test running goctl --version (should work even if binary is not installed)
	err = runner.runGoctl("test version command", "--version")
	if err != nil {
		t.Errorf("goctl --version failed: %v", err)
	}
}

func TestService_CreateMCPAPI(t *testing.T) {
	// Create test service
	service := NewService(true) // Enable debug mode to see errors
	
	// Create a temp directory for testing  
	tempDir, err := os.MkdirTemp("", "gozero-mcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Check if goctl is available
	goctlPath := config.Get(config.BinaryGoctl)
	if !filepath.IsAbs(goctlPath) {
		// Try to find the repo root by looking for go.work file
		wd, _ := os.Getwd()
		for dir := wd; dir != "/" && dir != "."; {
			if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
				goctlPath = filepath.Join(dir, goctlPath)
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	
	if _, err := os.Stat(goctlPath); os.IsNotExist(err) {
		t.Skip("goctl binary not found - run 'infra dep local install goctl' first")
	}
	
	// Test creating MCP API
	ctx := context.Background()
	err = service.CreateMCPAPI(ctx, "test-service", "Test MCP API Service", tempDir)
	if err != nil {
		t.Errorf("CreateMCPAPI failed: %v", err)
		return
	}
	
	// Verify that expected files were created
	expectedFiles := []string{
		"test-service.api",
		"go.mod",
		"internal/config/config.go",
		"internal/handler/mcphandler.go",
		"internal/handler/healthhandler.go",
	}
	
	for _, file := range expectedFiles {
		fullPath := filepath.Join(tempDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", file)
		}
	}
}

func TestService_GetProjectStructure(t *testing.T) {
	service := NewService(false)
	
	structure := service.GetProjectStructure("test-api")
	
	expectedFiles := []string{
		"test-api.api",
		"test-api.go", 
		"test-api.json",
		"go.mod",
		"go.sum",
		"etc/test-api_api.yaml",
		"internal/config/config.go",
		"internal/handler/",
		"internal/logic/",
		"internal/svc/servicecontext.go", 
		"internal/types/types.go",
	}
	
	if len(structure) != len(expectedFiles) {
		t.Errorf("Expected %d files, got %d", len(expectedFiles), len(structure))
	}
	
	for i, expected := range expectedFiles {
		if i < len(structure) && structure[i] != expected {
			t.Errorf("Expected file %d to be %s, got %s", i, expected, structure[i])
		}
	}
}