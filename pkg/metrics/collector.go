package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/joeblew999/infra/pkg/log"
)

// SystemMetrics holds all system performance metrics
type SystemMetrics struct {
	Timestamp    time.Time `json:"timestamp"`
	CPU          CPUMetrics `json:"cpu"`
	Memory       MemoryMetrics `json:"memory"`
	Goroutines   int `json:"goroutines"`
	GCStats      GCMetrics `json:"gc_stats"`
	Uptime       time.Duration `json:"uptime"`
}

type CPUMetrics struct {
	NumCPU       int     `json:"num_cpu"`
	NumGoroutine int     `json:"num_goroutine"`
	LoadAverage  float64 `json:"load_average"`
}

type MemoryMetrics struct {
	Alloc        uint64  `json:"alloc"`
	TotalAlloc   uint64  `json:"total_alloc"`
	Sys          uint64  `json:"sys"`
	HeapAlloc    uint64  `json:"heap_alloc"`
	HeapSys      uint64  `json:"heap_sys"`
	HeapIdle     uint64  `json:"heap_idle"`
	HeapInuse    uint64  `json:"heap_inuse"`
	StackInuse   uint64  `json:"stack_inuse"`
	StackSys     uint64  `json:"stack_sys"`
	Usage        float64 `json:"usage_percent"`
}

type GCMetrics struct {
	NumGC        uint32        `json:"num_gc"`
	PauseTotal   time.Duration `json:"pause_total"`
	LastPause    time.Duration `json:"last_pause"`
	NextGC       uint64        `json:"next_gc"`
}

// Collector manages metrics collection
type Collector struct {
	mu          sync.RWMutex
	startTime   time.Time
	metrics     []SystemMetrics
	maxHistory  int
	subscribers []chan<- SystemMetrics
}

var globalCollector *Collector
var once sync.Once

// GetCollector returns the singleton metrics collector
func GetCollector() *Collector {
	once.Do(func() {
		globalCollector = &Collector{
			startTime:   time.Now(),
			maxHistory:  100, // Keep last 100 data points
			subscribers: make([]chan<- SystemMetrics, 0),
		}
	})
	return globalCollector
}

// Start begins collecting metrics at the specified interval
func (c *Collector) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Info("Started metrics collection", "interval", interval)

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping metrics collection")
			return
		case <-ticker.C:
			metrics := c.collectMetrics()
			c.storeMetrics(metrics)
			c.notifySubscribers(metrics)
		}
	}
}

// collectMetrics gathers current system metrics
func (c *Collector) collectMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	now := time.Now()
	uptime := now.Sub(c.startTime)

	return SystemMetrics{
		Timestamp:  now,
		CPU: CPUMetrics{
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
		},
		Memory: MemoryMetrics{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			HeapAlloc:  m.HeapAlloc,
			HeapSys:    m.HeapSys,
			HeapIdle:   m.HeapIdle,
			HeapInuse:  m.HeapInuse,
			StackInuse: m.StackInuse,
			StackSys:   m.StackSys,
			Usage:      float64(m.HeapInuse) / float64(m.HeapSys) * 100,
		},
		Goroutines: runtime.NumGoroutine(),
		GCStats: GCMetrics{
			NumGC:      m.NumGC,
			PauseTotal: time.Duration(m.PauseTotalNs),
			LastPause:  time.Duration(m.PauseNs[(m.NumGC+255)%256]),
			NextGC:     m.NextGC,
		},
		Uptime: uptime,
	}
}

// storeMetrics adds metrics to the history buffer
func (c *Collector) storeMetrics(metrics SystemMetrics) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = append(c.metrics, metrics)
	
	// Trim to max history
	if len(c.metrics) > c.maxHistory {
		c.metrics = c.metrics[1:]
	}
}

// notifySubscribers sends metrics to all subscribers
func (c *Collector) notifySubscribers(metrics SystemMetrics) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, ch := range c.subscribers {
		select {
		case ch <- metrics:
		default:
			// Skip if channel is full
		}
	}
}

// GetLatest returns the most recent metrics
func (c *Collector) GetLatest() *SystemMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.metrics) == 0 {
		return nil
	}

	latest := c.metrics[len(c.metrics)-1]
	return &latest
}

// GetHistory returns all stored metrics
func (c *Collector) GetHistory() []SystemMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid race conditions
	history := make([]SystemMetrics, len(c.metrics))
	copy(history, c.metrics)
	return history
}

// Subscribe returns a channel that receives real-time metrics updates
func (c *Collector) Subscribe() <-chan SystemMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan SystemMetrics, 10) // Buffered channel
	c.subscribers = append(c.subscribers, ch)
	return ch
}

// GetMetricsJSON returns metrics as JSON string
func (c *Collector) GetMetricsJSON() (string, error) {
	latest := c.GetLatest()
	if latest == nil {
		return "{}", nil
	}

	data, err := json.Marshal(latest)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetHistoryJSON returns metrics history as JSON string
func (c *Collector) GetHistoryJSON() (string, error) {
	history := c.GetHistory()
	
	data, err := json.Marshal(history)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FormatBytes converts bytes to human readable format
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