package workflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// WorkflowMode describes how the automation should rebuild assets and run tests.
type WorkflowMode string

const (
	WorkflowBun  WorkflowMode = "bun"
	WorkflowNode WorkflowMode = "node"
)

// Config captures the file layout and commands required to rebuild assets,
// compile the server binary, and execute the Playwright suite.
type Config struct {
	TailwindInput   string
	TailwindOutput  string
	TailwindContent []string
	Binary          string
	ServerCommand   []string
	BaseURL         string
	Workflow        WorkflowMode
	Headed          bool
}

// DefaultConfig returns the conventions used by the DatastarUI sample app. Other
// projects can override individual fields before calling Prepare/Run.
func DefaultConfig() Config {
	return Config{
		TailwindInput:  "static/css/index.css",
		TailwindOutput: "static/css/out.css",
		TailwindContent: []string{
			"./pages/**/*",
			"../../.src/datastarui/fork/datastarui/components/**/*",
			"../../.src/datastarui/fork/datastarui/layouts/**/*",
			"../../.src/datastarui/fork/datastarui/pages/**/*",
		},
		Binary:        "datastarui-sample",
		ServerCommand: nil,
		BaseURL:       "http://localhost:4242",
		Workflow:      WorkflowBun,
		Headed:        false,
	}
}

// Run executes the selected Playwright workflow (Bun or Node) against the
// provided source directory. It installs dependencies, rebuilds assets, starts
// the server, runs the Playwright suite, and shuts everything down when
// finished.
func Run(ctx context.Context, sourceDir string, cfg Config) error {
	if err := Prepare(ctx, sourceDir, cfg); err != nil {
		return err
	}

	srvCmd, err := startServer(ctx, sourceDir, cfg)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer stopServer(srvCmd)

	if err := waitForHTTP(cfg.BaseURL, 30*time.Second); err != nil {
		return fmt.Errorf("wait for server: %w", err)
	}

	playwrightRunner, err := selectPlaywrightRunner(cfg.Workflow)
	if err != nil {
		return err
	}

	if err := runCmd(ctx, sourceDir, os.Environ(), playwrightRunner.install[0], playwrightRunner.install[1:]...); err != nil {
		return fmt.Errorf("playwright install failed: %w", err)
	}

	env := append(os.Environ(), fmt.Sprintf("PLAYWRIGHT_BASE_URL=%s", cfg.BaseURL))
	if cfg.Headed {
		env = append(env, "PLAYWRIGHT_HEADED=1")
	}
	if err := runCmd(ctx, sourceDir, env, playwrightRunner.test[0], playwrightRunner.test[1:]...); err != nil {
		return fmt.Errorf("playwright suite failed: %w", err)
	}

	return nil
}

// Prepare installs dependencies, regenerates templ output, rebuilds the
// Tailwind bundle (including hashed asset), and ensures the Go binary is up to
// date. Call this when a package only needs refreshed assets without running
// the Playwright suite.
func Prepare(ctx context.Context, sourceDir string, cfg Config) error {
	switch cfg.Workflow {
	case "", WorkflowBun:
		if err := runBunInstall(ctx, sourceDir); err != nil {
			return fmt.Errorf("bun install failed: %w", err)
		}

		if err := runTemplGenerate(ctx, sourceDir); err != nil {
			return fmt.Errorf("templ generate failed: %w", err)
		}

		if err := rebuildTailwindWith(ctx, sourceDir, cfg, []string{"bun", "x", "tailwindcss"}); err != nil {
			return fmt.Errorf("tailwind rebuild failed: %w", err)
		}

	case WorkflowNode:
		if err := runPnpmInstall(ctx, sourceDir); err != nil {
			return fmt.Errorf("pnpm install failed: %w", err)
		}

		if err := runTemplGenerate(ctx, sourceDir); err != nil {
			return fmt.Errorf("templ generate failed: %w", err)
		}

		if err := rebuildTailwindWith(ctx, sourceDir, cfg, []string{"pnpm", "exec", "tailwindcss"}); err != nil {
			return fmt.Errorf("tailwind rebuild failed: %w", err)
		}

	default:
		return fmt.Errorf("unsupported workflow: %s", cfg.Workflow)
	}

	if err := runGoBuild(ctx, sourceDir, cfg); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	return nil
}

func runBunInstall(ctx context.Context, sourceDir string) error {
	return runCmd(ctx, sourceDir, os.Environ(), "bun", "install")
}

func runPnpmInstall(ctx context.Context, sourceDir string) error {
	return runCmd(ctx, sourceDir, os.Environ(), "pnpm", "install")
}

func runTemplGenerate(ctx context.Context, sourceDir string) error {
	return runCmd(ctx, sourceDir, os.Environ(), "templ", "generate")
}

func rebuildTailwindWith(ctx context.Context, sourceDir string, cfg Config, runner []string) error {
	if len(runner) == 0 {
		return errors.New("tailwind runner not specified")
	}

	args := append([]string{}, runner...)
	args = append(args, "-i", cfg.TailwindInput, "-o", cfg.TailwindOutput)
	for _, content := range cfg.TailwindContent {
		args = append(args, "--content", content)
	}

	if err := runCmd(ctx, sourceDir, os.Environ(), args[0], args[1:]...); err != nil {
		return err
	}

	outPath := filepath.Join(sourceDir, cfg.TailwindOutput)
	data, err := os.ReadFile(outPath)
	if err != nil {
		return err
	}

	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])[:8]

	dir := filepath.Dir(outPath)
	base := filepath.Base(outPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if ext == "" {
		ext = ""
	}

	pattern := filepath.Join(dir, fmt.Sprintf("%s.*%s", name, ext))
	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {
		if match == outPath {
			continue
		}
		_ = os.Remove(match)
	}

	hashedPath := filepath.Join(dir, fmt.Sprintf("%s.%s%s", name, hash, ext))
	if err := os.WriteFile(hashedPath, data, 0o644); err != nil {
		return err
	}

	return nil
}

func runGoBuild(ctx context.Context, sourceDir string, cfg Config) error {
	env := append(os.Environ(), "GOWORK=off")
	args := []string{"build"}
	if cfg.Binary != "" {
		args = append(args, "-o", cfg.Binary)
	}
	args = append(args, "main.go")
	return runCmd(ctx, sourceDir, env, "go", args...)
}

func startServer(ctx context.Context, sourceDir string, cfg Config) (*exec.Cmd, error) {
	env := os.Environ()
	var cmd *exec.Cmd

	if len(cfg.ServerCommand) > 0 {
		cmd = exec.CommandContext(ctx, cfg.ServerCommand[0], cfg.ServerCommand[1:]...)
	} else {
		binaryPath := cfg.Binary
		if binaryPath != "" {
			binaryPath = filepath.Join(sourceDir, binaryPath)
			if _, err := os.Stat(binaryPath); err == nil {
				cmd = exec.CommandContext(ctx, binaryPath)
			}
		}
		if cmd == nil {
			cmd = exec.CommandContext(ctx, "go", "run", ".")
			env = append(env, "GOWORK=off")
		}
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

func stopServer(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Signal(os.Interrupt)
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
	}
}

func runCmd(ctx context.Context, sourceDir string, env []string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = sourceDir
	if len(env) > 0 {
		cmd.Env = env
	} else {
		cmd.Env = os.Environ()
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func waitForHTTP(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return errors.New("timeout waiting for http endpoint")
}

func selectPlaywrightRunner(mode WorkflowMode) (*playwrightRunner, error) {
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

type playwrightRunner struct {
	install []string
	test    []string
}
