package log

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

var ( 
	once sync.Once
)

// InitLogger initializes the global structured logger.
// It configures output to stdout and optionally to a file.
// logFilePath: path to the log file. If empty, no file logging.
// logLevel: minimum level to log (e.g., "debug", "info", "warn", "error").
// jsonFormat: true for JSON output, false for text output.
func InitLogger(logFilePath string, logLevel string, jsonFormat bool) {
	once.Do(func() {
		var level slog.Level
		switch logLevel {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			level = slog.LevelInfo // Default to Info
		}

		var writers []io.Writer
		writers = append(writers, os.Stdout)

		if logFilePath != "" {
			file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				slog.Default().Error("Failed to open log file", "path", logFilePath, "error", err)
			} else {
				writers = append(writers, file)
			}
		}

		multiWriter := io.MultiWriter(writers...)

		var handler slog.Handler
		if jsonFormat {
			handler = slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
				Level: level,
			})
		} else {
			handler = slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
				Level: level,
			})
		}

		slog.SetDefault(slog.New(handler))
	})
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	slog.Default().Debug(msg, args...)
}

// Info logs an info message.
func Info(msg string, args ...any) {
	slog.Default().Info(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	slog.Default().Warn(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	slog.Default().Error(msg, args...)
}
