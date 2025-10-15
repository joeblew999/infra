package ui

import "time"

// Snapshot represents the aggregate state rendered by UI surfaces. Until the
// live event pipeline is available the snapshot may be populated with test data
// to exercise templates.
type Snapshot struct {
	Environment string
	DataDir     string
	GeneratedAt time.Time
	Build       BuildInfo

	Services       []ServiceCard
	Metrics        []MetricCard
	Events         []EventLog
	Tips           []string
	TextIslands    []TextIsland
	Navigation     []NavigationItem
	ServiceDetails map[string]ServiceDetail
	Processes      map[string]ProcessDetail
}

// ServiceCard describes one managed service for display in dashboards.
type ServiceCard struct {
	ID            string
	Status        string
	Command       string
	Ports         []string
	Health        string
	LastEvent     string
	Description   string
	Scalable      bool
	ScaleStrategy string
}

// MetricCard summarises a key runtime metric.
type MetricCard struct {
	Label string
	Value string
	Hint  string
}

// EventLog captures entries displayed in the activity feed.
type EventLog struct {
	Timestamp string
	Message   string
	Level     string
}

// TextIsland holds localized text fragments surfaced by the UI templates.
type TextIsland struct {
	Key    string
	Locale string
	Title  string
	Body   string
}

// NavigationItem represents a linkable page in the UI shells.
type NavigationItem struct {
	Title       string
	Route       string
	Description string
}

// ServiceDetail provides focused information for a single service page.
type ServiceDetail struct {
	Card      ServiceCard
	Notes     []string
	Checklist []string
}

// ProcessDetail captures runtime metadata and buffered log output for a
// process-compose managed process. The master-detail UIs consume this structure
// to render process lists and drill-down panes.
type ProcessDetail struct {
	Runtime       ProcessRuntime
	Logs          ProcessLogs
	Scalable      bool
	ScaleStrategy string
}

// ProcessRuntime holds the latest runtime metadata reported by the supervisor.
type ProcessRuntime struct {
	ID        string
	Namespace string
	Status    string
	Health    string
	HasHealth bool
	Restarts  int
	ExitCode  int
	Command   string
	Ports     []string
	Replicas  int
	UpdatedAt time.Time
}

// ProcessLogs caches recent log output for a process so UI surfaces can render
// tails without querying the supervisor on every frame.
type ProcessLogs struct {
	Lines       []string
	Offset      int
	Limit       int
	CollectedAt time.Time
	Truncated   bool
}

// BuildInfo summarises the orchestrator build metadata for UI surfaces.
type BuildInfo struct {
	Version   string
	Revision  string
	BuildTime string
	Dirty     bool
}
