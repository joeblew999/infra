package caddy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/service"
)

func init() {
	goreman.RegisterService("caddy", func() error {
		return StartSupervised(nil)
	})
}

// StartSupervised launches Caddy under goreman supervision, generating a Caddyfile when needed.
func StartSupervised(cfg *CaddyConfig) error {
	configPath, err := ensureCaddyfile(cfg)
	if err != nil {
		return err
	}

	if err := dep.InstallBinary(config.BinaryCaddy, false); err != nil {
		return fmt.Errorf("failed to ensure caddy binary: %w", err)
	}

	processCfg := service.NewConfig(
		config.GetCaddyBinPath(),
		[]string{"run", "--config", configPath, "--adapter", "caddyfile"},
		service.WithEnv("CADDY_LOG_LEVEL=ERROR"),
	)

	return service.Start("caddy", processCfg)
}

// StartWithConfig writes the provided configuration then starts Caddy in the background using the Runner.
func StartWithConfig(cfg *CaddyConfig) *Runner {
	if cfg == nil {
		configValue := NewPresetConfig(PresetDevelopment, defaultCaddyPort())
		cfg = &configValue
	}

	if err := cfg.GenerateAndSave("Caddyfile"); err != nil {
		fmt.Printf("Failed to generate Caddyfile: %v\n", err)
		return nil
	}

	runner := New()
	runner.StartInBackground(defaultCaddyfilePath())
	return runner
}

// ReloadWithConfig regenerates the Caddyfile using the provided configuration and reloads the running process.
func ReloadWithConfig(cfg *CaddyConfig) error {
	configPath, err := ensureCaddyfile(cfg)
	if err != nil {
		return err
	}

	runner := New()
	return runner.Reload(configPath)
}

func ensureCaddyfile(cfg *CaddyConfig) (string, error) {
	if cfg != nil {
		if err := cfg.GenerateAndSave("Caddyfile"); err != nil {
			return "", fmt.Errorf("failed to generate Caddyfile: %w", err)
		}
		return defaultCaddyfilePath(), nil
	}

	configPath := defaultCaddyfilePath()
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return "", fmt.Errorf("failed to prepare caddy config directory: %w", err)
	}

	defaultConfig := NewPresetConfig(PresetDevelopment, defaultCaddyPort())
	if err := defaultConfig.GenerateAndSave("Caddyfile"); err != nil {
		return "", fmt.Errorf("failed to create default Caddyfile: %w", err)
	}

	return configPath, nil
}
