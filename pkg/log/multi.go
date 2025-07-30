package log

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	gonats "github.com/nats-io/nats.go"
	"github.com/joeblew999/infra/pkg/config"
	slogmulti "github.com/samber/slog-multi"
	slognats "github.com/samber/slog-nats"
)

// MultiConfig holds configuration for multi-destination logging
type MultiConfig struct {
	Destinations []DestinationConfig `json:"destinations"`
}

// LoadConfig loads logging configuration from centralized config path
// Returns empty config if file doesn't exist or is invalid
func LoadConfig() MultiConfig {
	var cfg MultiConfig
	
	configFile := config.GetLoggingConfigFile()
	data, err := os.ReadFile(configFile)
	if err != nil {
		return cfg // Empty config = use defaults
	}
	
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg // Invalid JSON = use defaults
	}
	
	// Filter out NATS destinations when no NATS server available
	// This prevents trying to connect to non-existent NATS
	filtered := []DestinationConfig{}
	for _, dest := range cfg.Destinations {
		if dest.Type != "nats" {
			filtered = append(filtered, dest)
		}
	}
	
	return MultiConfig{Destinations: filtered}
}

// DestinationConfig defines a single log destination
type DestinationConfig struct {
	Type    string `json:"type"`    // "stdout", "file", "nats"
	Level   string `json:"level"`   // "debug", "info", "warn", "error"
	Format  string `json:"format"`  // "json", "text"
	Path    string `json:"path,omitempty"`    // for file destinations
	Subject string `json:"subject,omitempty"` // for NATS
}

// InitMultiLogger initializes multi-destination logging with slog-multi
func InitMultiLogger(config MultiConfig) error {
	var handlers []slog.Handler

	for _, dest := range config.Destinations {
		handler, err := createHandler(dest)
		if err != nil {
			return fmt.Errorf("failed to create handler for %s: %w", dest.Type, err)
		}
		handlers = append(handlers, handler)
	}

	if len(handlers) == 0 {
		return fmt.Errorf("no valid destinations configured")
	}

	// Use slog-multi to fan out to all destinations
	multiHandler := slogmulti.Fanout(handlers...)
	slog.SetDefault(slog.New(multiHandler))

	return nil
}

// InitMultiLoggerWithNATS initializes multi-destination logging with NATS support
func InitMultiLoggerWithNATS(config MultiConfig, nc *gonats.Conn) error {
	var handlers []slog.Handler

	for _, dest := range config.Destinations {
		handler, err := createHandlerWithNATS(dest, nc)
		if err != nil {
			return fmt.Errorf("failed to create handler for %s: %w", dest.Type, err)
		}
		handlers = append(handlers, handler)
	}

	if len(handlers) == 0 {
		return fmt.Errorf("no valid destinations configured")
	}

	// Use slog-multi to fan out to all destinations
	multiHandler := slogmulti.Fanout(handlers...)
	slog.SetDefault(slog.New(multiHandler))

	return nil
}

// createHandler creates a slog.Handler based on destination config
func createHandler(dest DestinationConfig) (slog.Handler, error) {
	level, err := parseLevel(dest.Level)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch dest.Type {
	case "stdout":
		if dest.Format == "json" {
			return slog.NewJSONHandler(os.Stdout, opts), nil
		}
		return slog.NewTextHandler(os.Stdout, opts), nil

	case "stderr":
		if dest.Format == "json" {
			return slog.NewJSONHandler(os.Stderr, opts), nil
		}
		return slog.NewTextHandler(os.Stderr, opts), nil

	case "file":
		path := dest.Path
		if path == "" {
			path = config.GetLogsPath()
			// Ensure logs directory exists
			os.MkdirAll(path, 0755)
			path = filepath.Join(path, "app.log")
		}
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		if dest.Format == "json" {
			return slog.NewJSONHandler(file, opts), nil
		}
		return slog.NewTextHandler(file, opts), nil

	case "nats":
		// NATS logging requires a valid connection
		return nil, fmt.Errorf("nats destination requires nats connection - use InitMultiLoggerWithNATS")
	default:
		return nil, fmt.Errorf("unsupported destination type: %s", dest.Type)
	}
}

// createHandlerWithNATS creates a slog.Handler with NATS support
func createHandlerWithNATS(dest DestinationConfig, nc *gonats.Conn) (slog.Handler, error) {
	level, err := parseLevel(dest.Level)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch dest.Type {
	case "stdout":
		if dest.Format == "json" {
			return slog.NewJSONHandler(os.Stdout, opts), nil
		}
		return slog.NewTextHandler(os.Stdout, opts), nil

	case "stderr":
		if dest.Format == "json" {
			return slog.NewJSONHandler(os.Stderr, opts), nil
		}
		return slog.NewTextHandler(os.Stderr, opts), nil

	case "file":
		path := dest.Path
		if path == "" {
			path = config.GetLogsPath()
			// Ensure logs directory exists
			os.MkdirAll(path, 0755)
			path = filepath.Join(path, "app.log")
		}
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		if dest.Format == "json" {
			return slog.NewJSONHandler(file, opts), nil
		}
		return slog.NewTextHandler(file, opts), nil

	case "nats":
		if nc == nil {
			return nil, fmt.Errorf("nats connection required for nats destination")
		}
		subject := dest.Subject
		if subject == "" {
			subject = config.NATSLogStreamSubject
		}
		// Create encoded connection for JSON encoding
		ec, err := gonats.NewEncodedConn(nc, gonats.JSON_ENCODER)
		if err != nil {
			return nil, fmt.Errorf("failed to create encoded connection: %w", err)
		}
		return slognats.Option{
			Level:             opts.Level,
			EncodedConnection: ec,
			Subject:           subject,
		}.NewNATSHandler(), nil

	default:
		return nil, fmt.Errorf("unsupported destination type: %s", dest.Type)
	}
}

// parseLevel converts string level to slog.Level
func parseLevel(levelStr string) (slog.Level, error) {
	switch levelStr {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, nil
	}
}