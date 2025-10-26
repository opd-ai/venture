//go:build !test
// +build !test

package sprites

import (
	"container/list"
	"fmt"
	"hash/fnv"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// Cache is an LRU (Least Recently Used) cache for generated sprites.
// It stores sprites by their configuration hash to avoid regenerating
// identical sprites, improving performance during gameplay.
type Cache struct {
	capacity int
	cache    map[uint64]*cacheEntry
	lruList  *list.List
	mutex    sync.RWMutex
	hits     uint64
	misses   uint64
}

// cacheEntry represents a cached sprite with its LRU list element.
type cacheEntry struct {
	key     uint64
	sprite  *ebiten.Image
	element *list.Element
}

// CacheStats contains cache performance statistics.
type CacheStats struct {
	Hits     uint64
	Misses   uint64
	Size     int
	Capacity int
	HitRate  float64
}

// NewCache creates a new sprite cache with the specified capacity.
// capacity specifies the maximum number of sprites to cache.
// Recommended: 100-200 sprites (~500KB-2MB memory).
func NewCache(capacity int) *Cache {
	if capacity <= 0 {
		capacity = 100 // Default capacity
	}

	return &Cache{
		capacity: capacity,
		cache:    make(map[uint64]*cacheEntry, capacity),
		lruList:  list.New(),
		hits:     0,
		misses:   0,
	}
}

// Get retrieves a sprite from the cache by configuration.
// Returns nil if not found.
func (c *Cache) Get(config Config) *ebiten.Image {
	key := c.hashConfig(config)

	c.mutex.RLock()
	entry, found := c.cache[key]
	c.mutex.RUnlock()

	if found {
		c.mutex.Lock()
		// Move to front of LRU list (most recently used)
		c.lruList.MoveToFront(entry.element)
		c.hits++
		c.mutex.Unlock()
		return entry.sprite
	}

	c.mutex.Lock()
	c.misses++
	c.mutex.Unlock()
	return nil
}

// Put adds a sprite to the cache with the given configuration.
// If cache is full, evicts the least recently used sprite.
func (c *Cache) Put(config Config, sprite *ebiten.Image) {
	if sprite == nil {
		return
	}

	key := c.hashConfig(config)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if already exists
	if entry, found := c.cache[key]; found {
		// Update sprite and move to front
		entry.sprite = sprite
		c.lruList.MoveToFront(entry.element)
		return
	}

	// Evict if at capacity
	if c.lruList.Len() >= c.capacity {
		c.evictLRU()
	}

	// Add new entry
	element := c.lruList.PushFront(key)
	c.cache[key] = &cacheEntry{
		key:     key,
		sprite:  sprite,
		element: element,
	}
}

// evictLRU removes the least recently used sprite from the cache.
// Must be called with mutex locked.
func (c *Cache) evictLRU() {
	element := c.lruList.Back()
	if element == nil {
		return
	}

	key := element.Value.(uint64)
	c.lruList.Remove(element)
	delete(c.cache, key)
}

// Clear removes all sprites from the cache.
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[uint64]*cacheEntry, c.capacity)
	c.lruList.Init()
	c.hits = 0
	c.misses = 0
}

// Stats returns cache performance statistics.
func (c *Cache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:     c.hits,
		Misses:   c.misses,
		Size:     len(c.cache),
		Capacity: c.capacity,
		HitRate:  hitRate,
	}
}

// Size returns the current number of cached sprites.
func (c *Cache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

// Capacity returns the maximum number of sprites the cache can hold.
func (c *Cache) Capacity() int {
	return c.capacity
}

// SetCapacity changes the cache capacity.
// If new capacity is smaller, evicts LRU entries until size <= capacity.
func (c *Cache) SetCapacity(capacity int) {
	if capacity <= 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.capacity = capacity

	// Evict entries if over capacity
	for c.lruList.Len() > c.capacity {
		c.evictLRU()
	}
}

// hashConfig generates a hash key for a sprite configuration.
// Uses FNV-1a hash for fast, deterministic hashing.
func (c *Cache) hashConfig(config Config) uint64 {
	h := fnv.New64a()

	// Hash all relevant config fields
	fmt.Fprintf(h, "%d|%d|%d|%d|%s|%f|%d",
		config.Type,
		config.Width,
		config.Height,
		config.Seed,
		config.GenreID,
		config.Complexity,
		config.Variation,
	)

	// Hash custom parameters
	if config.Custom != nil {
		// Hash important custom fields that affect sprite generation
		for key, value := range config.Custom {
			fmt.Fprintf(h, "|%s=%v", key, value)
		}
	}

	return h.Sum64()
}

// CachedGenerator wraps a Generator with caching functionality.
type CachedGenerator struct {
	generator *Generator
	cache     *Cache
	enabled   bool
}

// NewCachedGenerator creates a generator with sprite caching.
func NewCachedGenerator(capacity int) *CachedGenerator {
	return &CachedGenerator{
		generator: NewGenerator(),
		cache:     NewCache(capacity),
		enabled:   true,
	}
}

// Generate generates a sprite using cache when possible.
func (cg *CachedGenerator) Generate(config Config) (*ebiten.Image, error) {
	// Try cache first if enabled
	if cg.enabled {
		if sprite := cg.cache.Get(config); sprite != nil {
			return sprite, nil
		}
	}

	// Cache miss - generate new sprite
	sprite, err := cg.generator.Generate(config)
	if err != nil {
		return nil, err
	}

	// Store in cache if enabled
	if cg.enabled {
		cg.cache.Put(config, sprite)
	}

	return sprite, nil
}

// Cache returns the underlying cache for direct access.
func (cg *CachedGenerator) Cache() *Cache {
	return cg.cache
}

// SetCacheEnabled enables or disables caching.
func (cg *CachedGenerator) SetCacheEnabled(enabled bool) {
	cg.enabled = enabled
}

// IsCacheEnabled returns whether caching is currently enabled.
func (cg *CachedGenerator) IsCacheEnabled() bool {
	return cg.enabled
}

// ClearCache clears the sprite cache.
func (cg *CachedGenerator) ClearCache() {
	cg.cache.Clear()
}

// Stats returns cache performance statistics.
func (cg *CachedGenerator) Stats() CacheStats {
	return cg.cache.Stats()
}

// BatchConfig contains configuration for batch sprite generation.
type BatchConfig struct {
	Configs    []Config
	Concurrent bool
	MaxWorkers int
	OnProgress func(completed, total int)
	OnError    func(index int, err error)
}

// BatchGenerate generates multiple sprites, optionally in parallel.
// Returns a slice of sprites in the same order as configs.
func (cg *CachedGenerator) BatchGenerate(batchConfig BatchConfig) ([]*ebiten.Image, error) {
	if len(batchConfig.Configs) == 0 {
		return nil, nil
	}

	results := make([]*ebiten.Image, len(batchConfig.Configs))

	if !batchConfig.Concurrent || batchConfig.MaxWorkers <= 1 {
		// Sequential generation
		for i, config := range batchConfig.Configs {
			sprite, err := cg.Generate(config)
			if err != nil {
				if batchConfig.OnError != nil {
					batchConfig.OnError(i, err)
				}
				continue
			}
			results[i] = sprite

			if batchConfig.OnProgress != nil {
				batchConfig.OnProgress(i+1, len(batchConfig.Configs))
			}
		}
		return results, nil
	}

	// Parallel generation
	workers := batchConfig.MaxWorkers
	if workers <= 0 {
		workers = 4 // Default
	}

	type job struct {
		index  int
		config Config
	}

	type result struct {
		index  int
		sprite *ebiten.Image
		err    error
	}

	jobs := make(chan job, len(batchConfig.Configs))
	resultsChan := make(chan result, len(batchConfig.Configs))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				sprite, err := cg.Generate(job.config)
				resultsChan <- result{
					index:  job.index,
					sprite: sprite,
					err:    err,
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for i, config := range batchConfig.Configs {
			jobs <- job{index: i, config: config}
		}
		close(jobs)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	completed := 0
	for res := range resultsChan {
		if res.err != nil {
			if batchConfig.OnError != nil {
				batchConfig.OnError(res.index, res.err)
			}
			continue
		}
		results[res.index] = res.sprite
		completed++

		if batchConfig.OnProgress != nil {
			batchConfig.OnProgress(completed, len(batchConfig.Configs))
		}
	}

	return results, nil
}

// Prewarm generates and caches sprites for common configurations.
// Useful during loading screens to populate cache before gameplay.
func (cg *CachedGenerator) Prewarm(configs []Config) error {
	_, err := cg.BatchGenerate(BatchConfig{
		Configs:    configs,
		Concurrent: true,
		MaxWorkers: 4,
	})
	return err
}
