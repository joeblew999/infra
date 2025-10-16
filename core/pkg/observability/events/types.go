package events

import (
	"fmt"
	"encoding/json"

	"strings"
	"time"

	"github.com/joeblew999/infra/core/pkg/runtime/process"
)

// EventType represents the type of process event.
type EventType string

const (
	// Lifecycle events
	EventTypeStarted  EventType = "started"
	EventTypeStopped  EventType = "stopped"
	EventTypeCrashed  EventType = "crashed"
	EventTypeRestarted EventType = "restarted"

	// Health events
	EventTypeHealthy   EventType = "healthy"
	EventTypeUnhealthy EventType = "unhealthy"

	// Status events
	EventTypeStatusChanged EventType = "status_changed"

	// Log events (for future WebSocket integration)
	EventTypeLog EventType = "log"
)

// Event represents a process lifecycle or health event.
type Event struct {
	// Event metadata
	Type      EventType `json:"type"`
	Process   string    `json:"process"`
	Namespace string    `json:"namespace,omitempty"`
	Timestamp time.Time `json:"timestamp"`

	// Current process state snapshot
	State process.ComposeProcessState `json:"state"`

	// Event-specific fields
	ExitCode  *int   `json:"exit_code,omitempty"`  // For crashed/stopped events
	Restarts  int    `json:"restarts,omitempty"`   // For restarted events
	Health    string `json:"health,omitempty"`     // For health events
	OldStatus string `json:"old_status,omitempty"` // For status_changed events
	NewStatus string `json:"new_status,omitempty"` // For status_changed events
	LogLine   string `json:"log_line,omitempty"`   // For log events
}

// Subject returns the NATS subject for this event.
// Format: core.process.{name}.{type}
// Examples:
//   - core.process.nats.started
//   - core.process.pocketbase.crashed
//   - core.process.datastar.healthy
func (e Event) Subject() string {
	processName := e.Process
	if e.Namespace != "" {
		// Replace slashes with dots for NATS subject hierarchy
		processName = e.Namespace + "." + e.Process
	}
	// Sanitize process name for NATS subject
	processName = strings.ReplaceAll(processName, "/", ".")
	processName = strings.ReplaceAll(processName, " ", "_")

	return fmt.Sprintf("core.process.%s.%s", processName, e.Type)
}

// String returns a human-readable description of the event.
func (e Event) String() string {
	prefix := e.Process
	if e.Namespace != "" {
		prefix = e.Namespace + "/" + e.Process
	}

	switch e.Type {
	case EventTypeStarted:
		return fmt.Sprintf("%s started ", prefix)
	case EventTypeStopped:
		if e.ExitCode != nil {
			return fmt.Sprintf("%s stopped (exit=%d)", prefix, *e.ExitCode)
		}
		return fmt.Sprintf("%s stopped", prefix)
	case EventTypeCrashed:
		if e.ExitCode != nil {
			return fmt.Sprintf("%s crashed (exit=%d)", prefix, *e.ExitCode)
		}
		return fmt.Sprintf("%s crashed", prefix)
	case EventTypeRestarted:
		return fmt.Sprintf("%s restarted (count=%d)", prefix, e.Restarts)
	case EventTypeHealthy:
		return fmt.Sprintf("%s healthy", prefix)
	case EventTypeUnhealthy:
		return fmt.Sprintf("%s unhealthy (health=%s)", prefix, e.Health)
	case EventTypeStatusChanged:
		return fmt.Sprintf("%s status: %s → %s", prefix, e.OldStatus, e.NewStatus)
	case EventTypeLog:
		return fmt.Sprintf("%s: %s", prefix, e.LogLine)
	default:
		return fmt.Sprintf("%s %s", prefix, e.Type)
	}
}

// Severity returns the severity level of this event.
func (e Event) Severity() Severity {
	switch e.Type {
	case EventTypeCrashed:
		return SeverityError
	case EventTypeUnhealthy:
		return SeverityWarning
	case EventTypeStopped:
		return SeverityInfo
	case EventTypeStarted, EventTypeHealthy, EventTypeRestarted:
		return SeverityInfo
	case EventTypeStatusChanged, EventTypeLog:
		return SeverityDebug
	default:
		return SeverityInfo
	}
}

// Severity represents the severity level of an event.
type Severity string

const (
	SeverityDebug   Severity = "debug"
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// MarshalJSON implements json.Marshaler.
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(struct {
		Alias
		Subject  string   `json:"subject"`
		Severity Severity `json:"severity"`
	}{
		Alias:    (Alias)(e),
		Subject:  e.Subject(),
		Severity: e.Severity(),
	})
}

// SubjectPattern returns the NATS subject pattern for subscribing to events.
// Examples:
//   - AllEvents() -> "core.process.>"
//   - ForProcess("nats") -> "core.process.nats.*"
//   - ForEventType(EventTypeCrashed) -> "core.process.*.crashed"
func SubjectPattern(opts ...SubjectOption) string {
	pattern := &subjectPattern{
		process: "*",
		event:   "*",
	}
	for _, opt := range opts {
		opt(pattern)
	}
	return pattern.build()
}

type subjectPattern struct {
	process string
	event   string
}

func (p *subjectPattern) build() string {
	if p.event == "" {
		return fmt.Sprintf("core.process.%s", p.process)
	}
	return fmt.Sprintf("core.process.%s.%s", p.process, p.event)
}

// SubjectOption configures a subject pattern.
type SubjectOption func(*subjectPattern)

// AllEvents returns a pattern matching all process events.
func AllEvents() SubjectOption {
	return func(p *subjectPattern) {
		p.process = ">"
		p.event = ""
	}
}

// ForProcess returns a pattern matching all events for a specific process.
func ForProcess(name string) SubjectOption {
	return func(p *subjectPattern) {
		p.process = strings.ReplaceAll(name, "/", ".")
	}
}

// ForEventType returns a pattern matching a specific event type across all processes.
func ForEventType(eventType EventType) SubjectOption {
	return func(p *subjectPattern) {
		p.event = string(eventType)
	}
}

// ForProcessAndType returns a pattern matching a specific event type for a specific process.
func ForProcessAndType(name string, eventType EventType) SubjectOption {
	return func(p *subjectPattern) {
		p.process = strings.ReplaceAll(name, "/", ".")
		p.event = string(eventType)
	}
}
