// Package federation provides server-to-server federation protocol.
package federation

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialDelay is the delay before the first retry
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Multiplier is the backoff multiplier (typically 2.0 for exponential backoff)
	Multiplier float64
	// Jitter adds randomness to prevent thundering herd (0.0-1.0)
	Jitter float64
}

// DefaultRetryConfig returns a retry configuration with sensible defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1, // 10% jitter
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// IsRetryable is a function that determines if an error is retryable
type IsRetryable func(error) bool

// RetryStrategy implements exponential backoff with jitter for retry logic
type RetryStrategy struct {
	config RetryConfig
	rng    *rand.Rand
	logger *logrus.Entry
}

// NewRetryStrategy creates a new retry strategy with the given configuration
func NewRetryStrategy(config RetryConfig) *RetryStrategy {
	return &RetryStrategy{
		config: config,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		logger: logrus.WithFields(logrus.Fields{
			"component": "retry_strategy",
		}),
	}
}

// Execute executes the given function with retry logic
// If isRetryable is nil, all errors are considered retryable
func (r *RetryStrategy) Execute(fn RetryableFunc, isRetryable IsRetryable) error {
	if isRetryable == nil {
		// Default: all errors are retryable
		isRetryable = func(err error) bool { return err != nil }
	}

	var lastErr error
	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := r.calculateDelay(attempt)
			r.logger.WithFields(logrus.Fields{
				"attempt": attempt,
				"delay":   delay,
			}).Debug("Retrying after delay")
			time.Sleep(delay)
		}

		lastErr = fn()
		if lastErr == nil {
			if attempt > 0 {
				r.logger.WithFields(logrus.Fields{
					"attempt": attempt,
				}).Info("Retry succeeded")
			}
			return nil
		}

		if !isRetryable(lastErr) {
			r.logger.WithFields(logrus.Fields{
				"attempt": attempt,
				"error":   lastErr,
			}).Debug("Error is not retryable, aborting")
			return lastErr
		}

		r.logger.WithFields(logrus.Fields{
			"attempt": attempt,
			"error":   lastErr,
		}).Debug("Attempt failed, will retry if attempts remain")
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", r.config.MaxRetries, lastErr)
}

// calculateDelay calculates the delay before the next retry attempt
// Uses exponential backoff with jitter to prevent thundering herd
func (r *RetryStrategy) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff: initialDelay * multiplier^(attempt-1)
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.Multiplier, float64(attempt-1))

	// Cap at max delay
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	// Add jitter: delay * (1 ± jitter)
	if r.config.Jitter > 0 {
		jitterRange := delay * r.config.Jitter
		jitter := (r.rng.Float64() * 2 - 1) * jitterRange // Random value in [-jitterRange, +jitterRange]
		delay += jitter
	}

	// Ensure non-negative
	if delay < 0 {
		delay = 0
	}

	return time.Duration(delay)
}

// IsNetworkError is a helper function that determines if an error is a network error
// that should be retried (connection refused, timeout, temporary errors)
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Check error message for common network error patterns
	errMsg := err.Error()
	networkPatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"no route to host",
		"broken pipe",
	}

	for _, pattern := range networkPatterns {
		if containsString(errMsg, pattern) {
			return true
		}
	}

	return false
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && findSubstring(s, substr)))
}

// findSubstring performs a case-insensitive substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			// Simple case-insensitive comparison
			if c1 != c2 && (c1|32) != (c2|32) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
