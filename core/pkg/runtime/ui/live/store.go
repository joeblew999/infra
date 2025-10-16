package live

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/joeblew999/infra/core/pkg/runtime/observability"
	runtimeprocess "github.com/joeblew999/infra/core/pkg/runtime/process"
	runtimeui "github.com/joeblew999/infra/core/pkg/runtime/ui"
)

// Store maintains an in-memory snapshot and broadcasts updates to subscribers.
type Store struct {
	mu       sync.RWMutex
	snapshot runtimeui.Snapshot

	subs   map[int]chan runtimeui.Snapshot
	nextID int
	ticks  int

	composePort int
}

// NewStore constructs a store with an initial snapshot.
func NewStore(initial runtimeui.Snapshot) *Store {
	return &Store{
		snapshot: runtimeui.CloneSnapshot(initial),
		subs:     make(map[int]chan runtimeui.Snapshot),
	}
}

// Snapshot returns a cloned snapshot suitable for read-only consumers.
func (s *Store) Snapshot() runtimeui.Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return runtimeui.CloneSnapshot(s.snapshot)
}

// Subscribe registers a listener for snapshot updates. The returned cancel
// function must be called to release resources.
func (s *Store) Subscribe() (<-chan runtimeui.Snapshot, func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan runtimeui.Snapshot, 1)
	ch <- runtimeui.CloneSnapshot(s.snapshot)

	id := s.nextID
	s.nextID++
	s.subs[id] = ch

	cancel := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if sub, ok := s.subs[id]; ok {
			delete(s.subs, id)
			close(sub)
		}
	}
	return ch, cancel
}

// Update applies a mutation function to the snapshot and notifies subscribers.
func (s *Store) Update(fn func(*runtimeui.Snapshot)) {
	s.mu.Lock()
	next := runtimeui.CloneSnapshot(s.snapshot)
	fn(&next)
	s.snapshot = next

	subs := make([]chan runtimeui.Snapshot, 0, len(s.subs))
	for _, ch := range s.subs {
		subs = append(subs, ch)
	}
	s.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- runtimeui.CloneSnapshot(next):
		default:
		}
	}
}

// StartSimulator mutates the snapshot at the provided interval to simulate
// runtime activity. The simulation stops when ctx is cancelled.
func (s *Store) StartSimulator(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.Update(func(snapshot *runtimeui.Snapshot) {
					s.ticks++
					mutateGeneratedAt(snapshot)
					mutateServices(snapshot)
					mutateMetrics(snapshot)
					mutateEvents(snapshot, s.ticks)
				})
			}
		}
	}()
}

// AppendEvent inserts a manual event into the snapshot and notifies
// subscribers.
func (s *Store) AppendEvent(message string) {
	s.Update(func(snapshot *runtimeui.Snapshot) {
		mutateGeneratedAt(snapshot)
		addEvent(snapshot, message)
	})
}

// ApplyProcessLogs updates the cached log buffer for the provided process ID.
func (s *Store) ApplyProcessLogs(processID string, logs []string, offset, limit int, truncated bool) {
	s.Update(func(snapshot *runtimeui.Snapshot) {
		runtimeui.ApplyProcessLogs(snapshot, processID, logs, offset, limit, truncated)
	})
}

// setComposePort records the Process Compose port used for live actions.
func (s *Store) setComposePort(port int) {
	s.mu.Lock()
	s.composePort = port
	s.mu.Unlock()
}

// ComposePort returns the port the store uses when communicating with Process Compose.
func (s *Store) ComposePort() int {
	s.mu.RLock()
	port := s.composePort
	s.mu.RUnlock()
	if port <= 0 {
		return runtimeprocess.ComposePort(nil)
	}
	return port
}

func mutateServices(snapshot *runtimeui.Snapshot) {
	for i := range snapshot.Services {
		svc := &snapshot.Services[i]
		switch svc.Status {
		case "running":
			if rand.Intn(10) == 0 {
				svc.Status = "restarting"
				svc.Health = "degraded"
				svc.LastEvent = "restart requested"
			}
		case "restarting":
			svc.Status = "running"
			svc.Health = "healthy"
			svc.LastEvent = "back online"
		default:
			svc.LastEvent = time.Now().Format("15:04:05") + " heartbeat"
		}
	}
}

func mutateMetrics(snapshot *runtimeui.Snapshot) {
	if len(snapshot.Metrics) == 0 {
		return
	}
	metric := &snapshot.Metrics[0]
	metric.Value = fmt.Sprintf("%d", rand.Intn(5)+3)
}

func mutateEvents(snapshot *runtimeui.Snapshot, tick int) {
	addEvent(snapshot, fmt.Sprintf("simulator tick #%d", tick))
}

func mutateGeneratedAt(snapshot *runtimeui.Snapshot) {
	snapshot.GeneratedAt = time.Now().Round(time.Second)
}

func addEvent(snapshot *runtimeui.Snapshot, message string) {
	snapshot.Events = append([]runtimeui.EventLog{{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     "info",
		Message:   message,
	}}, snapshot.Events...)
	if len(snapshot.Events) > 100 {
		snapshot.Events = snapshot.Events[:100]
	}
}

// StartEventStream subscribes to process events from NATS and adds them to the
// event log. This provides rich event history showing process lifecycle transitions.
func (s *Store) StartEventStream(ctx context.Context, natsURL string) error {
	consumer, err := observability.NewConsumer(natsURL)
	if err != nil {
		return fmt.Errorf("create event consumer: %w", err)
	}

	if err := consumer.Connect(); err != nil {
		return fmt.Errorf("connect to nats: %w", err)
	}

	// Subscribe to all process events
	if err := consumer.SubscribeAll(func(evt observability.Event) error {
		s.appendObservabilityEvent(evt)
		return nil
	}); err != nil {
		return fmt.Errorf("subscribe to events: %w", err)
	}

	// Close consumer when context is cancelled
	go func() {
		<-ctx.Done()
		consumer.Close()
	}()

	return nil
}

// appendObservabilityEvent adds an observability event to the snapshot event log.
func (s *Store) appendObservabilityEvent(evt observability.Event) {
	s.Update(func(snapshot *runtimeui.Snapshot) {
		icon := eventIcon(evt.Type)
		message := fmt.Sprintf("%s %s", icon, evt.String())

		snapshot.Events = append([]runtimeui.EventLog{{
			Timestamp: evt.Timestamp.Format("15:04:05"),
			Level:     string(evt.Severity()),
			Message:   message,
		}}, snapshot.Events...)

		if len(snapshot.Events) > 100 {
			snapshot.Events = snapshot.Events[:100]
		}
	})
}

// eventIcon returns an icon for the given event type.
func eventIcon(eventType observability.EventType) string {
	switch eventType {
	case observability.EventTypeStarted:
		return "▶️"
	case observability.EventTypeStopped:
		return "⏹️"
	case observability.EventTypeCrashed:
		return "❌"
	case observability.EventTypeRestarted:
		return "🔄"
	case observability.EventTypeHealthy:
		return "✅"
	case observability.EventTypeUnhealthy:
		return "⚠️"
	case observability.EventTypeStatusChanged:
		return "📊"
	default:
		return "ℹ️"
	}
}
