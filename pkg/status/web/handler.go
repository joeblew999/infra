package web

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	datastarlib "github.com/starfederation/datastar-go/datastar"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	runtimeevents "github.com/joeblew999/infra/pkg/runtime/events"
	"github.com/joeblew999/infra/pkg/status"
	"github.com/joeblew999/infra/pkg/webapp/templates"
)

func init() {
	templates.RegisterNavItem(templates.NavItem{
		Href:  config.StatusHTTPPath,
		Text:  "Status",
		Icon:  "⚡",
		Color: "gray",
		Order: 90,
	})
}

//go:embed templates/status-page.html
var statusPageTemplate string

const statusStreamInterval = 4 * time.Second

// StatusWebService provides web interface for system status monitoring
type StatusWebService struct{}

// NewStatusWebService creates a new status web service
func NewStatusWebService() *StatusWebService {
	return &StatusWebService{}
}

// RegisterRoutes mounts all status routes on the provided router
func (s *StatusWebService) RegisterRoutes(r chi.Router) {
	// Status page routes
	r.Get("/", HandleStatusPage)
	r.Get("/api/stream", HandleStatusStream)
	r.Get("/api/events", HandleServiceEvents)
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

	eventsCh, cancel := runtimeevents.Subscribe(128)
	defer cancel()

	keepAlive := time.NewTicker(statusStreamInterval)
	defer keepAlive.Stop()

	var lastUpdate time.Time
	for {
		select {
		case <-r.Context().Done():
			return
		case evt := <-eventsCh:
			if evt == nil {
				continue
			}
			if !lastUpdate.IsZero() && time.Since(lastUpdate) < 75*time.Millisecond {
				continue
			}
			if err := sendStatusSnapshot(sse); err != nil {
				log.Error("error streaming status snapshot", "error", err)
				return
			}
			lastUpdate = time.Now()
		case <-keepAlive.C:
			if err := sendStatusSnapshot(sse); err != nil {
				log.Error("error streaming keepalive snapshot", "error", err)
				return
			}
			lastUpdate = time.Now()
		}
	}
}

func HandleServiceEvents(w http.ResponseWriter, r *http.Request) {
	sse := datastarlib.NewSSE(w, r)

	if err := sendInitialStateEvent(sse); err != nil {
		log.Error("error sending initial state event", "error", err)
		return
	}

	eventsCh, cancel := runtimeevents.Subscribe(256)
	defer cancel()

	for {
		select {
		case <-r.Context().Done():
			return
		case evt := <-eventsCh:
			if evt == nil {
				continue
			}
			if err := sendRuntimeEvent(sse, evt); err != nil {
				log.Error("error streaming runtime event", "error", err)
				return
			}
		}
	}
}

// HandleSystemStatus is an HTTP handler for system status endpoint
func HandleSystemStatus(w http.ResponseWriter, r *http.Request) {
	systemStatus, err := status.GetCurrentStatus()
	if err != nil {
		log.Error("Error getting system status", "error", err)
		http.Error(w, "Failed to get system status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(systemStatus); err != nil {
		log.Error("Error encoding system status", "error", err)
		http.Error(w, "Failed to encode system status", http.StatusInternalServerError)
		return
	}
}

// Helper functions

func sendStatusSnapshot(sse *datastarlib.ServerSentEventGenerator) error {
	systemStatus, err := status.GetCurrentStatus()
	if err != nil {
		return err
	}

	snapshot := mapStatusToTemplate(*systemStatus)

	html, err := RenderStatusCards(snapshot)
	if err != nil {
		return err
	}

	return sse.PatchElements(html)
}

func sendInitialStateEvent(sse *datastarlib.ServerSentEventGenerator) error {
	systemStatus, err := status.GetCurrentStatus()
	if err != nil {
		return err
	}

	envelope := map[string]any{
		"type":      "initial-state",
		"timestamp": systemStatus.Timestamp,
		"services":  systemStatus.Services,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	return sse.Send("infra-service-event", []string{string(data)})
}

func sendRuntimeEvent(sse *datastarlib.ServerSentEventGenerator, evt runtimeevents.Event) error {
	envelope := map[string]any{
		"type":      evt.Type(),
		"timestamp": evt.Timestamp(),
	}

	switch e := evt.(type) {
	case runtimeevents.ServiceStatus:
		envelope["service_id"] = e.ServiceID
		envelope["running"] = e.Running
		envelope["pid"] = e.PID
		envelope["port"] = e.Port
		envelope["ownership"] = e.Ownership
		envelope["state"] = e.State
		envelope["message"] = e.Message
	case runtimeevents.ServiceAction:
		envelope["service_id"] = e.ServiceID
		envelope["message"] = e.Message
		envelope["kind"] = e.Kind
	case runtimeevents.ServiceRegistered:
		envelope["service_id"] = e.ServiceID
		envelope["name"] = e.Name
		envelope["description"] = e.Description
		envelope["required"] = e.Required
		envelope["port"] = e.Port
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	return sse.Send("infra-service-event", []string{string(data)})
}

func mapStatusToTemplate(systemStatus status.SystemStatus) StatusTemplateData {
	runtimeData := StatusRuntime{
		Goroutines:     systemStatus.Runtime.NumGoroutines,
		NumCPU:         systemStatus.Runtime.NumCPU,
		HeapAlloc:      status.FormatBytes(systemStatus.HeapAllocBytes),
		HeapSys:        status.FormatBytes(systemStatus.HeapSysBytes),
		StackInuse:     status.FormatBytes(systemStatus.StackInuseBytes),
		NextGC:         status.FormatBytes(systemStatus.NextGCBytes),
		LastGCPause:    systemStatus.LastPause,
		TotalGC:        systemStatus.Runtime.NumGC,
		GoVersion:      systemStatus.GoVersion,
		GOOS:           systemStatus.Runtime.GOOS,
		GOARCH:         systemStatus.Runtime.GOARCH,
		MemoryPercent:  systemStatus.MemoryPercent,
		MemoryBarClass: pickMemoryBarClass(systemStatus.MemoryPercent),
	}

	services := make([]StatusService, len(systemStatus.Services))
	for i, svc := range systemStatus.Services {
		border, pill := serviceStyles(svc.Level)
		services[i] = StatusService{
			Name:         svc.Name,
			State:        svc.State,
			Status:       svc.Status,
			Description:  svc.Description,
			LastAction:   svc.LastAction,
			Message:      svc.Message,
			Icon:         svc.Icon,
			Border:       border,
			Pill:         pill,
			Port:         svc.Port,
			Ownership:    svc.Ownership,
			MessageClass: messageClassForOwnership(svc.Ownership),
		}
	}

	summary := buildSummary(systemStatus.Services)

	return StatusTemplateData{
		LastUpdatedDisplay: systemStatus.Timestamp.Format("15:04:05"),
		LastUpdatedISO:     systemStatus.Timestamp.Format(time.RFC3339),
		SummaryIcon:        summary.Icon,
		SummaryHeadline:    summary.Headline,
		SummaryBody:        summary.Body,
		SummaryBorder:      summary.Border,
		SummaryGradient:    summary.Gradient,
		Uptime:             systemStatus.Uptime,
		LoadAverage:        systemStatus.Load,
		Runtime:            runtimeData,
		Services:           services,
	}
}

func messageClassForOwnership(ownership string) string {
	switch ownership {
	case "external":
		return "text-xs text-rose-600 dark:text-rose-300"
	case "infra":
		return "text-xs text-amber-600 dark:text-amber-300"
	case "this":
		return "text-xs text-emerald-600 dark:text-emerald-300"
	default:
		return "text-xs text-gray-500 dark:text-gray-400"
	}
}

type statusSummary struct {
	Icon     string
	Headline string
	Body     string
	Border   string
	Gradient string
}

func buildSummary(services []status.ServiceStatus) statusSummary {
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
