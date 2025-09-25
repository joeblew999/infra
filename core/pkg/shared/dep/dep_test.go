package dep

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	manifestJSON := `{"binaries":[{"name":"alpha","source":"placeholder"}]}`
	manifest, err := LoadManifest(strings.NewReader(manifestJSON))
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	if len(manifest.Binaries) != 1 || manifest.Binaries[0].Name != "alpha" {
		t.Fatalf("unexpected manifest contents: %#v", manifest.Binaries)
	}
}

func TestEnsureAll(t *testing.T) {
	t.Setenv("CORE_APP_ROOT", t.TempDir())
	manifest := &Manifest{Binaries: []BinarySpec{{Name: "alpha", Source: SourcePlaceholder}}}
	installer := DefaultInstaller{}
	paths, err := manifest.EnsureAll(installer)
	if err != nil {
		t.Fatalf("EnsureAll error: %v", err)
	}
	placeholderPath := ResolveBinaryPath("alpha")
	if got := paths["alpha"]; got != placeholderPath {
		t.Fatalf("expected path %q, got %q", placeholderPath, got)
	}
	data, err := os.ReadFile(placeholderPath)
	if err != nil {
		t.Fatalf("read stub: %v", err)
	}
	if !strings.Contains(string(data), "placeholder binary: alpha") {
		t.Fatalf("unexpected stub content: %q", data)
	}
}

func TestEnsureAllWithExternalPath(t *testing.T) {
	t.Setenv("CORE_APP_ROOT", t.TempDir())
	binFile := filepath.Join(t.TempDir(), "alpha")
	if err := os.WriteFile(binFile, []byte("#!/bin/sh\nexit 0"), 0o755); err != nil {
		t.Fatalf("write source: %v", err)
	}
	manifest := &Manifest{Binaries: []BinarySpec{{Name: "alpha", Source: SourcePlaceholder, Path: binFile}}}
	paths, err := manifest.EnsureAll(DefaultInstaller{})
	if err != nil {
		t.Fatalf("EnsureAll error: %v", err)
	}
	dest := ResolveBinaryPath("alpha")
	if paths["alpha"] != dest {
		t.Fatalf("expected dest %q, got %q", dest, paths["alpha"])
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(data) != "#!/bin/sh\nexit 0" {
		t.Fatalf("unexpected dest contents: %q", data)
	}
}
