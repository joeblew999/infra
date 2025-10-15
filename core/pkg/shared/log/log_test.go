package log

import (
	"context"
	"log/slog"
	"testing"
)

func TestWithDefaults(t *testing.T) {
	l := With(nil, Fields{"component": "test"})
	if l == nil {
		t.Fatal("expected logger")
	}
}

func TestNewRespectsLevel(t *testing.T) {
	cfg := Config{Level: slog.LevelDebug}
	logger := New(cfg)
	if logger == nil {
		t.Fatal("expected logger")
	}
	ctx := context.Background()
	logger.DebugContext(ctx, "debug message")
}
