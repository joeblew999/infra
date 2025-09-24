package runtime

import (
	"os"

	"github.com/joeblew999/infra/pkg/bento"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	svcports "github.com/joeblew999/infra/pkg/service/ports"
)

func startBento() (func(), error) {
	log.Info("🚀 Starting Bento stream processing service...")
	if err := bento.StartSupervised(svcports.ParsePort(config.GetBentoPort())); err != nil {
		return nil, err
	}
	log.Info("✅ Bento service started supervised", "port", config.GetBentoPort())
	NotifyCaddyRoutesChanged()
	return nil, nil
}

func ensureBentoDirectories() error {
	return os.MkdirAll(config.GetBentoPath(), 0o755)
}
