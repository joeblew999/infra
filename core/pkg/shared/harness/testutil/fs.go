package testutil

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a deterministic temporary directory for tests. When a prefix
// is provided a child directory is created beneath the testing temp root.
func TempDir(t *testing.T, prefix string) string {
	t.Helper()
	root := t.TempDir()
	if prefix == "" {
		return root
	}
	dir := filepath.Join(root, prefix)
	if err := os.Mkdir(dir, 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	return dir
}
