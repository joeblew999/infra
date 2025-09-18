package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	datastarlib "github.com/starfederation/datastar-go/datastar"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/webapp/templates"
)

func init() {
	templates.RegisterNavItem(templates.NavItem{
		Href:  config.ProcessesHTTPPath,
		Text:  "Processes",
		Icon:  "🔍",
		Color: "indigo",
		Order: 60,
	})
}

const processesStreamInterval = 3 * time.Second

//go:embed templates/processes-page.html
var processesPageTemplate string

// ProcessStatusWeb represents the process status payload for the JSON API.
type ProcessStatusWeb struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Indicator string `json:"indicator"`
	Uptime    string `json:"uptime"`
	CanStart  bool   `json:"can_start"`
	CanStop   bool   `json:"can_stop"`
}

// WebHandler provides HTTP handlers for process monitoring.
type WebHandler struct{}

// NewWebHandler creates a new web handler. The webDir parameter is retained for compatibility.
func NewWebHandler(_ string) *WebHandler {
	return &WebHandler{}
}

// SetupRoutes configures the process monitoring routes on a chi router.
func (w *WebHandler) SetupRoutes(r chi.Router) {
	r.Get("/", w.ProcessesPageHandler)
	r.Get("/api", w.ProcessesAPIHandler)
	r.Get("/api/stream", w.ProcessesStreamHandler)
	r.Post("/api/{name}/{action}", w.ProcessActionHandler)
	r.Post("/api/demo/register", w.RegisterDemoProcessHandler)
}

// ProcessesPageHandler serves the main processes page.
func (w *WebHandler) ProcessesPageHandler(rw http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("processes-page").Parse(processesPageTemplate)
	if err != nil {
		log.Error("error parsing processes template", "error", err)
		http.Error(rw, "failed to render page", http.StatusInternalServerError)
		return
	}

	var content strings.Builder
	data := struct {
		BasePath string
	}{BasePath: config.ProcessesHTTPPath}

	if err := tmpl.Execute(&content, data); err != nil {
		log.Error("error executing processes template", "error", err)
		http.Error(rw, "failed to render page", http.StatusInternalServerError)
		return
	}

	fullHTML, err := templates.RenderBasePage("Processes", content.String(), config.ProcessesHTTPPath)
	if err != nil {
		log.Error("error rendering base page", "error", err)
		http.Error(rw, "failed to render page", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = rw.Write([]byte(fullHTML))
}

// ProcessesStreamHandler streams live process updates via DataStar SSE.
func (w *WebHandler) ProcessesStreamHandler(rw http.ResponseWriter, r *http.Request) {
	sse := datastarlib.NewSSE(rw, r)

	if err := w.sendProcessSnapshot(sse); err != nil {
		log.Error("error sending initial process snapshot", "error", err)
		return
	}

	ticker := time.NewTicker(processesStreamInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if err := w.sendProcessSnapshot(sse); err != nil {
				log.Error("error streaming process snapshot", "error", err)
				return
			}
		}
	}
}

func (w *WebHandler) sendProcessSnapshot(sse *datastarlib.ServerSentEventGenerator) error {
	snapshot := collectProcessSnapshot()

	html, err := RenderProcessCards(mapProcessSnapshot(snapshot))
	if err != nil {
		return fmt.Errorf("render process cards: %w", err)
	}

	return sse.PatchElements(html)
}

// ProcessesAPIHandler provides JSON API for process status.
func (w *WebHandler) ProcessesAPIHandler(rw http.ResponseWriter, r *http.Request) {
	snapshot := collectProcessSnapshot()

	processes := make([]ProcessStatusWeb, len(snapshot.Processes))
	for i, proc := range snapshot.Processes {
		processes[i] = ProcessStatusWeb{
			Name:      proc.Name,
			Status:    proc.Status,
			Indicator: proc.Indicator,
			Uptime:    proc.Uptime,
			CanStart:  proc.CanStart,
			CanStop:   proc.CanStop,
		}
	}

	response := map[string]any{
		"processes": processes,
		"summary": map[string]int{
			"total":   snapshot.Summary.Total,
			"running": snapshot.Summary.Running,
			"stopped": snapshot.Summary.Stopped,
		},
		"timestamp": snapshot.Timestamp.Unix(),
	}

	rw.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(rw).Encode(response); err != nil {
		log.Error("error encoding process status", "error", err)
	}
}

// ProcessActionHandler handles process start/stop/restart actions.
func (w *WebHandler) ProcessActionHandler(rw http.ResponseWriter, r *http.Request) {
	processName := chi.URLParam(r, "name")
	action := chi.URLParam(r, "action")

	if processName == "" || action == "" {
		http.Error(rw, "missing name or action", http.StatusBadRequest)
		return
	}

	var err error
	switch action {
	case "start":
		err = goreman.Start(processName)
	case "stop":
		err = goreman.Stop(processName)
	case "restart":
		err = goreman.Restart(processName)
	default:
		http.Error(rw, "invalid action", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Error("process action failed", "name", processName, "action", action, "error", err)
		http.Error(rw, fmt.Sprintf("action failed: %v", err), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(map[string]string{"status": "success"})
}

// RegisterDemoProcessHandler registers a long-running demo process for testing.
func (w *WebHandler) RegisterDemoProcessHandler(rw http.ResponseWriter, r *http.Request) {
	goreman.Register("test-process", &goreman.ProcessConfig{
		Command:    "sh",
		Args:       []string{"-c", "while true; do echo 'Test process running...'; sleep 5; done"},
		WorkingDir: ".",
	})

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Demo process registered. Use %s/api/test-process/start to start it.", config.ProcessesHTTPPath),
	})
}

// Internal snapshot types used to generate both JSON and SSE responses.
type processSnapshot struct {
	Processes []processState
	Summary   processSummary
	Timestamp time.Time
}

type processSummary struct {
	Total   int
	Running int
	Stopped int
}

type processState struct {
	Name      string
	Status    string
	Indicator string
	Uptime    string
	CanStart  bool
	CanStop   bool
}

func collectProcessSnapshot() processSnapshot {
	statusMap := goreman.GetAllStatus()
	names := make([]string, 0, len(statusMap))
	for name := range statusMap {
		names = append(names, name)
	}
	sort.Strings(names)

	states := make([]processState, 0, len(names))
	running := 0
	stopped := 0

	for _, name := range names {
		status := statusMap[name]
		state := processState{
			Name:      name,
			Status:    status,
			Indicator: getStatusIndicator(status),
			Uptime:    formatProcessUptime(status),
			CanStart:  status != "running",
			CanStop:   status == "running",
		}

		if status == "running" {
			running++
		} else {
			stopped++
		}

		states = append(states, state)
	}

	return processSnapshot{
		Processes: states,
		Summary: processSummary{
			Total:   len(states),
			Running: running,
			Stopped: stopped,
		},
		Timestamp: time.Now(),
	}
}

func mapProcessSnapshot(snapshot processSnapshot) ProcessTemplateData {
	processes := make([]ProcessCard, len(snapshot.Processes))
	base := strings.TrimSuffix(config.ProcessesHTTPPath, "/")

	for i, proc := range snapshot.Processes {
		border, badge := processCardClasses(proc.Status)
		encoded := url.PathEscape(proc.Name)

		processes[i] = ProcessCard{
			Name:             proc.Name,
			StatusLabel:      strings.ToUpper(proc.Status),
			Indicator:        proc.Indicator,
			Uptime:           proc.Uptime,
			ShowStart:        proc.CanStart,
			ShowStop:         proc.CanStop,
			ShowRestart:      proc.CanStop,
			StartAction:      fmt.Sprintf("%s/api/%s/start", base, encoded),
			StopAction:       fmt.Sprintf("%s/api/%s/stop", base, encoded),
			RestartAction:    fmt.Sprintf("%s/api/%s/restart", base, encoded),
			CardBorderClass:  border,
			StatusBadgeClass: badge,
		}
	}

	return ProcessTemplateData{
		LastUpdatedDisplay: snapshot.Timestamp.Format("15:04:05"),
		LastUpdatedISO:     snapshot.Timestamp.Format(time.RFC3339),
		Summary: ProcessSummary{
			Total:   snapshot.Summary.Total,
			Running: snapshot.Summary.Running,
			Stopped: snapshot.Summary.Stopped,
		},
		Processes:    processes,
		HasProcesses: len(processes) > 0,
	}
}

func formatProcessUptime(status string) string {
	if status == "running" {
		return "Running"
	}
	return "Stopped"
}

func processCardClasses(status string) (string, string) {
	switch status {
	case "running":
		return "border-emerald-200 dark:border-emerald-600", "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-200"
	case "starting", "stopping":
		return "border-amber-200 dark:border-amber-600", "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-200"
	case "stopped", "killed":
		return "border-rose-200 dark:border-rose-600", "bg-rose-100 text-rose-800 dark:bg-rose-900/40 dark:text-rose-200"
	default:
		return "border-gray-200 dark:border-gray-700", "bg-gray-200 text-gray-800 dark:bg-gray-700 dark:text-gray-200"
	}
}

func getStatusIndicator(status string) string {
	switch status {
	case "running":
		return "🟢"
	case "stopped":
		return "🔴"
	case "starting":
		return "🟡"
	case "stopping":
		return "🟠"
	case "killed":
		return "💀"
	default:
		return "❓"
	}
}
