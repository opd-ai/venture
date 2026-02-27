// Package learning - time_provider.go
// TimeProvider interface enables deterministic timestamps for testing and reproducible state.
// In production, use RealTimeProvider; in tests, use a mock implementation.

package learning

import "time"

// TimeProvider is an interface for obtaining the current time.
// This enables deterministic timestamps for testing, save/load consistency,
// and network synchronization across game clients.
type TimeProvider interface {
	// Now returns the current time
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using the actual system clock.
// Use this in production when real wall-clock time is acceptable.
type RealTimeProvider struct{}

// Now returns the current system time.
// This uses time.Now() intentionally for production behavior - companion learning
// progression is based on real elapsed time, not procedural generation seeds.
// For deterministic testing, use MockTimeProvider instead.
func (RealTimeProvider) Now() time.Time {
	return time.Now() // Intentional exception to determinism guideline - companion metadata, not procgen
}

// DefaultTimeProvider returns the default TimeProvider (real system time).
func DefaultTimeProvider() TimeProvider {
	return RealTimeProvider{}
}

// MockTimeProvider implements TimeProvider with a fixed time for testing.
// Useful for deterministic tests and save/load verification.
type MockTimeProvider struct {
	CurrentTime time.Time
}

// Now returns the mock time.
func (m MockTimeProvider) Now() time.Time {
	return m.CurrentTime
}

// SetTime updates the mock time.
func (m *MockTimeProvider) SetTime(t time.Time) {
	m.CurrentTime = t
}

// Advance moves the mock time forward by the specified duration.
func (m *MockTimeProvider) Advance(d time.Duration) {
	m.CurrentTime = m.CurrentTime.Add(d)
}
