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
//
// WARNING: This function is TEST-ONLY and NOT thread-safe. It modifies a
// package-level variable without synchronization. This function should only
// be called from test files during test setup (before any concurrent goroutines
// access the time provider) via t.Cleanup(ResetTimeProvider).
//
// Production code should NEVER call this function. The default RealTimeProvider
// is initialized at package load time and should remain unchanged in production.
//
// Example test usage:
//
//	func TestMyFeature(t *testing.T) {
//	    SetTimeProvider(FixedTimeProvider{Timestamp: 1640000000})
//	    t.Cleanup(ResetTimeProvider)
//	    // ... test code ...
//	}
func SetTimeProvider(tp TimeProvider) {
	defaultTimeProvider = tp
}

// ResetTimeProvider resets to the default real time provider.
// This function is TEST-ONLY and should be used with t.Cleanup() in tests.
func ResetTimeProvider() {
	defaultTimeProvider = RealTimeProvider{}
}

// now returns the current timestamp from the configured time provider.
func now() int64 {
	return defaultTimeProvider.Now()
}
