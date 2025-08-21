package goreman

import (
	"fmt"
	"sync"
)

// ServiceFactory is a function that can register and start a service
type ServiceFactory func() error

// Global process registry - singleton pattern for automatic process management
var (
	globalManager   *Manager
	serviceRegistry = make(map[string]ServiceFactory)
	registryMutex   sync.RWMutex
	once            sync.Once
)

// GetManager returns the global process manager (singleton)
func GetManager() *Manager {
	once.Do(func() {
		globalManager = NewManager()
	})
	return globalManager
}

// Register adds a process to the global manager (idempotent)
// If the process already exists, it updates the configuration
func Register(name string, config *ProcessConfig) {
	manager := GetManager()
	manager.AddProcess(name, config)
}

// Start starts a process by name (idempotent)
// If already running, this is a no-op
func Start(name string) error {
	manager := GetManager()
	return manager.StartProcess(name)
}

// Stop stops a process by name (idempotent) 
// If not running, this is a no-op
func Stop(name string) error {
	manager := GetManager()
	return manager.StopProcess(name)
}

// Restart restarts a process by name (idempotent)
func Restart(name string) error {
	manager := GetManager()
	return manager.RestartProcess(name)
}

// IsRunning checks if a process is currently running
func IsRunning(name string) bool {
	manager := GetManager()
	status, err := manager.GetStatus(name)
	if err != nil {
		return false
	}
	return status == "running"
}

// RegisterAndStart is a convenience function for idempotent process management
// This is the main function packages should use:
// 1. Register the process configuration (updates if exists)  
// 2. Start the process if not already running
func RegisterAndStart(name string, config *ProcessConfig) error {
	Register(name, config)
	return Start(name)
}

// StopAll stops all registered processes gracefully
func StopAll() error {
	manager := GetManager()
	return manager.Stop()
}

// GetAllStatus returns status of all registered processes
func GetAllStatus() map[string]string {
	manager := GetManager()
	return manager.GetAllStatus()
}

// RegisterGroup adds a process group for coordinated startup/shutdown
func RegisterGroup(name string, processes []string) {
	manager := GetManager()
	manager.AddGroup(name, processes)
}

// StartGroup starts all processes in a group
func StartGroup(name string) error {
	manager := GetManager()
	return manager.StartGroup(name)
}

// StopGroup stops all processes in a group
func StopGroup(name string) error {
	manager := GetManager()
	return manager.StopGroup(name)
}

// RegisterService registers a service factory that can be invoked by name
// This allows decoupled service registration without direct package imports
func RegisterService(name string, factory ServiceFactory) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	serviceRegistry[name] = factory
}

// StartService starts a service by name using its registered factory
func StartService(name string) error {
	registryMutex.RLock()
	factory, exists := serviceRegistry[name]
	registryMutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("service %s not registered. Available services: %v", name, GetAvailableServices())
	}
	
	return factory()
}

// GetAvailableServices returns a list of registered service names
func GetAvailableServices() []string {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	
	services := make([]string, 0, len(serviceRegistry))
	for name := range serviceRegistry {
		services = append(services, name)
	}
	return services
}