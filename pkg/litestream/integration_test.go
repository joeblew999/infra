package litestream

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joeblew999/infra/pkg/config"
)

// TestLitestreamFilesystemIntegration tests Litestream replication with local SQLite and filesystem
func TestLitestreamFilesystemIntegration(t *testing.T) {
	t.Log("🧪 Starting LOCAL_FILESYSTEM_TEST for Litestream + SQLite integration")

	// Check if litestream is available via dep system
	t.Log("🔧 Checking litestream binary via dep system...")
	litestreamBinary, err := config.GetAbsoluteDepPath("litestream")
	if err != nil {
		t.Skip("Litestream binary not available")
	}

	// Create temporary test directories
	testDir, err := os.MkdirTemp("", "litestream-test-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Setup paths
	dbPath := filepath.Join(testDir, "pb_data", "test.db")
	backupPath := filepath.Join(testDir, "backups", "test.db")
	configPath := filepath.Join(testDir, "litestream.yml")

	t.Logf("📁 Test directory: %s", testDir)
	t.Logf("📊 Database: %s", dbPath)
	t.Logf("💾 Backup: %s", backupPath)

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		t.Fatalf("Failed to create db directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	// Create Litestream config for filesystem
	litestreamConfig := fmt.Sprintf(`
dbs:
  - path: %s
    replicas:
      - type: file
        path: %s
        sync-interval: 100ms
        retention: 1h
`, dbPath, backupPath)

	if err := os.WriteFile(configPath, []byte(litestreamConfig), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Create test SQLite database
	t.Log("🗄️  Creating test SQLite database...")
	if err := createTestDatabase(dbPath); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Start Litestream replication
	t.Log("🔄 Starting Litestream replication...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, litestreamBinary, "replicate", "-config", configPath)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start litestream: %v", err)
	}
	defer cmd.Process.Kill()

	// Give Litestream time to start
	time.Sleep(500 * time.Millisecond)

	// Test 1: Verify initial backup creation
	t.Log("✅ Verifying initial backup creation...")
	if err := waitForBackup(backupPath, 2*time.Second); err != nil {
		t.Fatalf("Backup not created: %v", err)
	}

	// Test 2: Write data and verify replication
	t.Log("📝 Writing test data...")
	if err := writeTestData(dbPath, "test-data-1"); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	// Wait for replication
	time.Sleep(200 * time.Millisecond)

	// Test 3: Verify data in backup
	t.Log("🔍 Verifying data in backup...")
	backupData, err := readTestData(backupPath)
	if err != nil {
		t.Fatalf("Failed to read from backup: %v", err)
	}
	if !strings.Contains(backupData, "test-data-1") {
		t.Fatalf("Data not found in backup: %q", backupData)
	}

	// Test 4: Simulate data loss and restore
	t.Log("🚨 Simulating data loss...")
	if err := os.Remove(dbPath); err != nil {
		t.Fatalf("Failed to remove database: %v", err)
	}

	t.Log("🔄 Restoring from backup...")
	restoreCmd := exec.Command(litestreamBinary, "restore", "-config", configPath)
	restoreCmd.Dir = filepath.Dir(dbPath) // Restore to same directory
	if err := restoreCmd.Run(); err != nil {
		t.Fatalf("Failed to restore: %v", err)
	}

	// Test 5: Verify restored data
	t.Log("✅ Verifying restored data...")
	restoredData, err := readTestData(dbPath)
	if err != nil {
		t.Fatalf("Failed to read restored data: %v", err)
	}
	if !strings.Contains(restoredData, "test-data-1") {
		t.Fatalf("Restored data missing: %q", restoredData)
	}

	t.Log("🎉 LOCAL_FILESYSTEM_TEST passed - Litestream + SQLite integration working!")
}

// TestLitestreamMultiWrite tests continuous replication with multiple writes
func TestLitestreamMultiWrite(t *testing.T) {
	t.Log("🧪 Testing continuous replication with multiple writes")

	testDir, err := os.MkdirTemp("", "litestream-multi-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	dbPath := filepath.Join(testDir, "pb_data", "multi.db")
	backupPath := filepath.Join(testDir, "backups", "multi.db")
	configPath := filepath.Join(testDir, "litestream.yml")

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	// Create config with faster sync for testing
	litestreamConfig := fmt.Sprintf(`
dbs:
  - path: %s
    replicas:
      - type: file
        path: %s
        sync-interval: 50ms
        retention: 30m
`, dbPath, backupPath)

	if err := os.WriteFile(configPath, []byte(litestreamConfig), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Check if litestream is available via dep system
	t.Log("🔧 Checking litestream binary via dep system...")
	litestreamBinary, err := config.GetAbsoluteDepPath("litestream")
	if err != nil {
		t.Skip("Litestream binary not available")
	}

	// Create initial database
	if err := createTestDatabase(dbPath); err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Start replication
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, litestreamBinary, "replicate", "-config", configPath)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start litestream: %v", err)
	}
	defer cmd.Process.Kill()

	time.Sleep(200 * time.Millisecond)

	// Write multiple records
	for i := 0; i < 5; i++ {
		record := fmt.Sprintf("record-%d", i)
		t.Logf("📝 Writing: %s", record)
		
		if err := writeTestData(dbPath, record); err != nil {
		t.Fatalf("Failed to write %s: %v", record, err)
		}
		
		time.Sleep(100 * time.Millisecond) // Allow replication
	}

	// Verify all records in backup
	backupData, err := readTestData(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}

	for i := 0; i < 5; i++ {
		expected := fmt.Sprintf("record-%d", i)
		if !strings.Contains(backupData, expected) {
			t.Errorf("Missing record in backup: %s", expected)
		}
	}

	if !strings.Contains(backupData, "record-4") {
		t.Errorf("Final record not found: %s", backupData)
	}

	t.Log("🎉 Multi-write test passed - continuous replication working!")
}

// createTestDatabase creates a simple test SQLite database
func createTestDatabase(path string) error {
	cmd := exec.Command("sqlite3", path, `
		PRAGMA journal_mode=WAL;
		CREATE TABLE IF NOT EXISTS test_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			data TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return cmd.Run()
}

// writeTestData inserts test data into the database
func writeTestData(path, data string) error {
	cmd := exec.Command("sqlite3", path, fmt.Sprintf(
		"INSERT INTO test_data (data) VALUES ('%s');", data))
	return cmd.Run()
}

// readTestData reads all test data from database
func readTestData(path string) (string, error) {
	cmd := exec.Command("sqlite3", path, "SELECT GROUP_CONCAT(data, ',') FROM test_data;")
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

// waitForBackup waits for backup file to exist
func waitForBackup(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("backup file %s not created within timeout", path)
}