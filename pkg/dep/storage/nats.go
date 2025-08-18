package storage

import (
	"fmt"
)

// NATS provides unified storage for NATS Object Store
// Handles both uploads and downloads with platform-specific naming
// TODO: Implement actual NATS Object Store integration
// For now: stubs for future implementation

type NATS struct {
	// Object store client (future)
	// bucket string
	// conn    *nats.Conn
	// js      nats.JetStreamContext
}

// Upload stores a binary in NATS Object Store
func (n *NATS) Upload(binaryName, version, platform, sourcePath string) error {
	// TODO: Implement actual NATS Object Store upload
	// Naming convention: binaries/{name}/{version}/{platform}
	fmt.Printf("[NATS] Would upload %s v%s for %s from %s\n", binaryName, version, platform, sourcePath)
	return nil
}

// Download retrieves a binary from NATS Object Store
func (n *NATS) Download(binaryName, version, platform, destPath string) error {
	// TODO: Implement actual NATS Object Store download
	// Naming convention: binaries/{name}/{version}/{platform}
	fmt.Printf("[NATS] Would download %s v%s for %s to %s\n", binaryName, version, platform, destPath)
	return nil
}

// Exists checks if a binary exists in NATS Object Store
func (n *NATS) Exists(binaryName, version, platform string) bool {
	// TODO: Implement actual NATS Object Store existence check
	return false
}

// List returns available binaries for a platform
func (n *NATS) List(platform string) ([]string, error) {
	// TODO: Implement actual NATS Object Store listing
	return []string{}, nil
}