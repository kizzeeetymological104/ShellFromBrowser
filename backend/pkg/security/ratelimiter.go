package security

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ============================================
// Rate Limiter — Per-User Token Bucket
// ============================================
// Prevents abuse with configurable rate limiting
// Uses golang.org/x/time/rate for token bucket algorithm

// RateLimiter manages per-user rate limits
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     int // requests per second
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(ratePerSecond int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     ratePerSecond,
	}
}

// Allow checks if a user is allowed to proceed (rate limit check)
func (rl *RateLimiter) Allow(userID string) bool {
	rl.mu.Lock()
	limiter, exists := rl.limiters[userID]
	if !exists {
		// Create new limiter for this user
		// rate: N requests per second, burst: N (allow burst up to N)
		limiter = rate.NewLimiter(rate.Limit(rl.rate), rl.rate)
		rl.limiters[userID] = limiter
	}
	rl.mu.Unlock()

	return limiter.Allow()
}

// Cleanup removes inactive limiters (call periodically to prevent memory leak)
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Simple cleanup: remove all limiters
	// TODO Phase 2: Track last activity and remove only inactive users
	rl.limiters = make(map[string]*rate.Limiter)
}

// StartCleanupRoutine starts a goroutine that periodically cleans up inactive limiters
func (rl *RateLimiter) StartCleanupRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			rl.Cleanup()
		}
	}()
}
