package goreman

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"

	logpkg "github.com/joeblew999/infra/pkg/log"
)

// ProcessConfig defines configuration for a single process
type ProcessConfig struct {
	Name        string
	Command     string
	Args        []string
	Env         []string
	WorkingDir  string
	Port        int
	HealthCheck *HealthCheck
}

// HealthCheck defines health check configuration
type HealthCheck struct {
	URL     string
	Timeout time.Duration
}

// Process represents a running process
type Process struct {
	Config    *ProcessConfig
	Cmd       *exec.Cmd
	Status    string
	ExitCode  int
	PID       int
	StartTime time.Time
	mu        sync.RWMutex
}

// ProcessEvent represents an event related to a process
type ProcessEvent struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	PID       int       `json:"pid"`
	StartTime time.Time `json:"start_time"`
	ExitCode  int       `json:"exit_code"`
	Timestamp time.Time `json:"timestamp"`
}

// Manager manages multiple processes
type Manager struct {
	processes map[string]*Process
	groups    map[string][]string
	natsConn  *nats.Conn // NATS connection for publishing events
	mu        sync.RWMutex
}

// NewManager creates a new process manager
func NewManager() *Manager {
	return &Manager{
		processes: make(map[string]*Process),
		groups:    make(map[string][]string),
	}
}

// SetNATSConn sets the NATS connection for the manager
func (m *Manager) SetNATSConn(nc *nats.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.natsConn = nc
}

// publishEvent publishes a process event to NATS
func (m *Manager) publishEvent(event ProcessEvent) {
	m.mu.RLock()
	nc := m.natsConn
	m.mu.RUnlock()

	if nc == nil {
		return // NATS connection not set
	}

	subject := fmt.Sprintf("goreman.events.%s", event.Name)
	data, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal process event", "error", err)
		return
	}

	if err := nc.Publish(subject, data); err != nil {
		slog.Error("Failed to publish process event to NATS", "subject", subject, "error", err)
	}
}

// AddProcess adds a process to the manager (idempotent)
// If process already exists and is running, preserves the running state
func (m *Manager) AddProcess(name string, config *ProcessConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	config.Name = name

	// If process already exists and is running, preserve its state
	if existing, exists := m.processes[name]; exists && existing.Status == "running" {
		existing.Config = config // Update config but keep running
		return
	}

	// Create new process or replace stopped process
	m.processes[name] = &Process{
		Config: config,
		Status: "stopped",
	}
}

// AddGroup adds a process group
func (m *Manager) AddGroup(name string, processes []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.groups[name] = processes
}

// StartProcess starts a single process
func (m *Manager) StartProcess(name string) error {
	m.mu.RLock()
	proc, exists := m.processes[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("process %s not found", name)
	}

	proc.mu.Lock()
	defer proc.mu.Unlock()

	if proc.Status == "running" {
		return nil
	}

	cmd := exec.Command(proc.Config.Command, proc.Config.Args...)
	cmd.Env = append(os.Environ(), proc.Config.Env...)
	cmd.Dir = proc.Config.WorkingDir

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to attach stdout pipe for %s: %w", name, err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to attach stderr pipe for %s: %w", name, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process %s: %w", name, err)
	}

	proc.Cmd = cmd
	proc.PID = cmd.Process.Pid
	proc.Status = "running"
	proc.StartTime = time.Now()

	go streamProcessOutput(name, "stdout", stdoutPipe)
	go streamProcessOutput(name, "stderr", stderrPipe)

	m.publishEvent(ProcessEvent{
		Name:      proc.Config.Name,
		Status:    proc.Status,
		PID:       proc.PID,
		StartTime: proc.StartTime,
		Timestamp: time.Now(),
	})

	// Monitor the process in a goroutine
	go m.monitorProcess(proc)

	return nil
}

func streamProcessOutput(processName, stream string, pipe io.ReadCloser) {
	defer pipe.Close()

	scanner := bufio.NewScanner(pipe)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 256*1024)
	for scanner.Scan() {
		line := scanner.Text()
		logpkg.Info(line, "process", processName, "stream", stream)
	}
	if err := scanner.Err(); err != nil {
		logpkg.Warn("Process log stream error", "process", processName, "stream", stream, "error", err)
	}
}

// monitorProcess monitors a running process and updates its status
func (m *Manager) monitorProcess(proc *Process) {
	_ = proc.Cmd.Wait() // Wait for the process to exit

	proc.mu.Lock()
	defer proc.mu.Unlock()

	proc.Status = "stopped"
	if proc.Cmd.ProcessState != nil {
		proc.ExitCode = proc.Cmd.ProcessState.ExitCode()
	}

	m.publishEvent(ProcessEvent{
		Name:      proc.Config.Name,
		Status:    proc.Status,
		PID:       proc.PID,
		StartTime: proc.StartTime,
		ExitCode:  proc.ExitCode,
		Timestamp: time.Now(),
	})
}

// Start starts all processes
func (m *Manager) Start() error {
	m.mu.RLock()
	names := make([]string, 0, len(m.processes))
	for name := range m.processes {
		names = append(names, name)
	}
	m.mu.RUnlock()

	for _, name := range names {
		if err := m.StartProcess(name); err != nil {
			return fmt.Errorf("failed to start process %s: %w", name, err)
		}
	}

	return nil
}

// StopProcess stops a single process
func (m *Manager) StopProcess(name string) error {
	m.mu.RLock()
	proc, exists := m.processes[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("process %s not found", name)
	}

	proc.mu.Lock()
	defer proc.mu.Unlock()

	if proc.Status != "running" || proc.Cmd == nil || proc.Cmd.Process == nil {
		return nil
	}

	if err := proc.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop process %s: %w", name, err)
	}

	// Wait for process to exit
	done := make(chan error, 1)
	go func() {
		done <- proc.Cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			proc.ExitCode = 1
		}
		proc.Status = "stopped"
	case <-time.After(10 * time.Second):
		// Force kill after timeout
		proc.Cmd.Process.Kill()
		proc.Status = "killed"
		proc.ExitCode = -1
	}

	m.publishEvent(ProcessEvent{
		Name:      proc.Config.Name,
		Status:    proc.Status,
		PID:       proc.PID,
		StartTime: proc.StartTime,
		ExitCode:  proc.ExitCode,
		Timestamp: time.Now(),
	})

	return nil
}

// GetProcessPID returns the PID for a named process if it is known.
func (m *Manager) GetProcessPID(name string) (int, bool) {
	m.mu.RLock()
	proc, exists := m.processes[name]
	m.mu.RUnlock()

	if !exists {
		return 0, false
	}

	proc.mu.RLock()
	defer proc.mu.RUnlock()

	if proc.Cmd != nil && proc.Cmd.Process != nil {
		return proc.Cmd.Process.Pid, true
	}

	if proc.PID != 0 {
		return proc.PID, true
	}

	return 0, false
}

// Stop stops all processes gracefully
func (m *Manager) Stop() error {
	m.mu.RLock()
	names := make([]string, 0, len(m.processes))
	for name := range m.processes {
		names = append(names, name)
	}
	m.mu.RUnlock()

	// Stop processes in reverse order
	for i := len(names) - 1; i >= 0; i-- {
		_ = m.StopProcess(names[i])
	}

	return nil
}

// StartGroup starts all processes in a group
func (m *Manager) StartGroup(name string) error {
	m.mu.RLock()
	group, exists := m.groups[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("group %s not found", name)
	}

	for _, procName := range group {
		if err := m.StartProcess(procName); err != nil {
			return fmt.Errorf("failed to start process %s in group %s: %w", procName, name, err)
		}
	}

	return nil
}

// StopGroup stops all processes in a group
func (m *Manager) StopGroup(name string) error {
	m.mu.RLock()
	group, exists := m.groups[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("group %s not found", name)
	}

	// Stop processes in reverse order
	for i := len(group) - 1; i >= 0; i-- {
		_ = m.StopProcess(group[i])
	}

	return nil
}

// GetStatus returns the status of a process
func (m *Manager) GetStatus(name string) (string, error) {
	m.mu.RLock()
	proc, exists := m.processes[name]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("process %s not found", name)
	}

	proc.mu.RLock()
	defer proc.mu.RUnlock()

	return proc.Status, nil
}

// GetAllStatus returns status of all processes
func (m *Manager) GetAllStatus() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]string)
	for name, proc := range m.processes {
		proc.mu.RLock()
		status[name] = proc.Status
		proc.mu.RUnlock()
	}

	return status
}

// RestartProcess restarts a process
func (m *Manager) RestartProcess(name string) error {
	if err := m.StopProcess(name); err != nil {
		return err
	}

	return m.StartProcess(name)
}

// Restart restarts all processes
func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		return err
	}

	return m.Start()
}
