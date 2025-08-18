package collection

import (
	"fmt"
)

// Factory creates and configures collection system components

// NewDefaultCollector creates a collector with default configuration
func NewDefaultCollector() (Collector, error) {
	config := DefaultConfig()
	return NewCollector(config)
}

// NewCollectorWithConfig creates a collector with custom configuration
func NewCollectorWithConfig(config *Config) (Collector, error) {
	return NewCollector(config)
}

// NewDefaultPublisher creates a publisher with default configuration
func NewDefaultPublisher() (Publisher, error) {
	// TODO: Implement publisher factory
	return nil, fmt.Errorf("publisher not implemented yet")
}

// NewDefaultManagedDownloader creates a managed downloader with default configuration
func NewDefaultManagedDownloader() (ManagedDownloader, error) {
	// TODO: Implement managed downloader factory
	return nil, fmt.Errorf("managed downloader not implemented yet")
}