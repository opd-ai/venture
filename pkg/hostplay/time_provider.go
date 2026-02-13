// Package hostplay provides time abstraction for deterministic testing.
package hostplay

import "time"

// TimeProvider is an interface for obtaining the current time.
// This enables deterministic testing by injecting mock time providers.
// In production, use RealTimeProvider; in tests, use a mock implementation.
type TimeProvider interface {
	// Now returns the current time.
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using the actual system clock.
type RealTimeProvider struct{}

// Now returns the current system time.
func (RealTimeProvider) Now() time.Time {
	return time.Now()
}

// DefaultTimeProvider returns the default TimeProvider (real system time).
func DefaultTimeProvider() TimeProvider {
	return RealTimeProvider{}
}
