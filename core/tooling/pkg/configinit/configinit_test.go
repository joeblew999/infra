package configinit

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
)

func writeTemplates(t *testing.T, coreDir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(coreDir, "config", "templates"), 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "go.mod"), []byte("module example.com/core\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "config", "templates", "ko.yaml.tmpl"), []byte("repository: {{ .Repository }}\n"), 0o644); err != nil {
		t.Fatalf("write ko template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "config", "templates", "fly.toml.tmpl"), []byte("app = \"{{ .AppName }}\"\n"), 0o644); err != nil {
		t.Fatalf("write fly template: %v", err)
	}
}

func TestPrepareAndRender(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CORE_APP_ROOT", dir)

	coreDir := filepath.Join(dir, "core")
	writeTemplates(t, coreDir)

	opts := Options{
		Profile:    sharedcfg.ToolingProfile{FlyApp: "test-app", FlyRegion: "syd"},
		RepoRoot:   dir,
		CoreDir:    coreDir,
		Repository: "registry.fly.io/test",
	}

	plan, err := Prepare(context.Background(), opts)
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if len(plan.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(plan.Targets))
	}

	rendered, err := Render(plan)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if len(rendered) != 2 {
		t.Fatalf("expected 2 rendered outputs, got %d", len(rendered))
	}

	var foundRepo bool
	for _, output := range rendered {
		if strings.Contains(string(output), "registry.fly.io/test") {
			foundRepo = true
		}
	}
	if !foundRepo {
		t.Fatalf("rendered outputs missing repository reference")
	}
}

func TestRunWritesFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CORE_APP_ROOT", dir)

	coreDir := filepath.Join(dir, "core")
	writeTemplates(t, coreDir)

	opts := Options{
		Profile:    sharedcfg.ToolingProfile{FlyApp: "test-app", FlyRegion: "syd"},
		RepoRoot:   dir,
		CoreDir:    coreDir,
		Repository: "registry.fly.io/test",
		Stdout:     io.Discard,
		Stderr:     io.Discard,
	}

	res, err := Run(context.Background(), opts)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(res.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(res.Files))
	}

	for _, file := range res.Files {
		if _, err := os.Stat(file.Path); err != nil {
			t.Fatalf("expected file %s to exist: %v", file.Path, err)
		}
	}
}
