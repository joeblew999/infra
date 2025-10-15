package cloudflare

import (
	"os"
	"path/filepath"
	"testing"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
)

func TestDefaultTokenPathFallsBackToProfile(t *testing.T) {
	t.Setenv(sharedcfg.EnvVarToolingProfile, "")
	cfg := sharedcfg.Tooling()
	expected := filepath.Join(sharedcfg.GetDataPath(), "core", "secrets", cloudflareSubdir, defaultFileName)
	if path := DefaultTokenPath(); path != expected {
		t.Fatalf("DefaultTokenPath() = %q, want %q", path, expected)
	}

	override := filepath.Join(t.TempDir(), "override")
	t.Setenv(sharedcfg.EnvVarCloudflareTokenPath, override)
	cfg = sharedcfg.Tooling()
	if cfg.Active.CloudflareTokenPath != override {
		t.Fatalf("expected profile to surface override path %q, got %q", override, cfg.Active.CloudflareTokenPath)
	}
	if path := DefaultTokenPath(); path != override {
		t.Fatalf("DefaultTokenPath() = %q, want %q", path, override)
	}
}

func TestSaveAndLoadToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token")
	token := "cf-test-token"
	if err := SaveToken(path, token); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved token: %v", err)
	}
	if string(data) != token {
		t.Fatalf("saved token mismatch, got %q", string(data))
	}

	loaded, err := LoadToken(path)
	if err != nil {
		t.Fatalf("LoadToken: %v", err)
	}
	if loaded != token {
		t.Fatalf("loaded token mismatch, got %q", loaded)
	}
}

func TestLoadTokenEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token")
	if err := os.WriteFile(path, []byte(" \n"), 0o600); err != nil {
		t.Fatalf("write token: %v", err)
	}
	if _, err := LoadToken(path); err == nil {
		t.Fatal("expected error when token file is empty")
	}
}
