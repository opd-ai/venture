package federation

import (
	"errors"
	"testing"
	"time"
)

func TestNewRetryStrategy(t *testing.T) {
	config := DefaultRetryConfig()
	rs := NewRetryStrategy(config)

	if rs == nil {
		t.Fatal("NewRetryStrategy returned nil")
	}

	if rs.config.MaxRetries != config.MaxRetries {
		t.Error("Config not properly set")
	}
}

func TestRetryStrategySuccessFirstAttempt(t *testing.T) {
	config := DefaultRetryConfig()
	rs := NewRetryStrategy(config)

	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := rs.Execute(fn, nil)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestRetryStrategySuccessAfterRetries(t *testing.T) {
	config := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.0,
	}
	rs := NewRetryStrategy(config)

	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("connection refused")
		}
		return nil
	}

	start := time.Now()
	err := rs.Execute(fn, IsNetworkError)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error after retries, got: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}

	// Should have at least some delay from retries (2 retries with 10ms + 20ms base = 30ms)
	if duration < 20*time.Millisecond {
		t.Errorf("Expected some delay from retries, got %v", duration)
	}
}

func TestRetryStrategyMaxRetriesExceeded(t *testing.T) {
	config := RetryConfig{
		MaxRetries:   2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.0,
	}
	rs := NewRetryStrategy(config)

	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("connection timeout")
	}

	err := rs.Execute(fn, IsNetworkError)
	if err == nil {
		t.Error("Expected error after max retries exceeded")
	}

	// Should be initial attempt + 2 retries = 3 calls
	if callCount != 3 {
		t.Errorf("Expected 3 calls (1 initial + 2 retries), got %d", callCount)
	}
}

func TestRetryStrategyNonRetryableError(t *testing.T) {
	config := DefaultRetryConfig()
	rs := NewRetryStrategy(config)

	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("non-retryable error")
	}

	isRetryable := func(err error) bool {
		return false // Never retry
	}

	err := rs.Execute(fn, isRetryable)
	if err == nil {
		t.Error("Expected error")
	}

	if callCount != 1 {
		t.Errorf("Expected only 1 call for non-retryable error, got %d", callCount)
	}
}

func TestRetryStrategyExponentialBackoff(t *testing.T) {
	config := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.0, // No jitter for predictable testing
	}
	rs := NewRetryStrategy(config)

	delays := []time.Duration{}
	lastTime := time.Now()

	callCount := 0
	fn := func() error {
		callCount++
		now := time.Now()
		if callCount > 1 {
			delays = append(delays, now.Sub(lastTime))
		}
		lastTime = now
		return errors.New("connection refused")
	}

	rs.Execute(fn, IsNetworkError)

	if len(delays) < 2 {
		t.Fatal("Not enough delays recorded")
	}

	// First retry should be ~10ms, second should be ~20ms
	// Allow 5ms tolerance for timing variance
	if delays[0] < 8*time.Millisecond || delays[0] > 15*time.Millisecond {
		t.Errorf("First retry delay should be ~10ms, got %v", delays[0])
	}

	if delays[1] < 18*time.Millisecond || delays[1] > 25*time.Millisecond {
		t.Errorf("Second retry delay should be ~20ms, got %v", delays[1])
	}
}

func TestRetryStrategyMaxDelay(t *testing.T) {
	config := RetryConfig{
		MaxRetries:   5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     150 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.0,
	}
	rs := NewRetryStrategy(config)

	// Calculate what delay would be without cap
	// Attempt 4: 100ms * 2^3 = 800ms
	// But should be capped at 150ms
	delay := rs.calculateDelay(4)

	if delay > config.MaxDelay {
		t.Errorf("Delay %v exceeds max delay %v", delay, config.MaxDelay)
	}

	// Should be close to max delay
	if delay < config.MaxDelay-10*time.Millisecond {
		t.Errorf("Expected delay close to max delay %v, got %v", config.MaxDelay, delay)
	}
}

func TestRetryStrategyJitter(t *testing.T) {
	config := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.5, // 50% jitter
		Seed:         0,   // Use time-based seed
	}
	rs := NewRetryStrategy(config)

	// Calculate multiple delays and check they vary
	delays := make(map[time.Duration]bool)
	for i := 0; i < 10; i++ {
		delay := rs.calculateDelay(1)
		delays[delay] = true
	}

	// With 50% jitter, delays should vary
	// (though there's a small chance they could be identical)
	if len(delays) < 2 {
		t.Log("Warning: Jitter may not be working as expected (could be random chance)")
	}

	// All delays should be within reasonable range
	// Base is 100ms, with 50% jitter: [50ms, 150ms]
	for delay := range delays {
		if delay < 40*time.Millisecond || delay > 160*time.Millisecond {
			t.Errorf("Delay %v is outside expected jitter range [50ms, 150ms]", delay)
		}
	}
}

func TestRetryStrategyDeterministicJitter(t *testing.T) {
	config := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.5,   // 50% jitter
		Seed:         12345, // Fixed seed for deterministic behavior
	}

	// Create two retry strategies with the same seed
	rs1 := NewRetryStrategy(config)
	rs2 := NewRetryStrategy(config)

	// Calculate delays multiple times and verify they match
	for i := 1; i <= 5; i++ {
		delay1 := rs1.calculateDelay(i)
		delay2 := rs2.calculateDelay(i)

		if delay1 != delay2 {
			t.Errorf("Attempt %d: Delays should be identical with same seed, got %v and %v", i, delay1, delay2)
		}
	}
}

func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil_error",
			err:      nil,
			expected: false,
		},
		{
			name:     "connection_refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "connection_reset",
			err:      errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "timeout",
			err:      errors.New("i/o timeout"),
			expected: true,
		},
		{
			name:     "network_unreachable",
			err:      errors.New("network is unreachable"),
			expected: true,
		},
		{
			name:     "broken_pipe",
			err:      errors.New("broken pipe"),
			expected: true,
		},
		{
			name:     "non_network_error",
			err:      errors.New("invalid input"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNetworkError(tt.err)
			if result != tt.expected {
				t.Errorf("IsNetworkError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries <= 0 {
		t.Error("MaxRetries should be positive")
	}

	if config.InitialDelay <= 0 {
		t.Error("InitialDelay should be positive")
	}

	if config.MaxDelay <= 0 {
		t.Error("MaxDelay should be positive")
	}

	if config.Multiplier <= 1.0 {
		t.Error("Multiplier should be greater than 1.0 for exponential backoff")
	}

	if config.Jitter < 0 || config.Jitter > 1.0 {
		t.Error("Jitter should be between 0.0 and 1.0")
	}
}
