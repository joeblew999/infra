package composecfg

import "encoding/json"

// Config describes Process Compose specific options for a process entry.
type Config struct {
	IsDaemon     *bool                 `json:"is_daemon,omitempty"`
	Availability *Availability         `json:"availability,omitempty"`
	DependsOn    map[string]Dependency `json:"depends_on,omitempty"`
	Readiness    *Probe                `json:"readiness_probe,omitempty"`
	Startup      *Probe                `json:"startup_probe,omitempty"`
	Shutdown     *Probe                `json:"shutdown_probe,omitempty"`
	Log          *Log                  `json:"log,omitempty"`
}

// Availability configures restart behaviour.
type Availability struct {
	Restart string `json:"restart,omitempty"`
}

// Dependency describes a Process Compose dependency condition.
type Dependency struct {
	Condition string `json:"condition,omitempty"`
}

// Probe models Process Compose probes (http/tcp/exec).
type Probe struct {
	HTTPGet              *HTTPGet   `json:"http_get,omitempty"`
	TCP                  *TCPProbe  `json:"tcp,omitempty"`
	Exec                 *ExecProbe `json:"exec,omitempty"`
	InitialDelaySeconds  int        `json:"initial_delay_seconds,omitempty"`
	PeriodSeconds        int        `json:"period_seconds,omitempty"`
	TimeoutSeconds       int        `json:"timeout_seconds,omitempty"`
	Interval             string     `json:"interval,omitempty"`        // Deprecated: use PeriodSeconds  
	Timeout              string     `json:"timeout,omitempty"`         // Deprecated: use TimeoutSeconds
	SuccessThreshold     int        `json:"success_threshold,omitempty"`
	FailureThreshold     int        `json:"failure_threshold,omitempty"`
}

// HTTPGet defines an HTTP probe.
type HTTPGet struct {
	URL string `json:"url,omitempty"`
}

// TCPProbe defines a TCP probe.
type TCPProbe struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
}

// ExecProbe defines an exec probe command.
type ExecProbe struct {
	Command    string `json:"command,omitempty"`
	WorkingDir string `json:"working_dir,omitempty"`
}

// Log describes Process Compose log rotation options.
type Log struct {
	MaxFiles   *int   `json:"max_files,omitempty"`
	MaxSize    string `json:"max_size,omitempty"`
	Timestamps *bool  `json:"timestamps,omitempty"`
	Level      string `json:"level,omitempty"`
}

// Map converts the configuration into a generic map for template merging.
func (c *Config) Map() map[string]any {
	if c == nil {
		return nil
	}
	data, err := json.Marshal(c)
	if err != nil {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return out
}
