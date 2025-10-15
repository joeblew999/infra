package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	cloudflareprefs "github.com/joeblew999/infra/core/tooling/pkg/cloudflare"
	flyprefs "github.com/joeblew999/infra/core/tooling/pkg/fly"
	profiles "github.com/joeblew999/infra/core/tooling/pkg/profiles"
)

func setupRepo(t *testing.T, root string) {
	t.Helper()
	coreDir := filepath.Join(root, "core")
	if err := os.MkdirAll(coreDir, 0o755); err != nil {
		t.Fatalf("mkdir core: %v", err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "go.mod"), []byte("module example.com/core\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
}

func TestStatusSnapshot(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CORE_APP_ROOT", dir)

	setupRepo(t, dir)

	flyprefs.SaveSettings(flyprefs.Settings{OrgSlug: "org", RegionCode: "syd", RegionName: "Sydney", UpdatedAt: time.Now()})
	cloudflareprefs.SaveSettings(cloudflareprefs.Settings{ZoneName: "example.com", AccountID: "acct", UpdatedAt: time.Now()})

	status, err := StatusSnapshot(context.Background(), profiles.ContextOptions{RepoRoot: dir})
	if err != nil {
		t.Fatalf("StatusSnapshot error: %v", err)
	}

	if status.RepoRoot != dir {
		t.Fatalf("expected repo root %q, got %q", dir, status.RepoRoot)
	}
	if status.Fly.OrgSlug != "org" {
		t.Fatalf("expected org 'org', got %q", status.Fly.OrgSlug)
	}
	if status.Cloudflare.ZoneName != "example.com" {
		t.Fatalf("expected zone example.com, got %q", status.Cloudflare.ZoneName)
	}
}
