package status

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/joeblew999/infra/pkg/log"
)

// RuntimeStats represents Go runtime statistics
type RuntimeStats struct {
	NumGoroutines int     `json:"num_goroutines"`
	NumCPU        int     `json:"num_cpu"`
	MemAlloc      uint64  `json:"mem_alloc_mb"`
	MemTotal      uint64  `json:"mem_total_mb"`
	MemSys        uint64  `json:"mem_sys_mb"`
	NumGC         uint32  `json:"num_gc"`
	GOOS          string  `json:"goos"`
	GOARCH        string  `json:"goarch"`
}

// GetRuntimeStats collects basic runtime statistics using only Go runtime (no CGO)
func GetRuntimeStats() RuntimeStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return RuntimeStats{
		NumGoroutines: runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
		MemAlloc:      m.Alloc / (1024 * 1024),
		MemTotal:      m.TotalAlloc / (1024 * 1024),
		MemSys:        m.Sys / (1024 * 1024),
		NumGC:         m.NumGC,
		GOOS:          runtime.GOOS,
		GOARCH:        runtime.GOARCH,
	}
}

// SimpleSystemMetrics represents basic system metrics without CGO dependencies
type SimpleSystemMetrics struct {
	ServerID  string       `json:"server_id"`
	Timestamp string       `json:"timestamp"`
	Runtime   RuntimeStats `json:"runtime"`
}

// GetSimpleSystemMetrics collects basic system metrics using only Go runtime (no CGO)
func GetSimpleSystemMetrics() (SimpleSystemMetrics, error) {
	metrics := SimpleSystemMetrics{
		Timestamp: time.Now().Format(time.RFC3339),
		Runtime:   GetRuntimeStats(),
	}

	// Server ID
	hostname, err := os.Hostname()
	if err != nil {
		return metrics, fmt.Errorf("failed to get hostname: %w", err)
	}
	metrics.ServerID = hostname

	return metrics, nil
}

// StartSimpleMetricCollection starts metric collection using only Go runtime stats (CGO-free)
func StartSimpleMetricCollection(ctx context.Context, nc *nats.Conn, interval time.Duration) {
	log.Info("Starting simple metric collection (CGO-free)", "interval", interval)

	hostname, err := os.Hostname()
	if err != nil {
		log.Error("Failed to get hostname for metric collection", "error", err)
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Simple metric collection stopped")
			return
		case <-ticker.C:
			metrics, err := GetSimpleSystemMetrics()
			if err != nil {
				log.Error("Failed to collect simple system metrics", "error", err)
				continue
			}

			jsonMetrics, err := json.Marshal(metrics)
			if err != nil {
				log.Error("Failed to marshal metrics to JSON", "error", err)
				continue
			}

			topic := fmt.Sprintf("metrics.server.%s", hostname)
			if err := nc.Publish(topic, jsonMetrics); err != nil {
				log.Error("Failed to publish metrics to NATS", "topic", topic, "error", err)
				continue
			}
			log.Debug("Simple metrics published", "topic", topic, "server_id", hostname)
		}
	}
}