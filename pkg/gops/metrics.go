package gops

import (
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemMetrics represents the overall system metrics collected from a server.
type SystemMetrics struct {
	ServerID  string `json:"server_id"`
	Timestamp string `json:"timestamp"`
	CPU       CPU    `json:"cpu"`
	Memory    Memory `json:"memory"`
	Disk      Disk   `json:"disk"`
}

// CPU represents CPU usage metrics.
type CPU struct {
	Percent float64 `json:"percent"`
}

// Memory represents memory usage metrics.
type Memory struct {
	TotalMB   uint64  `json:"total_mb"`
	UsedMB    uint64  `json:"used_mb"`
	UsedPercent float64 `json:"used_percent"`
}

// Disk represents disk usage metrics, keyed by path (e.g., "/").
type Disk struct {
	// Key is the mount point (e.g., "/")
	MountPoints map[string]DiskUsage `json:"mount_points"`
}

// DiskUsage represents usage statistics for a single disk mount point.
type DiskUsage struct {
	TotalGB   uint64  `json:"total_gb"`
	UsedGB    uint64  `json:"used_gb"`
	UsedPercent float64 `json:"used_percent"`
}

// GetSystemMetrics collects and returns current system metrics.
func GetSystemMetrics() (SystemMetrics, error) {
	metrics := SystemMetrics{}

	// Server ID
	hostname, err := os.Hostname()
	if err != nil {
		return metrics, fmt.Errorf("failed to get hostname: %w", err)
	}
	metrics.ServerID = hostname

	// Timestamp
	metrics.Timestamp = time.Now().Format(time.RFC3339)

	// CPU
	cpuPercents, err := cpu.Percent(time.Second, false) // Average CPU over 1 second
	if err != nil {
		return metrics, fmt.Errorf("failed to get CPU percent: %w", err)
	}
	if len(cpuPercents) > 0 {
		metrics.CPU.Percent = cpuPercents[0]
	}

	// Memory
	vMem, err := mem.VirtualMemory()
	if err != nil {
		return metrics, fmt.Errorf("failed to get virtual memory info: %w", err)
	}
	metrics.Memory.TotalMB = vMem.Total / (1024 * 1024)
	metrics.Memory.UsedMB = vMem.Used / (1024 * 1024)
	metrics.Memory.UsedPercent = vMem.UsedPercent

	// Disk
	metrics.Disk.MountPoints = make(map[string]DiskUsage)
	diskUsage, err := disk.Usage("/") // Root filesystem
	if err != nil {
		return metrics, fmt.Errorf("failed to get disk usage for /: %w", err)
	}
	metrics.Disk.MountPoints["/"] = DiskUsage{
		TotalGB:   diskUsage.Total / (1024 * 1024 * 1024),
		UsedGB:    diskUsage.Used / (1024 * 1024 * 1024),
		UsedPercent: diskUsage.UsedPercent,
	}

	return metrics, nil
}