package playwright

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// InstallPlaywright installs Playwright browsers for the specified workflow.
func InstallPlaywright(ctx context.Context, sourceDir string, workflow WorkflowMode) error {
	runner, err := selectRunner(workflow)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, runner.install[0], runner.install[1:]...)
	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}

// RunPlaywrightTests executes the Playwright test suite.
func RunPlaywrightTests(ctx context.Context, sourceDir, baseURL string, workflow WorkflowMode, headed bool) error {
	runner, err := selectRunner(workflow)
	if err != nil {
		return err
	}

	// Set up environment variables
	env := os.Environ()
	env = append(env, fmt.Sprintf("PLAYWRIGHT_BASE_URL=%s", baseURL))
	if headed {
		env = append(env, "PLAYWRIGHT_HEADED=1")
	}

	cmd := exec.CommandContext(ctx, runner.test[0], runner.test[1:]...)
	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	return cmd.Run()
}

// selectRunner returns the appropriate Playwright commands for the workflow mode.
func selectRunner(mode WorkflowMode) (*playwrightRunner, error) {
	switch mode {
	case "", WorkflowBun:
		return &playwrightRunner{
			install: []string{"bun", "x", "playwright", "install"},
			test:    []string{"bun", "x", "playwright", "test"},
		}, nil
	case WorkflowNode:
		return &playwrightRunner{
			install: []string{"pnpm", "exec", "playwright", "install"},
			test:    []string{"pnpm", "exec", "playwright", "test"},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported workflow: %s", mode)
	}
}

// VerifyWorkflow checks that the required tools are available for the workflow.
func VerifyWorkflow(workflow WorkflowMode) error {
	switch workflow {
	case "", WorkflowBun:
		if _, err := exec.LookPath("bun"); err != nil {
			return errors.New("bun runtime not found: install with 'brew install oven-sh/bun/bun'")
		}
	case WorkflowNode:
		if _, err := exec.LookPath("pnpm"); err != nil {
			return errors.New("pnpm not found: install with 'npm install -g pnpm'")
		}
	default:
		return fmt.Errorf("unsupported workflow: %s", workflow)
	}
	return nil
}

type playwrightRunner struct {
	install []string
	test    []string
}
