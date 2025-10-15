package ui

func CloneSnapshot(src Snapshot) Snapshot {
	dst := src
	dst.Services = cloneServices(src.Services)
	dst.Metrics = cloneMetrics(src.Metrics)
	dst.Events = cloneEvents(src.Events)
	dst.Tips = append([]string(nil), src.Tips...)
	dst.TextIslands = cloneTextIslands(src.TextIslands)
	dst.Navigation = cloneNavigation(src.Navigation)
	if src.ServiceDetails != nil {
		dst.ServiceDetails = make(map[string]ServiceDetail, len(src.ServiceDetails))
		for key, detail := range src.ServiceDetails {
			dst.ServiceDetails[key] = cloneServiceDetail(detail)
		}
	}
	if src.Processes != nil {
		dst.Processes = cloneProcessDetails(src.Processes)
	}
	return dst
}

func cloneServices(in []ServiceCard) []ServiceCard {
	out := make([]ServiceCard, len(in))
	for i, svc := range in {
		svc.Ports = append([]string(nil), svc.Ports...)
		out[i] = svc
	}
	return out
}

func cloneMetrics(in []MetricCard) []MetricCard {
	out := make([]MetricCard, len(in))
	copy(out, in)
	return out
}

func cloneEvents(in []EventLog) []EventLog {
	out := make([]EventLog, len(in))
	copy(out, in)
	return out
}

func cloneTextIslands(in []TextIsland) []TextIsland {
	out := make([]TextIsland, len(in))
	copy(out, in)
	return out
}

func cloneNavigation(in []NavigationItem) []NavigationItem {
	out := make([]NavigationItem, len(in))
	copy(out, in)
	return out
}

func cloneServiceDetail(in ServiceDetail) ServiceDetail {
	out := in
	out.Card = ServiceCard{
		ID:          in.Card.ID,
		Status:      in.Card.Status,
		Command:     in.Card.Command,
		Ports:       append([]string(nil), in.Card.Ports...),
		Health:      in.Card.Health,
		LastEvent:   in.Card.LastEvent,
		Description: in.Card.Description,
		Scalable:    in.Card.Scalable,
		ScaleStrategy: in.Card.ScaleStrategy,
	}
	out.Notes = append([]string(nil), in.Notes...)
	out.Checklist = append([]string(nil), in.Checklist...)
	return out
}

func cloneProcessDetails(in map[string]ProcessDetail) map[string]ProcessDetail {
	out := make(map[string]ProcessDetail, len(in))
	for key, detail := range in {
		out[key] = ProcessDetail{
			Runtime:       cloneProcessRuntime(detail.Runtime),
			Logs:          cloneProcessLogs(detail.Logs),
			Scalable:      detail.Scalable,
			ScaleStrategy: detail.ScaleStrategy,
		}
	}
	return out
}

func cloneProcessRuntime(in ProcessRuntime) ProcessRuntime {
	out := in
	out.Ports = append([]string(nil), in.Ports...)
	return out
}

func cloneProcessLogs(in ProcessLogs) ProcessLogs {
	out := in
	out.Lines = append([]string(nil), in.Lines...)
	return out
}
