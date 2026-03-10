package security

import (
	"testing"
	"time"
)

func TestNewAuthenticator(t *testing.T) {
	auth := NewAuthenticator("")
	if auth == nil {
		t.Error("NewAuthenticator should not return nil")
	}
	if auth.enabled {
		t.Error("Authenticator should not be enabled with empty password")
	}
}

func TestNewAuthenticatorWithPassword(t *testing.T) {
	auth := NewAuthenticator("testpassword")
	if !auth.enabled {
		t.Error("Authenticator should be enabled with non-empty password")
	}
	if auth.passwordHash == nil {
		t.Error("passwordHash should not be nil")
	}
	if auth.password != "testpassword" {
		t.Errorf("Expected password 'testpassword', got '%s'", auth.password)
	}
}

func TestAuthenticatorIsEnabled(t *testing.T) {
	authEnabled := NewAuthenticator("password")
	authDisabled := NewAuthenticator("")

	if !authEnabled.IsEnabled() {
		t.Error("Authenticator with password should be enabled")
	}
	if authDisabled.IsEnabled() {
		t.Error("Authenticator without password should not be enabled")
	}
}

func TestAuthenticatorAuthenticate(t *testing.T) {
	auth := NewAuthenticator("mypassword")

	if !auth.Authenticate("mypassword") {
		t.Error("Should authenticate with correct password")
	}
	if auth.Authenticate("wrongpassword") {
		t.Error("Should not authenticate with wrong password")
	}
}

func TestAuthenticatorAuthenticateDisabled(t *testing.T) {
	auth := NewAuthenticator("")
	if !auth.Authenticate("anypassword") {
		t.Error("Should authenticate any password when disabled")
	}
}

func TestAuthenticatorCreateSession(t *testing.T) {
	auth := NewAuthenticator("password")
	auth.CreateSession("session1")

	auth.mu.RLock()
	session, exists := auth.sessions["session1"]
	auth.mu.RUnlock()

	if !exists {
		t.Error("Session should exist after CreateSession")
	}
	if session.ID != "session1" {
		t.Errorf("Expected session ID 'session1', got '%s'", session.ID)
	}
}

func TestAuthenticatorCreateSessionDisabled(t *testing.T) {
	auth := NewAuthenticator("")
	auth.CreateSession("session1")

	auth.mu.RLock()
	_, exists := auth.sessions["session1"]
	auth.mu.RUnlock()

	if exists {
		t.Error("Session should not be created when auth is disabled")
	}
}

func TestAuthenticatorValidateSession(t *testing.T) {
	auth := NewAuthenticator("password")
	auth.CreateSession("session1")

	if !auth.ValidateSession("session1") {
		t.Error("Should validate existing session")
	}
	if auth.ValidateSession("nonexistent") {
		t.Error("Should not validate non-existent session")
	}
}

func TestAuthenticatorValidateSessionDisabled(t *testing.T) {
	auth := NewAuthenticator("")
	if !auth.ValidateSession("any") {
		t.Error("Should validate any session when disabled")
	}
}

func TestAuthenticatorDestroySession(t *testing.T) {
	auth := NewAuthenticator("password")
	auth.CreateSession("session1")
	auth.DestroySession("session1")

	auth.mu.RLock()
	_, exists := auth.sessions["session1"]
	auth.mu.RUnlock()

	if exists {
		t.Error("Session should not exist after DestroySession")
	}
}

func TestAuthenticatorCleanupExpiredSessions(t *testing.T) {
	auth := NewAuthenticator("password")
	now := time.Now().Unix()

	auth.mu.Lock()
	auth.sessions["old"] = &Session{
		ID:        "old",
		CreatedAt: now - 100,
		LastSeen:  now - 100,
	}
	auth.sessions["new"] = &Session{
		ID:        "new",
		CreatedAt: now,
		LastSeen:  now,
	}
	auth.mu.Unlock()

	auth.CleanupExpiredSessions(50)

	auth.mu.RLock()
	_, oldExists := auth.sessions["old"]
	_, newExists := auth.sessions["new"]
	auth.mu.RUnlock()

	if oldExists {
		t.Error("Old session should be cleaned up")
	}
	if !newExists {
		t.Error("New session should still exist")
	}
}

func TestConstantTimeCompare(t *testing.T) {
	if !ConstantTimeCompare("test", "test") {
		t.Error("Should return true for equal strings")
	}
	if ConstantTimeCompare("test", " Test") {
		t.Error("Should return false for different strings")
	}
	if !ConstantTimeCompare("", "") {
		t.Error("Should return true for empty strings (subtle.ConstantTimeCompare behavior)")
	}
	if !ConstantTimeCompare("a", "a") {
		t.Error("Should return true for equal single char")
	}
}
