package live

import (
	"context"
	"errors"
	"fmt"
	"time"

	runtimeprocess "github.com/joeblew999/infra/core/pkg/runtime/process"
	runtimeui "github.com/joeblew999/infra/core/pkg/runtime/ui"
)

const composeLogTailLines = 120

// StartComposeSync polls the Process Compose HTTP API and applies the live
// service state to the snapshot store. When the supervisor is unavailable the
// method leaves the previous snapshot intact, emitting an event only for
// unexpected failures.
func (s *Store) StartComposeSync(ctx context.Context, port int, interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	s.setComposePort(port)
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				states, err := runtimeprocess.FetchComposeProcesses(ctx, port)
				if err != nil {
					if errors.Is(err, runtimeprocess.ErrComposeUnavailable) {
						continue
					}
					s.AppendEvent(fmt.Sprintf("process-compose error: %v", err))
					continue
				}
				serviceStates := runtimeui.ServiceStatusesFromCompose(states)
				logCaptures := make(map[string]struct {
					lines     []string
					truncated bool
				})
				reportedErr := false
				for _, st := range states {
					if composeLogTailLines <= 0 {
						continue
					}
					lines, err := runtimeprocess.FetchComposeProcessLogs(ctx, port, st.Name, 0, composeLogTailLines)
					if err != nil {
						if !errors.Is(err, runtimeprocess.ErrComposeUnavailable) && !reportedErr {
							s.AppendEvent(fmt.Sprintf("process log fetch failed for %s: %v", st.Name, err))
							reportedErr = true
						}
						continue
					}
					truncated := composeLogTailLines > 0 && len(lines) >= composeLogTailLines
					logCaptures[st.Name] = struct {
						lines     []string
						truncated bool
					}{lines: lines, truncated: truncated}
				}
				s.Update(func(snapshot *runtimeui.Snapshot) {
					runtimeui.ApplyServiceStatus(snapshot, serviceStates)
					for name, capture := range logCaptures {
						runtimeui.ApplyProcessLogs(snapshot, name, capture.lines, 0, composeLogTailLines, capture.truncated)
					}
				})
			}
		}
	}()
}
