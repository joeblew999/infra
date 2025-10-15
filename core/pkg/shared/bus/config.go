package bus

import (
	"fmt"
	"time"
)

// Config captures high-level settings for the embedded JetStream bus.
type Config struct {
	Name            string
	Host            string
	Port            int
	HTTPPort        int
	EnableJetStream bool
	StoreDir        string
	MemoryStore     bool
	Clustered       bool
	SeedServers     []string
	CredentialsPath string
	StartupTimeout  time.Duration
}

// DefaultConfig returns sensible defaults for local development and tests.
func DefaultConfig() Config {
	return Config{
		Name:            "core-bus",
		Host:            "127.0.0.1",
		Port:            4222,
		HTTPPort:        8222,
		EnableJetStream: true,
		MemoryStore:     true,
		StartupTimeout:  10 * time.Second,
	}
}

// Address returns the host:port endpoint used by clients.
func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Validate performs basic configuration validation.
func (c Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("bus host required")
	}
	if c.Port == 0 {
		return fmt.Errorf("bus port required")
	}
	if c.EnableJetStream && !c.MemoryStore && c.StoreDir == "" {
		return fmt.Errorf("store dir required for persistent JetStream")
	}
	return nil
}

// InMemory toggles the configuration for ephemeral runs.
func (c Config) InMemory() Config {
	c.MemoryStore = true
	c.StoreDir = ""
	return c
}

// Persistent ensures JetStream writes to disk using the provided directory.
func (c Config) Persistent(dir string) Config {
	c.MemoryStore = false
	c.StoreDir = dir
	return c
}
