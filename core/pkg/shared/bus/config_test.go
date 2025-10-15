package bus

import "testing"

func TestValidate(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg.MemoryStore = false
	cfg.StoreDir = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPersistent(t *testing.T) {
	cfg := DefaultConfig().Persistent("/tmp/bus")
	if cfg.MemoryStore {
		t.Fatal("expected persistent store")
	}
	if cfg.StoreDir != "/tmp/bus" {
		t.Fatal("unexpected store dir")
	}
}
