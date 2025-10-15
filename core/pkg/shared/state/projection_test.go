package state

import (
	"testing"

	"github.com/joeblew999/infra/core/pkg/shared/events"
)

type counter struct {
	Count int
}

func TestProjectionApply(t *testing.T) {
	reducer := func(s counter, evt events.Envelope) (counter, error) {
		s.Count++
		return s, nil
	}
	proj := NewProjection(counter{}, reducer)
	evt, err := events.Wrap("core.counter", "increment", nil)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	if err := proj.Apply(evt); err != nil {
		t.Fatalf("apply: %v", err)
	}
	snapshot := proj.Snapshot()
	if snapshot.Count != 1 {
		t.Fatalf("expected count 1 got %d", snapshot.Count)
	}
}
