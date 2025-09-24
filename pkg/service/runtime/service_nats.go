package runtime

import (
	"context"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/nats/orchestrator"
)

func startNATS(ctx context.Context, record func(error)) (func(), error) {
	leafURL, cleanup, err := orchestrator.StartWithEnvironment(ctx)
	if err != nil {
		return nil, err
	}

	log.Info("✅ NATS stack ready", "leaf_url", leafURL)
	return cleanup, nil
}

func ensureNATSDirectories() error {
	path := filepath.Join(config.GetDataPath(), "nats")
	return os.MkdirAll(path, 0o755)
}
