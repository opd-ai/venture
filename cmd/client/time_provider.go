//go:build !android && !ios
// +build !android,!ios

package main

import "time"

// TimeProvider abstracts time.Now() for deterministic testing of gameplay code.
// In production, use RealTimeProvider; in tests, use MockTimeProvider.
type TimeProvider interface {
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using the actual system clock.
type RealTimeProvider struct{}

// Now returns the current system time.
func (RealTimeProvider) Now() time.Time {
	return time.Now()
}

// MockTimeProvider implements TimeProvider with a fixed time for deterministic tests.
type MockTimeProvider struct {
	FixedTime time.Time
}

// Now returns the fixed time configured on the mock.
func (m MockTimeProvider) Now() time.Time {
	return m.FixedTime
}

// DefaultTimeProvider returns the default TimeProvider (real system time).
func DefaultTimeProvider() TimeProvider {
	return RealTimeProvider{}
}
