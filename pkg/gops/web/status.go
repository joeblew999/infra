package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/joeblew999/infra/pkg/gops"
	"github.com/joeblew999/infra/pkg/log"
)

// SystemStatus represents the system status for web display
type SystemStatus struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Disk    float64 `json:"disk"`
	Uptime  string  `json:"uptime"`
	Load    string  `json:"load"`
	Services []ServiceStatus `json:"services,omitempty"`
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Port   int    `json:"port,omitempty"`
}

// GetSystemStatus returns current system status for web display
func GetSystemStatus() (*SystemStatus, error) {
	// Get system metrics from gops
	metrics, err := gops.GetSystemMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get system metrics: %w", err)
	}
	
	status := &SystemStatus{
		CPU:    metrics.CPU.Percent,
		Memory: metrics.Memory.UsedPercent,
		Disk:   getDiskUsagePercent(metrics.Disk),
		Uptime: getSystemUptime(),
		Load:   getLoadAverage(),
	}
	
	return status, nil
}

// getDiskUsagePercent extracts root filesystem usage percentage
func getDiskUsagePercent(disk gops.Disk) float64 {
	if usage, exists := disk.MountPoints["/"]; exists {
		return usage.UsedPercent
	}
	return 0.0
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
			} else {
				return fmt.Sprintf("%dm", minutes)
			}
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

// HandleSystemStatus is an HTTP handler for system status endpoint
func HandleSystemStatus(w http.ResponseWriter, r *http.Request) {
	status, err := GetSystemStatus()
	if err != nil {
		log.Error("Error getting system status", "error", err)
		http.Error(w, "Failed to get system status", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Error("Error encoding system status", "error", err)
		http.Error(w, "Failed to encode system status", http.StatusInternalServerError)
		return
	}
}