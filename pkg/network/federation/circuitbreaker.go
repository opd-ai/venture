// Package federation provides server-to-server federation protocol.
package federation

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CircuitState represents the current state of a circuit breaker
type CircuitState int

const (
	// CircuitStateClosed means requests are allowed (normal operation)
	CircuitStateClosed CircuitState = iota
	// CircuitStateOpen means requests are blocked (system failing)
	CircuitStateOpen
	// CircuitStateHalfOpen means limited requests are allowed (testing recovery)
	CircuitStateHalfOpen
)

// String returns a human-readable name for the circuit state
func (s CircuitState) String() string {
	switch s {
	case CircuitStateClosed:
		return "Closed"
	case CircuitStateOpen:
		return "Open"
	case CircuitStateHalfOpen:
		return "HalfOpen"
	default:
		return "Unknown"
	}
}

// CircuitBreakerConfig holds configuration for a circuit breaker
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures before opening
	FailureThreshold int
	// SuccessThreshold is the number of consecutive successes needed to close from half-open
	SuccessThreshold int
	// OpenTimeout is how long to wait before transitioning from open to half-open
	OpenTimeout time.Duration
	// HalfOpenMaxRequests is the maximum number of requests allowed in half-open state
	HalfOpenMaxRequests int
}

// DefaultCircuitBreakerConfig returns a circuit breaker configuration with sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:    5,                // Open after 5 consecutive failures
		SuccessThreshold:    2,                // Close after 2 consecutive successes in half-open
		OpenTimeout:         30 * time.Second, // Wait 30s before trying half-open
		HalfOpenMaxRequests: 3,                // Allow 3 requests in half-open state
	}
}

// CircuitBreaker implements the circuit breaker pattern for remote server connections
type CircuitBreaker struct {
	mu                   sync.RWMutex
	state                CircuitState
	config               CircuitBreakerConfig
	consecutiveFailures  int
	consecutiveSuccesses int
	lastFailureTime      time.Time
	halfOpenRequests     int
	logger               *logrus.Entry
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		state:  CircuitStateClosed,
		config: config,
		logger: logrus.WithFields(logrus.Fields{
			"component": "circuit_breaker",
		}),
	}
}

// Call executes the given function if the circuit breaker allows it
// Returns an error if the circuit is open or if the function fails
//
// Note: The circuit breaker lock is released between beforeRequest() and afterRequest()
// to avoid holding the lock during the potentially long-running fn() execution.
// This means the circuit state may change between the check and the result recording,
// which is acceptable as it only affects edge cases and maintains better performance.
func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.beforeRequest() {
		return fmt.Errorf("circuit breaker is open")
	}

	err := fn()
	cb.afterRequest(err)
	return err
}

// beforeRequest checks if a request should be allowed based on current state
func (cb *CircuitBreaker) beforeRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitStateClosed:
		return true
	case CircuitStateOpen:
		// Check if enough time has passed to transition to half-open
		if time.Since(cb.lastFailureTime) > cb.config.OpenTimeout {
			cb.transitionTo(CircuitStateHalfOpen)
			cb.halfOpenRequests = 1 // Count this request
			return true
		}
		return false
	case CircuitStateHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenRequests < cb.config.HalfOpenMaxRequests {
			cb.halfOpenRequests++
			return true
		}
		return false
	default:
		return false
	}
}

// afterRequest updates the circuit breaker state based on request result
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure() {
	cb.consecutiveFailures++
	cb.consecutiveSuccesses = 0
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case CircuitStateClosed:
		if cb.consecutiveFailures >= cb.config.FailureThreshold {
			cb.transitionTo(CircuitStateOpen)
		}
	case CircuitStateHalfOpen:
		// Any failure in half-open state immediately opens the circuit
		cb.transitionTo(CircuitStateOpen)
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess() {
	cb.consecutiveSuccesses++
	cb.consecutiveFailures = 0

	if cb.state == CircuitStateHalfOpen {
		if cb.consecutiveSuccesses >= cb.config.SuccessThreshold {
			cb.transitionTo(CircuitStateClosed)
		}
	}
}

// transitionTo changes the circuit breaker state and logs the transition
func (cb *CircuitBreaker) transitionTo(newState CircuitState) {
	oldState := cb.state
	cb.state = newState

	// Reset half-open request counter when transitioning to closed or open
	if newState == CircuitStateClosed || newState == CircuitStateOpen {
		cb.halfOpenRequests = 0
	}

	cb.logger.WithFields(logrus.Fields{
		"old_state":             oldState.String(),
		"new_state":             newState.String(),
		"consecutive_failures":  cb.consecutiveFailures,
		"consecutive_successes": cb.consecutiveSuccesses,
	}).Info("Circuit breaker state transition")
}

// State returns the current circuit breaker state (thread-safe)
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state with zero failures
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitStateClosed
	cb.consecutiveFailures = 0
	cb.consecutiveSuccesses = 0
	cb.halfOpenRequests = 0
	cb.logger.Info("Circuit breaker reset to closed state")
}

// Stats returns current circuit breaker statistics
func (cb *CircuitBreaker) Stats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":                 cb.state.String(),
		"consecutive_failures":  cb.consecutiveFailures,
		"consecutive_successes": cb.consecutiveSuccesses,
		"last_failure_time":     cb.lastFailureTime,
		"half_open_requests":    cb.halfOpenRequests,
	}
}
