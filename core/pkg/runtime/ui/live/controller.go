package live

import (
    "bufio"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
    "time"

    controllerspec "github.com/joeblew999/infra/core/controller/pkg/spec"
    runtimeui "github.com/joeblew999/infra/core/pkg/runtime/ui"
)

type controllerEvent struct {
    Reason string                        `json:"reason"`
    Time   time.Time                     `json:"time"`
    State  controllerspec.DesiredState   `json:"state"`
}

// StartControllerStream opens an SSE stream against the controller API and
// applies desired state updates to the live snapshot. When the controller is
// unreachable the method logs an event and retries with exponential backoff.
func (s *Store) StartControllerStream(ctx context.Context, controllerAddr string, reconnect time.Duration) {
    addr := strings.TrimSpace(controllerAddr)
    if addr == "" {
        addr = strings.TrimSpace(os.Getenv("CONTROLLER_ADDR"))
    }
    if addr == "" {
        return
    }
    if reconnect <= 0 {
        reconnect = 3 * time.Second
    }

    go func() {
        backoff := reconnect
        for {
            select {
            case <-ctx.Done():
                return
            default:
            }

            url := buildControllerURL(addr)
            req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
            if err != nil {
                s.AppendEvent(fmt.Sprintf("controller stream error: %v", err))
                if !sleepBackoff(ctx, backoff) {
                    return
                }
                backoff = nextBackoff(backoff, reconnect)
                continue
            }

            resp, err := http.DefaultClient.Do(req)
            if err != nil {
                s.AppendEvent(fmt.Sprintf("controller connect failed: %v", err))
                if !sleepBackoff(ctx, backoff) {
                    return
                }
                backoff = nextBackoff(backoff, reconnect)
                continue
            }

            backoff = reconnect
            scanner := bufio.NewScanner(resp.Body)
            scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
            var data strings.Builder

            process := func() {
                if data.Len() == 0 {
                    return
                }
                payload := data.String()
                data.Reset()
                var event controllerEvent
                if err := json.Unmarshal([]byte(payload), &event); err != nil {
                    s.AppendEvent(fmt.Sprintf("controller decode error: %v", err))
                    return
                }
                s.applyControllerEvent(event)
            }

            for scanner.Scan() {
                line := scanner.Text()
                switch {
                case strings.HasPrefix(line, "data:"):
                    if data.Len() > 0 {
                        data.WriteByte('\n')
                    }
                    data.WriteString(strings.TrimSpace(line[len("data:"):]))
                case strings.TrimSpace(line) == "":
                    process()
                }
            }
            process()
            if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, io.EOF) {
                s.AppendEvent(fmt.Sprintf("controller stream closed: %v", err))
            }
            _ = resp.Body.Close()
            if !sleepBackoff(ctx, backoff) {
                return
            }
            backoff = nextBackoff(backoff, reconnect)
        }
    }()
}

func (s *Store) applyControllerEvent(event controllerEvent) {
    s.Update(func(snapshot *runtimeui.Snapshot) {
        applyControllerSnapshot(snapshot, event)
    })
}

func applyControllerSnapshot(snapshot *runtimeui.Snapshot, event controllerEvent) {
    timestamp := event.Time
    if timestamp.IsZero() {
        timestamp = time.Now()
    }
    message := fmt.Sprintf("controller %s (%d services)", event.Reason, len(event.State.Services))
    prependSnapshotEvent(snapshot, timestamp, message)

    ensureControllerMetric(snapshot, len(event.State.Services), timestamp)

    for _, svc := range event.State.Services {
        updateServiceFromSpec(snapshot, svc, timestamp)
    }
}

func updateServiceFromSpec(snapshot *runtimeui.Snapshot, svc controllerspec.Service, timestamp time.Time) {
    ensureServiceDetails(snapshot)

    desiredRegions := len(svc.Scale.Regions)
    desiredText := fmt.Sprintf("desired replicas: %d regions", desiredRegions)

    for i := range snapshot.Services {
        card := &snapshot.Services[i]
        if strings.EqualFold(card.ID, svc.ID) {
            card.LastEvent = fmt.Sprintf("controller %s", svc.Scale.Strategy)
            card.Description = desiredText
            break
        }
    }

    detail := snapshot.ServiceDetails[svc.ID]
    if detail.Card.ID == "" {
        detail.Card.ID = svc.ID
    }
    detail.Card.Description = desiredText
    detail.Card.ScaleStrategy = svc.Scale.Strategy
    detail.Card.Scalable = len(svc.Scale.Regions) > 0

    notes := []string{}
    for _, region := range svc.Scale.Regions {
        notes = append(notes, fmt.Sprintf("%s: min %d desired %d max %d", strings.ToUpper(region.Name), region.Min, region.Desired, region.Max))
    }
    if svc.Routing.Provider != "" {
        notes = append(notes, fmt.Sprintf("Routing: %s zone %s", svc.Routing.Provider, svc.Routing.Zone))
    }
    if len(notes) > 0 {
        detail.Notes = notes
    }
    snapshot.ServiceDetails[svc.ID] = detail
}

func ensureServiceDetails(snapshot *runtimeui.Snapshot) {
    if snapshot.ServiceDetails == nil {
        snapshot.ServiceDetails = make(map[string]runtimeui.ServiceDetail)
    }
}

func ensureControllerMetric(snapshot *runtimeui.Snapshot, services int, ts time.Time) {
    label := "Controller Services"
    value := fmt.Sprintf("%d", services)
    for i := range snapshot.Metrics {
        metric := &snapshot.Metrics[i]
        if strings.EqualFold(metric.Label, label) {
            metric.Value = value
            metric.Hint = "Services defined in controller desired state"
            return
        }
    }
    snapshot.Metrics = append([]runtimeui.MetricCard{{
        Label: label,
        Value: value,
        Hint:  "Services defined in controller desired state",
    }}, snapshot.Metrics...)
}

func prependSnapshotEvent(snapshot *runtimeui.Snapshot, ts time.Time, message string) {
    entry := runtimeui.EventLog{
        Timestamp: ts.Format("15:04:05"),
        Level:     "info",
        Message:   message,
    }
    snapshot.Events = append([]runtimeui.EventLog{entry}, snapshot.Events...)
    if len(snapshot.Events) > 10 {
        snapshot.Events = snapshot.Events[:10]
    }
    snapshot.GeneratedAt = ts.Local().Round(time.Second)
}

func buildControllerURL(addr string) string {
    base := strings.TrimSpace(addr)
    if base == "" {
        return ""
    }
    if !startsWithHTTP(base) {
        base = "http://" + base
    }
    return strings.TrimRight(base, "/") + "/v1/events"
}

func startsWithHTTP(value string) bool {
    lower := strings.ToLower(value)
    return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func sleepBackoff(ctx context.Context, d time.Duration) bool {
    timer := time.NewTimer(d)
    defer timer.Stop()
    select {
    case <-ctx.Done():
        return false
    case <-timer.C:
        return true
    }
}

func nextBackoff(current, base time.Duration) time.Duration {
    next := current * 2
    if next > 30*time.Second {
        next = 30 * time.Second
    }
    if next < base {
        next = base
    }
    return next
}
