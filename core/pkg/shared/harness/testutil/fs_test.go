package testutil

import (
	"os"
	"testing"
)

func TestTempDir(t *testing.T) {
	dir := TempDir(t, "demo")
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected directory: %s", dir)
	}
}
