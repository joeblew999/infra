package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joeblew999/infra/pkg/goreman"
)

// ProcessStatusWeb represents process status for web display
type ProcessStatusWeb struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Indicator string `json:"indicator"`
	Uptime    string `json:"uptime"`
	CanStart  bool   `json:"can_start"`
	CanStop   bool   `json:"can_stop"`
}

// WebHandler provides HTTP handlers for process monitoring
type WebHandler struct {
	manager   *goreman.Manager
	templates *template.Template
	webDir    string
}

// NewWebHandler creates a new web handler
func NewWebHandler(webDir string) *WebHandler {
	// Load templates from the web directory
	templates := template.New("")
	if webDir != "" {
		globPattern := webDir + "/templates/*.html"
		template.Must(templates.ParseGlob(globPattern))
	}
	
	return &WebHandler{
		manager:   goreman.GetManager(),
		templates: templates,
		webDir:    webDir,
	}
}

// SetupRoutes configures the process monitoring routes on a chi router
func (w *WebHandler) SetupRoutes(r chi.Router) {
	r.Get("/", w.ProcessesPageHandler)
	r.Get("/api", w.ProcessesAPIHandler) 
	r.Post("/api/{name}/{action}", w.ProcessActionHandler)
	r.Post("/api/demo/register", w.RegisterDemoProcessHandler)
}

// ProcessesPageHandler serves the main processes page
func (w *WebHandler) ProcessesPageHandler(rw http.ResponseWriter, r *http.Request) {
	// Try to use template if available, otherwise fallback to direct file serving
	if w.templates != nil {
		if tmpl := w.templates.Lookup("processes.html"); tmpl != nil {
			rw.Header().Set("Content-Type", "text/html")
			if err := tmpl.Execute(rw, nil); err != nil {
				http.Error(rw, fmt.Sprintf("Template execution failed: %v", err), http.StatusInternalServerError)
			}
			return
		}
	}
	
	// Fallback: serve file directly if template not loaded
	if w.webDir != "" {
		http.ServeFile(rw, r, w.webDir+"/templates/processes.html")
		return
	}
	
	// Final fallback error
	http.Error(rw, "Template not found and webDir not configured", http.StatusInternalServerError)
}

// ProcessesAPIHandler provides JSON API for process status
func (w *WebHandler) ProcessesAPIHandler(rw http.ResponseWriter, r *http.Request) {
	status := goreman.GetAllStatus()
	processes := make([]ProcessStatusWeb, 0, len(status))
	
	running := 0
	stopped := 0
	
	for name, stat := range status {
		indicator := getStatusIndicator(stat)
		uptime := "--"
		
		if stat == "running" {
			running++
			uptime = "Running"
		} else {
			stopped++
		}
		
		processes = append(processes, ProcessStatusWeb{
			Name:      name,
			Status:    stat,
			Indicator: indicator,
			Uptime:    uptime,
			CanStart:  stat != "running",
			CanStop:   stat == "running",
		})
	}
	
	response := map[string]interface{}{
		"processes": processes,
		"summary": map[string]int{
			"total":   len(status),
			"running": running,
			"stopped": stopped,
		},
		"timestamp": time.Now().Unix(),
	}
	
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(response)
}

// ProcessActionHandler handles process start/stop/restart actions
func (w *WebHandler) ProcessActionHandler(rw http.ResponseWriter, r *http.Request) {
	// Extract process name and action from chi URL parameters
	processName := chi.URLParam(r, "name")
	action := chi.URLParam(r, "action")
	
	if processName == "" || action == "" {
		http.Error(rw, "Missing name or action parameter", http.StatusBadRequest)
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
		http.Error(rw, "Invalid action", http.StatusBadRequest)
		return
	}
	
	if err != nil {
		http.Error(rw, fmt.Sprintf("Action failed: %v", err), http.StatusInternalServerError)
		return
	}
	
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(map[string]string{"status": "success"})
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

// RegisterDemoProcessHandler registers a long-running demo process for testing
func (w *WebHandler) RegisterDemoProcessHandler(rw http.ResponseWriter, r *http.Request) {
	// Register a long-running test process
	goreman.Register("test-process", &goreman.ProcessConfig{
		Command:    "sh",
		Args:       []string{"-c", "while true; do echo 'Test process running...'; sleep 5; done"},
		WorkingDir: ".",
		Env:        []string{},
	})

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]string{
		"status": "success",
		"message": "Demo process registered. Use /api/processes/test-process/start to start it.",
	})
}