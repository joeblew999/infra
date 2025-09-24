package runtime

import (
	"os"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/mox"
)

func startMox() (func(), error) {
	log.Info("🚀 Starting mox mail server...")
	if err := mox.StartSupervised("localhost", "admin@localhost"); err != nil {
		return nil, err
	}
	log.Info("✅ Mox mail server started supervised")
	return nil, nil
}

func ensureMoxDirectories() error {
	return os.MkdirAll(config.GetMoxDataPath(), 0o755)
}
