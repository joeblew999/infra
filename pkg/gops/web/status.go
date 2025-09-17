package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	datastarlib "github.com/starfederation/datastar-go/datastar"

	"github.com/joeblew999/infra/pkg/config"
	infodatastar "github.com/joeblew999/infra/pkg/datastar"
	"github.com/joeblew999/infra/pkg/gops"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/metrics"
	"github.com/joeblew999/infra/web/templates"
)

//go:embed templates/status-page.html
var statusPageTemplate string

const statusStreamInterval = 4 * time.Second

// SystemStatus represents the system status for web display
// Memory is expressed in MB, CPU is the number of active goroutines.
type SystemStatus struct {
	CPU       float64         `json:"cpu"`
	Memory    float64         `json:"memory"`
	Disk      float64         `json:"disk"`
	Uptime    string          `json:"uptime"`
	Load      string          `json:"load"`
	Services  []ServiceStatus `json:"services,omitempty"`
	Timestamp time.Time       `json:"timestamp"`

	Runtime gops.RuntimeStats `json:"runtime"`

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
	Name     string `json:"name"`
	Status   string `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Port     int    `json:"port,omitempty"`
	Icon     string `json:"icon,omitempty"`
	Healthy  bool   `json:"healthy"`
	Required bool   `json:"required"`
	Level    string `json:"level"`
}

// GopsWebService provides web interface for system monitoring
type GopsWebService struct{}

// NewGopsWebService creates a new gops web service
func NewGopsWebService() *GopsWebService {
	return &GopsWebService{}
}

// RegisterRoutes mounts all gops routes on the provided router
func (s *GopsWebService) RegisterRoutes(r chi.Router) {
	r.Get("/", HandleStatusPage)
	r.Get("/api/stream", HandleStatusStream)
	r.Get("/api/system-status", HandleSystemStatus)
}

// HandleStatusPage renders the runtime status dashboard with the shared base layout.
func HandleStatusPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("status-page").Parse(statusPageTemplate)
	if err != nil {
		log.Error("error parsing status page template", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var content strings.Builder
	data := struct {
		BasePath string
	}{BasePath: config.StatusHTTPPath}

	if err := tmpl.Execute(&content, data); err != nil {
		log.Error("error executing status page template", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	fullHTML, err := templates.RenderBasePage("Status", content.String(), config.StatusHTTPPath)
	if err != nil {
		log.Error("error rendering base page", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(fullHTML))
}

// HandleStatusStream streams live status updates via DataStar SSE to the browser.
func HandleStatusStream(w http.ResponseWriter, r *http.Request) {
	sse := datastarlib.NewSSE(w, r)

	if err := sendStatusSnapshot(sse); err != nil {
		log.Error("error sending initial status snapshot", "error", err)
		return
	}

	ticker := time.NewTicker(statusStreamInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if err := sendStatusSnapshot(sse); err != nil {
				log.Error("error streaming status snapshot", "error", err)
				return
			}
		}
	}
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

// GetSystemStatus returns current system status for web display
func GetSystemStatus() (*SystemStatus, error) {
	metricsSnapshot, err := gops.GetSystemMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get simple system metrics: %w", err)
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	status := &SystemStatus{
		CPU:       float64(metricsSnapshot.Runtime.NumGoroutines),
		Memory:    float64(metricsSnapshot.Runtime.MemAlloc),
		Disk:      0.0,
		Uptime:    getSystemUptime(),
		Load:      getLoadAverage(),
		Timestamp: time.Now(),
		Runtime:   metricsSnapshot.Runtime,

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

func sendStatusSnapshot(sse *datastarlib.ServerSentEventGenerator) error {
	status, err := GetSystemStatus()
	if err != nil {
		return err
	}

	snapshot := mapStatusToTemplate(*status)

	html, err := infodatastar.RenderStatusCards(snapshot)
	if err != nil {
		return err
	}

	return sse.PatchElements(html)
}

func mapStatusToTemplate(status SystemStatus) infodatastar.StatusTemplateData {
	runtimeData := infodatastar.StatusRuntime{
		Goroutines:     status.Runtime.NumGoroutines,
		NumCPU:         status.Runtime.NumCPU,
		HeapAlloc:      metrics.FormatBytes(status.HeapAllocBytes),
		HeapSys:        metrics.FormatBytes(status.HeapSysBytes),
		StackInuse:     metrics.FormatBytes(status.StackInuseBytes),
		NextGC:         metrics.FormatBytes(status.NextGCBytes),
		LastGCPause:    status.LastPause,
		TotalGC:        status.Runtime.NumGC,
		GoVersion:      status.GoVersion,
		GOOS:           status.Runtime.GOOS,
		GOARCH:         status.Runtime.GOARCH,
		MemoryPercent:  status.MemoryPercent,
		MemoryBarClass: pickMemoryBarClass(status.MemoryPercent),
	}

	services := make([]infodatastar.StatusService, len(status.Services))
	for i, svc := range status.Services {
		border, pill := serviceStyles(svc.Level)
		services[i] = infodatastar.StatusService{
			Name:   svc.Name,
			Status: svc.Status,
			Detail: svc.Detail,
			Icon:   svc.Icon,
			Border: border,
			Pill:   pill,
			Port:   svc.Port,
		}
	}

	summary := buildSummary(status.Services)

	return infodatastar.StatusTemplateData{
		LastUpdatedDisplay: status.Timestamp.Format("15:04:05"),
		LastUpdatedISO:     status.Timestamp.Format(time.RFC3339),
		SummaryIcon:        summary.Icon,
		SummaryHeadline:    summary.Headline,
		SummaryBody:        summary.Body,
		SummaryBorder:      summary.Border,
		SummaryGradient:    summary.Gradient,
		Uptime:             status.Uptime,
		LoadAverage:        status.Load,
		Runtime:            runtimeData,
		Services:           services,
	}
}

type statusSummary struct {
	Icon     string
	Headline string
	Body     string
	Border   string
	Gradient string
}

func buildSummary(services []ServiceStatus) statusSummary {
	hasError := false
	hasWarn := false
	for _, svc := range services {
		switch svc.Level {
		case "error":
			hasError = true
		case "warn":
			hasWarn = true
		}
	}

	switch {
	case hasError:
		return statusSummary{
			Icon:     "🛑",
			Headline: "Attention required",
			Body:     "One or more required services are down.",
			Border:   "border-red-200 dark:border-red-700",
			Gradient: "from-red-100 to-red-50 dark:from-red-900/20 dark:to-red-800/20",
		}
	case hasWarn:
		return statusSummary{
			Icon:     "⚠️",
			Headline: "Degraded",
			Body:     "Optional services are unavailable; functionality may be limited.",
			Border:   "border-amber-200 dark:border-amber-600",
			Gradient: "from-amber-100 to-amber-50 dark:from-amber-900/20 dark:to-amber-800/20",
		}
	default:
		return statusSummary{
			Icon:     "✅",
			Headline: "All systems operational",
			Body:     "Core infrastructure components are reachable.",
			Border:   "border-emerald-200 dark:border-emerald-600",
			Gradient: "from-emerald-100 to-emerald-50 dark:from-emerald-900/20 dark:to-emerald-800/20",
		}
	}
}

func serviceStyles(level string) (border, pill string) {
	switch level {
	case "error":
		return "border border-red-200 dark:border-red-700", "bg-red-500/90 text-white"
	case "warn":
		return "border border-amber-200 dark:border-amber-600", "bg-amber-500/80 text-gray-900"
	default:
		return "border border-emerald-200 dark:border-emerald-600", "bg-emerald-500/80 text-white"
	}
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

func pickMemoryBarClass(percent float64) string {
	switch {
	case percent >= 85:
		return "bg-rose-500"
	case percent >= 65:
		return "bg-amber-500"
	default:
		return "bg-emerald-500"
	}
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
	var services []ServiceStatus

	for _, probe := range []struct {
		Name     string
		Port     int
		Icon     string
		Required bool
		Detail   string
	}{
		{
			Name:     "Web UI",
			Port:     atoiOrDefault(config.GetWebServerPort(), 1337),
			Icon:     "🌐",
			Required: true,
			Detail:   "HTTP server that hosts the control panel.",
		},
		{
			Name:     "NATS",
			Port:     atoiOrDefault(config.GetNATSPort(), 4222),
			Icon:     "📡",
			Required: true,
			Detail:   "Embedded messaging backbone (JetStream).",
		},
		{
			Name:     "Metrics",
			Port:     atoiOrDefault(config.GetMetricsPort(), 9091),
			Icon:     "📈",
			Required: false,
			Detail:   "Prometheus scrape endpoint.",
		},
		{
			Name:     "Deck API",
			Port:     atoiOrDefault(config.GetDeckAPIPort(), 8888),
			Icon:     "🃏",
			Required: false,
			Detail:   "On-demand PowerPoint generator.",
		},
	} {
		healthy := probePort(probe.Port)
		level := "ok"
		statusLabel := "Running"
		if !healthy {
			if probe.Required {
				level = "error"
				statusLabel = "Down"
			} else {
				level = "warn"
				statusLabel = "Standby"
			}
		}

		services = append(services, ServiceStatus{
			Name:     probe.Name,
			Status:   statusLabel,
			Detail:   probe.Detail,
			Port:     probe.Port,
			Icon:     probe.Icon,
			Healthy:  healthy,
			Required: probe.Required,
			Level:    level,
		})
	}

	return services
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
