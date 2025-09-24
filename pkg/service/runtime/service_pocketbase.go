package runtime

import (
	"context"
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/pocketbase"
)

func startPocketBase(ctx context.Context, mode string, record func(error)) (func(), error) {
	log.Info("🚀 Starting embedded PocketBase server...")
	pbEnv := "production"
	if mode == "development" {
		pbEnv = "development"
	}

	pbServer := pocketbase.NewServer(pbEnv)
	go func() {
		if err := pbServer.Start(ctx); err != nil {
			log.Warn("PocketBase failed to start", "error", err)
			if record != nil {
				record(fmt.Errorf("pocketbase failed to start: %w", err))
			}
		}
	}()

	NotifyCaddyRoutesChanged()
	return nil, nil
}

func ensurePocketBaseDirectories() error {
	return os.MkdirAll(config.GetPocketBaseDataPath(), 0o755)
}
