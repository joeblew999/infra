package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// WebAuthnConfig holds configuration for WebAuthn setup
type WebAuthnConfig struct {
	RPDisplayName string
	RPID          string
	RPOrigins     []string
}

// WebAuthnService handles WebAuthn operations
type WebAuthnService struct {
	webauthn *webauthn.WebAuthn
	users    UserStore
	sessions SessionStore
}

// NewWebAuthnService creates a new WebAuthn service
func NewWebAuthnService(config WebAuthnConfig, users UserStore, sessions SessionStore) (*WebAuthnService, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: config.RPDisplayName,
		RPID:          config.RPID,
		RPOrigins:     config.RPOrigins,
		Timeouts: webauthn.TimeoutsConfig{
			Login: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    120 * time.Second,
			},
			Registration: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    120 * time.Second,
			},
		},
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			// No AuthenticatorAttachment specified - allows all authenticator types
			// This works best across Safari, Chrome, Firefox, Edge
			RequireResidentKey:      protocol.ResidentKeyNotRequired(),
			UserVerification:        protocol.VerificationDiscouraged, // Most compatible setting
		},
	})
	if err != nil {
		return nil, err
	}

	return &WebAuthnService{
		webauthn: wa,
		users:    users,
		sessions: sessions,
	}, nil
}

// BeginRegistration starts the WebAuthn registration process
func (w *WebAuthnService) BeginRegistration(username string) (*protocol.CredentialCreation, string, error) {
	user, err := w.users.GetOrCreateUser(username)
	if err != nil {
		return nil, "", err
	}

	options, session, err := w.webauthn.BeginRegistration(user)
	if err != nil {
		return nil, "", err
	}

	token := uuid.New().String()
	if err := w.sessions.StoreWebAuthnSession(token, *session); err != nil {
		return nil, "", err
	}

	return options, token, nil
}

// BeginRegistrationWithPlatform starts registration with platform authenticator preference
func (w *WebAuthnService) BeginRegistrationWithPlatform(username string, usePlatform bool) (*protocol.CredentialCreation, string, error) {
	user, err := w.users.GetOrCreateUser(username)
	if err != nil {
		return nil, "", err
	}

	// Create custom options with platform preference
	options, session, err := w.webauthn.BeginRegistration(user)
	if err != nil {
		return nil, "", err
	}

	// Modify authenticator selection based on platform preference
	if usePlatform {
		options.Response.AuthenticatorSelection.AuthenticatorAttachment = protocol.Platform
		options.Response.AuthenticatorSelection.UserVerification = protocol.VerificationPreferred
	} else {
		// Allow all authenticator types for maximum compatibility
		options.Response.AuthenticatorSelection.UserVerification = protocol.VerificationDiscouraged
	}

	token := uuid.New().String()
	if err := w.sessions.StoreWebAuthnSession(token, *session); err != nil {
		return nil, "", err
	}

	return options, token, nil
}

// FinishRegistration completes the WebAuthn registration process
func (w *WebAuthnService) FinishRegistration(token string, response *protocol.ParsedCredentialCreationData) (*User, error) {
	session, err := w.sessions.GetWebAuthnSession(token)
	if err != nil {
		return nil, err
	}

	user, err := w.users.GetUserByID(string(session.UserID))
	if err != nil {
		return nil, err
	}

	credential, err := w.webauthn.CreateCredential(user, *session, response)
	if err != nil {
		return nil, err
	}

	if err := w.users.AddCredential(user.WebAuthnName(), credential); err != nil {
		return nil, err
	}

	w.sessions.DeleteWebAuthnSession(token)
	return user, nil
}

// BeginLogin starts the WebAuthn login process
func (w *WebAuthnService) BeginLogin(username string) (*protocol.CredentialAssertion, string, error) {
	user, err := w.users.GetUser(username)
	if err != nil {
		return nil, "", err
	}

	options, session, err := w.webauthn.BeginLogin(user)
	if err != nil {
		return nil, "", err
	}

	token := uuid.New().String()
	if err := w.sessions.StoreWebAuthnSession(token, *session); err != nil {
		return nil, "", err
	}

	return options, token, nil
}

// FinishLogin completes the WebAuthn login process
func (w *WebAuthnService) FinishLogin(token string, response *protocol.ParsedCredentialAssertionData) (*User, string, error) {
	session, err := w.sessions.GetWebAuthnSession(token)
	if err != nil {
		return nil, "", err
	}

	user, err := w.users.GetUserByID(string(session.UserID))
	if err != nil {
		return nil, "", err
	}

	_, err = w.webauthn.ValidateLogin(user, *session, response)
	if err != nil {
		return nil, "", err
	}

	w.sessions.DeleteWebAuthnSession(token)

	// Create user session
	sessionID := uuid.New().String()
	if err := w.sessions.CreateUserSession(sessionID, user.WebAuthnName(), 30*time.Minute); err != nil {
		return nil, "", err
	}

	return user, sessionID, nil
}

// GetUserSession retrieves the user for a given session ID
func (w *WebAuthnService) GetUserSession(sessionID string) (string, error) {
	return w.sessions.GetUserSession(sessionID)
}

// DeleteUserSession removes a user session
func (w *WebAuthnService) DeleteUserSession(sessionID string) error {
	return w.sessions.DeleteUserSession(sessionID)
}

// CreateTestUser creates a user with a mock WebAuthn credential for testing
func (w *WebAuthnService) CreateTestUser(username string) (*User, error) {
	user, err := w.users.GetOrCreateUser(username)
	if err != nil {
		return nil, err
	}

	// Create a mock credential for testing
	mockCredential := &webauthn.Credential{
		ID:              []byte("test-credential-id"),
		PublicKey:       []byte("mock-public-key"),
		AttestationType: "none",
		Authenticator: webauthn.Authenticator{
			AAGUID:       []byte("mock-aaguid"),
			SignCount:    0,
			CloneWarning: false,
		},
	}

	if err := w.users.AddCredential(username, mockCredential); err != nil {
		return nil, err
	}

	return user, nil
}

// FinishRegistrationFromJSON completes registration from JSON response
func (w *WebAuthnService) FinishRegistrationFromJSON(token string, responseData interface{}) (*User, error) {
	// Convert interface{} to JSON bytes and back to proper struct
	jsonBytes, err := json.Marshal(responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var response protocol.CredentialCreationResponse
	if err := json.Unmarshal(jsonBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	parsedResponse, err := response.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return w.FinishRegistration(token, parsedResponse)
}

// FinishLoginFromJSON completes login from JSON response
func (w *WebAuthnService) FinishLoginFromJSON(token string, responseData interface{}) (*User, string, error) {
	// Convert interface{} to JSON bytes and back to proper struct
	jsonBytes, err := json.Marshal(responseData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal response: %w", err)
	}

	var response protocol.CredentialAssertionResponse
	if err := json.Unmarshal(jsonBytes, &response); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	parsedResponse, err := response.Parse()
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse response: %w", err)
	}

	return w.FinishLogin(token, parsedResponse)
}

// CreateUserSession creates a new user session
func (w *WebAuthnService) CreateUserSession(sessionID, userID string, ttl time.Duration) error {
	return w.sessions.CreateUserSession(sessionID, userID, ttl)
}

// GetUserCredentials returns all credentials for a user
func (w *WebAuthnService) GetUserCredentials(username string) (*User, error) {
	return w.users.GetUser(username)
}

// DeleteUserCredential removes a credential by index
func (w *WebAuthnService) DeleteUserCredential(username string, index int) error {
	return w.users.RemoveCredentialByIndex(username, index)
}