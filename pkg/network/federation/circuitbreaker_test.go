package federation

import (
	"errors"
	"testing"
	"time"
)

func TestNewCircuitBreaker(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(config)

	if cb == nil {
		t.Fatal("NewCircuitBreaker returned nil")
	}

	if cb.State() != CircuitStateClosed {
		t.Errorf("Expected initial state to be Closed, got %v", cb.State())
	}
}

func TestCircuitBreakerStateTransitions(t *testing.T) {
	tests := []struct {
		name           string
		config         CircuitBreakerConfig
		operations     []bool // true = success, false = failure
		expectedState  CircuitState
	}{
		{
			name: "closed_to_open_after_failures",
			config: CircuitBreakerConfig{
				FailureThreshold:    3,
				SuccessThreshold:    2,
				OpenTimeout:         1 * time.Second,
				HalfOpenMaxRequests: 3,
			},
			operations:    []bool{false, false, false},
			expectedState: CircuitStateOpen,
		},
		{
			name: "stays_closed_with_successes",
			config: CircuitBreakerConfig{
				FailureThreshold:    3,
				SuccessThreshold:    2,
				OpenTimeout:         1 * time.Second,
				HalfOpenMaxRequests: 3,
			},
			operations:    []bool{true, true, true},
			expectedState: CircuitStateClosed,
		},
		{
			name: "stays_closed_below_threshold",
			config: CircuitBreakerConfig{
				FailureThreshold:    3,
				SuccessThreshold:    2,
				OpenTimeout:         1 * time.Second,
				HalfOpenMaxRequests: 3,
			},
			operations:    []bool{false, false, true},
			expectedState: CircuitStateClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCircuitBreaker(tt.config)

			for _, success := range tt.operations {
				var err error
				if !success {
					err = errors.New("simulated failure")
				}
				cb.Call(func() error { return err })
			}

			if cb.State() != tt.expectedState {
				t.Errorf("Expected state %v, got %v", tt.expectedState, cb.State())
			}
		})
	}
}

func TestCircuitBreakerOpenToHalfOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		OpenTimeout:         100 * time.Millisecond,
		HalfOpenMaxRequests: 3,
	}
	cb := NewCircuitBreaker(config)

	// Trigger circuit to open
	cb.Call(func() error { return errors.New("failure 1") })
	cb.Call(func() error { return errors.New("failure 2") })

	if cb.State() != CircuitStateOpen {
		t.Fatalf("Expected state Open, got %v", cb.State())
	}

	// Attempt should fail immediately
	err := cb.Call(func() error { return nil })
	if err == nil {
		t.Error("Expected call to fail when circuit is open")
	}

	// Wait for open timeout
	time.Sleep(150 * time.Millisecond)

	// Should transition to half-open
	err = cb.Call(func() error { return nil })
	if err != nil {
		t.Errorf("Expected call to succeed in half-open state, got error: %v", err)
	}

	if cb.State() != CircuitStateHalfOpen {
		t.Errorf("Expected state HalfOpen, got %v", cb.State())
	}
}

func TestCircuitBreakerHalfOpenToClosed(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		OpenTimeout:         100 * time.Millisecond,
		HalfOpenMaxRequests: 3,
	}
	cb := NewCircuitBreaker(config)

	// Trigger circuit to open
	cb.Call(func() error { return errors.New("failure 1") })
	cb.Call(func() error { return errors.New("failure 2") })

	// Wait for transition to half-open
	time.Sleep(150 * time.Millisecond)

	// Execute enough successful requests to close
	cb.Call(func() error { return nil })
	cb.Call(func() error { return nil })

	if cb.State() != CircuitStateClosed {
		t.Errorf("Expected state Closed after successful half-open requests, got %v", cb.State())
	}
}

func TestCircuitBreakerHalfOpenToOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		OpenTimeout:         100 * time.Millisecond,
		HalfOpenMaxRequests: 3,
	}
	cb := NewCircuitBreaker(config)

	// Trigger circuit to open
	cb.Call(func() error { return errors.New("failure 1") })
	cb.Call(func() error { return errors.New("failure 2") })

	// Wait for transition to half-open
	time.Sleep(150 * time.Millisecond)

	// One failure in half-open should re-open
	cb.Call(func() error { return errors.New("failure in half-open") })

	if cb.State() != CircuitStateOpen {
		t.Errorf("Expected state Open after failure in half-open, got %v", cb.State())
	}
}

func TestCircuitBreakerHalfOpenMaxRequests(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    5, // High threshold so we stay in half-open
		OpenTimeout:         100 * time.Millisecond,
		HalfOpenMaxRequests: 2,
	}
	cb := NewCircuitBreaker(config)

	// Trigger circuit to open
	cb.Call(func() error { return errors.New("failure 1") })
	cb.Call(func() error { return errors.New("failure 2") })

	// Wait for transition to half-open
	time.Sleep(150 * time.Millisecond)

	// Allow max requests
	cb.Call(func() error { return nil })
	cb.Call(func() error { return nil })

	// Third request should be blocked (still in half-open state)
	err := cb.Call(func() error { return nil })
	if err == nil {
		t.Error("Expected call to be blocked after max half-open requests")
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(config)

	// Trigger some failures
	cb.Call(func() error { return errors.New("failure 1") })
	cb.Call(func() error { return errors.New("failure 2") })

	// Reset
	cb.Reset()

	if cb.State() != CircuitStateClosed {
		t.Errorf("Expected state Closed after reset, got %v", cb.State())
	}

	stats := cb.Stats()
	if stats["consecutive_failures"].(int) != 0 {
		t.Error("Expected consecutive_failures to be 0 after reset")
	}
}

func TestCircuitBreakerStats(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(config)

	stats := cb.Stats()

	if stats["state"].(string) != "Closed" {
		t.Error("Expected initial state in stats to be Closed")
	}

	if stats["consecutive_failures"].(int) != 0 {
		t.Error("Expected consecutive_failures to be 0 initially")
	}

	// Trigger a failure
	cb.Call(func() error { return errors.New("failure") })

	stats = cb.Stats()
	if stats["consecutive_failures"].(int) != 1 {
		t.Error("Expected consecutive_failures to be 1 after one failure")
	}
}

func TestCircuitStateString(t *testing.T) {
	tests := []struct {
		state    CircuitState
		expected string
	}{
		{CircuitStateClosed, "Closed"},
		{CircuitStateOpen, "Open"},
		{CircuitStateHalfOpen, "HalfOpen"},
		{CircuitState(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.state.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.state.String())
			}
		})
	}
}

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	config := DefaultCircuitBreakerConfig()

	if config.FailureThreshold <= 0 {
		t.Error("FailureThreshold should be positive")
	}

	if config.SuccessThreshold <= 0 {
		t.Error("SuccessThreshold should be positive")
	}

	if config.OpenTimeout <= 0 {
		t.Error("OpenTimeout should be positive")
	}

	if config.HalfOpenMaxRequests <= 0 {
		t.Error("HalfOpenMaxRequests should be positive")
	}
}
