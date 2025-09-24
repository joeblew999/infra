package runtime

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/caddy"
	"github.com/joeblew999/infra/pkg/log"
)

var caddyReload = caddy.ReloadWithConfig

func generateCaddyConfig(specs []ServiceSpec) (caddy.CaddyConfig, error) {
	tpl, err := buildCaddyTemplate(specs)
	if err != nil {
		return caddy.CaddyConfig{}, err
	}

	cfg := caddy.CaddyConfig{
		Port:   tpl.ListenPort,
		Target: tpl.RootTarget,
	}
	for _, route := range tpl.Routes {
		cfg.Routes = append(cfg.Routes, caddy.ProxyRoute{Path: route.Path, Target: route.Target})
	}
	return cfg, nil
}

func reloadCaddyConfig() error {
	specs := activeServiceSpecs
	if len(specs) == 0 {
		specs = buildServiceSpecs(activeOptions)
	}

	if !hasService(specs, ServiceCaddy) {
		return nil
	}

	cfg, err := generateCaddyConfig(specs)
	if err != nil {
		return err
	}

	if err := caddyReload(&cfg); err != nil {
		return fmt.Errorf("reload caddy: %w", err)
	}
	log.Info("🔄 Caddy configuration reloaded", "routes", len(cfg.Routes))
	return nil
}

func hasService(specs []ServiceSpec, id ServiceID) bool {
	for _, spec := range specs {
		if spec.ID == id {
			return true
		}
	}
	return false
}

// NotifyCaddyRoutesChanged rebuilds the Caddy configuration and reloads the running process.
func NotifyCaddyRoutesChanged() {
	if err := reloadCaddyConfig(); err != nil {
		log.Warn("⚠️ Failed to reload Caddy configuration", "error", err)
	}
}
