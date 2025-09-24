package state

import (
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/joeblew999/infra/pkg/runtime/events"
)

// RuntimeState is the event-sourced view of an infrastructure service.
type RuntimeState struct {
	ID             string
	Name           string
	Description    string
	Icon           string
	Required       bool
	Port           int
	PortLabel      string
	Running        bool
	PID            int
	Ownership      string
	State          string
	Message        string
	LastAction     string
	LastActionKind string
	LastActionAt   time.Time
	UpdatedAt      time.Time
}

var (
	stateMu        sync.RWMutex
	stateByID      = make(map[string]*RuntimeState)
	subscriberMu   sync.Mutex
	subscriberStop func()
)

func init() {
	startSubscriber()
}

func startSubscriber() {
	subscriberMu.Lock()
	defer subscriberMu.Unlock()

	if subscriberStop != nil {
		subscriberStop()
		subscriberStop = nil
	}

	ch, cancel := events.Subscribe(64)
	subscriberStop = cancel

	go func() {
		for evt := range ch {
			applyEvent(evt)
		}
	}()
}

// Snapshot copies the currently known runtime state for all services.
func Snapshot() []RuntimeState {
	stateMu.RLock()
	defer stateMu.RUnlock()

	states := make([]RuntimeState, 0, len(stateByID))
	for _, st := range stateByID {
		if st == nil {
			continue
		}
		copy := *st
		states = append(states, copy)
	}

	sort.Slice(states, func(i, j int) bool {
		if states[i].Required == states[j].Required {
			return states[i].ID < states[j].ID
		}
		return states[i].Required && !states[j].Required
	})

	return states
}

func applyEvent(evt events.Event) {
	switch e := evt.(type) {
	case events.ServiceRegistered:
		applyServiceRegistered(e)
	case events.ServiceStatus:
		applyServiceStatus(e)
	case events.ServiceAction:
		applyServiceAction(e)
	}
}

func applyServiceRegistered(evt events.ServiceRegistered) {
	stateMu.Lock()
	defer stateMu.Unlock()

	st := ensureStateLocked(evt.ServiceID)
	st.ID = evt.ServiceID
	st.Name = evt.Name
	st.Description = evt.Description
	st.Icon = evt.Icon
	st.Required = evt.Required
	st.PortLabel = evt.Port
	if evt.Port != "" {
		if p, err := strconv.Atoi(evt.Port); err == nil {
			st.Port = p
		}
	}
	st.UpdatedAt = evt.Timestamp()
}

func applyServiceStatus(evt events.ServiceStatus) {
	stateMu.Lock()
	defer stateMu.Unlock()

	st := ensureStateLocked(evt.ServiceID)
	st.Running = evt.Running
	st.PID = evt.PID
	if evt.Port > 0 {
		st.Port = evt.Port
		st.PortLabel = strconv.Itoa(evt.Port)
	}
	st.Ownership = evt.Ownership
	st.State = evt.State
	st.Message = evt.Message
	st.UpdatedAt = evt.Timestamp()
	if !evt.Running && evt.Message == "" {
		st.Message = ""
	}
}

func applyServiceAction(evt events.ServiceAction) {
	stateMu.Lock()
	defer stateMu.Unlock()

	st := ensureStateLocked(evt.ServiceID)
	st.LastAction = evt.Message
	st.LastActionKind = evt.Kind
	st.LastActionAt = evt.Timestamp()
	st.UpdatedAt = evt.Timestamp()
}

func ensureStateLocked(id string) *RuntimeState {
	if st, ok := stateByID[id]; ok {
		return st
	}
	st := &RuntimeState{ID: id}
	stateByID[id] = st
	return st
}

// Reset clears all stored state. Intended for tests.
func Reset() {
	stateMu.Lock()
	for k := range stateByID {
		delete(stateByID, k)
	}
	stateMu.Unlock()
	events.Reset()
	startSubscriber()
}
