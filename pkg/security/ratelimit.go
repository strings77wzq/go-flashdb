package security

import (
	"sync"
	"time"
)

type RateLimiter struct {
	limit    int
	window   time.Duration
	requests map[string]*clientInfo
	mu       sync.RWMutex
}

type clientInfo struct {
	count     int
	windowEnd time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		limit:    limit,
		window:   window,
		requests: make(map[string]*clientInfo),
	}
	go limiter.cleanupLoop()
	return limiter
}

func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.requests[clientID]

	if !exists || now.After(info.windowEnd) {
		rl.requests[clientID] = &clientInfo{
			count:     1,
			windowEnd: now.Add(rl.window),
		}
		return true
	}

	if info.count >= rl.limit {
		return false
	}

	info.count++
	return true
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for clientID, info := range rl.requests {
		if now.After(info.windowEnd) {
			delete(rl.requests, clientID)
		}
	}
}
