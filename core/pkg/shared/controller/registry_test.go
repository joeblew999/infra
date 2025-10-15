package controller

import (
	"testing"

	proc "github.com/joeblew999/infra/core/pkg/shared/process"
)

func TestRegistryRegister(t *testing.T) {
	reg := NewRegistry()
	spec := ServiceSpec{
		ID: "demo",
		Process: proc.Spec{
			Command: "demo",
			Env:     map[string]string{"A": "1"},
			Args:    []string{"--flag"},
		},
		Ports: []Port{{Name: "http", Port: 8080, Protocol: "http"}},
	}
	if err := reg.Register(spec); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := reg.Register(spec); err == nil {
		t.Fatal("expected duplicate error")
	}
	items := reg.List()
	if len(items) != 1 {
		t.Fatalf("expected 1 service got %d", len(items))
	}
	svc, ok := reg.Get("demo")
	if !ok || svc.ID != "demo" {
		t.Fatal("missing service")
	}
	if &svc == &spec {
		t.Fatal("expected defensive copy")
	}
}
