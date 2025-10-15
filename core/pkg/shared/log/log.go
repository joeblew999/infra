package log

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

// Fields represents structured logging attributes.
type Fields map[string]any

// Config controls creation of loggers.
type Config struct {
	Level   slog.Level
	Handler slog.Handler
}

var (
	once          sync.Once
	defaultLogger *slog.Logger
)

// Default returns the shared default logger. The logger is initialised lazily
// to ensure packages depending on shared/log do not pay the cost when a custom
// logger is provided from the runtime.
func Default() *slog.Logger {
	once.Do(func() {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
		defaultLogger = slog.New(handler)
	})
	return defaultLogger
}

// New constructs a logger from the provided config, falling back to the shared
// defaults when options are omitted.
func New(cfg Config) *slog.Logger {
	handler := cfg.Handler
	if handler == nil {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Level})
	}
	return slog.New(handler)
}

// With attaches structured fields to the provided logger.
func With(logger *slog.Logger, fields Fields) *slog.Logger {
	if logger == nil {
		logger = Default()
	}
	if len(fields) == 0 {
		return logger
	}
	args := make([]any, 0, len(fields))
	for k, v := range fields {
		args = append(args, slog.Any(k, v))
	}
	return logger.With(args...)
}

// Info logs a message with optional structured fields using the provided logger
// or the shared default when nil is passed.
func Info(ctx context.Context, logger *slog.Logger, msg string, fields Fields) {
	if logger == nil {
		logger = Default()
	}
	if len(fields) == 0 {
		logger.InfoContext(ctx, msg)
		return
	}
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	logger.InfoContext(ctx, msg, attrs...)
}

// Error logs an error message using the shared logger helpers.
func Error(ctx context.Context, logger *slog.Logger, err error, fields Fields) {
	merged := make(Fields, len(fields)+1)
	for k, v := range fields {
		merged[k] = v
	}
	merged["error"] = err
	Info(ctx, logger, "error", merged)
}
