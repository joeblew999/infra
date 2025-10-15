package process

import (
	"strings"

	runtimecfg "github.com/joeblew999/infra/core/pkg/runtime/config"
	sharedservices "github.com/joeblew999/infra/core/pkg/shared/services"
)

// EnsureServiceBinaries builds or installs all service binaries required by the
// runtime stack. When appRoot is empty it falls back to the configured runtime
// path.
func EnsureServiceBinaries(appRoot string) error {
	root := strings.TrimSpace(appRoot)
	if root == "" {
		cfg := runtimecfg.Load()
		root = cfg.Paths.AppRoot
	}
	return sharedservices.EnsureRuntime(root)
}
