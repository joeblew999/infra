package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/metrics"
	"github.com/joeblew999/infra/pkg/webapp/templates"
)

func init() {
	templates.RegisterNavItem(templates.NavItem{
		Href:  config.MetricsHTTPPath,
		Text:  "Metrics",
		Icon:  "📊",
		Color: "purple",
		Order: 30,
	})
}

//go:embed templates/metrics-cards.html
var metricsCardsTemplate string

//go:embed templates/metrics-page.html
var metricsPageTemplate string

// MetricsTemplateData holds data for the metrics template
type MetricsTemplateData struct {
	MemUsage       string
	MemUsageFloat  float64
	HeapAlloc      string
	HeapAllocBytes uint64
	HeapSys        string
	StackInuse     string
	NumCPU         int
	Goroutines     int
	Uptime         string
	GCCount        uint32
	LastPause      string
	NextGC         string
	Timestamp      int64
}

// MetricsWebService provides web interface for metrics management
type MetricsWebService struct{}

// NewMetricsWebService creates a new metrics web service
func NewMetricsWebService() *MetricsWebService {
	return &MetricsWebService{}
}

// RegisterRoutes mounts all metrics routes on the provided router
func (s *MetricsWebService) RegisterRoutes(r chi.Router) {
	// Metrics page routes
	r.Get("/", HandleMetricsPage)
	r.Get("/api", HandleMetricsAPI)
	r.Get("/api/history", HandleMetricsHistory)
	r.Get("/api/stream", HandleMetricsStream)
}

// HandleMetricsAPI returns current system metrics as JSON
func HandleMetricsAPI(w http.ResponseWriter, r *http.Request) {
	collector := metrics.GetCollector()
	latest := collector.GetLatest()

	if latest == nil {
		// Return empty metrics if none collected yet
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latest)
}

// HandleMetricsHistory returns metrics history as JSON
func HandleMetricsHistory(w http.ResponseWriter, r *http.Request) {
	collector := metrics.GetCollector()
	history := collector.GetHistory()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// HandleMetricsStream provides real-time metrics updates via SSE
func HandleMetricsStream(w http.ResponseWriter, r *http.Request) {
	collector := metrics.GetCollector()
	sse := datastar.NewSSE(w, r)

	// Subscribe to metrics updates
	metricsCh := collector.Subscribe()

	// Send initial metrics
	if latest := collector.GetLatest(); latest != nil {
		if err := sendMetricsUpdate(sse, latest); err != nil {
			log.Error("Error sending initial metrics", "error", err)
			return
		}
	}

	// Stream updates
	for {
		select {
		case <-r.Context().Done():
			return
		case m := <-metricsCh:
			if err := sendMetricsUpdate(sse, &m); err != nil {
				log.Error("Error sending metrics update", "error", err)
				return
			}
		}
	}
}

// HandleMetricsPage renders the complete metrics page with navigation
func HandleMetricsPage(w http.ResponseWriter, r *http.Request) {
	// Parse the metrics page template
	tmpl, err := template.New("metrics-page").Parse(metricsPageTemplate)
	if err != nil {
		log.Error("Error parsing metrics page template", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Execute the template to get the content HTML
	var contentBuf strings.Builder
	err = tmpl.Execute(&contentBuf, nil) // No data needed for the page template
	if err != nil {
		log.Error("Error executing metrics page template", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Use the centralized base page renderer
	fullHTML, err := templates.RenderBasePage("Metrics", contentBuf.String(), "/metrics")
	if err != nil {
		log.Error("Error rendering base page", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fullHTML))
}

// Helper functions

func sendMetricsUpdate(sse *datastar.ServerSentEventGenerator, m *metrics.SystemMetrics) error {
	// Prepare template data
	data := MetricsTemplateData{
		MemUsage:       fmt.Sprintf("%.1f%%", m.Memory.Usage),
		MemUsageFloat:  m.Memory.Usage,
		HeapAlloc:      metrics.FormatBytes(m.Memory.HeapAlloc),
		HeapAllocBytes: m.Memory.HeapAlloc,
		HeapSys:        metrics.FormatBytes(m.Memory.HeapSys),
		StackInuse:     metrics.FormatBytes(m.Memory.StackInuse),
		NumCPU:         m.CPU.NumCPU,
		Goroutines:     m.Goroutines,
		Uptime:         formatDuration(m.Uptime),
		GCCount:        m.GCStats.NumGC,
		LastPause:      m.GCStats.LastPause.String(),
		NextGC:         metrics.FormatBytes(m.GCStats.NextGC),
		Timestamp:      m.Timestamp.Unix(),
	}

	// Parse and execute template
	tmpl, err := template.New("metrics-cards").Parse(metricsCardsTemplate)
	if err != nil {
		return fmt.Errorf("error parsing metrics template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return fmt.Errorf("error executing metrics template: %w", err)
	}

	// Send the rendered template via SSE
	return sse.PatchElements(result.String())
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	} else {
		days := int(d.Hours() / 24)
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%dd %dh", days, hours)
	}
}
