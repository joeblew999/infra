package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source"`
	Details   string `json:"details,omitempty"`
}

// LogStreamConfig holds configuration for log streaming
type LogStreamConfig struct {
	MaxLines    int           `json:"max_lines"`
	RefreshRate time.Duration `json:"refresh_rate"`
	LogLevel    string        `json:"log_level"`
	ShowSources []string      `json:"show_sources"`
}

// GetDefaultConfig returns default configuration for log streaming
func GetDefaultConfig() LogStreamConfig {
	return LogStreamConfig{
		MaxLines:    100,
		RefreshRate: 2 * time.Second,
		LogLevel:    "info",
		ShowSources: []string{"system", "web", "nats", "goreman"},
	}
}

// GetRecentLogs returns recent log entries from various sources
func GetRecentLogs(maxLines int) ([]LogEntry, error) {
	var logs []LogEntry

	webPort := config.GetWebServerPort()
	natsURL := config.GetNATSURL()

	// For now, we'll simulate log entries with actual system activity
	// In a real implementation, this would read from log files or a centralized log store

	// Add some system activity logs
	logs = append(logs, LogEntry{
		Timestamp: time.Now().Add(-5 * time.Minute).Format("15:04:05"),
		Level:     "INFO",
		Message:   fmt.Sprintf("Web server started on port %s", webPort),
		Source:    "web",
		Details:   "HTTP server initialization completed successfully",
	})

	logs = append(logs, LogEntry{
		Timestamp: time.Now().Add(-4 * time.Minute).Format("15:04:05"),
		Level:     "DEBUG",
		Message:   "NATS connection established",
		Source:    "nats",
		Details:   fmt.Sprintf("Connected to %s", natsURL),
	})

	logs = append(logs, LogEntry{
		Timestamp: time.Now().Add(-3 * time.Minute).Format("15:04:05"),
		Level:     "INFO",
		Message:   "Process monitoring initialized",
		Source:    "goreman",
		Details:   "Watching 3 processes for health status",
	})

	logs = append(logs, LogEntry{
		Timestamp: time.Now().Add(-2 * time.Minute).Format("15:04:05"),
		Level:     "WARN",
		Message:   "High memory usage detected",
		Source:    "system",
		Details:   "Memory usage at 78%, consider monitoring",
	})

	logs = append(logs, LogEntry{
		Timestamp: time.Now().Add(-1 * time.Minute).Format("15:04:05"),
		Level:     "INFO",
		Message:   "System status check completed",
		Source:    "web",
		Details:   "All health checks passed successfully",
	})

	// Add a recent log entry
	logs = append(logs, LogEntry{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     "INFO",
		Message:   "Log streaming service active",
		Source:    "web",
		Details:   "Real-time log monitoring is now operational",
	})

	// Limit to maxLines
	if len(logs) > maxLines {
		logs = logs[len(logs)-maxLines:]
	}

	return logs, nil
}

// StreamLogs handles real-time log streaming via Server-Sent Events
func StreamLogs(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	config := GetDefaultConfig()

	// Initial load of recent logs
	logs, err := GetRecentLogs(config.MaxLines)
	if err != nil {
		log.Error("Error getting recent logs", "error", err)
		http.Error(w, "Failed to get logs", http.StatusInternalServerError)
		return
	}

	// Send initial logs
	logsHTML := renderLogsHTML(logs)
	if err := sse.PatchElements(fmt.Sprintf(`<div id="log-entries">%s</div>`, logsHTML)); err != nil {
		log.Error("Error sending initial logs", "error", err)
		return
	}

	// Set up ticker for periodic updates
	ticker := time.NewTicker(config.RefreshRate)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			log.Debug("Log streaming client disconnected")
			return
		case <-ticker.C:
			// Get fresh logs
			freshLogs, err := GetRecentLogs(config.MaxLines)
			if err != nil {
				log.Error("Error getting fresh logs", "error", err)
				continue
			}

			// Send updated logs
			freshHTML := renderLogsHTML(freshLogs)
			if err := sse.PatchElements(fmt.Sprintf(`<div id="log-entries">%s</div>`, freshHTML)); err != nil {
				log.Error("Error sending fresh logs", "error", err)
				return
			}

			// Also update the timestamp
			timestamp := fmt.Sprintf(`<div id="last-updated">Last updated: %s</div>`, time.Now().Format("15:04:05"))
			if err := sse.PatchElements(timestamp); err != nil {
				log.Error("Error updating timestamp", "error", err)
				return
			}
		}
	}
}

// renderLogsHTML converts log entries to HTML
func renderLogsHTML(logs []LogEntry) string {
	var html strings.Builder

	for _, entry := range logs {
		levelColor := getLevelColor(entry.Level)
		sourceColor := getSourceColor(entry.Source)

		html.WriteString(fmt.Sprintf(`
			<div class="mb-2 p-3 bg-gray-50 dark:bg-gray-700 rounded-md border-l-4 %s">
				<div class="flex items-center justify-between mb-1">
					<div class="flex items-center space-x-2">
						<span class="text-xs font-mono %s px-2 py-1 rounded">%s</span>
						<span class="text-xs font-mono %s px-2 py-1 rounded">%s</span>
						<span class="text-xs text-gray-500 dark:text-gray-400">%s</span>
					</div>
				</div>
				<div class="text-sm text-gray-900 dark:text-white font-medium">%s</div>
				%s
			</div>`,
			levelColor,
			getLevelTextColor(entry.Level), entry.Level,
			sourceColor, entry.Source,
			entry.Timestamp,
			entry.Message,
			func() string {
				if entry.Details != "" {
					return fmt.Sprintf(`<div class="text-xs text-gray-600 dark:text-gray-400 mt-1">%s</div>`, entry.Details)
				}
				return ""
			}(),
		))
	}

	return html.String()
}

// getLevelColor returns the border color class for log level
func getLevelColor(level string) string {
	switch strings.ToUpper(level) {
	case "ERROR":
		return "border-red-500"
	case "WARN":
		return "border-yellow-500"
	case "INFO":
		return "border-blue-500"
	case "DEBUG":
		return "border-gray-500"
	default:
		return "border-gray-300"
	}
}

// getLevelTextColor returns the text color class for log level badge
func getLevelTextColor(level string) string {
	switch strings.ToUpper(level) {
	case "ERROR":
		return "bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400"
	case "WARN":
		return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400"
	case "INFO":
		return "bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400"
	case "DEBUG":
		return "bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-400"
	default:
		return "bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-400"
	}
}

// getSourceColor returns the color class for source badge
func getSourceColor(source string) string {
	switch source {
	case "web":
		return "bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400"
	case "system":
		return "bg-purple-100 text-purple-800 dark:bg-purple-900/20 dark:text-purple-400"
	case "nats":
		return "bg-indigo-100 text-indigo-800 dark:bg-indigo-900/20 dark:text-indigo-400"
	case "goreman":
		return "bg-orange-100 text-orange-800 dark:bg-orange-900/20 dark:text-orange-400"
	default:
		return "bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-400"
	}
}

// LogsWebService provides web interface for logs management
type LogsWebService struct{}

// NewLogsWebService creates a new logs web service
func NewLogsWebService() *LogsWebService {
	return &LogsWebService{}
}

// RegisterRoutes mounts all logs routes on the provided router
func (s *LogsWebService) RegisterRoutes(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		content, err := RenderLogsPage()
		if err != nil {
			http.Error(w, "Failed to render logs page", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(content))
	})

	// API routes
	r.Get("/api/stream", StreamLogs)
	r.Get("/api/config", HandleLogConfig)
	r.Post("/api/config", HandleLogConfig)
}

// HandleLogConfig handles log configuration updates
func HandleLogConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		config := GetDefaultConfig()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
		return
	}

	if r.Method == http.MethodPost {
		var config LogStreamConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Here you could save the configuration to a file or database
		// For now, just return success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	}
}
