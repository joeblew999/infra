package bento

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
)

type Service struct {
	cmd       *exec.Cmd
	ctx       context.Context
	cancel    context.CancelFunc
	configDir string
	port      int
}

func NewService(port int) (*Service, error) {
	configDir := config.GetBentoPath()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create bento config directory: %w", err)
	}

	return &Service{
		configDir: configDir,
		port:      port,
	}, nil
}

func (s *Service) Start() error {
	bentoPath, err := dep.Get("bento")
	if err != nil {
		return fmt.Errorf("bento binary not found: %w", err)
	}
	
	// Ensure absolute path for reliability
	bentoPath, err = filepath.Abs(bentoPath)
	if err != nil {
		return fmt.Errorf("failed to resolve bento path: %w", err)
	}

	configFile := filepath.Join(s.configDir, "bento.yaml")
	
	// Ensure config directory exists
	if err := os.MkdirAll(s.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create bento config directory: %w", err)
	}
	
	// Check if config file exists, create if missing
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Info("Bento config not found, creating default", "path", configFile)
		if err := s.createDefaultConfig(configFile); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check config file: %w", err)
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.cmd = exec.CommandContext(s.ctx, bentoPath, "run", configFile)
	s.cmd.Dir = s.configDir
	s.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start bento: %w", err)
	}

	log.Info("Bento service started", "port", s.port, "config", configFile)
	return nil
}

func (s *Service) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}

	if s.cmd != nil && s.cmd.Process != nil {
		if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Warn("Failed to send SIGTERM to bento", "error", err)
		}

		// Wait for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()

		select {
		case <-time.After(5 * time.Second):
			log.Warn("Bento did not stop gracefully, killing")
			if err := s.cmd.Process.Kill(); err != nil {
				return fmt.Errorf("failed to kill bento: %w", err)
			}
		case err := <-done:
			if err != nil {
				log.Info("Bento stopped", "error", err)
			} else {
				log.Info("Bento stopped gracefully")
			}
		}
	}

	return nil
}

func (s *Service) Wait() error {
	if s.cmd != nil {
		return s.cmd.Wait()
	}
	return nil
}

func (s *Service) createDefaultConfig(configPath string) error {
	config := fmt.Sprintf(`
http:
  address: 0.0.0.0:%d
  enabled: true

input:
  generate:
    mapping: |
      root = { "message": "hello world", "timestamp": now() }
    interval: 5s

output:
  stdout: {}
`, s.port)

	return os.WriteFile(configPath, []byte(config), 0644)
}