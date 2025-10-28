package engine

import "sync"

// statusEffectPool is a global object pool for StatusEffectComponent instances.
// It reduces GC pressure during combat by reusing StatusEffectComponent objects
// instead of allocating new ones on every status effect application.
//
// The pool uses sync.Pool which provides automatic memory management:
// - Objects may be automatically removed during GC if memory pressure is high
// - Thread-safe concurrent access
// - Zero-allocation for Get/Put in the happy path
var statusEffectPool = sync.Pool{
	New: func() interface{} {
		// Allocate new StatusEffectComponent when pool is empty
		return &StatusEffectComponent{}
	},
}

// NewStatusEffectComponent acquires a StatusEffectComponent from the pool
// and initializes it with the provided values.
//
// This replaces direct allocation: &StatusEffectComponent{...}
//
// The component must be returned to the pool via ReleaseStatusEffect when
// it's no longer needed to enable reuse and reduce GC pressure.
//
// Usage:
//
//	effect := NewStatusEffectComponent("poison", 5.0, 10.0, 1.0)
//	entity.AddComponent(effect)
//	// ... later when effect expires ...
//	ReleaseStatusEffect(effect)
func NewStatusEffectComponent(effectType string, magnitude, duration, tickInterval float64) *StatusEffectComponent {
	effect := statusEffectPool.Get().(*StatusEffectComponent)

	// Initialize with provided values
	effect.EffectType = effectType
	effect.Magnitude = magnitude
	effect.Duration = duration
	effect.TickInterval = tickInterval
	effect.NextTick = tickInterval

	return effect
}

// ReleaseStatusEffect returns a StatusEffectComponent to the pool for reuse.
// The component is reset to clear its state before being returned to the pool.
//
// This method should be called when the status effect expires or is removed
// from an entity. After calling this method, the component should not be used
// anymore as it may be reused by another part of the system.
//
// Usage:
//
//	entity.RemoveComponent(effect.Type())
//	ReleaseStatusEffect(effect)
//
// Note: It's safe to call this method multiple times on the same component
// (idempotent), but it's wasteful as it adds unnecessary overhead.
func ReleaseStatusEffect(effect *StatusEffectComponent) {
	if effect == nil {
		return
	}

	// Reset state to prevent memory leaks and ensure clean reuse
	effect.Reset()

	// Return to pool
	statusEffectPool.Put(effect)
}

// StatusEffectPoolStats provides statistics about the status effect pool.
// Note: sync.Pool doesn't expose internal metrics, so these are tracking
// statistics based on pool operations.
type StatusEffectPoolStats struct {
	// Acquired is the number of effects acquired from the pool (lifetime)
	Acquired uint64

	// Released is the number of effects returned to the pool (lifetime)
	Released uint64

	// Active is the approximate number of effects currently in use
	// (Acquired - Released)
	Active uint64
}

var (
	poolStatsLock sync.Mutex
	poolStats     StatusEffectPoolStats
)

// GetStatusEffectPoolStats returns current statistics about the pool.
// These stats are useful for monitoring memory usage and pool effectiveness.
//
// Note: Stats tracking is disabled by default for performance. Enable by
// uncommenting the trackAcquire/trackRelease calls in New/Release functions.
func GetStatusEffectPoolStats() StatusEffectPoolStats {
	poolStatsLock.Lock()
	defer poolStatsLock.Unlock()
	return poolStats
}

// ResetStatusEffectPoolStats resets the pool statistics to zero.
// This is useful for testing and benchmarking.
func ResetStatusEffectPoolStats() {
	poolStatsLock.Lock()
	defer poolStatsLock.Unlock()
	poolStats = StatusEffectPoolStats{}
}
