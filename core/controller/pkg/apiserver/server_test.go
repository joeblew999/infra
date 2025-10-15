package apiserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleListServices(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.yaml")
	if err := os.WriteFile(specPath, []byte("services: []\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	server, err := New(specPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/services", nil)
	w := httptest.NewRecorder()
	server.handleListServices(w, req)

	if status := w.Result().StatusCode; status != http.StatusOK {
		t.Fatalf("unexpected status %d", status)
	}
	if err := server.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestHandleUpdateService(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.yaml")
	if err := os.WriteFile(specPath, []byte("services: []\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	server, err := New(specPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	body := strings.NewReader(`{"service":{"id":"worker","scale":{"strategy":"local","regions":[{"name":"iad","min":1,"desired":1,"max":2}]}}}`)
	req := httptest.NewRequest(http.MethodPatch, "/v1/services/update", body)
	w := httptest.NewRecorder()
	server.handleUpdateService(w, req)
	if status := w.Result().StatusCode; status != http.StatusCreated {
		t.Fatalf("unexpected status %d", status)
	}
}
