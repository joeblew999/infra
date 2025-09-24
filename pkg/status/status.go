package status

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/joeblew999/infra/pkg/log"
	servicestate "github.com/joeblew999/infra/pkg/service/state"
)

// SystemStatus represents the current system status for web display
type SystemStatus struct {
	CPU       float64         `json:"cpu"`
	Memory    float64         `json:"memory"`
	Disk      float64         `json:"disk"`
	Uptime    string          `json:"uptime"`
	Load      string          `json:"load"`
	Services  []ServiceStatus `json:"services,omitempty"`
	Timestamp time.Time       `json:"timestamp"`

	Runtime RuntimeStats `json:"runtime"`

	HeapAllocBytes  uint64  `json:"heap_alloc_bytes"`
	HeapSysBytes    uint64  `json:"heap_sys_bytes"`
	StackInuseBytes uint64  `json:"stack_inuse_bytes"`
	NextGCBytes     uint64  `json:"next_gc_bytes"`
	LastPause       string  `json:"last_pause"`
	MemoryPercent   float64 `json:"memory_percent"`
	GoVersion       string  `json:"go_version"`
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name         string    `json:"name"`
	State        string    `json:"state"`
	Status       string    `json:"status"`
	Description  string    `json:"description"`
	Detail       string    `json:"detail,omitempty"`
	Port         int       `json:"port,omitempty"`
	Icon         string    `json:"icon,omitempty"`
	Healthy      bool      `json:"healthy"`
	Required     bool      `json:"required"`
	Level        string    `json:"level"`
	PID          int       `json:"pid,omitempty"`
	LastAction   string    `json:"last_action,omitempty"`
	LastActionAt time.Time `json:"last_action_at,omitempty"`
	Message      string    `json:"message,omitempty"`
	Ownership    string    `json:"ownership,omitempty"`
}

// GetCurrentStatus returns current system status for web display
func GetCurrentStatus() (*SystemStatus, error) {
	simpleMetrics, err := GetSimpleSystemMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get simple system metrics: %w", err)
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	status := &SystemStatus{
		CPU:       float64(simpleMetrics.Runtime.NumGoroutines),
		Memory:    float64(simpleMetrics.Runtime.MemAlloc),
		Disk:      0.0,
		Uptime:    getSystemUptime(),
		Load:      getLoadAverage(),
		Timestamp: time.Now(),
		Runtime:   simpleMetrics.Runtime,

		HeapAllocBytes:  mem.HeapAlloc,
		HeapSysBytes:    mem.HeapSys,
		StackInuseBytes: mem.StackInuse,
		NextGCBytes:     mem.NextGC,
		LastPause:       formatLastPause(mem),
		MemoryPercent:   computeMemoryPercent(mem),
		GoVersion:       runtime.Version(),
	}

	status.Services = probeServices()

	return status, nil
}

func computeMemoryPercent(mem runtime.MemStats) float64 {
	if mem.Sys == 0 {
		return 0
	}
	percent := (float64(mem.Alloc) / float64(mem.Sys)) * 100
	if percent < 0 {
		return 0
	}
	return percent
}

func formatLastPause(mem runtime.MemStats) string {
	if mem.NumGC == 0 {
		return "n/a"
	}
	idx := (mem.NumGC + 255) % 256
	pause := time.Duration(mem.PauseNs[idx])
	if pause == 0 {
		return "<1µs"
	}
	return pause.String()
}

func probeServices() []ServiceStatus {
	states := servicestate.Snapshot()
	services := make([]ServiceStatus, 0, len(states))

	for _, state := range states {
		healthy := state.Running
		ownership := state.Ownership
		if ownership == "" {
			ownership = "free"
		}

		canonicalState := state.State
		if canonicalState == "" {
			if state.Running {
				canonicalState = "running"
			} else {
				canonicalState = "stopped"
			}
		}

		statusLabel := labelForServiceState(canonicalState)
		level := levelForServiceState(canonicalState, state.Required, ownership)

		detailLines := []string{state.Description}
		if state.LastAction != "" {
			detailLines = append(detailLines, fmt.Sprintf("Last: %s", state.LastAction))
		}
		if state.Message != "" {
			detailLines = append(detailLines, state.Message)
		}
		if !state.UpdatedAt.IsZero() {
			detailLines = append(detailLines, fmt.Sprintf("Updated: %s", state.UpdatedAt.Format(time.RFC3339)))
		}
		detail := strings.Join(detailLines, "\n")

		switch {
		case state.Running:
			statusLabel = "Running"
			level = "ok"
		case ownership == "external":
			statusLabel = "Conflict"
			level = "error"
		case ownership == "infra":
			statusLabel = "In Use"
			level = "warn"
		case ownership == "this":
			statusLabel = "Stale"
			level = "warn"
		case state.Required:
			statusLabel = "Stopped"
			level = "warn"
		default:
			statusLabel = "Standby"
			level = "warn"
		}

		port := state.Port
		if port == 0 && state.PortLabel != "" {
			if parsed, err := strconv.Atoi(state.PortLabel); err == nil {
				port = parsed
			}
		}

		services = append(services, ServiceStatus{
			Name:         state.Name,
			State:        canonicalState,
			Status:       statusLabel,
			Description:  state.Description,
			Detail:       detail,
			Port:         port,
			Icon:         state.Icon,
			Healthy:      healthy,
			Required:     state.Required,
			Level:        level,
			PID:          state.PID,
			LastAction:   state.LastAction,
			LastActionAt: state.LastActionAt,
			Message:      state.Message,
			Ownership:    ownership,
		})
	}

	return services
}

func labelForServiceState(state string) string {
	switch state {
	case "running":
		return "Running"
	case "ready":
		return "Ready"
	case "pending":
		return "Pending"
	case "reclaimed":
		return "Reclaimed"
	case "blocked":
		return "Blocked"
	case "error":
		return "Error"
	case "stopped":
		return "Stopped"
	default:
		if state == "" {
			return "Unknown"
		}
		return strings.Title(state)
	}
}

func levelForServiceState(state string, required bool, ownership string) string {
	switch state {
	case "running", "ready":
		return "ok"
	case "blocked", "error":
		return "error"
	case "reclaimed", "pending":
		return "warn"
	case "stopped":
		if required {
			return "warn"
		}
		return "warn"
	default:
		if ownership == "external" {
			return "error"
		}
		if required {
			return "warn"
		}
		return "warn"
	}
}

func probePort(port int) bool {
	if port <= 0 {
		return false
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 250*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func atoiOrDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

// getSystemUptime returns system uptime as a formatted string
func getSystemUptime() string {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("uptime")
		output, err := cmd.Output()
		if err != nil {
			log.Debug("Failed to get uptime", "error", err)
			return "unknown"
		}

		// Parse uptime output: "16:23  up  1:23, 2 users, load averages: 1.23 1.45 1.67"
		uptimeStr := strings.TrimSpace(string(output))
		if parts := strings.Split(uptimeStr, "up "); len(parts) > 1 {
			if commaParts := strings.Split(parts[1], ","); len(commaParts) > 0 {
				return strings.TrimSpace(commaParts[0])
			}
		}
	} else if runtime.GOOS == "linux" {
		// Read from /proc/uptime
		cmd := exec.Command("cat", "/proc/uptime")
		output, err := cmd.Output()
		if err != nil {
			log.Debug("Failed to get uptime from /proc/uptime", "error", err)
			return "unknown"
		}

		uptimeStr := strings.TrimSpace(string(output))
		if parts := strings.Fields(uptimeStr); len(parts) > 0 {
			seconds, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return "unknown"
			}

			days := int(seconds) / 86400
			hours := (int(seconds) % 86400) / 3600
			minutes := (int(seconds) % 3600) / 60

			if days > 0 {
				return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
			} else if hours > 0 {
				return fmt.Sprintf("%dh %dm", hours, minutes)
			}
			return fmt.Sprintf("%dm", minutes)
		}
	}

	return "unknown"
}

// getLoadAverage returns system load average as a formatted string
func getLoadAverage() string {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("uptime")
		output, err := cmd.Output()
		if err != nil {
			log.Debug("Failed to get load average", "error", err)
			return "unknown"
		}

		// Parse load averages from uptime output
		uptimeStr := strings.TrimSpace(string(output))
		if parts := strings.Split(uptimeStr, "load averages: "); len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("cat", "/proc/loadavg")
		output, err := cmd.Output()
		if err != nil {
			log.Debug("Failed to get load average from /proc/loadavg", "error", err)
			return "unknown"
		}

		// Format: "0.52 0.58 0.59 1/178 12345"
		loadStr := strings.TrimSpace(string(output))
		if parts := strings.Fields(loadStr); len(parts) >= 3 {
			return fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2])
		}
	}

	return "unknown"
}

// FormatBytes formats byte values in human readable format
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
