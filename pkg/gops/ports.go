package gops

import (
	"fmt"
	"net"
	"time"
)

// IsPortAvailable checks if a given port is available for listening.
// It attempts to bind to the port and immediately closes the listener.
func IsPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

// WaitForPortAvailable waits for a port to become available within a timeout.
func WaitForPortAvailable(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if IsPortAvailable(port) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}
