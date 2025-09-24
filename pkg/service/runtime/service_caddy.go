package runtime

import (
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/caddy"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

func startCaddy() (func(), error) {
	specs := activeServiceSpecs
	if len(specs) == 0 {
		specs = buildServiceSpecs(activeOptions)
	}

	cfg, err := generateCaddyConfig(specs)
	if err != nil {
		return nil, fmt.Errorf("build caddy config: %w", err)
	}

	log.Info("🚀 Starting Caddy reverse proxy...", "target", cfg.Target, "routes", len(cfg.Routes))
	if err := caddy.StartSupervised(&cfg); err != nil {
		return nil, err
	}
	log.Info("✅ Caddy reverse proxy started supervised")
	return nil, nil
}

func ensureCaddyDirectories() error {
	return os.MkdirAll(config.GetCaddyPath(), 0o755)
}
