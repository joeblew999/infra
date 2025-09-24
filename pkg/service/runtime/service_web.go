package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/webapp"
)

func startWebServer(ctx context.Context, opts Options, record func(error)) (func(), error) {
	webPort := config.GetWebServerPort()
	log.Info("🌐 Starting web server", "address", "http://0.0.0.0:"+webPort)

	go func() {
		svc := webapp.NewService(
			webapp.WithPort(webPort),
			webapp.WithNATSURL(config.GetNATSURL()),
		)

		if err := svc.Start(ctx); err != nil {
			log.Error("❌ Failed to start web server", "error", err)
			if record != nil {
				record(fmt.Errorf("web server failed to start: %w", err))
			}
		}
	}()

	time.Sleep(500 * time.Millisecond)
	return nil, nil
}

func ensureWebDirectories() error {
	return config.EnsureAppDirectories()
}
