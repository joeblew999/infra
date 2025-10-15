package config

import (
	"path/filepath"
	"testing"
)

func TestToolingDefaults(t *testing.T) {
	t.Setenv(EnvVarToolingProfile, "")
	t.Setenv(EnvVarToolingKORepository, "")
	t.Setenv(EnvVarToolingSupportsDocker, "")

	cfg := Tooling()

	if cfg.Active.Name != defaultToolingProfileName {
		t.Fatalf("expected active profile %q, got %q", defaultToolingProfileName, cfg.Active.Name)
	}
	if cfg.Active.KORepository != "ko.local/core" {
		t.Fatalf("expected default KO repository ko.local/core, got %q", cfg.Active.KORepository)
	}
	if cfg.Active.Mode != ToolingModeLocal {
		t.Fatalf("expected local mode, got %q", cfg.Active.Mode)
	}
	if cfg.Active.KoConfig != ".ko.yaml" {
		t.Fatalf("expected default ko config .ko.yaml, got %q", cfg.Active.KoConfig)
	}
	if cfg.Active.KoTemplate != "config/templates/ko.yaml.tmpl" {
		t.Fatalf("expected default ko template path config/templates/ko.yaml.tmpl, got %q", cfg.Active.KoTemplate)
	}
	if cfg.Active.FlyTemplate != "config/templates/fly.toml.tmpl" {
		t.Fatalf("expected default fly template path config/templates/fly.toml.tmpl, got %q", cfg.Active.FlyTemplate)
	}
	if cfg.Active.FlyRegion != "syd" {
		t.Fatalf("expected default fly region syd, got %q", cfg.Active.FlyRegion)
	}
	if cfg.Active.ImportPath != "./cmd/core" {
		t.Fatalf("expected default import path ./cmd/core, got %q", cfg.Active.ImportPath)
	}
	if len(cfg.Profiles) == 0 {
		t.Fatal("expected at least one profile in configuration")
	}
	if !cfg.Active.SupportsDocker {
		t.Fatal("expected local profile to allow docker by default")
	}
	expectedToken := filepath.Join(GetDataPath(), "core", "secrets", "fly", defaultTokenFileName)
	if cfg.Active.TokenPath != expectedToken {
		t.Fatalf("expected token path %q, got %q", expectedToken, cfg.Active.TokenPath)
	}
	cloudflareToken := filepath.Join(GetDataPath(), "core", "secrets", "cloudflare", defaultCloudflareTokenFileName)
	if cfg.Active.CloudflareTokenPath != cloudflareToken {
		t.Fatalf("expected cloudflare token path %q, got %q", cloudflareToken, cfg.Active.CloudflareTokenPath)
	}
}

func TestToolingEnvOverrides(t *testing.T) {
	t.Setenv(EnvVarToolingProfile, "local")
	t.Setenv(EnvVarToolingKORepository, "registry.example.com/core")
	t.Setenv(EnvVarToolingFlyApp, "core-app")
	t.Setenv(EnvVarToolingFlyOrg, "example-org")
	t.Setenv(EnvVarToolingTagTemplate, "custom")
	t.Setenv(EnvVarToolingSupportsDocker, "false")
	t.Setenv(EnvVarFlyTokenPath, "/tmp/custom-token")
	t.Setenv(EnvVarCloudflareTokenPath, "/tmp/cf-token")

	cfg := Tooling()

	if cfg.Active.KORepository != "registry.example.com/core" {
		t.Fatalf("expected KO repository override applied, got %q", cfg.Active.KORepository)
	}
	if cfg.Active.FlyApp != "core-app" {
		t.Fatalf("expected Fly app override, got %q", cfg.Active.FlyApp)
	}
	if cfg.Active.FlyOrg != "example-org" {
		t.Fatalf("expected Fly org override, got %q", cfg.Active.FlyOrg)
	}
	if cfg.Active.TagTemplate != "custom" {
		t.Fatalf("expected tag template override, got %q", cfg.Active.TagTemplate)
	}
	if cfg.Active.SupportsDocker {
		t.Fatalf("expected SupportsDocker to be false after override")
	}
	if cfg.Active.TokenPath != "/tmp/custom-token" {
		t.Fatalf("expected token path override, got %q", cfg.Active.TokenPath)
	}
	if cfg.Active.CloudflareTokenPath != "/tmp/cf-token" {
		t.Fatalf("expected cloudflare token path override, got %q", cfg.Active.CloudflareTokenPath)
	}
	if flyProfile, ok := cfg.Profiles["fly"]; ok {
		if flyProfile.Mode != ToolingModeFly {
			t.Fatalf("expected fly profile mode %q, got %q", ToolingModeFly, flyProfile.Mode)
		}
		if flyProfile.KORepository == "" {
			t.Fatal("expected fly profile to include KO repository")
		}
		if flyProfile.FlyConfig != "fly.toml" {
			t.Fatalf("expected fly profile fly config fly.toml, got %q", flyProfile.FlyConfig)
		}
		if flyProfile.FlyTemplate != "config/templates/fly.toml.tmpl" {
			t.Fatalf("expected fly profile fly template path config/templates/fly.toml.tmpl, got %q", flyProfile.FlyTemplate)
		}
		if flyProfile.FlyRegion != "syd" {
			t.Fatalf("expected fly profile fly region syd, got %q", flyProfile.FlyRegion)
		}
		if flyProfile.ImportPath != "./cmd/core" {
			t.Fatalf("expected fly profile import path ./cmd/core, got %q", flyProfile.ImportPath)
		}
	}
}
