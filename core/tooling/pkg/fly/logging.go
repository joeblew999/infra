package fly

import (
	"fmt"
	"log/slog"

	client "github.com/superfly/fly-go"

	sharedlog "github.com/joeblew999/infra/core/pkg/shared/log"
)

// NewLogger returns a fly-go compatible logger backed by the shared slog setup.
// Debug output is only emitted when verbose is true; messages are tagged with
// the supplied component to make tooling sources easier to filter.
func NewLogger(component string, verbose bool) client.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	base := sharedlog.New(sharedlog.Config{Level: level})
	if component != "" {
		base = base.With(slog.String("component", component))
	}
	return &loggerAdapter{verbose: verbose, logger: base}
}

type loggerAdapter struct {
	verbose bool
	logger  *slog.Logger
}

func (l *loggerAdapter) ensureLogger() {
	if l.logger == nil {
		l.logger = sharedlog.Default()
	}
}

func (l *loggerAdapter) Debug(v ...interface{}) {
	if !l.verbose {
		return
	}
	l.ensureLogger()
	l.logger.Debug(fmt.Sprint(v...))
}

func (l *loggerAdapter) Debugf(format string, v ...interface{}) {
	if !l.verbose {
		return
	}
	l.ensureLogger()
	l.logger.Debug(fmt.Sprintf(format, v...))
}
