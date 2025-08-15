package gops

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
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

// KillProcessByPort kills any process listening on the given port
func KillProcessByPort(port int) error {
	pid := GetProcessByPort(port)
	if pid == "" {
		return nil // No process found
	}

	return KillProcess(pid)
}

// GetProcessByPort returns the PID of the process listening on the given port
func GetProcessByPort(port int) string {
	portStr := strconv.Itoa(port)
	
	if runtime.GOOS == "windows" {
		// Windows: use netstat and tasklist
		cmd := exec.Command("cmd", "/c", fmt.Sprintf(`netstat -ano | findstr :%s | findstr LISTENING`, portStr))
		output, err := cmd.Output()
		if err != nil {
			return ""
		}
		
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "LISTENING") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					return parts[len(parts)-1] // PID is last column
				}
			}
		}
		return ""
	}

	// Unix-like systems: use lsof
	cmd := exec.Command("sh", "-c", fmt.Sprintf(`lsof -ti :%s 2>/dev/null | head -1`, portStr))
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(output))
}

// KillProcess kills a process by PID
func KillProcess(pid string) error {
	if pid == "" {
		return nil
	}

	if runtime.GOOS == "windows" {
		// Windows: use taskkill
		cmd := exec.Command("taskkill", "/F", "/PID", pid)
		return cmd.Run()
	}

	// Unix-like systems: use kill
	cmd := exec.Command("kill", "-9", pid)
	return cmd.Run()
}

// KillProcessByName kills processes by name pattern
func KillProcessByName(name string) error {
	if runtime.GOOS == "windows" {
		// Windows: use taskkill with filter
		cmd := exec.Command("taskkill", "/F", "/IM", name+"*")
		cmd.Run() // Ignore errors
		return nil
	}

	// Unix-like systems: use pkill
	cmd := exec.Command("pkill", "-9", "-f", name)
	cmd.Run() // Ignore errors
	return nil
}
