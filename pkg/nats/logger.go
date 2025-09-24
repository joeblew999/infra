package nats

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/log"
)

const natsLogProcess = "nats"

type structuredNATSLogger struct {
	process string
}

func newStructuredNATSLogger() *structuredNATSLogger {
	return &structuredNATSLogger{process: natsLogProcess}
}

func (l *structuredNATSLogger) Noticef(format string, args ...any) {
	log.Info(l.formatMessage(format, args...), "process", l.process)
}

func (l *structuredNATSLogger) Warnf(format string, args ...any) {
	log.Warn(l.formatMessage(format, args...), "process", l.process)
}

func (l *structuredNATSLogger) Fatalf(format string, args ...any) {
	log.Error(l.formatMessage(format, args...), "process", l.process, "fatal", true)
}

func (l *structuredNATSLogger) Errorf(format string, args ...any) {
	log.Error(l.formatMessage(format, args...), "process", l.process)
}

func (l *structuredNATSLogger) Debugf(format string, args ...any) {
	log.Debug(l.formatMessage(format, args...), "process", l.process)
}

func (l *structuredNATSLogger) Tracef(format string, args ...any) {
	log.Debug(l.formatMessage(format, args...), "process", l.process, "trace", true)
}

func (l *structuredNATSLogger) formatMessage(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
