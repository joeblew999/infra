package runtime

import (
	"os"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/deck"
	"github.com/joeblew999/infra/pkg/log"
	svcports "github.com/joeblew999/infra/pkg/service/ports"
)

func startDeckAPI() (func(), error) {
	log.Info("🚀 Starting deck API service...")
	if err := deck.StartAPISupervised(svcports.ParsePort(config.GetDeckAPIPort())); err != nil {
		return nil, err
	}
	log.Info("✅ Deck API service started supervised", "port", config.GetDeckAPIPort())
	NotifyCaddyRoutesChanged()
	return nil, nil
}

func ensureDeckDirectories() error {
	paths := []string{
		config.GetDeckPath(),
		config.GetDeckBinPath(),
		config.GetDeckCachePath(),
	}
	for _, p := range paths {
		if p == "" {
			continue
		}
		if err := os.MkdirAll(p, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func startDeckWatcher() (func(), error) {
	log.Info("🚀 Starting deck asset watcher...")
	if err := deck.StartWatcherSupervised([]string{"test/deck"}, []string{"svg", "png", "pdf"}); err != nil {
		return nil, err
	}
	log.Info("✅ Deck watcher service started supervised")
	return nil, nil
}
