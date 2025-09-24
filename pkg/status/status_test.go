package status

import (
	"testing"
	"time"

	runtimeevents "github.com/joeblew999/infra/pkg/runtime/events"
	servicestate "github.com/joeblew999/infra/pkg/service/state"
)

func TestGetCurrentStatusReflectsRuntimeEvents(t *testing.T) {
	servicestate.Reset()

	runtimeevents.Publish(runtimeevents.ServiceRegistered{
		TS:          time.Now(),
		ServiceID:   "web",
		Name:        "Web Server",
		Description: "Test web",
		Icon:        "🌐",
		Required:    true,
		Port:        "1337",
	})

	runtimeevents.Publish(runtimeevents.ServiceStatus{
		TS:        time.Now(),
		ServiceID: "web",
		Running:   true,
		PID:       4242,
		Port:      1337,
		Ownership: "this",
		State:     "running",
	})

	runtimeevents.Publish(runtimeevents.ServiceAction{
		TS:        time.Now(),
		ServiceID: "web",
		Kind:      "started",
		Message:   "Service running on port 1337",
	})

	deadline := time.After(time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for status snapshot")
		default:
			sys, err := GetCurrentStatus()
			if err != nil {
				t.Fatalf("GetCurrentStatus returned error: %v", err)
			}
			if len(sys.Services) == 0 {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			svc := sys.Services[0]
			if svc.State != "running" {
				t.Fatalf("expected state running, got %s", svc.State)
			}
			if svc.PID != 4242 {
				t.Fatalf("expected pid 4242, got %d", svc.PID)
			}
			if svc.LastAction == "" {
				t.Fatalf("expected last action to be recorded")
			}
			return
		}
	}
}
