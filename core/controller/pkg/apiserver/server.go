package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	controllerspec "github.com/joeblew999/infra/core/controller/pkg/spec"
)

// Server wraps desired state storage and exposes HTTP handlers.
type Server struct {
	mu         sync.RWMutex
	specPath   string
	state      controllerspec.DesiredState
	watchers   map[chan struct{}]struct{}
	watchersMu sync.RWMutex
}

// New loads the desired state from disk.
func New(specPath string) (*Server, error) {
	state, err := controllerspec.LoadFile(specPath)
	if err != nil {
		return nil, err
	}
	return &Server{specPath: specPath, state: state, watchers: make(map[chan struct{}]struct{})}, nil
}

// Close persists the state back to disk.
func (s *Server) Close() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.specPath, data, 0o644)
}

// Router builds the HTTP router.
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/services", s.handleListServices)
	mux.HandleFunc("/v1/services/update", s.handleUpdateService)
	mux.HandleFunc("/v1/events", s.handleEvents)
	return mux
}

func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	payload := struct {
		Services []controllerspec.Service `json:"services"`
	}{Services: s.state.Services}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// State returns a copy of the desired state.
func (s *Server) State() controllerspec.DesiredState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Subscribe returns a channel that is notified when the desired state changes.
func (s *Server) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	s.watchersMu.Lock()
	s.watchers[ch] = struct{}{}
	s.watchersMu.Unlock()
	return ch, func() {
		s.watchersMu.Lock()
		delete(s.watchers, ch)
		s.watchersMu.Unlock()
		close(ch)
	}
}

func (s *Server) notifyWatchers() {
	s.watchersMu.RLock()
	for ch := range s.watchers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	s.watchersMu.RUnlock()
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch, cancel := s.Subscribe()
	defer cancel()
	writeState := func(reason string) error {
		state := s.State()
		payload := struct {
			Reason string                      `json:"reason"`
			Time   time.Time                   `json:"time"`
			State  controllerspec.DesiredState `json:"state"`
		}{
			Reason: reason,
			Time:   time.Now().UTC(),
			State:  state,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "event: state\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}
	if err := writeState("initial"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case _, ok := <-ch:
			if !ok {
				return
			}
			if err := writeState("update"); err != nil {
				return
			}
		}
	}
}
