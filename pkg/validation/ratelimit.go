package validation

import (
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting per client.
// Safe for concurrent use.
//
// Note: This package intentionally uses time.Now() for security/rate limiting purposes,
// which is an acceptable exception to the deterministic generation guideline (Coding Guideline #2).
// Rate limiting MUST use real time to be effective against actual network attacks.
// The determinism requirement applies to procedural generation (terrain, items, quests),
// not security/network infrastructure.
type RateLimiter struct {
	// rate is the number of requests allowed per interval
	rate int

	// interval is the time window for rate limiting
	interval time.Duration

	// clients tracks request timestamps for each client
	clients map[uint64]*clientBucket

	// mu protects concurrent access to clients map
	mu sync.RWMutex

	// cleanupInterval is how often to clean up old client entries
	cleanupInterval time.Duration

	// lastCleanup tracks when we last cleaned up
	lastCleanup time.Time
}

// clientBucket tracks request timestamps for a single client
type clientBucket struct {
	// timestamps of recent requests (within the interval window)
	timestamps []time.Time

	// lastRequest tracks the most recent request time
	lastRequest time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per interval (defaults to 1 if <= 0)
// interval: time window for rate limiting (defaults to 1 second if <= 0)
func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	// Validate and set defaults for invalid values
	if rate <= 0 {
		rate = 1
	}
	if interval <= 0 {
		interval = time.Second
	}

	return &RateLimiter{
		rate:            rate,
		interval:        interval,
		clients:         make(map[uint64]*clientBucket),
		cleanupInterval: 5 * time.Minute,
		lastCleanup:     time.Now(),
	}
}

// Allow checks if a request from the given client should be allowed
// Returns true if the request is within rate limits, false otherwise
func (rl *RateLimiter) Allow(clientID uint64) bool {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Periodic cleanup of inactive clients
	if now.Sub(rl.lastCleanup) > rl.cleanupInterval {
		rl.cleanup(now)
		rl.lastCleanup = now
	}

	// Get or create client bucket
	bucket, exists := rl.clients[clientID]
	if !exists {
		bucket = &clientBucket{
			timestamps:  make([]time.Time, 0, rl.rate),
			lastRequest: now,
		}
		rl.clients[clientID] = bucket
	}

	// Remove timestamps outside the current interval window
	cutoff := now.Add(-rl.interval)
	validTimestamps := make([]time.Time, 0, len(bucket.timestamps))
	for _, ts := range bucket.timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	bucket.timestamps = validTimestamps

	// Check if rate limit is exceeded
	if len(bucket.timestamps) >= rl.rate {
		return false
	}

	// Allow request and record timestamp
	bucket.timestamps = append(bucket.timestamps, now)
	bucket.lastRequest = now

	return true
}

// Reset clears the rate limit state for a specific client
// Useful for testing or manual override
func (rl *RateLimiter) Reset(clientID uint64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.clients, clientID)
}

// ResetAll clears all rate limit state
// Useful for testing or server restart
func (rl *RateLimiter) ResetAll() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.clients = make(map[uint64]*clientBucket)
}

// GetStats returns statistics about the rate limiter
// Returns number of tracked clients
func (rl *RateLimiter) GetStats() map[string]int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]int{
		"tracked_clients": len(rl.clients),
	}
}

// cleanup removes client entries that haven't made requests recently
// Must be called with mu locked
func (rl *RateLimiter) cleanup(now time.Time) {
	// Remove clients inactive for 10 minutes
	inactiveThreshold := now.Add(-10 * time.Minute)

	for clientID, bucket := range rl.clients {
		if bucket.lastRequest.Before(inactiveThreshold) {
			delete(rl.clients, clientID)
		}
	}
}

// GetClientRequestCount returns the number of requests in the current window for a client
// Useful for monitoring and debugging
func (rl *RateLimiter) GetClientRequestCount(clientID uint64) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	bucket, exists := rl.clients[clientID]
	if !exists {
		return 0
	}

	// Count valid timestamps
	now := time.Now()
	cutoff := now.Add(-rl.interval)
	count := 0
	for _, ts := range bucket.timestamps {
		if ts.After(cutoff) {
			count++
		}
	}

	return count
}
