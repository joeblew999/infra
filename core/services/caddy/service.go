package caddyservice

import (
	"context"
	"embed"
	json "encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	_ "github.com/caddy-dns/acmedns"
	"github.com/caddyserver/caddy/v2"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/mholt/caddy-l4"

	runtimedep "github.com/joeblew999/infra/core/pkg/runtime/dep"
	composecfg "github.com/joeblew999/infra/core/pkg/runtime/process/composecfg"
)

//go:embed service.json
var manifestFS embed.FS

// Config models our embedded Caddy configuration manifest.
type Config struct {
	Binaries []runtimedep.BinarySpec `json:"binaries"`
	Process  struct {
		Env     map[string]string  `json:"env"`
		Compose *composecfg.Config `json:"compose,omitempty"`
	} `json:"process"`
	Ports struct {
		HTTP Port `json:"http"`
	} `json:"ports"`
	Config struct {
		Target string `json:"target"`
	} `json:"config"`
	Scalable      bool   `json:"scalable,omitempty"`
	ScaleStrategy string `json:"scale_strategy,omitempty"`
}

type Port struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// LoadConfig loads the embedded config manifest.
func LoadConfig() (*Config, error) {
	data, err := manifestFS.ReadFile("service.json")
	if err != nil {
		return nil, fmt.Errorf("read caddy manifest: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("decode caddy manifest: %w", err)
	}
	return &cfg, nil
}

// ComposeOverrides returns the Process Compose overrides defined in the manifest.
func (c *Config) ComposeOverrides() map[string]any {
	if c == nil || c.Process.Compose == nil {
		return nil
	}
	return c.Process.Compose.Map()
}

// EnsureBinaries builds or stages the caddy binary defined in the manifest.
func (c *Config) EnsureBinaries() (map[string]string, error) {
	manifest := &runtimedep.Manifest{Binaries: c.Binaries}
	return runtimedep.EnsureManifest(manifest, runtimedep.DefaultInstaller)
}

// Run starts an embedded Caddy instance with a simple reverse proxy to target.
func Run(ctx context.Context, extraArgs []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// Ignore extra args - Caddy is configured programmatically via service.json
	if len(extraArgs) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: extra args ignored (configured via service.json): %v\n", extraArgs)
	}

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if _, err := cfg.EnsureBinaries(); err != nil {
		return err
	}

	config, err := buildConfig(cfg)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- withEnv(cfg.Process.Env, func() error {
			return caddy.Run(&config)
		})
	}()

	if err := waitForTCP(cfg.Ports.HTTP.Port, 10*time.Second); err != nil {
		_ = caddy.Stop()
		return err
	}
	fmt.Fprintf(os.Stderr, "[caddy] Server started on http://127.0.0.1:%d\n", cfg.Ports.HTTP.Port)
	fmt.Fprintf(os.Stderr, "[caddy] Proxying to target: %s\n", cfg.Config.Target)
	fmt.Fprintf(os.Stderr, "[caddy] Waiting for shutdown signal...\n")

	select {
	case err := <-errCh:
		if err != nil {
			fmt.Fprintf(os.Stderr, "[caddy] Error: %v\n", err)
		}
		return err
	case <-ctx.Done():
		done := make(chan error, 1)
		go func() { done <- caddy.Stop() }()
		select {
		case err := <-done:
			return err
		case <-time.After(5 * time.Second):
			return fmt.Errorf("shutdown caddy: timeout waiting for stop")
		}
	}
}

func buildConfig(cfg *Config) (caddy.Config, error) {
	listen := fmt.Sprintf(":%d", cfg.Ports.HTTP.Port)
	target := cfg.Config.Target

	httpConfig := map[string]any{
		"servers": map[string]any{
			"core": map[string]any{
				"listen": []string{listen},
				"routes": []map[string]any{
					{
						"handle": []map[string]any{
							{
								"handler":   "reverse_proxy",
								"upstreams": []map[string]any{{"dial": target}},
							},
						},
					},
				},
			},
		},
	}

	httpRaw, err := json.Marshal(httpConfig)
	if err != nil {
		return caddy.Config{}, fmt.Errorf("marshal http config: %w", err)
	}

	return caddy.Config{
		Admin: &caddy.AdminConfig{Disabled: true},
		AppsRaw: caddy.ModuleMap{
			"http": json.RawMessage(httpRaw),
		},
	}, nil
}

func withEnv(overrides map[string]string, fn func() error) error {
	if len(overrides) == 0 {
		return fn()
	}
	originals := make(map[string]*string, len(overrides))
	for key, value := range overrides {
		if value == "" {
			continue
		}
		if existing, ok := os.LookupEnv(key); ok {
			val := existing
			originals[key] = &val
		} else {
			originals[key] = nil
		}
		_ = os.Setenv(key, value)
	}

	defer func() {
		for key, val := range originals {
			if val == nil {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, *val)
			}
		}
	}()

	return fn()
}

func waitForTCP(port int, timeout time.Duration) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("tcp", addr, 250*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("caddy port %d not ready: %w", port, err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}
