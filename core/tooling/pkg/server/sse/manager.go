package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/joeblew999/infra/core/tooling/pkg/app"
	"github.com/joeblew999/infra/core/tooling/pkg/orchestrator"
	profiles "github.com/joeblew999/infra/core/tooling/pkg/profiles"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

// Manager coordinates SSE sessions backed by the orchestrator launch adapter.
type Manager struct {
	svc      *app.Service
	mu       sync.RWMutex
	sessions map[string]*session
}

type session struct {
	adapter *orchestrator.StreamAdapter
}

// NewManager constructs an SSE manager using the provided service.
func NewManager(svc *app.Service) *Manager {
	return &Manager{
		svc:      svc,
		sessions: make(map[string]*session),
	}
}

// LaunchRequest describes the JSON payload accepted by LaunchHandler.
type LaunchRequest struct {
	Profile string              `json:"profile"`
	Request types.DeployRequest `json:"request"`
	Timeout time.Duration       `json:"timeout,omitempty"`
}

// LaunchHandler starts a deploy and streams progress/prompt events via SSE.
func (m *Manager) LaunchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	var req LaunchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	opts := orchestrator.DeployOptions{
		ProfileOverride: req.Profile,
		DeployRequest:   req.Request,
		Timeout:         req.Timeout,
	}

	adapter, resultCh, errCh := m.svc.Launch(r.Context(), opts)
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())

	m.mu.Lock()
	m.sessions[sessionID] = &session{adapter: adapter}
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.sessions, sessionID)
		m.mu.Unlock()
		adapter.Close()
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	writeEvent := func(event string, payload any) error {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	if err := writeEvent("init", map[string]string{"id": sessionID}); err != nil {
		return
	}

	progressCh := adapter.Progress
	promptCh := adapter.Prompts

	for progressCh != nil || promptCh != nil {
		select {
		case msg, ok := <-progressCh:
			if !ok {
				progressCh = nil
				continue
			}
			if err := writeEvent("progress", msg); err != nil {
				return
			}
		case prompt, ok := <-promptCh:
			if !ok {
				promptCh = nil
				continue
			}
			if err := writeEvent("prompt", prompt); err != nil {
				return
			}
		case res := <-resultCh:
			if res != nil {
				_ = writeEvent("result", res)
			}
			return
		case err := <-errCh:
			if err != nil {
				_ = writeEvent("error", map[string]string{"error": err.Error()})
			}
			return
		case <-r.Context().Done():
			_ = writeEvent("error", map[string]string{"error": r.Context().Err().Error()})
			return
		}
	}
}

// PromptResponseHandler accepts prompt responses for a running session.
type PromptResponseHandler struct {
	Manager *Manager
}

// ServeHTTP implements http.Handler.
func (h PromptResponseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	var payload types.PromptResponse
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("invalid payload: %v", err), http.StatusBadRequest)
		return
	}
	if payload.ID == "" {
		http.Error(w, "missing prompt id", http.StatusBadRequest)
		return
	}

	h.Manager.mu.RLock()
	sess, ok := h.Manager.sessions[sessionID]
	h.Manager.mu.RUnlock()
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	sess.adapter.Respond(payload.ID, payload)
	w.WriteHeader(http.StatusNoContent)
}

// StatusHandler emits the status snapshot as JSON.
func (m *Manager) StatusHandler(w http.ResponseWriter, r *http.Request) {
	profile := r.URL.Query().Get("profile")
	status, err := m.svc.Status(r.Context(), profiles.ContextOptions{ProfileOverride: profile})
	if err != nil {
		http.Error(w, fmt.Sprintf("status error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, fmt.Sprintf("encode error: %v", err), http.StatusInternalServerError)
		return
	}
}
