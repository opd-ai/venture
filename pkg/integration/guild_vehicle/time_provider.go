// time_provider.go defines the TimeProvider interface for deterministic timestamp generation.
// This enables multiplayer synchronization and reproducible testing by replacing time.Now() calls.
package guild_vehicle

import "time"

// TimeProvider is an interface for obtaining timestamps, enabling deterministic testing.
// This interface is package-local by design to avoid coupling with other packages.
// The same pattern exists in cmd/client/, cmd/server/, and several other packages.
type TimeProvider interface {
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

// Now returns the fixed time configured for testing.
func (p FixedTimeProvider) Now() time.Time {
	return p.FixedTime
}

// defaultTimeProvider is the package-level time provider instance.
var defaultTimeProvider TimeProvider = RealTimeProvider{}

// SetTimeProvider sets the package-level time provider.
// This is primarily used for testing to inject a FixedTimeProvider for deterministic timestamps.
// It is NOT thread-safe; call only during test setup or initialization, not concurrently.
// Use ResetTimeProvider to restore default behavior after tests.
func SetTimeProvider(tp TimeProvider) {
	defaultTimeProvider = tp
}

// ResetTimeProvider resets the package-level time provider to the default RealTimeProvider.
// Call this in test cleanup (e.g., defer ResetTimeProvider()) to avoid test pollution.
// It is NOT thread-safe; call only during test teardown, not concurrently.
func ResetTimeProvider() {
	defaultTimeProvider = RealTimeProvider{}
}

// now returns the current time from the configured time provider.
// This is an internal helper used by FleetManager operations for timestamp generation.
// In production, it returns real wall-clock time; in tests with FixedTimeProvider,
// it returns the configured fixed timestamp for deterministic behavior.
func now() time.Time {
	return defaultTimeProvider.Now()
}
