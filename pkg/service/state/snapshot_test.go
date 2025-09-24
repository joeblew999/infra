package state

import (
	"testing"
	"time"

	runtimeevents "github.com/joeblew999/infra/pkg/runtime/events"
)

func TestSnapshotTracksEventStream(t *testing.T) {
	Reset()

	runtimeevents.Publish(runtimeevents.ServiceRegistered{
		TS:          time.Now(),
		ServiceID:   "web",
		Name:        "Web Server",
		Description: "test",
		Icon:        "🌐",
		Required:    true,
		Port:        "8080",
	})

	runtimeevents.Publish(runtimeevents.ServiceStatus{
		TS:        time.Now(),
		ServiceID: "web",
		Running:   true,
		PID:       100,
		Port:      8080,
		Ownership: "this",
		State:     "running",
		Message:   "running",
	})

	runtimeevents.Publish(runtimeevents.ServiceAction{
		TS:        time.Now(),
		ServiceID: "web",
		Kind:      "started",
		Message:   "Service started",
	})

	deadline := time.After(time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("state did not observe events")
		default:
			states := Snapshot()
			if len(states) == 0 {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			st := states[0]
			if st.ID != "web" {
				t.Fatalf("unexpected service id: %s", st.ID)
			}
			if !st.Running || st.PID != 100 || st.Port != 8080 {
				t.Fatalf("unexpected runtime state: %#v", st)
			}
			if st.LastAction != "Service started" {
				t.Fatalf("expected last action to be recorded, got %q", st.LastAction)
			}
			if st.LastActionKind != "started" {
				t.Fatalf("expected last action kind to be recorded, got %q", st.LastActionKind)
			}
			if st.Ownership != "this" {
				t.Fatalf("expected ownership 'this', got %q", st.Ownership)
			}
			if st.State != "running" {
				t.Fatalf("expected state 'running', got %q", st.State)
			}
			return
		}
	}
}

func TestResetClearsState(t *testing.T) {
	Reset()

	runtimeevents.Publish(runtimeevents.ServiceRegistered{
		TS:          time.Now(),
		ServiceID:   "web",
		Name:        "Web",
		Description: "desc",
	})

	waitForService(t)

	Reset()

	if got := Snapshot(); len(got) != 0 {
		t.Fatalf("expected empty state after reset, got %v", got)
	}
}

func waitForService(t *testing.T) {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for service registration")
		default:
			states := Snapshot()
			if len(states) > 0 {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}
