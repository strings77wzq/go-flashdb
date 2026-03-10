package security

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type Authenticator struct {
	password     string
	passwordHash string
	enabled      bool
	sessions     map[string]*Session
	mu           sync.RWMutex
}

type Session struct {
	ID        string
	CreatedAt int64
	LastSeen  int64
}

func NewAuthenticator(password string) *Authenticator {
	auth := &Authenticator{
		password:     password,
		passwordHash: hashPassword(password),
		enabled:      password != "",
		sessions:     make(map[string]*Session),
	}
	return auth
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (a *Authenticator) IsEnabled() bool {
	return a.enabled
}

func (a *Authenticator) Authenticate(password string) bool {
	if !a.enabled {
		return true
	}
	return hashPassword(password) == a.passwordHash
}

func (a *Authenticator) CreateSession(sessionID string) {
	if !a.enabled {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now().Unix()
	a.sessions[sessionID] = &Session{
		ID:        sessionID,
		CreatedAt: now,
		LastSeen:  now,
	}
}

func (a *Authenticator) ValidateSession(sessionID string) bool {
	if !a.enabled {
		return true
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	session, exists := a.sessions[sessionID]
	if !exists {
		return false
	}
	session.LastSeen = time.Now().Unix()
	return true
}

func (a *Authenticator) DestroySession(sessionID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.sessions, sessionID)
}

func (a *Authenticator) CleanupExpiredSessions(maxAge int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now().Unix()
	for id, session := range a.sessions {
		if now-session.LastSeen > maxAge {
			delete(a.sessions, id)
		}
	}
}
