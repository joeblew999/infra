package ui

import (
	"strings"
	"time"
)

// ApplyProcessLogs records the provided log lines for the named process. The
// data is stored in the snapshot so UI surfaces can render recent output
// without querying the supervisor on every frame.
func ApplyProcessLogs(snapshot *Snapshot, processID string, logs []string, offset, limit int, truncated bool) {
	if snapshot == nil {
		return
	}
	id := NormalizeProcessID(strings.TrimSpace(processID), "")
	if id == "" {
		return
	}
	if snapshot.Processes == nil {
		snapshot.Processes = make(map[string]ProcessDetail)
	}
	detail := snapshot.Processes[id]
	detail.Logs = ProcessLogs{
		Lines:       append([]string(nil), logs...),
		Offset:      offset,
		Limit:       limit,
		CollectedAt: time.Now().Round(time.Second),
		Truncated:   truncated,
	}
	snapshot.Processes[id] = detail
}
