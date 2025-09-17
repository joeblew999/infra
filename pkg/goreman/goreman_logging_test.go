package goreman

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	logpkg "github.com/joeblew999/infra/pkg/log"
)

func TestStreamProcessOutput(t *testing.T) {
	// Capture log output
	buf := new(bytes.Buffer)
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo})

	orig := logpkg.GetLogger()
	logpkg.SetLogger(slog.New(handler))
	t.Cleanup(func() {
		if orig != nil {
			logpkg.SetLogger(orig)
		}
	})

	m := NewManager()

	cfg := &ProcessConfig{
		Command: os.Args[0],
		Args:    []string{"-test.run=TestGoremanLogHelper", "--"},
		Env:     append(os.Environ(), "GO_TEST_GOREMAN_HELPER=1"),
	}

	m.AddProcess("helper", cfg)

	if err := m.StartProcess("helper"); err != nil {
		t.Fatalf("failed to start helper process: %v", err)
	}

	// Allow some time for the process to run and logs to flush
	requireLog := func(sub string) {
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if strings.Contains(buf.String(), sub) {
				return
			}
			time.Sleep(25 * time.Millisecond)
		}
		t.Fatalf("log output did not contain %q; captured: %s", sub, buf.String())
	}

	requireLog(`"process":"helper"`)
	requireLog(`"stream":"stdout"`)
	requireLog(`"stream":"stderr"`)
}

func TestGoremanLogHelper(t *testing.T) {
	if os.Getenv("GO_TEST_GOREMAN_HELPER") != "1" {
		return
	}
	// Emit both stdout and stderr before exiting
	os.Stdout.WriteString("stdout line from helper\n")
	os.Stderr.WriteString("stderr line from helper\n")
	os.Exit(0)
}
