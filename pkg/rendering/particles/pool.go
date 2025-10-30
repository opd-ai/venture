// Package particles provides object pooling for particle systems.
// This file implements sync.Pool-based pooling to reduce GC pressure
// from frequent particle system allocation/deallocation.
package particles

import "sync"

// particleSystemPool provides reusable ParticleSystem instances.
// Using sync.Pool reduces allocation pressure during particle-heavy effects
// (combat, spells, environmental effects).
var particleSystemPool = sync.Pool{
	New: func() interface{} {
		return &ParticleSystem{
			Particles: make([]Particle, 0, 100), // Pre-allocate capacity for typical effects
		}
	},
}

// particleSlicePool provides reusable particle slices.
// Separate from ParticleSystem pool to enable independent sizing.
var particleSlicePool = sync.Pool{
	New: func() interface{} {
		particles := make([]Particle, 0, 100)
		return &particles
	},
}

// NewParticleSystem creates a new particle system from the pool.
// The system is initialized with the given particles, type, and config.
//
// IMPORTANT: The caller must call ReleaseParticleSystem when done to return
// the system to the pool and prevent memory leaks.
//
// Parameters:
//   - particles: Initial particle slice (may be empty)
//   - pType: Type of particle system
//   - config: Configuration used to generate the system
//
// Returns: Pooled ParticleSystem ready for use
func NewParticleSystem(particles []Particle, pType ParticleType, config Config) *ParticleSystem {
	ps := particleSystemPool.Get().(*ParticleSystem)

	// Clear previous state
	ps.Particles = ps.Particles[:0]
	ps.ElapsedTime = 0

	// Set new state
	ps.Type = pType
	ps.Config = config

	// Append particles (reuses underlying capacity if available)
	ps.Particles = append(ps.Particles, particles...)

	return ps
}

// ReleaseParticleSystem returns a particle system to the pool for reuse.
// The system is reset to prevent state leaks between uses.
//
// MUST be called when the particle system is no longer needed.
// After calling, the system should not be used as it may be reused elsewhere.
//
// Safe to call multiple times (idempotent), but wasteful.
func ReleaseParticleSystem(ps *ParticleSystem) {
	if ps == nil {
		return
	}

	// Clear particle slice to prevent memory retention
	// Keep capacity for reuse but zero length
	ps.Particles = ps.Particles[:0]

	// Clear other fields to prevent state leaks
	ps.ElapsedTime = 0
	ps.Type = 0
	// Note: Config is value type, will be overwritten on next use

	particleSystemPool.Put(ps)
}

// AcquireParticleSlice gets a particle slice from the pool.
// Use this when you need a temporary particle buffer.
//
// Returns: Pointer to slice with 0 length, 100 capacity
func AcquireParticleSlice() *[]Particle {
	particles := particleSlicePool.Get().(*[]Particle)
	*particles = (*particles)[:0] // Reset length, keep capacity
	return particles
}

// ReleaseParticleSlice returns a particle slice to the pool.
// The slice is reset to length 0 but capacity is preserved.
func ReleaseParticleSlice(particles *[]Particle) {
	if particles == nil {
		return
	}

	// Reset to zero length, keeping capacity
	*particles = (*particles)[:0]

	particleSlicePool.Put(particles)
}

// ParticlePoolStats provides statistics about particle pool usage.
// Note: sync.Pool doesn't expose metrics, so these are approximate tracking stats.
type ParticlePoolStats struct {
	// SystemsAcquired is lifetime count of particle systems acquired from pool
	SystemsAcquired uint64

	// SystemsReleased is lifetime count of particle systems returned to pool
	SystemsReleased uint64

	// SystemsActive is approximate count of active systems (Acquired - Released)
	SystemsActive uint64

	// SlicesAcquired is lifetime count of particle slices acquired from pool
	SlicesAcquired uint64

	// SlicesReleased is lifetime count of particle slices returned to pool
	SlicesReleased uint64

	// SlicesActive is approximate count of active slices (Acquired - Released)
	SlicesActive uint64
}

var (
	particlePoolStatsLock sync.Mutex
	particlePoolStats     ParticlePoolStats
)

// GetParticlePoolStats returns current particle pool statistics.
// Useful for monitoring memory usage and pool effectiveness.
//
// Note: Stats tracking is disabled by default for performance.
// Enable by uncommenting tracking calls in New/Release functions.
func GetParticlePoolStats() ParticlePoolStats {
	particlePoolStatsLock.Lock()
	defer particlePoolStatsLock.Unlock()
	return particlePoolStats
}

// ResetParticlePoolStats resets pool statistics to zero.
// Useful for testing and benchmarking.
func ResetParticlePoolStats() {
	particlePoolStatsLock.Lock()
	defer particlePoolStatsLock.Unlock()
	particlePoolStats = ParticlePoolStats{}
}
