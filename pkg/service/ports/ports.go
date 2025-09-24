package ports

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Ownership describes who currently holds a port.
type Ownership int

const (
	OwnershipFree Ownership = iota
	OwnershipThisService
	OwnershipOtherInfra
	OwnershipExternal
)

// Probe captures details about the process holding a port.
type Probe struct {
	Port      int
	PID       string
	Command   string
	Ownership Ownership
}

// ParsePort converts a port string to an integer, returning 0 on failure.
func ParsePort(portStr string) int {
	port, err := strconv.Atoi(strings.TrimSpace(portStr))
	if err != nil {
		return 0
	}
	return port
}

// IsAvailable checks if a given port is available for listening.
func IsAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

// WaitAvailable waits for a port to become available within a timeout.
func WaitAvailable(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if IsAvailable(port) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// Inspect inspects the port and classifies ownership against the expected PID (if any).
func Inspect(port int, expectedPID string) (*Probe, error) {
	probe := &Probe{Port: port}

	pid := GetProcessByPort(port)
	if pid == "" {
		probe.Ownership = OwnershipFree
		return probe, nil
	}
	probe.PID = pid

	if cmd, err := GetCommandForPID(pid); err == nil {
		probe.Command = cmd
	}

	if expectedPID != "" && pid == expectedPID {
		probe.Ownership = OwnershipThisService
		return probe, nil
	}

	if strings.Contains(probe.Command, "infra") {
		probe.Ownership = OwnershipOtherInfra
	} else {
		probe.Ownership = OwnershipExternal
	}

	return probe, nil
}

// FormatConflictMessage renders a human-readable explanation for port conflicts.
func FormatConflictMessage(service string, probe *Probe) string {
	if probe == nil || probe.PID == "" {
		return ""
	}

	msg := fmt.Sprintf("%s port %d is in use by PID %s", service, probe.Port, probe.PID)
	if probe.Command != "" {
		msg += fmt.Sprintf(" (%s)", shortenCommand(probe.Command))
	}

	switch probe.Ownership {
	case OwnershipThisService:
		msg += ". This looks like a stale infra-managed process — try 'infra shutdown' or rerun the command to reclaim it."
	case OwnershipOtherInfra:
		msg += ". Another infra session is using this port; run 'infra shutdown' in that session or stop the PID manually."
	case OwnershipExternal:
		msg += ". Stop that process or update infra's configuration to use a different port."
	}

	return msg
}

// GetProcessByPort returns the PID of the process listening on the given port.
func GetProcessByPort(port int) string {
	portStr := strconv.Itoa(port)

	if runtime.GOOS == "windows" {
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
					return parts[len(parts)-1]
				}
			}
		}
		return ""
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf(`lsof -ti :%s -sTCP:LISTEN 2>/dev/null | head -1`, portStr))
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// GetCommandForPID returns the command line for a given PID, if available.
func GetCommandForPID(pid string) (string, error) {
	if pid == "" {
		return "", fmt.Errorf("pid is empty")
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %s", pid), "/FO", "CSV", "/NH")
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		line := strings.TrimSpace(string(output))
		if line == "" {
			return "", fmt.Errorf("no command for pid %s", pid)
		}
		fields := strings.Split(line, ",")
		if len(fields) == 0 {
			return "", fmt.Errorf("unexpected tasklist output for pid %s", pid)
		}
		return strings.Trim(fields[0], "\""), nil
	}

	cmd := exec.Command("ps", "-o", "command=", "-p", pid)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

var (
	wdOnce   sync.Once
	wdPath   string
	homeOnce sync.Once
	homePath string
)

func workingDir() string {
	wdOnce.Do(func() {
		if cwd, err := os.Getwd(); err == nil {
			wdPath = filepath.Clean(cwd)
		}
	})
	return wdPath
}

func homeDir() string {
	homeOnce.Do(func() {
		if home, err := os.UserHomeDir(); err == nil {
			homePath = filepath.Clean(home)
		}
	})
	return homePath
}

func shortenCommand(command string) string {
	if command == "" {
		return command
	}
	tokens := strings.Fields(command)
	if len(tokens) == 0 {
		return command
	}
	for i, token := range tokens {
		trimmed := strings.Trim(token, `"'`)
		if !strings.ContainsAny(trimmed, "/\\") {
			continue
		}
		short := shortenPath(trimmed)
		if short == trimmed {
			continue
		}
		tokens[i] = strings.Replace(token, trimmed, short, 1)
	}
	return strings.Join(tokens, " ")
}

func shortenPath(path string) string {
	if path == "" || strings.Contains(path, "://") {
		return path
	}
	clean := filepath.Clean(path)
	if wd := workingDir(); wd != "" {
		if rel, err := filepath.Rel(wd, clean); err == nil && !strings.HasPrefix(rel, "..") {
			clean = filepath.Join(".", rel)
		}
	}
	clean = filepath.ToSlash(clean)
	if home := homeDir(); home != "" {
		homeSlash := filepath.ToSlash(home)
		clean = strings.ReplaceAll(clean, homeSlash, "~")
	}
	if idx := strings.Index(clean, "go-build/"); idx != -1 {
		clean = "go-build/" + filepath.ToSlash(filepath.Base(clean))
	}
	return clean
}

// KillProcessByPort kills any process listening on the given port.
func KillProcessByPort(port int) error {
	pid := GetProcessByPort(port)
	if pid == "" {
		return nil
	}
	return KillProcess(pid)
}

// KillProcess terminates a process by PID.
func KillProcess(pid string) error {
	if pid == "" {
		return nil
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/PID", pid)
		return cmd.Run()
	}

	cmd := exec.Command("kill", "-9", pid)
	return cmd.Run()
}

// KillProcessByName kills processes by name pattern.
func KillProcessByName(name string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/IM", name+"*")
		cmd.Run()
		return nil
	}

	cmd := exec.Command("pkill", "-9", "-x", name)
	cmd.Run()
	return nil
}

// KillInfraGoRunProcess specifically kills "go run ." processes in the infra directory.
func KillInfraGoRunProcess() error {
	if runtime.GOOS == "windows" {
		return nil
	}

	pwd := exec.Command("pwd")
	pwdOutput, err := pwd.Output()
	if err != nil {
		return err
	}
	currentDir := strings.TrimSpace(string(pwdOutput))

	cmd := exec.Command("sh", "-c", fmt.Sprintf(`
		pgrep -af "go.*run" | while read pid cmdline; do
			if echo "$cmdline" | grep -q "go.*run.*\\."; then
				pwdx_output=$(pwdx $pid 2>/dev/null | cut -d: -f2 | tr -d ' ' 2>/dev/null)
				if [ "$pwdx_output" = "%s" ]; then
					kill -9 $pid 2>/dev/null
				fi
			fi
		done
	`, currentDir))
	cmd.Run()
	return nil
}
