package xtemplate_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joeblew999/infra/pkg/xtemplate"
)

func TestServiceStartsAndServesTemplates(t *testing.T) {
	t.Setenv("ENVIRONMENT", "test")

	templateDir := t.TempDir()
	port := freePort(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := xtemplate.NewService(
		xtemplate.WithTemplateDir(templateDir),
		xtemplate.WithPort(port),
		xtemplate.WithMinify(false),
		xtemplate.WithWatchTemplates(false),
		xtemplate.WithDebug(false),
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- svc.Start(ctx)
	}()

	url := fmt.Sprintf("http://127.0.0.1:%s/", port)
	if err := waitForHTTP(url, 2*time.Minute); err != nil {
		cancel()
		t.Fatalf("xtemplate server did not become ready: %v", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		cancel()
		t.Fatalf("failed to GET homepage: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		cancel()
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		cancel()
		t.Fatalf("read body: %v", err)
	}

	if !strings.Contains(string(body), "XTemplate") {
		cancel()
		t.Fatalf("expected response to mention XTemplate, got: %s", string(body))
	}

	if _, err := os.Stat(filepath.Join(templateDir, "index.html")); err != nil {
		cancel()
		t.Fatalf("expected seed index.html to be created: %v", err)
	}

	cancel()

    select {
    case err := <-errCh:
        if err != nil {
            msg := err.Error()
            if !strings.Contains(msg, "context canceled") && !strings.Contains(msg, "signal: killed") {
                t.Fatalf("xtemplate service returned error: %v", err)
            }
        }
	case <-time.After(5 * time.Second):
		t.Fatalf("xtemplate service did not exit after cancel")
	}
}

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer l.Close()
	_, port, _ := net.SplitHostPort(l.Addr().String())
	return port
}

func waitForHTTP(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", url)
}
