package cloudflare

import (
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
	// Set up temp data directory for secrets backend
	tempDir := t.TempDir()
	t.Setenv("CORE_DATA_PATH", tempDir)

	token := "cf-test-token"

	// SaveToken now uses secrets backend, path parameter is deprecated
	if err := SaveToken("", token); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	// LoadToken should retrieve the token from secrets backend
	loaded, err := LoadToken("")
	if err != nil {
		t.Fatalf("LoadToken: %v", err)
	}
	if loaded != token {
		t.Fatalf("loaded token mismatch, got %q want %q", loaded, token)
	}
}

func TestSaveEmptyToken(t *testing.T) {
	// Save empty/whitespace token should fail
	if err := SaveToken("", " \n"); err == nil {
		t.Fatal("expected error when saving empty token, got nil")
	}

	if err := SaveToken("", ""); err == nil {
		t.Fatal("expected error when saving empty token, got nil")
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	// Set up temp data directory for secrets backend
	tempDir := t.TempDir()
	t.Setenv("CORE_DATA_PATH", tempDir)

	settings := Settings{
		ZoneName:  "example.com",
		ZoneID:    "zone123",
		AccountID: "account456",
		R2Bucket:  "my-bucket",
		R2Region:  "us-east-1",
		AppDomain: "app.example.com",
	}

	// Save settings
	if err := SaveSettings(settings); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	// Load settings
	loaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}

	// Verify all fields match
	if loaded.ZoneName != settings.ZoneName {
		t.Errorf("ZoneName mismatch: got %q, want %q", loaded.ZoneName, settings.ZoneName)
	}
	if loaded.ZoneID != settings.ZoneID {
		t.Errorf("ZoneID mismatch: got %q, want %q", loaded.ZoneID, settings.ZoneID)
	}
	if loaded.AccountID != settings.AccountID {
		t.Errorf("AccountID mismatch: got %q, want %q", loaded.AccountID, settings.AccountID)
	}
	if loaded.R2Bucket != settings.R2Bucket {
		t.Errorf("R2Bucket mismatch: got %q, want %q", loaded.R2Bucket, settings.R2Bucket)
	}
	if loaded.R2Region != settings.R2Region {
		t.Errorf("R2Region mismatch: got %q, want %q", loaded.R2Region, settings.R2Region)
	}
	if loaded.AppDomain != settings.AppDomain {
		t.Errorf("AppDomain mismatch: got %q, want %q", loaded.AppDomain, settings.AppDomain)
	}
}
