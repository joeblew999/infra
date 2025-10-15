package state

import (
	"sync"

	"github.com/joeblew999/infra/core/pkg/shared/events"
)

// Reducer applies event envelopes to mutate state. Returning an error aborts
// the projection update.
type Reducer[T any] func(T, events.Envelope) (T, error)

// Projection maintains a copy of state by applying incoming events through a
// reducer function.
type Projection[T any] struct {
	mu      sync.RWMutex
	current T
	reduce  Reducer[T]
}

// NewProjection constructs a projection with an initial state and reducer.
func NewProjection[T any](initial T, reducer Reducer[T]) *Projection[T] {
	return &Projection[T]{current: initial, reduce: reducer}
}

// Apply feeds an event into the projection.
func (p *Projection[T]) Apply(evt events.Envelope) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	next, err := p.reduce(p.current, evt)
	if err != nil {
		return err
	}
	p.current = next
	return nil
}

// Snapshot returns a copy of the current state.
func (p *Projection[T]) Snapshot() T {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.current
}
