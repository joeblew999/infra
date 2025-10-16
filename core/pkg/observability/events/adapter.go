package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/joeblew999/infra/core/pkg/runtime/process"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

// Adapter polls process-compose for state changes and publishes events to NATS.
type Adapter struct {
	composePort  int
	natsURL      string
	pollInterval time.Duration
	nc           *nats.Conn
	js           nats.JetStreamContext
	lastStates   map[string]process.ComposeProcessState
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// Config configures the event adapter.
type Config struct {
	ComposePort  int           // Port for process-compose API (default: 28081)
	NATSURL      string        // NATS server URL (default: nats://127.0.0.1:4222)
	PollInterval time.Duration // How often to poll for state changes (default: 2s)
}

// NewAdapter creates a new event adapter.
func NewAdapter(cfg Config) (*Adapter, error) {
	if cfg.ComposePort == 0 {
		cfg.ComposePort = 28081
	}
	if cfg.NATSURL == "" {
		cfg.NATSURL = "nats://127.0.0.1:4222"
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 2 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Adapter{
		composePort:  cfg.ComposePort,
		natsURL:      cfg.NATSURL,
		pollInterval: cfg.PollInterval,
		lastStates:   make(map[string]process.ComposeProcessState),
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

// Start connects to NATS and begins polling process-compose.
func (a *Adapter) Start() error {
	// Connect to NATS
	nc, err := nats.Connect(a.natsURL,
		nats.Name("core-event-adapter"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(1*time.Second),
	)
	if err != nil {
		return fmt.Errorf("connect to nats: %w", err)
	}
	a.nc = nc

	// Setup JetStream
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return fmt.Errorf("setup jetstream: %w", err)
	}
	a.js = js

	// Ensure the stream exists
	if err := a.ensureStream(); err != nil {
		nc.Close()
		return fmt.Errorf("ensure stream: %w", err)
	}

	log.Info().
		Str("nats_url", a.natsURL).
		Int("compose_port", a.composePort).
		Dur("poll_interval", a.pollInterval).
		Msg("Event adapter started")

	// Start polling in background
	go a.pollLoop()

	return nil
}

// Stop gracefully stops the adapter.
func (a *Adapter) Stop() error {
	a.cancel()
	if a.nc != nil {
		a.nc.Close()
	}
	log.Info().Msg("Event adapter stopped")
	return nil
}

// ensureStream creates the NATS JetStream stream for process events if it doesn't exist.
func (a *Adapter) ensureStream() error {
	streamName := "PROCESS_EVENTS"

	// Check if stream exists
	_, err := a.js.StreamInfo(streamName)
	if err == nil {
		return nil // Stream already exists
	}

	// Create stream
	_, err = a.js.AddStream(&nats.StreamConfig{
		Name:        streamName,
		Description: "Process lifecycle and health events from process-compose",
		Subjects:    []string{"core.process.>"},
		Retention:   nats.LimitsPolicy,
		MaxAge:      24 * time.Hour, // Retain events for 24 hours
		Storage:     nats.FileStorage,
		Replicas:    1,
	})
	if err != nil {
		return fmt.Errorf("create stream: %w", err)
	}

	log.Info().Str("stream", streamName).Msg("Created JetStream stream")
	return nil
}

// pollLoop continuously polls process-compose and detects changes.
func (a *Adapter) pollLoop() {
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	// Initialize state on first poll
	a.poll()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.poll()
		}
	}
}

// poll fetches current state and publishes change events.
func (a *Adapter) poll() {
	states, err := process.FetchComposeProcesses(a.ctx, a.composePort)
	if err != nil {
		if err == process.ErrComposeUnavailable {
			log.Debug().Msg("Process-compose unavailable, skipping poll")
			return
		}
		log.Error().Err(err).Msg("Failed to fetch process states")
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Create a map of current states
	currentStates := make(map[string]process.ComposeProcessState)
	for _, state := range states {
		key := a.processKey(state)
		currentStates[key] = state
	}

	// Detect changes
	for key, current := range currentStates {
		last, existed := a.lastStates[key]

		if !existed {
			// New process detected
			a.publishEvent(Event{
				Type:      EventTypeStarted,
				Process:   current.Name,
				Namespace: current.Namespace,
				Timestamp: time.Now(),
				State:     current,
			})
		} else {
			// Check for state transitions
			a.detectTransitions(last, current)
		}
	}

	// Detect removed processes
	for key, last := range a.lastStates {
		if _, exists := currentStates[key]; !exists {
			a.publishEvent(Event{
				Type:      EventTypeStopped,
				Process:   last.Name,
				Namespace: last.Namespace,
				Timestamp: time.Now(),
				State:     last,
			})
		}
	}

	// Update last states
	a.lastStates = currentStates
}

// detectTransitions identifies state changes and publishes events.
func (a *Adapter) detectTransitions(last, current process.ComposeProcessState) {
	now := time.Now()

	// Running state changed
	if !last.IsRunning && current.IsRunning {
		a.publishEvent(Event{
			Type:      EventTypeStarted,
			Process:   current.Name,
			Namespace: current.Namespace,
			Timestamp: now,
			State:     current,
		})
	} else if last.IsRunning && !current.IsRunning {
		// Determine if crashed or stopped gracefully
		eventType := EventTypeStopped
		if current.ExitCode != 0 {
			eventType = EventTypeCrashed
		}
		a.publishEvent(Event{
			Type:      eventType,
			Process:   current.Name,
			Namespace: current.Namespace,
			Timestamp: now,
			State:     current,
			ExitCode:  &current.ExitCode,
		})
	}

	// Restart count changed
	if current.Restarts > last.Restarts {
		a.publishEvent(Event{
			Type:      EventTypeRestarted,
			Process:   current.Name,
			Namespace: current.Namespace,
			Timestamp: now,
			State:     current,
			Restarts:  current.Restarts,
		})
	}

	// Health state changed (if health probe exists)
	if current.HasHealthProbe && last.Health != current.Health {
		eventType := EventTypeHealthy
		if current.Health != "Ready" {
			eventType = EventTypeUnhealthy
		}
		a.publishEvent(Event{
			Type:      eventType,
			Process:   current.Name,
			Namespace: current.Namespace,
			Timestamp: now,
			State:     current,
			Health:    current.Health,
		})
	}

	// Status changed
	if last.Status != current.Status {
		a.publishEvent(Event{
			Type:      EventTypeStatusChanged,
			Process:   current.Name,
			Namespace: current.Namespace,
			Timestamp: now,
			State:     current,
			OldStatus: last.Status,
			NewStatus: current.Status,
		})
	}
}

// publishEvent publishes an event to NATS JetStream.
func (a *Adapter) publishEvent(evt Event) {
	subject := evt.Subject()
	data, err := json.Marshal(evt)
	if err != nil {
		log.Error().Err(err).Str("event_type", string(evt.Type)).Msg("Failed to marshal event")
		return
	}

	_, err = a.js.Publish(subject, data)
	if err != nil {
		log.Error().Err(err).Str("subject", subject).Msg("Failed to publish event")
		return
	}

	log.Debug().
		Str("subject", subject).
		Str("event_type", string(evt.Type)).
		Str("process", evt.Process).
		Msg("Published event")
}

// processKey generates a unique key for a process state.
func (a *Adapter) processKey(state process.ComposeProcessState) string {
	if state.Namespace != "" {
		return state.Namespace + "/" + state.Name
	}
	return state.Name
}
