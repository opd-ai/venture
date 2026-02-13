package choice_consequences

import "time"

// time_provider.go defines the TimeProvider interface for deterministic timestamp generation.
// This enables testing with predictable timestamps and supports deterministic game state.

// TimeProvider is an interface for obtaining timestamps, enabling deterministic testing.
type TimeProvider interface {
	// Now returns the current Unix timestamp.
	Now() int64
}

// RealTimeProvider provides real wall-clock time via time.Now().
// This is the default provider for production use.
type RealTimeProvider struct{}

// Now returns the current Unix timestamp from the system clock.
func (RealTimeProvider) Now() int64 {
	return time.Now().Unix()
}

// FixedTimeProvider provides a fixed timestamp for deterministic testing.
type FixedTimeProvider struct {
	Timestamp int64
}

// Now returns the fixed timestamp.
func (p FixedTimeProvider) Now() int64 {
	return p.Timestamp
}

// defaultTimeProvider is the package-level time provider instance.
var defaultTimeProvider TimeProvider = RealTimeProvider{}

// SetTimeProvider sets the package-level time provider for testing.
// Not thread-safe; should only be called during test setup.
func SetTimeProvider(tp TimeProvider) {
	defaultTimeProvider = tp
}

// ResetTimeProvider resets to the default real time provider.
func ResetTimeProvider() {
	defaultTimeProvider = RealTimeProvider{}
}

// now returns the current timestamp from the configured time provider.
func now() int64 {
	return defaultTimeProvider.Now()
}
