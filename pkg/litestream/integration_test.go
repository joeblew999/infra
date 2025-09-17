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
	"github.com/joeblew999/infra/pkg/dep"
)

// TestLitestreamFilesystemIntegration tests Litestream replication with local SQLite and filesystem
func TestLitestreamFilesystemIntegration(t *testing.T) {
	t.Skip("litestream restore path still being stabilized")
	t.Log("🧪 Starting LOCAL_FILESYSTEM_TEST for Litestream + SQLite integration")

	// Ensure litestream binary is available via dep system
	t.Log("🔧 Checking litestream binary via dep system...")
	if err := dep.InstallBinary("litestream", false); err != nil {
		t.Fatalf("Failed to install litestream: %v", err)
	}
	litestreamBinary, err := config.GetAbsoluteDepPath("litestream")
	if err != nil {
		t.Fatalf("Failed to resolve litestream binary path: %v", err)
	}

	// Create temporary test directories
	testDir := setupTestDir(t)

	// Setup paths
	dbPath := filepath.Join(testDir, "pb_data", "test.db")
	backupDir := filepath.Join(testDir, "backups")
	configPath := filepath.Join(testDir, "litestream.yml")

	t.Logf("📁 Test directory: %s", testDir)
	t.Logf("📊 Database: %s", dbPath)
	t.Logf("💾 Backup directory: %s", backupDir)

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		t.Fatalf("Failed to create db directory: %v", err)
	}
	if err := os.MkdirAll(backupDir, 0755); err != nil {
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
`, dbPath, backupDir)

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
	cmd := exec.CommandContext(ctx, litestreamBinary, "replicate", "-config", configPath)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start litestream: %v", err)
	}
	stopped := false
	stopReplication := func() {
		if stopped {
			return
		}
		cancel()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		stopped = true
	}
	defer stopReplication()

	// Give Litestream time to start
	time.Sleep(500 * time.Millisecond)

	// Write data and verify replication
	t.Log("📝 Writing test data...")
	if err := writeTestData(dbPath, "test-data-1"); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	if err := waitForBackup(backupDir, 5*time.Second); err != nil {
		t.Fatalf("Backup not created: %v", err)
	}

	// Stop replication before running restore commands
	stopReplication()

	// Verify data by restoring to a preview database
	previewPath := filepath.Join(testDir, "preview.db")
	previewCmd := exec.Command(litestreamBinary, "restore", "-config", configPath, "-o", previewPath, dbPath)
	previewCmd.Dir = filepath.Dir(dbPath)
	if output, err := previewCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to restore preview: %v\nOutput: %s", err, string(output))
	}

	backupData, err := readTestData(previewPath)
	if err != nil {
		t.Fatalf("Failed to read from restored preview: %v", err)
	}
	if !strings.Contains(backupData, "test-data-1") {
		t.Fatalf("Data not found in restored preview: %q", backupData)
	}
	_ = os.Remove(previewPath)

	// Test 4: Simulate data loss and restore
	t.Log("🚨 Simulating data loss...")
	if err := os.Remove(dbPath); err != nil {
		t.Fatalf("Failed to remove database: %v", err)
	}

	t.Log("🔄 Restoring from backup...")
	restoreCmd := exec.Command(litestreamBinary, "restore", "-config", configPath, dbPath)
	restoreCmd.Dir = filepath.Dir(dbPath) // Restore to same directory
	if output, err := restoreCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to restore: %v\nOutput: %s", err, string(output))
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
	t.Skip("litestream restore path still being stabilized")
	t.Log("🧪 Testing continuous replication with multiple writes")

	testDir := setupTestDir(t)

	dbPath := filepath.Join(testDir, "pb_data", "multi.db")
	backupDir := filepath.Join(testDir, "backups")
	configPath := filepath.Join(testDir, "litestream.yml")

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}
	if err := os.MkdirAll(backupDir, 0755); err != nil {
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
`, dbPath, backupDir)

	if err := os.WriteFile(configPath, []byte(litestreamConfig), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Ensure litestream binary is available via dep system
	t.Log("🔧 Checking litestream binary via dep system...")
	if err := dep.InstallBinary("litestream", false); err != nil {
		t.Fatalf("Failed to install litestream: %v", err)
	}
	litestreamBinary, err := config.GetAbsoluteDepPath("litestream")
	if err != nil {
		t.Fatalf("Failed to resolve litestream binary path: %v", err)
	}

	// Create initial database
	if err := createTestDatabase(dbPath); err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Start replication
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, litestreamBinary, "replicate", "-config", configPath)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start litestream: %v", err)
	}
	stopped := false
	stopReplication := func() {
		if stopped {
			return
		}
		cancel()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		stopped = true
	}
	defer stopReplication()

	time.Sleep(500 * time.Millisecond)

	// Write multiple records
	for i := 0; i < 5; i++ {
		record := fmt.Sprintf("record-%d", i)
		t.Logf("📝 Writing: %s", record)

		if err := writeTestData(dbPath, record); err != nil {
			t.Fatalf("Failed to write %s: %v", record, err)
		}

		time.Sleep(100 * time.Millisecond) // Allow replication
	}

	if err := waitForBackup(backupDir, 5*time.Second); err != nil {
		t.Fatalf("Backup not ready: %v", err)
	}

	stopReplication()

	previewPath := filepath.Join(testDir, "preview.db")
	previewCmd := exec.Command(litestreamBinary, "restore", "-config", configPath, "-o", previewPath, dbPath)
	previewCmd.Dir = filepath.Dir(dbPath)
	if output, err := previewCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to restore preview: %v\nOutput: %s", err, string(output))
	}

	backupData, err := readTestData(previewPath)
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
	_ = os.Remove(previewPath)

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
func waitForBackup(dir string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if hasBackupArtifacts(dir) {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("backup artifacts not created in %s within timeout", dir)
}

func hasBackupArtifacts(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			if strings.HasSuffix(entry.Name(), ".db") {
				return true
			}
			continue
		}

		if entry.Name() != "generations" {
			continue
		}

		gens, err := os.ReadDir(filepath.Join(dir, "generations"))
		if err != nil || len(gens) == 0 {
			continue
		}

		for _, gen := range gens {
			snapshotsDir := filepath.Join(dir, "generations", gen.Name(), "snapshots")
			snaps, err := os.ReadDir(snapshotsDir)
			if err == nil && len(snaps) > 0 {
				return true
			}
		}
	}

	return false
}

func setupTestDir(t *testing.T) string {
	sanitized := strings.NewReplacer("/", "_", "\\", "_").Replace(t.Name())
	baseDir := filepath.Join(config.GetTestDataPath(), "litestream", sanitized)
	_ = os.RemoveAll(baseDir)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		t.Fatalf("Failed to determine absolute test directory: %v", err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(absDir) })
	return absDir
}
