package service

import (
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/goreman"
)

// Option customizes a goreman ProcessConfig.
type Option func(*goreman.ProcessConfig)

// WithWorkingDir sets the working directory for the supervised process.
func WithWorkingDir(dir string) Option {
	return func(cfg *goreman.ProcessConfig) {
		cfg.WorkingDir = dir
	}
}

// WithEnv appends additional environment variables in KEY=VALUE form.
func WithEnv(extra ...string) Option {
	return func(cfg *goreman.ProcessConfig) {
		cfg.Env = append(cfg.Env, extra...)
	}
}

// NewConfig constructs a ProcessConfig with sensible defaults (working dir '.', current environment).
func NewConfig(command string, args []string, opts ...Option) *goreman.ProcessConfig {
	cfg := &goreman.ProcessConfig{
		Command:    command,
		Args:       append([]string(nil), args...),
		WorkingDir: ".",
		Env:        append([]string(nil), os.Environ()...),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Start registers and starts a process with goreman after ensuring defaults are set.
func Start(name string, cfg *goreman.ProcessConfig) error {
	if cfg == nil {
		return fmt.Errorf("supervisor: nil process config for %s", name)
	}
	if cfg.WorkingDir == "" {
		cfg.WorkingDir = "."
	}
	if len(cfg.Env) == 0 {
		cfg.Env = append([]string(nil), os.Environ()...)
	}
	return goreman.RegisterAndStart(name, cfg)
}
