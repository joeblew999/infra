package runtime

import (
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/hugo"
	"github.com/joeblew999/infra/pkg/log"
)

func startHugo(record func(error)) (func(), error) {
	log.Info("🚀 Starting Hugo documentation server...")
	if err := hugo.StartSupervised(); err != nil {
		return nil, err
	}
	log.Info("✅ Hugo documentation server started supervised", "port", config.GetHugoPort())
	NotifyCaddyRoutesChanged()
	return nil, nil
}

func ensureHugoDirectories() error {
	return config.EnsureAppDirectories()
}
