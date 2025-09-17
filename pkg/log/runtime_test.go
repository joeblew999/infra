package log

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReconfigureMultiLogger(t *testing.T) {
	tempDir := t.TempDir()

	orig := GetLogger()
	t.Cleanup(func() {
		if orig != nil {
			SetLogger(orig)
		}
	})

	// Initial configuration writes to file one
	pathOne := filepath.Join(tempDir, "one.log")
	cfg := MultiConfig{Destinations: []DestinationConfig{{Type: "file", Level: "info", Format: "json", Path: pathOne}}}
	if err := ReconfigureMultiLogger(cfg); err != nil {
		t.Fatalf("initial reconfigure failed: %v", err)
	}

	Info("first message", "id", 1)

	dataOne, err := os.ReadFile(pathOne)
	if err != nil {
		t.Fatalf("failed to read first log file: %v", err)
	}
	if !strings.Contains(string(dataOne), "first message") {
		t.Fatalf("expected first log file to contain entry, got: %s", dataOne)
	}

	// Reconfigure to write to file two only
	pathTwo := filepath.Join(tempDir, "two.log")
	cfg = MultiConfig{Destinations: []DestinationConfig{{Type: "file", Level: "info", Format: "json", Path: pathTwo}}}
	if err := ReconfigureMultiLogger(cfg); err != nil {
		t.Fatalf("second reconfigure failed: %v", err)
	}

	Info("second message", "id", 2)

	dataTwo, err := os.ReadFile(pathTwo)
	if err != nil {
		t.Fatalf("failed to read second log file: %v", err)
	}
	if !strings.Contains(string(dataTwo), "second message") {
		t.Fatalf("expected second log file to contain entry, got: %s", dataTwo)
	}

	// Ensure the first file did not receive the second message
	dataOneAfter, err := os.ReadFile(pathOne)
	if err != nil {
		t.Fatalf("failed to reread first log file: %v", err)
	}
	if strings.Contains(string(dataOneAfter), "second message") {
		t.Fatalf("expected first log file to remain untouched after reconfigure")
	}
}
