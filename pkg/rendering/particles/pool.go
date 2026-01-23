// Package particles provides object pooling for particle systems.
// This file implements sync.Pool-based pooling to reduce GC pressure
// from frequent particle system allocation/deallocation.
package particles

import (
	"math/rand"
	"sync"
	"time"
)

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
	ps, ok := particleSystemPool.Get().(*ParticleSystem)
	if !ok {
		// Pool returned unexpected type, create new instance
		ps = &ParticleSystem{}
	}

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
	obj := particleSlicePool.Get()
	particles, ok := obj.(*[]Particle)
	if !ok {
		// Should not happen but create new slice if type assertion fails
		newSlice := make([]Particle, 0, 100)
		return &newSlice
	}
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

// rngSourcePool provides reusable *rand.Rand instances.
// This reduces allocations from math/rand.newSource which accounts for
// ~29% of memory allocation in particle generation (per AUDIT.md).
var rngSourcePool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(0))
	},
}

// AcquireRNG gets a seeded random number generator from the pool.
// The RNG is seeded with the provided seed for deterministic generation.
//
// IMPORTANT: Call ReleaseRNG when done to return to pool.
func AcquireRNG(seed int64) *rand.Rand {
	rng, ok := rngSourcePool.Get().(*rand.Rand)
	if !ok {
		return rand.New(rand.NewSource(seed))
	}
	rng.Seed(seed)
	return rng
}

// ReleaseRNG returns a random number generator to the pool.
// After calling, the RNG should not be used as it may be reused elsewhere.
func ReleaseRNG(rng *rand.Rand) {
	if rng == nil {
		return
	}
	rngSourcePool.Put(rng)
}

// weatherSystemPool provides reusable WeatherSystem instances.
// Weather generation allocates ~457KB per event; pooling reduces GC pressure.
var weatherSystemPool = sync.Pool{
	New: func() interface{} {
		return &WeatherSystem{
			Particles: make([]Particle, 0, 2400), // Pre-allocate for typical weather
			Effects:   NewWeatherEffect(),
		}
	},
}

// AcquireWeatherSystem gets a WeatherSystem from the pool.
// The system is reset and ready for configuration.
//
// IMPORTANT: Call ReleaseWeatherSystem when done.
func AcquireWeatherSystem() *WeatherSystem {
	ws, ok := weatherSystemPool.Get().(*WeatherSystem)
	if !ok {
		return &WeatherSystem{
			Particles: make([]Particle, 0, 2400),
			Effects:   NewWeatherEffect(),
		}
	}
	// Reset state
	ws.Particles = ws.Particles[:0]
	ws.ElapsedTime = 0
	if ws.Effects == nil {
		ws.Effects = NewWeatherEffect()
	} else {
		// Clear maps
		for k := range ws.Effects.Puddles {
			delete(ws.Effects.Puddles, k)
		}
		for k := range ws.Effects.SnowLevel {
			delete(ws.Effects.SnowLevel, k)
		}
		ws.Effects.VisibilityModifier = 1.0
		ws.Effects.WindDriftX = 0
		ws.Effects.WindDriftY = 0
	}
	return ws
}

// ReleaseWeatherSystem returns a WeatherSystem to the pool.
// After calling, the system should not be used.
func ReleaseWeatherSystem(ws *WeatherSystem) {
	if ws == nil {
		return
	}
	// Keep the particle slice capacity, just reset length
	ws.Particles = ws.Particles[:0]
	ws.rng = nil // Don't keep reference to RNG
	weatherSystemPool.Put(ws)
}

// ambienceSystemPool provides reusable AmbienceSystem instances.
// Ambience generation allocates ~12.6KB per event; pooling reduces GC pressure.
var ambienceSystemPool = sync.Pool{
	New: func() interface{} {
		return &AmbienceSystem{
			Particles: make([]Particle, 0, 100), // Pre-allocate for typical ambience
		}
	},
}

// AcquireAmbienceSystem gets an AmbienceSystem from the pool.
// The system is reset and ready for configuration.
//
// IMPORTANT: Call ReleaseAmbienceSystem when done.
func AcquireAmbienceSystem() *AmbienceSystem {
	as, ok := ambienceSystemPool.Get().(*AmbienceSystem)
	if !ok {
		return &AmbienceSystem{
			Particles: make([]Particle, 0, 100),
		}
	}
	// Reset state
	as.Particles = as.Particles[:0]
	as.ElapsedTime = 0
	as.respawnCounter = 0
	as.rng = nil
	return as
}

// ReleaseAmbienceSystem returns an AmbienceSystem to the pool.
// After calling, the system should not be used.
func ReleaseAmbienceSystem(as *AmbienceSystem) {
	if as == nil {
		return
	}
	// Keep the particle slice capacity, just reset length
	as.Particles = as.Particles[:0]
	as.rng = nil
	ambienceSystemPool.Put(as)
}

// ambienceCacheEntry stores a cached ambience system with metadata.
type ambienceCacheEntry struct {
	system    *AmbienceSystem
	lastUsed  int64 // Unix nano timestamp
	accessCnt uint64
}

// ambienceCache provides an LRU cache for ambience systems by environment type.
// This reduces allocation during area transitions by reusing previously generated ambience.
type ambienceCache struct {
	mu        sync.RWMutex
	entries   map[ambienceCacheKey]*ambienceCacheEntry
	maxSize   int
	hitCount  uint64
	missCount uint64
}

// ambienceCacheKey identifies a unique ambience configuration.
type ambienceCacheKey struct {
	Type    EnvironmentType
	Width   int
	Height  int
	Seed    int64
	GenreID string
}

// globalAmbienceCache is the singleton cache instance.
var globalAmbienceCache = &ambienceCache{
	entries: make(map[ambienceCacheKey]*ambienceCacheEntry),
	maxSize: 16, // Cache up to 16 different ambience configurations
}

// makeAmbienceCacheKey creates a cache key from an AmbienceConfig.
func makeAmbienceCacheKey(config AmbienceConfig) ambienceCacheKey {
	return ambienceCacheKey{
		Type:    config.Type,
		Width:   config.Width,
		Height:  config.Height,
		Seed:    config.Seed,
		GenreID: config.GenreID,
	}
}

// get retrieves a cached ambience system, returning nil if not found.
func (c *ambienceCache) get(key ambienceCacheKey) *AmbienceSystem {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		c.missCount++
		return nil
	}

	c.hitCount++
	entry.lastUsed = nanoTime()
	entry.accessCnt++
	return entry.system
}

// put stores an ambience system in the cache, evicting LRU entries if necessary.
func (c *ambienceCache) put(key ambienceCacheKey, system *AmbienceSystem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entry if at capacity
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[key] = &ambienceCacheEntry{
		system:    system,
		lastUsed:  nanoTime(),
		accessCnt: 1,
	}
}

// evictOldest removes the least recently used entry.
// Caller must hold the lock.
func (c *ambienceCache) evictOldest() {
	var oldestKey ambienceCacheKey
	var oldestTime int64 = 1<<63 - 1 // max int64

	for k, v := range c.entries {
		if v.lastUsed < oldestTime {
			oldestTime = v.lastUsed
			oldestKey = k
		}
	}

	delete(c.entries, oldestKey)
}

// nanoTime returns current time in nanoseconds for LRU ordering.
func nanoTime() int64 {
	return time.Now().UnixNano()
}

// Stats returns cache hit/miss statistics.
func (c *ambienceCache) Stats() (hits, misses uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hitCount, c.missCount
}

// Clear removes all entries from the cache.
func (c *ambienceCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[ambienceCacheKey]*ambienceCacheEntry)
	c.hitCount = 0
	c.missCount = 0
}

// Size returns the number of cached entries.
func (c *ambienceCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// GetAmbienceCacheStats returns hit rate statistics for the global cache.
func GetAmbienceCacheStats() (hits, misses uint64, size int) {
	hits, misses = globalAmbienceCache.Stats()
	size = globalAmbienceCache.Size()
	return hits, misses, size
}

// ClearAmbienceCache clears the global ambience cache.
func ClearAmbienceCache() {
	globalAmbienceCache.Clear()
}

// DebrisContext provides pre-allocated buffers for debris particle updates.
// Reusing this context across UpdateDebrisPooled calls reduces allocations
// from ~47KB to near-zero per update.
type DebrisContext struct {
	// SpatialHash is a reusable spatial hash for collision detection
	SpatialHash *SpatialHash

	// candidateBuffer stores candidate particle indices from spatial queries
	candidateBuffer []int

	// neighborsBuffer stores confirmed neighbor particle indices
	neighborsBuffer []int
}

// debrisContextPool provides reusable DebrisContext instances.
// Debris updates allocate ~47KB per call (per AUDIT.md); pooling eliminates this.
var debrisContextPool = sync.Pool{
	New: func() interface{} {
		return &DebrisContext{
			SpatialHash:     NewSpatialHash(8.0, -1000, -1000, 1000, 1000), // Default cell size
			candidateBuffer: make([]int, 0, 128),
			neighborsBuffer: make([]int, 0, 64),
		}
	},
}

// AcquireDebrisContext gets a DebrisContext from the pool.
// The context is reset and ready for use with UpdateDebrisPooled.
//
// IMPORTANT: Call ReleaseDebrisContext when done to return to pool.
func AcquireDebrisContext() *DebrisContext {
	ctx, ok := debrisContextPool.Get().(*DebrisContext)
	if !ok {
		return &DebrisContext{
			SpatialHash:     NewSpatialHash(8.0, -1000, -1000, 1000, 1000),
			candidateBuffer: make([]int, 0, 128),
			neighborsBuffer: make([]int, 0, 64),
		}
	}
	// Reset buffers but keep capacity
	ctx.candidateBuffer = ctx.candidateBuffer[:0]
	ctx.neighborsBuffer = ctx.neighborsBuffer[:0]
	// Spatial hash will be cleared in UpdateDebrisPooled
	return ctx
}

// ReleaseDebrisContext returns a DebrisContext to the pool.
// After calling, the context should not be used as it may be reused elsewhere.
func ReleaseDebrisContext(ctx *DebrisContext) {
	if ctx == nil {
		return
	}
	// Clear spatial hash to free map entries
	if ctx.SpatialHash != nil {
		ctx.SpatialHash.Clear()
	}
	ctx.candidateBuffer = ctx.candidateBuffer[:0]
	ctx.neighborsBuffer = ctx.neighborsBuffer[:0]
	debrisContextPool.Put(ctx)
}
