package auth

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/nats-io/nkeys"
	"github.com/joeblew999/infra/pkg/log"
)

// AuthService provides complete authentication functionality
type AuthService struct {
	webauthn        *WebAuthnService
	datastarHandler *DatastarHandlers
}

// NewAuthService creates a complete auth service with all handlers
func NewAuthService(config WebAuthnConfig, users UserStore, sessions SessionStore, webDir string) (*AuthService, error) {
	webauthnService, err := NewWebAuthnService(config, users, sessions)
	if err != nil {
		return nil, err
	}

	datastarHandler := NewDatastarHandlers(webauthnService, webDir)

	return &AuthService{
		webauthn:        webauthnService,
		datastarHandler: datastarHandler,
	}, nil
}

// RegisterRoutes mounts all auth routes on the provided router
func (s *AuthService) RegisterRoutes(r chi.Router) {
	// Datastar SSE routes
	s.datastarHandler.RegisterRoutes(r)

	// Legacy JSON API routes
	r.Post("/register/begin", s.beginRegister)
	r.Post("/register/finish", s.finishRegister)
	r.Post("/login/begin", s.beginLogin)
	r.Post("/login/finish", s.finishLogin)
	
	// Additional routes
	r.Get("/dashboard", s.dashboard)
	r.Post("/logout", s.logout)
	r.Post("/login/conditional", s.conditionalLogin)
	// SECURITY: Test user creation route removed for production safety
}

// NewAuthRouter creates a subrouter with all auth routes configured
func (s *AuthService) NewAuthRouter() chi.Router {
	r := chi.NewRouter()
	s.RegisterRoutes(r)
	return r
}

// Mount mounts the auth subrouter at the specified path
func (s *AuthService) Mount(mainRouter chi.Router, path string) {
	mainRouter.Mount(path, s.NewAuthRouter())
}

// GetUserSession wraps the webauthn service method
func (s *AuthService) GetUserSession(sessionID string) (string, error) {
	return s.webauthn.GetUserSession(sessionID)
}

// DeleteUserSession wraps the webauthn service method
func (s *AuthService) DeleteUserSession(sessionID string) error {
	return s.webauthn.DeleteUserSession(sessionID)
}

// SECURITY: Removed CreateTestUser method - should only be used in tests

// CreateUserSession wraps the webauthn service method
func (s *AuthService) CreateUserSession(sessionID, userID string, ttl time.Duration) error {
	return s.webauthn.CreateUserSession(sessionID, userID, ttl)
}

func (s *AuthService) writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (s *AuthService) decode(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		log.Debug("JSON decode error", "error", err)
		return err
	}
	return nil
}

func (s *AuthService) beginRegister(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	options, token, err := s.webauthn.BeginRegistration(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, map[string]any{
		"token":   token,
		"options": options,
	})
}

func (s *AuthService) finishRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string                                `json:"token"`
		Resp  protocol.ParsedCredentialCreationData `json:"response"`
	}
	if err := s.decode(r, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := s.webauthn.FinishRegistration(req.Token, &req.Resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate NATS credentials for CLI access
	kp, _ := nkeys.CreateUser()
	pub, _ := kp.PublicKey()
	seed, _ := kp.Seed()

	s.writeJSON(w, map[string]any{
		"status":   "ok",
		"username": user.WebAuthnName(),
		"seed":     string(seed),
		"public":   string(pub),
	})
}

func (s *AuthService) beginLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")

	options, token, err := s.webauthn.BeginLogin(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, map[string]any{
		"token":   token,
		"options": options,
	})
}

func (s *AuthService) finishLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string                                 `json:"token"`
		Resp  protocol.ParsedCredentialAssertionData `json:"response"`
	}
	if err := s.decode(r, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, sessionID, err := s.webauthn.FinishLogin(req.Token, &req.Resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate fresh NATS credentials
	kp, _ := nkeys.CreateUser()
	seed, _ := kp.Seed()

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	s.writeJSON(w, map[string]any{
		"status":   "ok",
		"username": user.WebAuthnName(),
		"seed":     string(seed),
	})
}

func (s *AuthService) dashboard(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Please log in first", http.StatusUnauthorized)
		return
	}

	userID, err := s.webauthn.GetUserSession(sessionCookie.Value)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	tpl := template.Must(template.ParseFiles(s.datastarHandler.webDir + "/dashboard.html"))
	tpl.Execute(w, map[string]any{
		"UserID":    userID,
		"SessionID": sessionCookie.Value,
	})
}

func (s *AuthService) logout(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session")
	if err == nil {
		s.webauthn.DeleteUserSession(sessionCookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	s.writeJSON(w, map[string]string{"status": "logged out"})
}

func (s *AuthService) conditionalLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CredentialID string `json:"credentialId"`
	}
	if err := s.decode(r, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// TODO: This should get the actual user from the credential lookup
	// For now, return an error as this is not fully implemented
	http.Error(w, "Conditional login not fully implemented", http.StatusNotImplemented)
}

// SECURITY: Removed createTestUser function - it was a backdoor allowing
// authentication bypass by creating sessions without proper WebAuthn flow