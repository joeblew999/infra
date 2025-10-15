package controller

import (
	"fmt"
	"sort"
	"sync"

	proc "github.com/joeblew999/infra/core/pkg/shared/process"
)

// Port defines an exposed service port.
type Port struct {
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// ServiceSpec captures the contract runtime packages rely on when supervising
// a service.
type ServiceSpec struct {
	ID           string            `json:"id"`
	DisplayName  string            `json:"displayName"`
	Summary      string            `json:"summary"`
	Process      proc.Spec         `json:"process"`
	Ports        []Port            `json:"ports,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Profiles     []string          `json:"profiles,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Registry holds all registered services.
type Registry struct {
	mu       sync.RWMutex
	services map[string]ServiceSpec
}

// NewRegistry constructs an empty registry.
func NewRegistry() *Registry {
	return &Registry{services: make(map[string]ServiceSpec)}
}

// Register adds a service to the registry.
func (r *Registry) Register(spec ServiceSpec) error {
	if spec.ID == "" {
		return fmt.Errorf("service id is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.services[spec.ID]; exists {
		return fmt.Errorf("service %s already registered", spec.ID)
	}
	// defensive copies
	copySpec := spec
	if len(copySpec.Process.Env) > 0 {
		envCopy := make(map[string]string, len(copySpec.Process.Env))
		for k, v := range copySpec.Process.Env {
			envCopy[k] = v
		}
		copySpec.Process.Env = envCopy
	}
	if len(copySpec.Process.Args) > 0 {
		argsCopy := make([]string, len(copySpec.Process.Args))
		copy(argsCopy, copySpec.Process.Args)
		copySpec.Process.Args = argsCopy
	}
	if len(copySpec.Ports) > 0 {
		portsCopy := make([]Port, len(copySpec.Ports))
		copy(portsCopy, copySpec.Ports)
		copySpec.Ports = portsCopy
	}
	if len(copySpec.Dependencies) > 0 {
		depsCopy := make([]string, len(copySpec.Dependencies))
		copy(depsCopy, copySpec.Dependencies)
		copySpec.Dependencies = depsCopy
	}
	if len(copySpec.Profiles) > 0 {
		profilesCopy := make([]string, len(copySpec.Profiles))
		copy(profilesCopy, copySpec.Profiles)
		copySpec.Profiles = profilesCopy
	}
	if len(copySpec.Metadata) > 0 {
		metaCopy := make(map[string]string, len(copySpec.Metadata))
		for k, v := range copySpec.Metadata {
			metaCopy[k] = v
		}
		copySpec.Metadata = metaCopy
	}
	r.services[spec.ID] = copySpec
	return nil
}

// List returns registered services in deterministic order.
func (r *Registry) List() []ServiceSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ServiceSpec, 0, len(r.services))
	for _, spec := range r.services {
		out = append(out, spec)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Get returns a service by id.
func (r *Registry) Get(id string) (ServiceSpec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	spec, ok := r.services[id]
	return spec, ok
}
