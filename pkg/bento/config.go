package bento

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/config"
)

// Config represents bento configuration structure
type Config struct {
	HTTP    HTTPConfig    `yaml:"http"`
	Input   InputConfig   `yaml:"input"`
	Output  OutputConfig  `yaml:"output"`
	Buffer  BufferConfig  `yaml:"buffer,omitempty"`
	Metrics MetricsConfig `yaml:"metrics,omitempty"`
}

type HTTPConfig struct {
	Address string `yaml:"address"`
	Enabled bool   `yaml:"enabled"`
}

type InputConfig struct {
	Generate *GenerateConfig `yaml:"generate,omitempty"`	
	HTTP     *HTTPInputConfig `yaml:"http,omitempty"`
	Kafka    *KafkaInputConfig `yaml:"kafka,omitempty"`
}

type GenerateConfig struct {
	Mapping  string `yaml:"mapping"`
	Interval string `yaml:"interval"`
	Count    int    `yaml:"count,omitempty"`
}

type HTTPInputConfig struct {
	Address string `yaml:"address"`
}

type KafkaInputConfig struct {
	Addresses []string `yaml:"addresses"`
	Topics    []string `yaml:"topics"`
}

type OutputConfig struct {
	Stdout *StdoutConfig `yaml:"stdout,omitempty"`
	HTTP   *HTTPOutputConfig `yaml:"http,omitempty"`
	Kafka  *KafkaOutputConfig `yaml:"kafka,omitempty"`
}

type StdoutConfig struct {
}

type HTTPOutputConfig struct {
	URL string `yaml:"url"`
}

type KafkaOutputConfig struct {
	Addresses []string `yaml:"addresses"`
	Topic     string   `yaml:"topic"`
}

type BufferConfig struct {
	Type   string `yaml:"type,omitempty"`
	Memory *MemoryBufferConfig `yaml:"memory,omitempty"`
}

type MemoryBufferConfig struct {
	Limit int64 `yaml:"limit,omitempty"`
}

type MetricsConfig struct {
	HTTP *HTTPMetricsConfig `yaml:"http,omitempty"`
}

type HTTPMetricsConfig struct {
	Address string `yaml:"address,omitempty"`
	Path    string `yaml:"path,omitempty"`
}

// GetConfigPath returns the path to the bento configuration directory
func GetConfigPath() string {
	return config.GetBentoPath()
}

// GetConfigFile returns the path to the default bento configuration file
func GetConfigFile() string {
	return filepath.Join(GetConfigPath(), "bento.yaml")
}

// EnsureConfigDir ensures the bento configuration directory exists
func EnsureConfigDir() error {
	return os.MkdirAll(GetConfigPath(), 0755)
}

// CreateDefaultConfig creates a default bento configuration file if it doesn't exist
func CreateDefaultConfig() error {
	configFile := GetConfigFile()
	
	// Check if config already exists
	if _, err := os.Stat(configFile); err == nil {
		return nil // Config already exists, nothing to do
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check config file: %w", err)
	}

	// Ensure directory exists
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to ensure config directory: %w", err)
	}

	defaultConfig := `http:
  address: 0.0.0.0:4195
  enabled: true

input:
  generate:
    mapping: |
      root = { "message": "hello world", "timestamp": now() }
    interval: 5s

output:
  stdout: {}

# Uncomment for HTTP server
#   http:
#     address: 0.0.0.0:8080
#     enabled: true

# Uncomment for metrics
# metrics:
#   http:
#     address: 0.0.0.0:4196
#     path: /metrics
`

	// Write default config atomically
	tmpFile := configFile + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write temporary config: %w", err)
	}
	
	return os.Rename(tmpFile, configFile)
}

// ValidateConfig validates bento configuration
func ValidateConfig(configPath string) error {
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("config file not found: %s", configPath)
	}
	return nil
}