package events

import (
	"sync"
	"time"
)

type EventType string

const (
	EventServiceRegistered EventType = "service.registered"
	EventServiceStatus     EventType = "service.status"
	EventServiceAction     EventType = "service.action"
)

type Event interface {
	Type() EventType
	Timestamp() time.Time
}

type ServiceRegistered struct {
	TS          time.Time
	ServiceID   string
	Name        string
	Description string
	Icon        string
	Required    bool
	Port        string
}

func (e ServiceRegistered) Type() EventType      { return EventServiceRegistered }
func (e ServiceRegistered) Timestamp() time.Time { return e.TS }

type ServiceStatus struct {
	TS        time.Time
	ServiceID string
	Running   bool
	PID       int
	Port      int
	Ownership string
	State     string
	Message   string
}

func (e ServiceStatus) Type() EventType      { return EventServiceStatus }
func (e ServiceStatus) Timestamp() time.Time { return e.TS }

type ServiceAction struct {
	TS        time.Time
	ServiceID string
	Message   string
	Kind      string
}

func (e ServiceAction) Type() EventType      { return EventServiceAction }
func (e ServiceAction) Timestamp() time.Time { return e.TS }

type Dispatcher struct {
	mu          sync.RWMutex
	subscribers map[int]chan Event
	nextID      int
}

var globalDispatcher = &Dispatcher{subscribers: make(map[int]chan Event)}

func Publish(evt Event) {
	if evt == nil {
		return
	}
	globalDispatcher.mu.RLock()
	defer globalDispatcher.mu.RUnlock()
	for id, ch := range globalDispatcher.subscribers {
		select {
		case ch <- evt:
		default:
			_ = id
		}
	}
}

func Subscribe(buffer int) (<-chan Event, func()) {
	if buffer <= 0 {
		buffer = 1
	}
	ch := make(chan Event, buffer)
	globalDispatcher.mu.Lock()
	id := globalDispatcher.nextID
	globalDispatcher.nextID++
	globalDispatcher.subscribers[id] = ch
	globalDispatcher.mu.Unlock()

	cancel := func() {
		globalDispatcher.mu.Lock()
		if subscriber, ok := globalDispatcher.subscribers[id]; ok {
			delete(globalDispatcher.subscribers, id)
			close(subscriber)
		}
		globalDispatcher.mu.Unlock()
	}
	return ch, cancel
}

func Reset() {
	globalDispatcher.mu.Lock()
	for id, ch := range globalDispatcher.subscribers {
		delete(globalDispatcher.subscribers, id)
		close(ch)
	}
	globalDispatcher.mu.Unlock()
}
