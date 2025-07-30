package log

import (
	"log/slog"
	"sync"

	slogmulti "github.com/samber/slog-multi"
)

var (
	loggerMutex sync.RWMutex
	currentLogger *slog.Logger
)

// SetLogger sets the global logger to a new instance
// This allows runtime reconfiguration of logging
func SetLogger(logger *slog.Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	
	currentLogger = logger
	slog.SetDefault(logger)
}

// GetLogger returns the current logger instance
func GetLogger() *slog.Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	
	if currentLogger == nil {
		return slog.Default()
	}
	return currentLogger
}

// ReconfigureMultiLogger reconfigures logging at runtime
func ReconfigureMultiLogger(config MultiConfig) error {
	var handlers []slog.Handler

	for _, dest := range config.Destinations {
		handler, err := createHandler(dest)
		if err != nil {
			return err
		}
		handlers = append(handlers, handler)
	}

	if len(handlers) == 0 {
		return nil // Keep current logger
	}

	newLogger := slog.New(slogmulti.Fanout(handlers...))
	SetLogger(newLogger)
	
	return nil
}