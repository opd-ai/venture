// Package trade provides time abstraction for deterministic testing.
package trade

import "time"

// TimeProvider is an interface for obtaining the current time.
// This enables deterministic testing by injecting mock time providers.
// In production, use RealTimeProvider; in tests, use MockTimeProvider.
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

// MockTimeProvider implements TimeProvider with a controllable time for testing.
type MockTimeProvider struct {
	current time.Time
}

// NewMockTimeProvider creates a mock time provider starting at the given time.
func NewMockTimeProvider(t time.Time) *MockTimeProvider {
	return &MockTimeProvider{current: t}
}

// Now returns the current mock time.
func (m *MockTimeProvider) Now() time.Time {
	return m.current
}

// Advance moves the mock time forward by the given duration.
func (m *MockTimeProvider) Advance(d time.Duration) {
	m.current = m.current.Add(d)
}

// Set sets the mock time to a specific value.
func (m *MockTimeProvider) Set(t time.Time) {
	m.current = t
}
