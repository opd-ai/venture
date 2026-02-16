// time_provider.go defines the TimeProvider interface for deterministic timestamp generation.
// This enables multiplayer synchronization and reproducible testing by replacing time.Now() calls.
package guild_housing

import "time"

// TimeProvider is an interface for obtaining timestamps, enabling deterministic testing.
type TimeProvider interface {
	// Now returns the current time.
	Now() time.Time
}

// RealTimeProvider provides real wall-clock time via time.Now().
type RealTimeProvider struct{}

// Now returns the current wall-clock time.
func (RealTimeProvider) Now() time.Time {
	return time.Now()
}

// FixedTimeProvider provides a fixed timestamp for deterministic testing.
type FixedTimeProvider struct {
	FixedTime time.Time
}

// Now returns the fixed time value.
func (p FixedTimeProvider) Now() time.Time {
	return p.FixedTime
}

// defaultTimeProvider is the package-level time provider instance.
var defaultTimeProvider TimeProvider = RealTimeProvider{}

// SetTimeProvider sets the package-level time provider for testing.
func SetTimeProvider(tp TimeProvider) {
	defaultTimeProvider = tp
}

// ResetTimeProvider resets to the default real time provider.
func ResetTimeProvider() {
	defaultTimeProvider = RealTimeProvider{}
}

// now returns the current time from the configured time provider.
func now() time.Time {
	return defaultTimeProvider.Now()
}
