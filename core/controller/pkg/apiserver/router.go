package apiserver

import (
	"encoding/json"
	"net/http"

	controllerspec "github.com/joeblew999/infra/core/controller/pkg/spec"
)

// UpdateRequest represents a partial update to the desired state.
type UpdateRequest struct {
	Service controllerspec.Service `json:"service"`
}

func (s *Server) handleUpdateService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Service.ID == "" {
		http.Error(w, "service id required", http.StatusBadRequest)
		return
	}
	state := controllerspec.DesiredState{Services: []controllerspec.Service{req.Service}}
	if err := state.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for idx, svc := range s.state.Services {
		if svc.ID == req.Service.ID {
			s.state.Services[idx] = req.Service
			s.notifyWatchers()
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	s.state.Services = append(s.state.Services, req.Service)
	s.notifyWatchers()
	w.WriteHeader(http.StatusCreated)
}
