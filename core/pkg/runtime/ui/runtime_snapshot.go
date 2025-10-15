package ui

import (
	"fmt"
	"strings"
	"time"

	runtimeprocess "github.com/joeblew999/infra/core/pkg/runtime/process"
)

// ServiceStatus captures the runtime state of a managed service for UI surfaces.
type ServiceStatus struct {
	ID        string
	Namespace string
	Status    string
	Health    string
	HasHealth bool
	Restarts  int
	ExitCode  int
	Replicas  int
}

// BuildSnapshotFromServiceStatus returns a snapshot seeded with live service
// state data while retaining the supporting metadata from the test snapshot.
func BuildSnapshotFromServiceStatus(states []ServiceStatus) Snapshot {
	snapshot := LoadTestSnapshot()
	ApplyServiceStatus(&snapshot, states)
	return snapshot
}

// ServiceStatusesFromCompose maps Process Compose process states into generic
// service status structures understood by the UI helpers.
func ServiceStatusesFromCompose(states []runtimeprocess.ComposeProcessState) []ServiceStatus {
	result := make([]ServiceStatus, 0, len(states))
	for _, st := range states {
		result = append(result, ServiceStatus{
			ID:        st.Name,
			Namespace: st.Namespace,
			Status:    st.Status,
			Health:    st.Health,
			HasHealth: st.HasHealthProbe,
			Restarts:  st.Restarts,
			ExitCode:  st.ExitCode,
			Replicas:  st.Replicas,
		})
	}
	return result
}

// ApplyServiceStatus mutates the provided snapshot in-place using the supplied
// service status information. Missing services fall back to "stopped".
func ApplyServiceStatus(snapshot *Snapshot, states []ServiceStatus) {
	if snapshot == nil {
		return
	}
	timestamp := time.Now().Round(time.Second)

	if snapshot.Processes == nil {
		snapshot.Processes = make(map[string]ProcessDetail)
	}
	if len(states) == 0 {
		snapshot.GeneratedAt = timestamp
		return
	}

	statusByID := make(map[string]ServiceStatus, len(states))
	for _, st := range states {
		id := normalizeServiceID(st.ID, st.Namespace)
		if id == "" {
			continue
		}
		statusByID[id] = st
	}

	running := 0
	totalRestarts := 0
	serviceByID := make(map[string]ServiceCard, len(snapshot.Services))
	for i := range snapshot.Services {
		svc := &snapshot.Services[i]
		st, ok := statusByID[svc.ID]
		if !ok {
			// mark as stopped when the supervisor does not report it
			svc.Status = "stopped"
			svc.Health = "-"
			svc.LastEvent = "not managed by process-compose"
			serviceByID[svc.ID] = *svc
			continue
		}

		status := strings.ToLower(strings.TrimSpace(st.Status))
		if status == "" {
			status = "running"
		}
		svc.Status = status

		health := strings.ToLower(strings.TrimSpace(st.Health))
		if health == "" {
			if st.HasHealth {
				health = "unknown"
			} else {
				health = "-"
			}
		}
		svc.Health = health

		svc.LastEvent = buildLastEventMessage(st)
		if status == "running" || status == "ready" {
			running++
		}
		if st.Restarts > 0 {
			totalRestarts += st.Restarts
		}
		serviceByID[svc.ID] = *svc
	}

	snapshot.GeneratedAt = timestamp
	snapshot.Navigation = buildNavigation(snapshot.Services)
	snapshot.ServiceDetails = buildServiceDetails(snapshot.Services)

	updateMetrics(snapshot, running, len(snapshot.Services), totalRestarts)
	prependEvent(snapshot, fmt.Sprintf("process-compose sync @ %s", snapshot.GeneratedAt.Format("15:04:05")))

	updateProcessRuntime(snapshot, states, serviceByID, timestamp)
}

func normalizeServiceID(name, namespace string) string {
	id := strings.TrimSpace(name)
	if id == "" {
		id = strings.TrimSpace(namespace)
	}
	if strings.Contains(id, "/") {
		parts := strings.Split(id, "/")
		id = parts[len(parts)-1]
	}
	switch id {
	case "caddy":
		return "core.caddy"
	default:
		return id
	}
}

// NormalizeProcessID exposes the internal normalization rules so other
// packages can align identifiers before mutating snapshot process maps.
func NormalizeProcessID(name, namespace string) string {
	return normalizeServiceID(name, namespace)
}

func buildLastEventMessage(st ServiceStatus) string {
	status := strings.TrimSpace(st.Status)
	if status == "" {
		status = "unknown"
	}
	if st.Restarts > 0 {
		return fmt.Sprintf("status %s (restarts %d)", status, st.Restarts)
	}
	return fmt.Sprintf("status %s", status)
}

func updateMetrics(snapshot *Snapshot, running, total, restarts int) {
	if total == 0 {
		return
	}

	activeValue := fmt.Sprintf("%d/%d", running, total)
	restartValue := fmt.Sprintf("%d", restarts)

	var activeFound, restartFound bool
	for i := range snapshot.Metrics {
		metric := &snapshot.Metrics[i]
		label := strings.ToLower(metric.Label)
		switch {
		case strings.Contains(label, "active") && strings.Contains(label, "service"):
			metric.Value = activeValue
			metric.Hint = "Services reported by process-compose"
			activeFound = true
		case strings.Contains(label, "restart"):
			metric.Value = restartValue
			metric.Hint = "Process Compose restart count"
			restartFound = true
		}
	}

	if !activeFound {
		snapshot.Metrics = append([]MetricCard{{
			Label: "Active Services",
			Value: activeValue,
			Hint:  "Services reported by process-compose",
		}}, snapshot.Metrics...)
	}

	if !restartFound {
		snapshot.Metrics = append(snapshot.Metrics, MetricCard{
			Label: "Process Restarts",
			Value: restartValue,
			Hint:  "Process Compose restart count",
		})
	}
}

func prependEvent(snapshot *Snapshot, message string) {
	entry := EventLog{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     "info",
		Message:   message,
	}
	snapshot.Events = append([]EventLog{entry}, snapshot.Events...)
	if len(snapshot.Events) > 10 {
		snapshot.Events = snapshot.Events[:10]
	}
}

func updateProcessRuntime(snapshot *Snapshot, states []ServiceStatus, serviceByID map[string]ServiceCard, timestamp time.Time) {
	if snapshot.Processes == nil {
		snapshot.Processes = make(map[string]ProcessDetail)
	}
	seen := make(map[string]struct{}, len(states))
	for _, st := range states {
		id := normalizeServiceID(st.ID, st.Namespace)
		if id == "" {
			continue
		}
		runtime := ProcessRuntime{
			ID:        st.ID,
			Namespace: st.Namespace,
			Status:    strings.ToLower(strings.TrimSpace(st.Status)),
			Health:    strings.ToLower(strings.TrimSpace(st.Health)),
			HasHealth: st.HasHealth,
			Restarts:  st.Restarts,
			ExitCode:  st.ExitCode,
			Replicas:  st.Replicas,
			UpdatedAt: timestamp,
		}
		if runtime.Status == "" {
			runtime.Status = "unknown"
		}
		if runtime.Health == "" {
			if st.HasHealth {
				runtime.Health = "unknown"
			} else {
				runtime.Health = "-"
			}
		}
		if svc, ok := serviceByID[id]; ok {
			runtime.Command = svc.Command
			runtime.Ports = append([]string(nil), svc.Ports...)
		}
		detail := snapshot.Processes[id]
		detail.Runtime = runtime
		snapshot.Processes[id] = detail
		seen[id] = struct{}{}
	}

	for id, svc := range serviceByID {
		detail := snapshot.Processes[id]
		if _, ok := seen[id]; !ok {
			detail.Runtime = ProcessRuntime{
				ID:        id,
				Status:    svc.Status,
				Health:    svc.Health,
				Command:   svc.Command,
				Ports:     append([]string(nil), svc.Ports...),
				Replicas:  1,
				UpdatedAt: timestamp,
			}
			detail.Scalable = svc.Scalable
			detail.ScaleStrategy = svc.ScaleStrategy
		} else {
			runtime := detail.Runtime
			runtime.Command = svc.Command
			runtime.Ports = append([]string(nil), svc.Ports...)
			if runtime.Replicas == 0 {
				runtime.Replicas = 1
			}
			detail.Runtime = runtime
			if detail.ScaleStrategy == "" {
				detail.ScaleStrategy = svc.ScaleStrategy
			}
			if !detail.Scalable {
				detail.Scalable = svc.Scalable
			}
		}
		snapshot.Processes[id] = detail
	}
}
