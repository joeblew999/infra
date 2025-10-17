package playwright

import (
	"context"
	"fmt"
)

// Run executes the complete Playwright test workflow:
// 1. Verify workflow tools are available
// 2. Run optional prepare hook (project-specific setup)
// 3. Install Playwright browsers
// 4. Start server (if configured)
// 5. Wait for server to be ready
// 6. Run Playwright tests
// 7. Stop server gracefully
func Run(ctx context.Context, cfg Config) error {
	// Verify required tools
	if err := VerifyWorkflow(cfg.Workflow); err != nil {
		return fmt.Errorf("workflow verification failed: %w", err)
	}

	// Run optional prepare hook
	if cfg.Prepare != nil {
		if err := cfg.Prepare(ctx, cfg.SourceDir); err != nil {
			return fmt.Errorf("prepare failed: %w", err)
		}
	}

	// Install Playwright browsers
	if err := InstallPlaywright(ctx, cfg.SourceDir, cfg.Workflow); err != nil {
		return fmt.Errorf("playwright install failed: %w", err)
	}

	// Start server if needed
	srvCmd, err := StartServer(ctx, cfg.SourceDir, cfg.ServerConfig)
	if err != nil {
		return fmt.Errorf("start server failed: %w", err)
	}
	defer StopServer(srvCmd)

	// Wait for server to be ready
	if srvCmd != nil {
		timeout := cfg.ServerConfig.StartTimeout
		if timeout == 0 {
			timeout = DefaultServerConfig().StartTimeout
		}
		if err := WaitForHTTP(cfg.BaseURL, timeout); err != nil {
			return fmt.Errorf("wait for server failed: %w", err)
		}
	}

	// Run Playwright tests
	if err := RunPlaywrightTests(ctx, cfg.SourceDir, cfg.BaseURL, cfg.Workflow, cfg.Headed); err != nil {
		return fmt.Errorf("playwright tests failed: %w", err)
	}

	return nil
}
