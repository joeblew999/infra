package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	cfg := Load()
	if cfg.Paths.AppRoot == "" {
		t.Fatal("expected AppRoot to be populated")
	}
	if cfg.Paths.Dep == "" || cfg.Paths.Bin == "" {
		t.Fatal("expected Dep and Bin paths to be populated")
	}
	if cfg.Environment == "" {
		t.Fatal("expected Environment to be populated")
	}
}
