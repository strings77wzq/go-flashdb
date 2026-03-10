package security

import (
	"crypto/subtle"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Authenticator struct {
	password     string
	passwordHash []byte
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
	var hash []byte
	var err error
	if password != "" {
		hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			hash = []byte{}
		}
	}

	auth := &Authenticator{
		password:     password,
		passwordHash: hash,
		enabled:      password != "",
		sessions:     make(map[string]*Session),
	}
	return auth
}

func (a *Authenticator) IsEnabled() bool {
	return a.enabled
}

func (a *Authenticator) Authenticate(password string) bool {
	if !a.enabled {
		return true
	}
	err := bcrypt.CompareHashAndPassword(a.passwordHash, []byte(password))
	return err == nil
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

func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
