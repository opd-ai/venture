package validation

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(3, time.Second)

	clientID := uint64(1)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !limiter.Allow(clientID) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied (rate limit exceeded)
	if limiter.Allow(clientID) {
		t.Error("4th request should be denied (rate limit 3/s)")
	}

	// Wait for rate limit window to pass
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again after window reset
	if !limiter.Allow(clientID) {
		t.Error("Request should be allowed after rate limit window reset")
	}
}

func TestRateLimiter_MultipleClients(t *testing.T) {
	limiter := NewRateLimiter(2, time.Second)

	client1 := uint64(1)
	client2 := uint64(2)

	// Each client should have independent rate limits
	if !limiter.Allow(client1) {
		t.Error("Client 1 request 1 should be allowed")
	}
	if !limiter.Allow(client1) {
		t.Error("Client 1 request 2 should be allowed")
	}

	if !limiter.Allow(client2) {
		t.Error("Client 2 request 1 should be allowed")
	}
	if !limiter.Allow(client2) {
		t.Error("Client 2 request 2 should be allowed")
	}

	// Both clients should be rate limited
	if limiter.Allow(client1) {
		t.Error("Client 1 should be rate limited")
	}
	if limiter.Allow(client2) {
		t.Error("Client 2 should be rate limited")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	limiter := NewRateLimiter(2, time.Second)

	clientID := uint64(1)

	// Use up rate limit
	limiter.Allow(clientID)
	limiter.Allow(clientID)

	// Should be denied
	if limiter.Allow(clientID) {
		t.Error("Request should be denied before reset")
	}

	// Reset client
	limiter.Reset(clientID)

	// Should be allowed after reset
	if !limiter.Allow(clientID) {
		t.Error("Request should be allowed after reset")
	}
}

func TestRateLimiter_ResetAll(t *testing.T) {
	limiter := NewRateLimiter(1, time.Second)

	client1 := uint64(1)
	client2 := uint64(2)

	// Use up rate limits for both clients
	limiter.Allow(client1)
	limiter.Allow(client2)

	// Both should be denied
	if limiter.Allow(client1) {
		t.Error("Client 1 should be denied before reset")
	}
	if limiter.Allow(client2) {
		t.Error("Client 2 should be denied before reset")
	}

	// Reset all
	limiter.ResetAll()

	// Both should be allowed after reset
	if !limiter.Allow(client1) {
		t.Error("Client 1 should be allowed after reset")
	}
	if !limiter.Allow(client2) {
		t.Error("Client 2 should be allowed after reset")
	}
}

func TestRateLimiter_GetStats(t *testing.T) {
	limiter := NewRateLimiter(10, time.Second)

	// Initially no clients
	stats := limiter.GetStats()
	if stats["tracked_clients"] != 0 {
		t.Errorf("Expected 0 tracked clients, got %d", stats["tracked_clients"])
	}

	// Add some clients
	limiter.Allow(uint64(1))
	limiter.Allow(uint64(2))
	limiter.Allow(uint64(3))

	stats = limiter.GetStats()
	if stats["tracked_clients"] != 3 {
		t.Errorf("Expected 3 tracked clients, got %d", stats["tracked_clients"])
	}
}

func TestRateLimiter_GetClientRequestCount(t *testing.T) {
	limiter := NewRateLimiter(10, time.Second)

	clientID := uint64(1)

	// Initially 0 requests
	if count := limiter.GetClientRequestCount(clientID); count != 0 {
		t.Errorf("Expected 0 requests, got %d", count)
	}

	// Make some requests
	limiter.Allow(clientID)
	limiter.Allow(clientID)
	limiter.Allow(clientID)

	// Should count 3 requests
	if count := limiter.GetClientRequestCount(clientID); count != 3 {
		t.Errorf("Expected 3 requests, got %d", count)
	}

	// Wait for window to pass
	time.Sleep(1100 * time.Millisecond)

	// Count should be 0 after window expires
	if count := limiter.GetClientRequestCount(clientID); count != 0 {
		t.Errorf("Expected 0 requests after window expiry, got %d", count)
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	limiter := NewRateLimiter(100, time.Second)

	clientID := uint64(1)
	numGoroutines := 10
	requestsPerGoroutine := 20

	var wg sync.WaitGroup
	successCount := make(chan int, numGoroutines)

	// Spawn concurrent goroutines making requests
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := 0
			for j := 0; j < requestsPerGoroutine; j++ {
				if limiter.Allow(clientID) {
					count++
				}
			}
			successCount <- count
		}()
	}

	wg.Wait()
	close(successCount)

	// Count total successful requests
	total := 0
	for count := range successCount {
		total += count
	}

	// Should allow exactly rate limit (100) requests
	if total != 100 {
		t.Errorf("Expected exactly 100 successful requests, got %d", total)
	}
}

func TestRateLimiter_WindowSliding(t *testing.T) {
	limiter := NewRateLimiter(2, 500*time.Millisecond)

	clientID := uint64(1)

	// Request 1 at t=0
	if !limiter.Allow(clientID) {
		t.Error("Request 1 should be allowed")
	}

	// Request 2 at t=~0
	if !limiter.Allow(clientID) {
		t.Error("Request 2 should be allowed")
	}

	// Request 3 at t=~0 (should be denied)
	if limiter.Allow(clientID) {
		t.Error("Request 3 should be denied")
	}

	// Wait half the window
	time.Sleep(300 * time.Millisecond)

	// Request 4 at t=300ms (should still be denied, window hasn't passed)
	if limiter.Allow(clientID) {
		t.Error("Request 4 should be denied (window not expired)")
	}

	// Wait for first window to fully expire
	time.Sleep(300 * time.Millisecond)

	// Request 5 at t=600ms (should be allowed, window expired)
	if !limiter.Allow(clientID) {
		t.Error("Request 5 should be allowed after window expired")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	// Use short cleanup interval for testing
	limiter := NewRateLimiter(10, time.Second)
	limiter.cleanupInterval = 100 * time.Millisecond

	// Add several clients
	for i := uint64(1); i <= 5; i++ {
		limiter.Allow(i)
	}

	stats := limiter.GetStats()
	if stats["tracked_clients"] != 5 {
		t.Errorf("Expected 5 clients before cleanup, got %d", stats["tracked_clients"])
	}

	// Manually trigger cleanup by advancing time
	// In real usage, cleanup happens automatically during Allow() calls
	time.Sleep(200 * time.Millisecond)

	// Make request to trigger cleanup check
	limiter.Allow(uint64(100))

	// Note: Cleanup only removes clients inactive for 10 minutes,
	// so the count won't change in this test (clients just added)
	// This test validates the cleanup mechanism doesn't crash
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	limiter := NewRateLimiter(1000000, time.Second)
	clientID := uint64(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(clientID)
	}
}

func BenchmarkRateLimiter_AllowMultipleClients(b *testing.B) {
	limiter := NewRateLimiter(1000000, time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clientID := uint64(i % 100) // Simulate 100 different clients
		limiter.Allow(clientID)
	}
}

func BenchmarkRateLimiter_GetClientRequestCount(b *testing.B) {
	limiter := NewRateLimiter(1000, time.Second)
	clientID := uint64(1)

	// Pre-populate with some requests
	for i := 0; i < 10; i++ {
		limiter.Allow(clientID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.GetClientRequestCount(clientID)
	}
}

func TestRateLimiter_InvalidRateValues(t *testing.T) {
	tests := []struct {
		name            string
		rate            int
		interval        time.Duration
		expectedRate    int
		expectedAllowed bool
	}{
		{
			name:            "zero rate defaults to 1",
			rate:            0,
			interval:        time.Second,
			expectedRate:    1,
			expectedAllowed: true,
		},
		{
			name:            "negative rate defaults to 1",
			rate:            -5,
			interval:        time.Second,
			expectedRate:    1,
			expectedAllowed: true,
		},
		{
			name:            "zero interval defaults to 1 second",
			rate:            5,
			interval:        0,
			expectedRate:    5,
			expectedAllowed: true,
		},
		{
			name:            "negative interval defaults to 1 second",
			rate:            5,
			interval:        -time.Second,
			expectedRate:    5,
			expectedAllowed: true,
		},
		{
			name:            "both zero defaults to 1 request per second",
			rate:            0,
			interval:        0,
			expectedRate:    1,
			expectedAllowed: true,
		},
		{
			name:            "both negative defaults to 1 request per second",
			rate:            -1,
			interval:        -time.Second,
			expectedRate:    1,
			expectedAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.rate, tt.interval)
			clientID := uint64(1)

			// First request should always be allowed with valid defaults
			allowed := limiter.Allow(clientID)
			if allowed != tt.expectedAllowed {
				t.Errorf("First request allowed = %v, want %v", allowed, tt.expectedAllowed)
			}

			// Verify rate was set correctly by checking how many requests are allowed
			for i := 1; i < tt.expectedRate; i++ {
				if !limiter.Allow(clientID) {
					t.Errorf("Request %d should be allowed (expected rate %d)", i+1, tt.expectedRate)
				}
			}

			// Next request should be denied (rate limit reached)
			if limiter.Allow(clientID) {
				t.Errorf("Request %d should be denied (rate limit %d reached)", tt.expectedRate+1, tt.expectedRate)
			}
		})
	}
}

func TestRateLimiter_ZeroRateDoesNotDenyAllRequests(t *testing.T) {
	// This test ensures the bug is fixed: zero rate should default to 1, not deny all requests
	limiter := NewRateLimiter(0, time.Second)
	clientID := uint64(1)

	// First request should be allowed (rate defaults to 1)
	if !limiter.Allow(clientID) {
		t.Error("First request should be allowed with rate=0 (defaults to 1)")
	}

	// Second request should be denied (rate limit of 1 reached)
	if limiter.Allow(clientID) {
		t.Error("Second request should be denied (rate limit 1 reached)")
	}
}

func TestRateLimiter_NegativeRateDoesNotDenyAllRequests(t *testing.T) {
	// This test ensures the bug is fixed: negative rate should default to 1, not deny all requests
	limiter := NewRateLimiter(-5, time.Second)
	clientID := uint64(1)

	// First request should be allowed (rate defaults to 1)
	if !limiter.Allow(clientID) {
		t.Error("First request should be allowed with rate=-5 (defaults to 1)")
	}

	// Second request should be denied (rate limit of 1 reached)
	if limiter.Allow(clientID) {
		t.Error("Second request should be denied (rate limit 1 reached)")
	}
}
