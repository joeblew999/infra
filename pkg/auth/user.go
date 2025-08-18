package auth

import (
	"errors"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// User represents a WebAuthn user
type User struct {
	ID          []byte
	Name        string
	DisplayName string
	Credentials []webauthn.Credential
}

// WebAuthnID returns the user's WebAuthn ID
func (u *User) WebAuthnID() []byte {
	return u.ID
}

// WebAuthnName returns the user's WebAuthn name
func (u *User) WebAuthnName() string {
	return u.Name
}

// WebAuthnDisplayName returns the user's WebAuthn display name
func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnCredentials returns the user's WebAuthn credentials
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

// AddCredential adds a credential to the user
func (u *User) AddCredential(credential *webauthn.Credential) {
	u.Credentials = append(u.Credentials, *credential)
}

// RemoveCredential removes a credential by its ID
func (u *User) RemoveCredential(credentialID []byte) bool {
	for i, cred := range u.Credentials {
		if string(cred.ID) == string(credentialID) {
			// Remove credential at index i
			u.Credentials = append(u.Credentials[:i], u.Credentials[i+1:]...)
			return true
		}
	}
	return false
}

// RemoveCredentialByIndex removes a credential by its index
func (u *User) RemoveCredentialByIndex(index int) error {
	if index < 0 || index >= len(u.Credentials) {
		return errors.New("credential index out of range")
	}
	u.Credentials = append(u.Credentials[:index], u.Credentials[index+1:]...)
	return nil
}

// UserStore interface for user storage operations
type UserStore interface {
	GetUser(username string) (*User, error)
	GetUserByID(userID string) (*User, error)
	GetOrCreateUser(username string) (*User, error)
	AddCredential(username string, credential *webauthn.Credential) error
	RemoveCredential(username string, credentialID []byte) error
	RemoveCredentialByIndex(username string, index int) error
}

// InMemoryUserStore implements UserStore using in-memory storage
type InMemoryUserStore struct {
	users map[string]*User
}

// NewInMemoryUserStore creates a new in-memory user store
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

// GetUser retrieves a user by username
func (s *InMemoryUserStore) GetUser(username string) (*User, error) {
	user, exists := s.users[username]
	if !exists {
		fmt.Printf("DEBUG: User %s not found in store\n", username)
		return nil, errors.New("user not found")
	}
	fmt.Printf("DEBUG: Found user %s with %d credentials\n", username, len(user.Credentials))
	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *InMemoryUserStore) GetUserByID(userID string) (*User, error) {
	for _, user := range s.users {
		if string(user.ID) == userID {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

// GetOrCreateUser retrieves an existing user or creates a new one
func (s *InMemoryUserStore) GetOrCreateUser(username string) (*User, error) {
	if user, exists := s.users[username]; exists {
		return user, nil
	}

	user := &User{
		ID:          []byte(uuid.New().String()),
		Name:        username,
		DisplayName: username,
		Credentials: []webauthn.Credential{},
	}
	s.users[username] = user
	return user, nil
}

// AddCredential adds a credential to a user
func (s *InMemoryUserStore) AddCredential(username string, credential *webauthn.Credential) error {
	user, exists := s.users[username]
	if !exists {
		return errors.New("user not found")
	}
	user.AddCredential(credential)
	fmt.Printf("DEBUG: Added credential for user %s, total credentials: %d\n", username, len(user.Credentials))
	return nil
}

// RemoveCredential removes a credential by its ID
func (s *InMemoryUserStore) RemoveCredential(username string, credentialID []byte) error {
	user, exists := s.users[username]
	if !exists {
		return errors.New("user not found")
	}
	if !user.RemoveCredential(credentialID) {
		return errors.New("credential not found")
	}
	fmt.Printf("DEBUG: Removed credential for user %s, remaining credentials: %d\n", username, len(user.Credentials))
	return nil
}

// RemoveCredentialByIndex removes a credential by its index
func (s *InMemoryUserStore) RemoveCredentialByIndex(username string, index int) error {
	user, exists := s.users[username]
	if !exists {
		return errors.New("user not found")
	}
	if err := user.RemoveCredentialByIndex(index); err != nil {
		return err
	}
	fmt.Printf("DEBUG: Removed credential at index %d for user %s, remaining credentials: %d\n", index, username, len(user.Credentials))
	return nil
}