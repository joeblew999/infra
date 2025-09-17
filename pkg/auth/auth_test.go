package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joeblew999/infra/pkg/config"
)

func TestAuthDataDirectory(t *testing.T) {
	// Test auth data directory creation and isolation
	authPath := config.GetAuthPath()
	
	// Create auth directory
	err := os.MkdirAll(authPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create auth data directory: %v", err)
	}
	
	t.Logf("✅ Auth data directory created: %s", authPath)
	
	// Create some test artifacts to simulate auth storage
	testFiles := []string{
		"users.json",
		"sessions.db",
		"credentials.data",
		"webauthn-sessions.cache",
	}
	
	for _, filename := range testFiles {
		testFile := filepath.Join(authPath, filename)
		err := os.WriteFile(testFile, []byte("test auth data"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		t.Logf("✅ Test auth file: %s", testFile)
	}
	
	t.Logf("📁 Test artifacts in: %s", authPath)
}

func TestInMemoryUserStore(t *testing.T) {
	// Test user storage with test isolation
	store := NewInMemoryUserStore()
	
	// Test user creation
	user, err := store.GetOrCreateUser("testuser")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	if user.Name != "testuser" {
		t.Errorf("User name mismatch: got %s, want testuser", user.Name)
	}
	
	t.Logf("✅ User created: %s (ID: %s)", user.Name, string(user.ID))
	
	// Test user retrieval
	retrievedUser, err := store.GetUser("testuser")
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}
	
	if string(retrievedUser.ID) != string(user.ID) {
		t.Errorf("Retrieved user ID mismatch")
	}
	
	t.Logf("✅ User retrieved: %s", retrievedUser.Name)
	
	// Test user by ID retrieval
	userByID, err := store.GetUserByID(string(user.ID))
	if err != nil {
		t.Fatalf("Failed to retrieve user by ID: %v", err)
	}
	
	if userByID.Name != "testuser" {
		t.Errorf("User by ID name mismatch")
	}
	
	t.Logf("✅ User retrieved by ID: %s", userByID.Name)
	
	// Save test user data to auth directory for inspection
	authPath := config.GetAuthPath()
	err = os.MkdirAll(authPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create auth directory: %v", err)
	}
	
	userDataFile := filepath.Join(authPath, "test-user-data.txt")
	userData := "Test User Data:\n"
	userData += "Name: " + user.Name + "\n"
	userData += "Display Name: " + user.DisplayName + "\n"
	userData += "ID: " + string(user.ID) + "\n"
	userData += "Credentials: " + string(rune(len(user.Credentials))) + "\n"
	
	err = os.WriteFile(userDataFile, []byte(userData), 0644)
	if err != nil {
		t.Fatalf("Failed to save user data: %v", err)
	}
	
	t.Logf("✅ User data saved: %s", userDataFile)
}

func TestInMemorySessionStore(t *testing.T) {
	// Test session storage with test isolation
	store := NewInMemorySessionStore()
	
	// Test user session creation
	sessionID := "test-session-123"
	userID := "test-user-456"
	ttl := 30 * time.Minute
	
	err := store.CreateUserSession(sessionID, userID, ttl)
	if err != nil {
		t.Fatalf("Failed to create user session: %v", err)
	}
	
	t.Logf("✅ User session created: %s -> %s", sessionID, userID)
	
	// Test user session retrieval
	retrievedUserID, err := store.GetUserSession(sessionID)
	if err != nil {
		t.Fatalf("Failed to retrieve user session: %v", err)
	}
	
	if retrievedUserID != userID {
		t.Errorf("Session user ID mismatch: got %s, want %s", retrievedUserID, userID)
	}
	
	t.Logf("✅ User session retrieved: %s", retrievedUserID)
	
	// Test session deletion
	err = store.DeleteUserSession(sessionID)
	if err != nil {
		t.Fatalf("Failed to delete user session: %v", err)
	}
	
	// Verify session is deleted
	_, err = store.GetUserSession(sessionID)
	if err == nil {
		t.Error("Session should have been deleted")
	}
	
	t.Logf("✅ User session deleted")
	
	// Save test session data to auth directory for inspection
	authPath := config.GetAuthPath()
	err = os.MkdirAll(authPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create auth directory: %v", err)
	}
	
	sessionDataFile := filepath.Join(authPath, "test-session-data.txt")
	sessionData := "Test Session Data:\n"
	sessionData += "Session ID: " + sessionID + "\n"
	sessionData += "User ID: " + userID + "\n"
	sessionData += "TTL: " + ttl.String() + "\n"
	sessionData += "Status: Deleted after test\n"
	
	err = os.WriteFile(sessionDataFile, []byte(sessionData), 0644)
	if err != nil {
		t.Fatalf("Failed to save session data: %v", err)
	}
	
	t.Logf("✅ Session data saved: %s", sessionDataFile)
}

func TestAuthServiceConfiguration(t *testing.T) {
	// Test auth service configuration with test isolation
	webauthnConfig := WebAuthnConfig{
		RPID:          "localhost",
		RPDisplayName: "Test App",
		RPOrigins:     []string{"https://localhost:8080"},
	}
	
	users := NewInMemoryUserStore()
	sessions := NewInMemorySessionStore()
	webDir := "test-web-dir"
	
	// Create auth service
	authService, err := NewAuthService(webauthnConfig, users, sessions, webDir)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}
	
	if authService == nil {
		t.Fatal("Auth service is nil")
	}
	
	t.Logf("✅ Auth service created successfully")
	
	// Test user creation through webauthn service directly (test-only)
	testUser, err := authService.webauthn.CreateTestUser("integration-test-user")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	if testUser.Name != "integration-test-user" {
		t.Errorf("Test user name mismatch")
	}
	
	t.Logf("✅ Test user created: %s", testUser.Name)
	
	// Save auth service config to test directory for inspection
	authPath := config.GetAuthPath()
	err = os.MkdirAll(authPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create auth directory: %v", err)
	}
	
	configFile := filepath.Join(authPath, "test-auth-config.txt")
	configData := "Test Auth Service Configuration:\n"
	configData += "RPID: " + webauthnConfig.RPID + "\n"
	configData += "Display Name: " + webauthnConfig.RPDisplayName + "\n"
	configData += "Origins: " + webauthnConfig.RPOrigins[0] + "\n"
	configData += "Web Directory: " + webDir + "\n"
	configData += "Test User: " + testUser.Name + "\n"
	
	err = os.WriteFile(configFile, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to save auth config: %v", err)
	}
	
	t.Logf("✅ Auth config saved: %s", configFile)
}