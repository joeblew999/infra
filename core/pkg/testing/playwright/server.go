package playwright

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// StartServer starts the test server based on the configuration.
// Returns the running command or nil if SkipServer is true.
func StartServer(ctx context.Context, sourceDir string, cfg ServerConfig) (*exec.Cmd, error) {
	if cfg.SkipServer {
		return nil, nil
	}

	env := os.Environ()
	var cmd *exec.Cmd

	// Use custom command if provided
	if len(cfg.Command) > 0 {
		cmd = exec.CommandContext(ctx, cfg.Command[0], cfg.Command[1:]...)
	} else if cfg.Binary != "" {
		// Use binary (absolute or relative to sourceDir)
		binaryPath := cfg.Binary
		if !filepath.IsAbs(binaryPath) {
			binaryPath = filepath.Join(sourceDir, binaryPath)
		}
		if _, err := os.Stat(binaryPath); err == nil {
			cmd = exec.CommandContext(ctx, binaryPath)
		} else {
			return nil, err
		}
	} else {
		// Default: run with "go run ."
		cmd = exec.CommandContext(ctx, "go", "run", ".")
		env = append(env, "GOWORK=off")
	}

	cmd.Dir = sourceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

// StopServer gracefully stops the server process.
// Sends SIGINT, waits 5 seconds, then kills if still running.
func StopServer(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	// Send interrupt signal
	_ = cmd.Process.Signal(os.Interrupt)

	// Wait for graceful shutdown
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Graceful shutdown succeeded
	case <-time.After(5 * time.Second):
		// Force kill after timeout
		_ = cmd.Process.Kill()
	}
}

// WaitForHTTP polls the given URL until it responds or timeout is reached.
// Returns nil when the server is ready, error on timeout.
func WaitForHTTP(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			// Accept any non-5xx status as "ready"
			if resp.StatusCode < 500 {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return errors.New("timeout waiting for http endpoint")
}
