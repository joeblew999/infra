package gateway

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/service"
)

// StartS3Gateway starts the nats-s3 gateway under goreman supervision (idempotent).
func StartS3Gateway(natsURL string) {
	if err := dep.InstallBinary(config.BinaryNatsS3, false); err != nil {
		log.Error("Failed to install nats-s3 binary, cannot start gateway", "error", err)
		return
	}

	listenAddr := fmt.Sprintf("0.0.0.0:%s", config.GetNatsS3Port())
	processCfg := service.NewConfig(config.Get(config.BinaryNatsS3), []string{"--listen", listenAddr, "--natsServers", natsURL})
	if err := service.Start("nats-s3", processCfg); err != nil {
		log.Error("Failed to start nats-s3 gateway with goreman", "error", err)
	}
}
