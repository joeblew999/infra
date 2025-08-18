package auth

import (
	"errors"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/nats-io/nats.go"
)

// SessionStore interface for session storage operations
type SessionStore interface {
	StoreWebAuthnSession(token string, session webauthn.SessionData) error
	GetWebAuthnSession(token string) (*webauthn.SessionData, error)
	DeleteWebAuthnSession(token string) error
	CreateUserSession(sessionID, userID string, ttl time.Duration) error
	GetUserSession(sessionID string) (string, error)
	DeleteUserSession(sessionID string) error
}

// InMemorySessionStore implements SessionStore using in-memory storage
type InMemorySessionStore struct {
	webauthnSessions map[string]webauthn.SessionData
	userSessions     map[string]string
}

// NewInMemorySessionStore creates a new in-memory session store
func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		webauthnSessions: make(map[string]webauthn.SessionData),
		userSessions:     make(map[string]string),
	}
}

// StoreWebAuthnSession stores a WebAuthn session
func (s *InMemorySessionStore) StoreWebAuthnSession(token string, session webauthn.SessionData) error {
	s.webauthnSessions[token] = session
	return nil
}

// GetWebAuthnSession retrieves a WebAuthn session
func (s *InMemorySessionStore) GetWebAuthnSession(token string) (*webauthn.SessionData, error) {
	session, exists := s.webauthnSessions[token]
	if !exists {
		return nil, errors.New("session not found")
	}
	return &session, nil
}

// DeleteWebAuthnSession deletes a WebAuthn session
func (s *InMemorySessionStore) DeleteWebAuthnSession(token string) error {
	delete(s.webauthnSessions, token)
	return nil
}

// CreateUserSession creates a user session
func (s *InMemorySessionStore) CreateUserSession(sessionID, userID string, ttl time.Duration) error {
	s.userSessions[sessionID] = userID
	return nil
}

// GetUserSession retrieves a user session
func (s *InMemorySessionStore) GetUserSession(sessionID string) (string, error) {
	userID, exists := s.userSessions[sessionID]
	if !exists {
		return "", errors.New("session not found")
	}
	return userID, nil
}

// DeleteUserSession deletes a user session
func (s *InMemorySessionStore) DeleteUserSession(sessionID string) error {
	delete(s.userSessions, sessionID)
	return nil
}

// NATSSessionStore implements SessionStore using NATS KV
type NATSSessionStore struct {
	kv               nats.KeyValue
	webauthnSessions *InMemorySessionStore // Fallback for WebAuthn sessions
}

// NewNATSSessionStore creates a new NATS-based session store
func NewNATSSessionStore(nc *nats.Conn) (*NATSSessionStore, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	kv, err := js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket:   "auth_sessions",
		Replicas: 1,
		TTL:      30 * time.Minute,
	})
	if err != nil {
		return nil, err
	}

	return &NATSSessionStore{
		kv:               kv,
		webauthnSessions: NewInMemorySessionStore(),
	}, nil
}

// StoreWebAuthnSession stores a WebAuthn session (in-memory for short-term use)
func (s *NATSSessionStore) StoreWebAuthnSession(token string, session webauthn.SessionData) error {
	return s.webauthnSessions.StoreWebAuthnSession(token, session)
}

// GetWebAuthnSession retrieves a WebAuthn session
func (s *NATSSessionStore) GetWebAuthnSession(token string) (*webauthn.SessionData, error) {
	return s.webauthnSessions.GetWebAuthnSession(token)
}

// DeleteWebAuthnSession deletes a WebAuthn session
func (s *NATSSessionStore) DeleteWebAuthnSession(token string) error {
	return s.webauthnSessions.DeleteWebAuthnSession(token)
}

// CreateUserSession creates a user session in NATS KV
func (s *NATSSessionStore) CreateUserSession(sessionID, userID string, ttl time.Duration) error {
	_, err := s.kv.Put(sessionID, []byte(userID))
	return err
}

// GetUserSession retrieves a user session from NATS KV
func (s *NATSSessionStore) GetUserSession(sessionID string) (string, error) {
	entry, err := s.kv.Get(sessionID)
	if err != nil {
		return "", err
	}
	return string(entry.Value()), nil
}

// DeleteUserSession deletes a user session from NATS KV
func (s *NATSSessionStore) DeleteUserSession(sessionID string) error {
	return s.kv.Delete(sessionID)
}