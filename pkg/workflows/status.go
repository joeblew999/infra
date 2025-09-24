package workflows

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// StatusOptions configures deployment status checks.
type StatusOptions struct {
	AppName string
	Verbose bool
	Logs    int
}

// CheckDeploymentStatus runs flyctl status/logs for the given application.
func CheckDeploymentStatus(opts StatusOptions) error {
	if opts.AppName == "" {
		opts.AppName = getEnvOrDefault("FLY_APP_NAME", "infra-mgmt")
	}

	statusArgs := []string{"status", "-a", opts.AppName}
	if opts.Verbose {
		statusArgs = append(statusArgs, "--json")
	}

	if err := runBinary(config.GetFlyctlBinPath(), statusArgs...); err != nil {
		return fmt.Errorf("failed to fetch status: %w", err)
	}

	if opts.Logs > 0 {
		log.Info("Fetching recent application logs", "app", opts.AppName, "lines", opts.Logs)
		logArgs := []string{"logs", "-a", opts.AppName, fmt.Sprintf("--max=%d", opts.Logs)}
		if opts.Verbose {
			logArgs = append(logArgs, "--full")
		}
		if err := runBinary(config.GetFlyctlBinPath(), logArgs...); err != nil {
			return fmt.Errorf("failed to fetch logs: %w", err)
		}
	}

	return nil
}
