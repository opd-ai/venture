package parallel

import (
	"sync"
	"sync/atomic"
)

// ThreadSafeCache wraps any cache with RWMutex for concurrent access.
// Provides thread-safe get/set operations with minimal lock contention.
type ThreadSafeCache struct {
	mu     sync.RWMutex
	cache  map[string]interface{}
	hits   int64 // Use atomic operations
	misses int64 // Use atomic operations
}

// NewThreadSafeCache creates a new thread-safe cache.
func NewThreadSafeCache() *ThreadSafeCache {
	return &ThreadSafeCache{
		cache: make(map[string]interface{}),
	}
}

// Get retrieves a value from the cache.
// Returns (value, true) if found, (nil, false) if not found.
func (c *ThreadSafeCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	value, found := c.cache[key]
	c.mu.RUnlock()

	if found {
		atomic.AddInt64(&c.hits, 1)
	} else {
		atomic.AddInt64(&c.misses, 1)
	}
	return value, found
}

// Set stores a value in the cache.
func (c *ThreadSafeCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = value
}

// Delete removes a value from the cache.
func (c *ThreadSafeCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
}

// Clear removes all entries from the cache.
func (c *ThreadSafeCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]interface{})
	atomic.StoreInt64(&c.hits, 0)
	atomic.StoreInt64(&c.misses, 0)
}

// Size returns the number of entries in the cache.
func (c *ThreadSafeCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

// HitRate returns the cache hit rate as a percentage (0.0 - 100.0).
func (c *ThreadSafeCache) HitRate() float64 {
	hits := atomic.LoadInt64(&c.hits)
	misses := atomic.LoadInt64(&c.misses)

	total := hits + misses
	if total == 0 {
		return 0.0
	}
	return float64(hits) / float64(total) * 100.0
}

// Stats returns cache statistics.
type CacheStats struct {
	Size    int     // Number of entries
	Hits    int64   // Number of cache hits
	Misses  int64   // Number of cache misses
	HitRate float64 // Hit rate percentage
}

// GetStats returns current cache statistics.
func (c *ThreadSafeCache) GetStats() CacheStats {
	c.mu.RLock()
	size := len(c.cache)
	c.mu.RUnlock()

	hits := atomic.LoadInt64(&c.hits)
	misses := atomic.LoadInt64(&c.misses)

	total := hits + misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100.0
	}

	return CacheStats{
		Size:    size,
		Hits:    hits,
		Misses:  misses,
		HitRate: hitRate,
	}
}

// GetOrCompute retrieves a value from cache or computes it if missing.
// The compute function is only called if the key is not found (cache miss).
// This is thread-safe and prevents redundant computation.
func (c *ThreadSafeCache) GetOrCompute(key string, compute func() interface{}) interface{} {
	// Try read-only access first (fast path)
	c.mu.RLock()
	value, found := c.cache[key]
	c.mu.RUnlock()

	if found {
		atomic.AddInt64(&c.hits, 1)
		return value
	}
	atomic.AddInt64(&c.misses, 1)

	// Compute value (without holding lock)
	computed := compute()

	// Store computed value
	c.mu.Lock()
	// Check again in case another goroutine computed it while we waited
	if existing, found := c.cache[key]; found {
		c.mu.Unlock()
		return existing // Use existing value
	}
	c.cache[key] = computed
	c.mu.Unlock()

	return computed
}

// Keys returns all keys in the cache.
// Returns a snapshot of current keys (safe for iteration).
func (c *ThreadSafeCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}
	return keys
}

// ContainsKey returns true if the cache contains the specified key.
func (c *ThreadSafeCache) ContainsKey(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, found := c.cache[key]
	return found
}
