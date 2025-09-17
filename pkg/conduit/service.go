// Package conduit provides process management for Conduit and its connectors
package conduit

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/goreman"
)

// Service manages Conduit processes using goreman
type Service struct {
	manager *goreman.Manager
	config  *ServiceConfig
	natsConn  *nats.Conn
}

// ServiceConfig defines service configuration
type ServiceConfig struct {
	WorkingDir string
	LogDir     string
	Timeout    time.Duration
}

// NewService creates a new Conduit service
func NewService(nc *nats.Conn) *Service {
	manager := goreman.NewManager()
	manager.SetNATSConn(nc)

	return &Service{
		manager: manager,
		config: &ServiceConfig{
			WorkingDir: ".",
			Timeout:    30 * time.Second,
		},
		natsConn: nc,
	}
}

// Initialize sets up the service with conduit processes
func (s *Service) Initialize() error {
	depPath := config.GetDepPath()
	
	// Add conduit core process
	s.manager.AddProcess("conduit", &goreman.ProcessConfig{
		Name:    "conduit",
		Command: filepath.Join(depPath, "conduit"),
		Args:    []string{},
		WorkingDir: s.config.WorkingDir,
	})
	
	// Add connector processes
	connectors := []string{
		"conduit-connector-s3",
		"conduit-connector-postgres",
		"conduit-connector-kafka",
		"conduit-connector-file",
	}
	
	for _, connector := range connectors {
		s.manager.AddProcess(connector, &goreman.ProcessConfig{
			Name:    connector,
			Command: filepath.Join(depPath, connector),
			Args:    []string{},
			WorkingDir: s.config.WorkingDir,
		})
	}
	
	// Create process groups
	s.manager.AddGroup("core", []string{"conduit"})
	s.manager.AddGroup("connectors", connectors)
	s.manager.AddGroup("all", append([]string{"conduit"}, connectors...))
	
	return nil
}

// Start starts all Conduit processes
func (s *Service) Start() error {
	return s.manager.Start()
}

// Stop stops all Conduit processes gracefully
func (s *Service) Stop() error {
	return s.manager.Stop()
}

// StartCore starts only the conduit core process
func (s *Service) StartCore() error {
	return s.manager.StartGroup("core")
}

// StartConnectors starts all connector processes
func (s *Service) StartConnectors() error {
	return s.manager.StartGroup("connectors")
}

// StopCore stops only the conduit core process
func (s *Service) StopCore() error {
	return s.manager.StopGroup("core")
}

// StopConnectors stops all connector processes
func (s *Service) StopConnectors() error {
	return s.manager.StopGroup("connectors")
}

// Restart restarts all processes
func (s *Service) Restart() error {
	return s.manager.Restart()
}

// Status returns the status of all processes
func (s *Service) Status() map[string]string {
	return s.manager.GetAllStatus()
}

// GetBinaryPath returns the path to a specific binary
func (s *Service) GetBinaryPath(name string) string {
	return filepath.Join(config.GetDepPath(), GetBinaryName(name))
}

// EnsureAndStart ensures binaries are available and starts the service
func (s *Service) EnsureAndStart(debug bool) error {
	// Ensure binaries are available
	if err := Ensure(debug); err != nil {
		return fmt.Errorf("failed to ensure binaries: %w", err)
	}
	
	// Initialize service
	if err := s.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}
	
	// Start all processes
	return s.Start()
}