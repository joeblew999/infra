package log

import (
	"context"
	"log/slog"

	sharedlog "github.com/joeblew999/infra/core/pkg/shared/log"
)

// Fields mirrors the shared log fields type.
type Fields = sharedlog.Fields

// Config mirrors the shared log configuration type.
type Config = sharedlog.Config

// Default returns the shared default logger.
func Default() *slog.Logger {
	return sharedlog.Default()
}

// New constructs a logger using shared defaults.
func New(cfg Config) *slog.Logger {
	return sharedlog.New(cfg)
}

// With attaches fields to a logger.
func With(logger *slog.Logger, fields Fields) *slog.Logger {
	return sharedlog.With(logger, fields)
}

// Info emits an informational log message.
func Info(ctx context.Context, logger *slog.Logger, msg string, fields Fields) {
	sharedlog.Info(ctx, logger, msg, fields)
}

// Error emits an error log message.
func Error(ctx context.Context, logger *slog.Logger, err error, fields Fields) {
	sharedlog.Error(ctx, logger, err, fields)
}
