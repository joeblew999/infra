package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	runtimecfg "github.com/joeblew999/infra/core/pkg/runtime/config"
	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	caddyservice "github.com/joeblew999/infra/core/services/caddy"
	natssvc "github.com/joeblew999/infra/core/services/nats"
	pocketbasesvc "github.com/joeblew999/infra/core/services/pocketbase"
)

// EnsureFunc represents a service-specific ensure routine.
type EnsureFunc func() error

// RuntimeEnsurers returns the default set of ensure functions needed for the
// orchestrator stack.
func RuntimeEnsurers() []EnsureFunc {
	return []EnsureFunc{ensureProcessCompose, ensureNATS, ensurePocketBase, ensureCaddy}
}

// EnsureRuntime ensures all default runtime services are available.
func EnsureRuntime(appRoot string) error {
	return Ensure(appRoot, RuntimeEnsurers()...)
}

// Ensure executes the provided ensure functions while temporarily updating the
// CORE_APP_ROOT environment variable when an application root is supplied.
func Ensure(appRoot string, ensures ...EnsureFunc) error {
	restore := overrideAppRoot(appRoot)
	defer restore()

	for _, fn := range ensures {
		if fn == nil {
			continue
		}
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func ensureNATS() error {
	spec, err := natssvc.LoadSpec()
	if err != nil {
		return fmt.Errorf("nats spec: %w", err)
	}
	if _, err := spec.EnsureBinaries(); err != nil {
		return fmt.Errorf("nats ensure binaries: %w", err)
	}
	return nil
}

func ensurePocketBase() error {
	spec, err := pocketbasesvc.LoadSpec()
	if err != nil {
		return fmt.Errorf("pocketbase spec: %w", err)
	}
	if _, err := spec.EnsureBinaries(); err != nil {
		return fmt.Errorf("pocketbase ensure binaries: %w", err)
	}
	return nil
}

func ensureCaddy() error {
	spec, err := caddyservice.LoadConfig()
	if err != nil {
		return fmt.Errorf("caddy config: %w", err)
	}
	if _, err := spec.EnsureBinaries(); err != nil {
		return fmt.Errorf("caddy ensure binaries: %w", err)
	}
	return nil
}

func ensureProcessCompose() error {
	cfg := runtimecfg.Load()
	depDir := filepath.Join(cfg.Paths.AppRoot, ".dep")
	binPath := filepath.Join(depDir, "process-compose")

	// Check if binary already exists
	if _, err := os.Stat(binPath); err == nil {
		return nil // Already exists
	}

	// Create .dep directory if it doesn't exist
	if err := os.MkdirAll(depDir, 0o755); err != nil {
		return fmt.Errorf("create dep dir: %w", err)
	}

	// Build process-compose binary
	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/processcompose")
	cmd.Dir = cfg.Paths.AppRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build process-compose: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func overrideAppRoot(appRoot string) func() {
	root := strings.TrimSpace(appRoot)
	if root == "" {
		return func() {}
	}
	original := os.Getenv(sharedcfg.EnvVarAppRoot)
	_ = os.Setenv(sharedcfg.EnvVarAppRoot, root)
	return func() {
		if original == "" {
			_ = os.Unsetenv(sharedcfg.EnvVarAppRoot)
		} else {
			_ = os.Setenv(sharedcfg.EnvVarAppRoot, original)
		}
	}
}
