package render

import (
	"fmt"
	"strings"
	"time"

	runtimeui "github.com/joeblew999/infra/core/pkg/runtime/ui"
)

// ViewModel is the shared data structure rendered by both web and TUI shells.
type ViewModel struct {
	Title          string
	Snapshot       runtimeui.Snapshot
	Navigation     []runtimeui.NavigationItem
	CurrentPage    string
	Generated      string
	Live           bool
	Debug          bool
	BuildSummary   string
	Processes      []ProcessItem
	CurrentProcess *ProcessItem
}

// NewViewModel normalizes the requested page and produces a view model using
// the provided snapshot and title.
func NewViewModel(title string, snapshot runtimeui.Snapshot, page string, live bool) ViewModel {
	normalized := runtimeui.NormalizePage(snapshot, page)
	processes, currentProcess := buildProcessItems(snapshot, normalized)
	return ViewModel{
		Title:          title,
		Snapshot:       snapshot,
		Navigation:     append([]runtimeui.NavigationItem(nil), snapshot.Navigation...),
		CurrentPage:    normalized,
		Generated:      snapshot.GeneratedAt.Format(time.RFC3339),
		Live:           live,
		Debug:          live,
		BuildSummary:   formatBuildSummary(snapshot.Build),
		Processes:      processes,
		CurrentProcess: currentProcess,
	}
}

// ProcessItem models a single process entry for master-detail layouts.
type ProcessItem struct {
	ID            string
	Route         string
	Title         string
	Runtime       runtimeui.ProcessRuntime
	Logs          runtimeui.ProcessLogs
	Selected      bool
	Scalable      bool
	ScaleStrategy string
}

func formatBuildSummary(info runtimeui.BuildInfo) string {
	parts := make([]string, 0, 3)
	version := strings.TrimSpace(info.Version)
	if version == "" {
		version = "development"
	}
	parts = append(parts, version)

	if info.Revision != "" {
		rev := info.Revision
		if len(rev) > 7 {
			rev = rev[:7]
		}
		if info.Dirty {
			rev = fmt.Sprintf("%s (dirty)", rev)
		}
		parts = append(parts, "commit "+rev)
	} else if info.Dirty {
		parts = append(parts, "dirty")
	}

	if info.BuildTime != "" {
		parts = append(parts, info.BuildTime)
	}

	return strings.Join(parts, " · ")
}

func buildProcessItems(snapshot runtimeui.Snapshot, currentRoute string) ([]ProcessItem, *ProcessItem) {
	processes := make([]ProcessItem, 0)
	var current *ProcessItem

	for _, nav := range snapshot.Navigation {
		if !strings.HasPrefix(nav.Route, "service/") {
			continue
		}
		id := strings.TrimPrefix(nav.Route, "service/")
		detail, ok := snapshot.Processes[id]
		item := ProcessItem{
			ID:       id,
			Route:    nav.Route,
			Title:    nav.Title,
			Selected: nav.Route == currentRoute,
		}
		if ok {
			item.Runtime = detail.Runtime
			item.Logs = detail.Logs
			item.Scalable = detail.Scalable
			item.ScaleStrategy = detail.ScaleStrategy
		}
		if item.Runtime.ID == "" {
			item.Runtime.ID = id
		}
		if item.Runtime.Replicas == 0 {
			item.Runtime.Replicas = 1
		}
		if item.ScaleStrategy == "" {
			if item.Scalable {
				item.ScaleStrategy = "local"
			} else {
				item.ScaleStrategy = "infra"
			}
		}
		if item.Runtime.Command == "" {
			if svcDetail, ok := snapshot.ServiceDetails[nav.Route]; ok {
				item.Runtime.Command = svcDetail.Card.Command
				item.Runtime.Ports = append([]string(nil), svcDetail.Card.Ports...)
				if item.Runtime.Status == "" {
					item.Runtime.Status = svcDetail.Card.Status
				}
				if item.Runtime.Health == "" {
					item.Runtime.Health = svcDetail.Card.Health
				}
			}
		}
		if item.Runtime.UpdatedAt.IsZero() {
			item.Runtime.UpdatedAt = snapshot.GeneratedAt
		}
		processes = append(processes, item)
		if item.Selected {
			current = &processes[len(processes)-1]
		}
	}

	return processes, current
}
