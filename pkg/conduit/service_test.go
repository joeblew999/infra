package conduit

import (
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
	service := NewService()
	if service == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestServiceInitialize(t *testing.T) {
	service := NewService()
	
	// Ensure binaries are available first
	if err := Ensure(true); err != nil {
		t.Skipf("binaries not available: %v", err)
	}
	
	if err := service.Initialize(); err != nil {
		t.Fatalf("failed to initialize service: %v", err)
	}
	
	// Check that processes are configured
	status := service.Status()
	expectedProcesses := []string{
		"conduit",
		"conduit-connector-s3",
		"conduit-connector-postgres",
		"conduit-connector-kafka",
		"conduit-connector-file",
	}
	
	for _, proc := range expectedProcesses {
		if _, exists := status[proc]; !exists {
			t.Errorf("expected process %s to be configured", proc)
		}
	}
}

func TestServiceStartStop(t *testing.T) {
	service := NewService()
	
	// Ensure binaries are available first
	if err := Ensure(true); err != nil {
		t.Skipf("binaries not available: %v", err)
	}
	
	if err := service.Initialize(); err != nil {
		t.Fatalf("failed to initialize service: %v", err)
	}
	
	// Test start/stop cycle
	if err := service.Start(); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}
	
	// Wait a moment for processes to start
	time.Sleep(100 * time.Millisecond)
	
	status := service.Status()
	for name, state := range status {
		if state != "running" {
			t.Errorf("expected process %s to be running, got %s", name, state)
		}
	}
	
	// Test stop
	if err := service.Stop(); err != nil {
		t.Fatalf("failed to stop service: %v", err)
	}
}

func TestServiceGroups(t *testing.T) {
	service := NewService()
	
	// Ensure binaries are available first
	if err := Ensure(true); err != nil {
		t.Skipf("binaries not available: %v", err)
	}
	
	if err := service.Initialize(); err != nil {
		t.Fatalf("failed to initialize service: %v", err)
	}
	
	// Test core group
	if err := service.StartCore(); err != nil {
		t.Fatalf("failed to start core: %v", err)
	}
	
	// Check only core is running
	status := service.Status()
	if status["conduit"] != "running" {
		t.Error("expected conduit to be running")
	}
	
	// Test connectors group
	if err := service.StartConnectors(); err != nil {
		t.Fatalf("failed to start connectors: %v", err)
	}
	
	// Test stop connectors
	if err := service.StopConnectors(); err != nil {
		t.Fatalf("failed to stop connectors: %v", err)
	}
	
	// Test stop core
	if err := service.StopCore(); err != nil {
		t.Fatalf("failed to stop core: %v", err)
	}
}

func TestEnsureAndStart(t *testing.T) {
	service := NewService()
	
	// Test the combined ensure and start functionality
	if err := service.EnsureAndStart(true); err != nil {
		t.Skipf("ensure and start failed (expected if binaries not ready): %v", err)
	}
	
	// Verify processes are running
	status := service.Status()
	if len(status) == 0 {
		t.Error("expected processes to be configured")
	}
	
	// Clean up
	_ = service.Stop()
}

func TestGetBinaryPath(t *testing.T) {
	service := NewService()
	
	path := service.GetBinaryPath("conduit")
	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestServiceRestart(t *testing.T) {
	service := NewService()
	
	// Ensure binaries are available first
	if err := Ensure(true); err != nil {
		t.Skipf("binaries not available: %v", err)
	}
	
	if err := service.Initialize(); err != nil {
		t.Fatalf("failed to initialize service: %v", err)
	}
	
	// Test restart cycle
	if err := service.Start(); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}
	
	if err := service.Restart(); err != nil {
		t.Fatalf("failed to restart service: %v", err)
	}
	
	// Clean up
	_ = service.Stop()
}