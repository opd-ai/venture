// Package webrtc time_provider.go
// TimeProvider interface enables deterministic timestamps for testing.
// In production, use RealTimeProvider; in tests, use MockTimeProvider.
package webrtc

import "time"

// TimeProvider abstracts time access for deterministic testing of
// networking timestamps, cache expiry, and latency measurements.
type TimeProvider interface {
	// Now returns the current time.
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using the system clock.
type RealTimeProvider struct{}

// Now returns the current system time.
func (RealTimeProvider) Now() time.Time {
	return time.Now()
}

// DefaultTimeProvider returns a RealTimeProvider for production use.
func DefaultTimeProvider() TimeProvider {
	return RealTimeProvider{}
}

// MockTimeProvider implements TimeProvider with a controllable time for testing.
type MockTimeProvider struct {
	CurrentTime time.Time
}

// Now returns the mock time.
func (m *MockTimeProvider) Now() time.Time {
	return m.CurrentTime
}

// SetTime updates the mock time to a specific value.
func (m *MockTimeProvider) SetTime(t time.Time) {
	m.CurrentTime = t
}

// Advance moves the mock time forward by the given duration.
func (m *MockTimeProvider) Advance(d time.Duration) {
	m.CurrentTime = m.CurrentTime.Add(d)
}
