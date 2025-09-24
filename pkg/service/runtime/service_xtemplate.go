package runtime

import (
	"os"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/xtemplate"
)

func startXTemplate() (func(), error) {
	log.Info("🚀 Starting XTemplate development server...")
	if err := xtemplate.StartSupervised(); err != nil {
		return nil, err
	}
	log.Info("✅ XTemplate development server started supervised", "port", config.GetXTemplatePort())
	NotifyCaddyRoutesChanged()
	return nil, nil
}

func ensureXTemplateDirectories() error {
	return os.MkdirAll(config.GetXTemplatePath(), 0o755)
}
