package security

import (
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(100, time.Minute)
	if limiter == nil {
		t.Error("NewRateLimiter should not return nil")
	}
	if limiter.limit != 100 {
		t.Errorf("Expected limit 100, got %d", limiter.limit)
	}
	if limiter.window != time.Minute {
		t.Errorf("Expected window 1m0s, got %v", limiter.window)
	}
	if limiter.requests == nil {
		t.Error("requests map should not be nil")
	}
}

func TestRateLimiterAllow(t *testing.T) {
	limiter := NewRateLimiter(5, time.Minute)

	for i := 0; i < 5; i++ {
		if !limiter.Allow("client1") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	if limiter.Allow("client1") {
		t.Error("6th request should be denied")
	}
}

func TestRateLimiterAllowDifferentClients(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)

	if !limiter.Allow("client1") {
		t.Error("client1 first request should be allowed")
	}
	if !limiter.Allow("client2") {
		t.Error("client2 first request should be allowed")
	}
	if !limiter.Allow("client1") {
		t.Error("client1 second request should be allowed")
	}
	if limiter.Allow("client1") {
		t.Error("client1 third request should be denied")
	}
	if !limiter.Allow("client2") {
		t.Error("client2 second request should be allowed")
	}
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	limiter := NewRateLimiter(2, 50*time.Millisecond)

	limiter.Allow("client1")
	limiter.Allow("client1")

	if limiter.Allow("client1") {
		t.Error("Should be rate limited before window expiry")
	}

	time.Sleep(60 * time.Millisecond)

	if !limiter.Allow("client1") {
		t.Error("Should be allowed after window expiry")
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	limiter := NewRateLimiter(10, 50*time.Millisecond)

	limiter.Allow("client1")

	time.Sleep(100 * time.Millisecond)

	limiter.cleanup()

	limiter.mu.RLock()
	_, exists := limiter.requests["client1"]
	limiter.mu.RUnlock()

	if exists {
		t.Error("Expired client should be cleaned up")
	}
}

func TestRateLimiterCleanupMultiple(t *testing.T) {
	limiter := NewRateLimiter(10, time.Hour)

	limiter.Allow("client1")

	limiter.mu.Lock()
	limiter.requests["client2"] = &clientInfo{
		count:     1,
		windowEnd: time.Now().Add(-time.Hour),
	}
	limiter.mu.Unlock()

	limiter.cleanup()

	limiter.mu.RLock()
	_, c1Exists := limiter.requests["client1"]
	_, c2Exists := limiter.requests["client2"]
	limiter.mu.RUnlock()

	if !c1Exists {
		t.Error("client1 should still exist")
	}
	if c2Exists {
		t.Error("client2 should be cleaned up (expired)")
	}
}

func TestRateLimiterZeroLimit(t *testing.T) {
	limiter := NewRateLimiter(0, time.Minute)

	if !limiter.Allow("client1") {
		t.Error("Request should be allowed when limit is 0 (no rate limiting)")
	}
}
