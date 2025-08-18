package platform

import (
	"context"
	"runtime"
	"sync"
)

// Simulator provides platform simulation for cross-platform binary collection
type Simulator struct {
	mu               sync.RWMutex
	overrideOS       string
	overrideArch     string
	simulationActive bool
}

// Global simulator instance
var globalSimulator = &Simulator{}

// NewSimulator creates a new platform simulator
func NewSimulator() *Simulator {
	return &Simulator{}
}

// SimulatePlatform sets the platform simulation for the current goroutine/context
func (s *Simulator) SimulatePlatform(os, arch string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.overrideOS = os
	s.overrideArch = arch
	s.simulationActive = true
}

// ClearSimulation removes platform simulation
func (s *Simulator) ClearSimulation() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.overrideOS = ""
	s.overrideArch = ""
	s.simulationActive = false
}

// GetOS returns the current OS (real or simulated)
func (s *Simulator) GetOS() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.simulationActive && s.overrideOS != "" {
		return s.overrideOS
	}
	return runtime.GOOS
}

// GetArch returns the current architecture (real or simulated)
func (s *Simulator) GetArch() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.simulationActive && s.overrideArch != "" {
		return s.overrideArch
	}
	return runtime.GOARCH
}

// IsSimulating returns true if platform simulation is active
func (s *Simulator) IsSimulating() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.simulationActive
}

// GetSimulatedPlatform returns the simulated platform
func (s *Simulator) GetSimulatedPlatform() (string, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.overrideOS, s.overrideArch
}

// WithPlatform executes a function with platform simulation
func (s *Simulator) WithPlatform(os, arch string, fn func() error) error {
	s.SimulatePlatform(os, arch)
	defer s.ClearSimulation()
	return fn()
}

// Global functions for easy access

// SimulatePlatform sets global platform simulation
func SimulatePlatform(os, arch string) {
	globalSimulator.SimulatePlatform(os, arch)
}

// ClearSimulation removes global platform simulation
func ClearSimulation() {
	globalSimulator.ClearSimulation()
}

// GetOS returns the current OS (real or simulated)
func GetOS() string {
	return globalSimulator.GetOS()
}

// GetArch returns the current architecture (real or simulated)
func GetArch() string {
	return globalSimulator.GetArch()
}

// IsSimulating returns true if platform simulation is active
func IsSimulating() bool {
	return globalSimulator.IsSimulating()
}

// GetSimulatedPlatform returns the simulated platform
func GetSimulatedPlatform() (string, string) {
	return globalSimulator.GetSimulatedPlatform()
}

// WithPlatform executes a function with platform simulation
func WithPlatform(os, arch string, fn func() error) error {
	return globalSimulator.WithPlatform(os, arch, fn)
}

// Context-based simulation for goroutine safety

type platformKey struct{}

// PlatformContext holds platform simulation data
type PlatformContext struct {
	OS   string
	Arch string
}

// WithPlatformContext adds platform simulation to context
func WithPlatformContext(ctx context.Context, os, arch string) context.Context {
	return context.WithValue(ctx, platformKey{}, &PlatformContext{
		OS:   os,
		Arch: arch,
	})
}

// FromContext gets platform simulation from context
func FromContext(ctx context.Context) (string, string, bool) {
	if pc, ok := ctx.Value(platformKey{}).(*PlatformContext); ok {
		return pc.OS, pc.Arch, true
	}
	return "", "", false
}

// GetOSFromContext returns OS from context or real OS
func GetOSFromContext(ctx context.Context) string {
	if os, _, ok := FromContext(ctx); ok {
		return os
	}
	return runtime.GOOS
}

// GetArchFromContext returns architecture from context or real architecture
func GetArchFromContext(ctx context.Context) string {
	if _, arch, ok := FromContext(ctx); ok {
		return arch
	}
	return runtime.GOARCH
}

// PlatformInfo represents platform information
type PlatformInfo struct {
	OS           string
	Arch         string
	Platform     string // "os-arch" format
	IsSimulated  bool
	RealOS       string
	RealArch     string
}

// GetPlatformInfo returns comprehensive platform information
func GetPlatformInfo() *PlatformInfo {
	realOS := runtime.GOOS
	realArch := runtime.GOARCH
	currentOS := GetOS()
	currentArch := GetArch()
	
	return &PlatformInfo{
		OS:          currentOS,
		Arch:        currentArch,
		Platform:    currentOS + "-" + currentArch,
		IsSimulated: IsSimulating(),
		RealOS:      realOS,
		RealArch:    realArch,
	}
}

// GetPlatformInfoFromContext returns platform information from context
func GetPlatformInfoFromContext(ctx context.Context) *PlatformInfo {
	realOS := runtime.GOOS
	realArch := runtime.GOARCH
	currentOS := GetOSFromContext(ctx)
	currentArch := GetArchFromContext(ctx)
	_, _, isSimulated := FromContext(ctx)
	
	return &PlatformInfo{
		OS:          currentOS,
		Arch:        currentArch,
		Platform:    currentOS + "-" + currentArch,
		IsSimulated: isSimulated,
		RealOS:      realOS,
		RealArch:    realArch,
	}
}