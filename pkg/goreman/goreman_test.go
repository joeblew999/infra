package goreman

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestAddProcess(t *testing.T) {
	manager := NewManager()
	
	config := &ProcessConfig{
		Name:    "test",
		Command: "echo",
		Args:    []string{"hello"},
	}
	
	manager.AddProcess("test", config)
	
	status, err := manager.GetStatus("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if status != "stopped" {
		t.Errorf("expected status 'stopped', got %s", status)
	}
}

func TestStartStopProcess(t *testing.T) {
	manager := NewManager()
	
	config := &ProcessConfig{
		Name:    "test",
		Command: "echo",
		Args:    []string{"hello"},
	}
	
	manager.AddProcess("test", config)
	
	// Start the process
	if err := manager.StartProcess("test"); err != nil {
		t.Fatalf("failed to start process: %v", err)
	}
	
	// Check status
	status, err := manager.GetStatus("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if status != "running" {
		t.Errorf("expected status 'running', got %s", status)
	}
	
	// Stop the process
	if err := manager.StopProcess("test"); err != nil {
		t.Fatalf("failed to stop process: %v", err)
	}
	
	// Check final status
	status, err = manager.GetStatus("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if status != "stopped" {
		t.Errorf("expected status 'stopped', got %s", status)
	}
}

func TestProcessGroups(t *testing.T) {
	manager := NewManager()
	
	// Add processes
	manager.AddProcess("proc1", &ProcessConfig{
		Name:    "proc1",
		Command: "echo",
		Args:    []string{"hello"},
	})
	
	manager.AddProcess("proc2", &ProcessConfig{
		Name:    "proc2",
		Command: "echo",
		Args:    []string{"world"},
	})
	
	// Add group
	manager.AddGroup("test-group", []string{"proc1", "proc2"})
	
	// Start group
	if err := manager.StartGroup("test-group"); err != nil {
		t.Fatalf("failed to start group: %v", err)
	}
	
	// Check statuses
	for _, name := range []string{"proc1", "proc2"} {
		status, err := manager.GetStatus(name)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "running" {
			t.Errorf("expected process %s to be running, got %s", name, status)
		}
	}
	
	// Stop group
	if err := manager.StopGroup("test-group"); err != nil {
		t.Fatalf("failed to stop group: %v", err)
	}
	
	// Check final statuses
	for _, name := range []string{"proc1", "proc2"} {
		status, err := manager.GetStatus(name)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "stopped" {
			t.Errorf("expected process %s to be stopped, got %s", name, status)
		}
	}
}

func TestGetAllStatus(t *testing.T) {
	manager := NewManager()
	
	manager.AddProcess("test1", &ProcessConfig{
		Name:    "test1",
		Command: "echo",
		Args:    []string{"hello"},
	})
	
	manager.AddProcess("test2", &ProcessConfig{
		Name:    "test2",
		Command: "echo",
		Args:    []string{"world"},
	})
	
	status := manager.GetAllStatus()
	if len(status) != 2 {
		t.Errorf("expected 2 processes, got %d", len(status))
	}
	
	if status["test1"] != "stopped" {
		t.Errorf("expected test1 to be stopped, got %s", status["test1"])
	}
	
	if status["test2"] != "stopped" {
		t.Errorf("expected test2 to be stopped, got %s", status["test2"])
	}
}

func TestRestartProcess(t *testing.T) {
	manager := NewManager()
	
	config := &ProcessConfig{
		Name:    "test",
		Command: "echo",
		Args:    []string{"hello"},
	}
	
	manager.AddProcess("test", config)
	
	// Start and then restart
	if err := manager.StartProcess("test"); err != nil {
		t.Fatalf("failed to start process: %v", err)
	}
	
	if err := manager.RestartProcess("test"); err != nil {
		t.Fatalf("failed to restart process: %v", err)
	}
	
	status, err := manager.GetStatus("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if status != "running" {
		t.Errorf("expected status 'running' after restart, got %s", status)
	}
	
	// Clean up
	_ = manager.StopProcess("test")
}

func TestInvalidProcess(t *testing.T) {
	manager := NewManager()
	
	_, err := manager.GetStatus("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent process")
	}
	
	err = manager.StartProcess("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent process")
	}
	
	err = manager.StopProcess("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent process")
	}
}

func TestInvalidGroup(t *testing.T) {
	manager := NewManager()
	
	err := manager.StartGroup("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent group")
	}
	
	err = manager.StopGroup("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent group")
	}
}