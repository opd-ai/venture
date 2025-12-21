package sprites

import (
	"container/list"
	"fmt"
	"hash"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// hasherPool pools FNV-64a hashers to avoid per-call allocations.
// This eliminates 1 allocation per hashConfig call in hot paths.
var hasherPool = sync.Pool{
	New: func() interface{} {
		return fnv.New64a()
	},
}

// hashBuffer is a reusable byte buffer for building hash strings.
// Thread-local via sync.Pool to avoid allocations in hashConfig.
var hashBufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 128)
		return &buf
	},
}

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
//
// Memory calculation (Phase 45: 64×64 default):
//   - 32×32 RGBA sprite: 4KB (32×32×4)
//   - 64×64 RGBA sprite: 16KB (64×64×4)
//   - 128×128 RGBA sprite: 64KB (128×128×4)
//
// Recommended capacity for <300MB memory target:
//   - 64×64 sprites: 100-200 sprites (~1.6MB-3.2MB)
//   - Mixed sizes: 150 sprites (~3MB average)
//   - Maximum: ~18,750 sprites (using cache.MaxCacheSize / cache.SpriteSize64)
//
// For byte-based caching, use pkg/rendering/cache.SpriteCache instead,
// which manages memory limits directly in bytes using cache.DefaultCacheSize.
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
// Optimized: Uses pooled hasher and buffer to avoid per-call allocations.
func (c *Cache) hashConfig(config Config) uint64 {
	// Get pooled hasher and buffer
	h := hasherPool.Get().(hash.Hash64)
	h.Reset()
	bufPtr := hashBufferPool.Get().(*[]byte)
	buf := (*bufPtr)[:0]

	// Build hash string without fmt.Fprintf allocations
	// Format: "%d|%d|%d|%d|%s|%f|%d"
	buf = strconv.AppendInt(buf, int64(config.Type), 10)
	buf = append(buf, '|')
	buf = strconv.AppendInt(buf, int64(config.Width), 10)
	buf = append(buf, '|')
	buf = strconv.AppendInt(buf, int64(config.Height), 10)
	buf = append(buf, '|')
	buf = strconv.AppendInt(buf, config.Seed, 10)
	buf = append(buf, '|')
	buf = append(buf, config.GenreID...)
	buf = append(buf, '|')
	buf = strconv.AppendFloat(buf, config.Complexity, 'f', -1, 64)
	buf = append(buf, '|')
	buf = strconv.AppendInt(buf, int64(config.Variation), 10)

	h.Write(buf)

	// Hash custom parameters in sorted key order for determinism
	// Go map iteration order is randomized, so we must sort keys
	if config.Custom != nil && len(config.Custom) > 0 {
		keys := make([]string, 0, len(config.Custom))
		for key := range config.Custom {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(h, "|%s=%v", key, config.Custom[key])
		}
	}

	sum := h.Sum64()

	// Return pooled resources
	*bufPtr = buf
	hashBufferPool.Put(bufPtr)
	hasherPool.Put(h)

	return sum
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

// cachedBatchJob represents a sprite generation job in the batch processing pipeline.
type cachedBatchJob struct {
	index  int
	config Config
}

// cachedBatchResult represents the result of a sprite generation job.
type cachedBatchResult struct {
	index  int
	sprite *ebiten.Image
	err    error
}

// generateSequential generates sprites sequentially without concurrency.
func (cg *CachedGenerator) generateSequential(configs []Config, onError func(int, error), onProgress func(int, int)) []*ebiten.Image {
	results := make([]*ebiten.Image, len(configs))
	for i, config := range configs {
		sprite, err := cg.Generate(config)
		if err != nil {
			if onError != nil {
				onError(i, err)
			}
			continue
		}
		results[i] = sprite
		if onProgress != nil {
			onProgress(i+1, len(configs))
		}
	}
	return results
}

// startWorkers starts worker goroutines for parallel sprite generation.
func (cg *CachedGenerator) startWorkers(workers int, jobs <-chan cachedBatchJob, results chan<- cachedBatchResult, wg *sync.WaitGroup) {
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				sprite, err := cg.Generate(job.config)
				results <- cachedBatchResult{
					index:  job.index,
					sprite: sprite,
					err:    err,
				}
			}
		}()
	}
}

// collectCachedResults collects sprite generation results and handles errors and progress.
func collectCachedResults(resultsChan <-chan cachedBatchResult, results []*ebiten.Image, onError func(int, error), onProgress func(int, int), total int) {
	completed := 0
	for res := range resultsChan {
		if res.err != nil {
			if onError != nil {
				onError(res.index, res.err)
			}
			continue
		}
		results[res.index] = res.sprite
		completed++
		if onProgress != nil {
			onProgress(completed, total)
		}
	}
}

// BatchGenerate generates multiple sprites, optionally in parallel.
// Returns a slice of sprites in the same order as configs.
func (cg *CachedGenerator) BatchGenerate(batchConfig BatchConfig) ([]*ebiten.Image, error) {
	if len(batchConfig.Configs) == 0 {
		return nil, nil
	}

	results := make([]*ebiten.Image, len(batchConfig.Configs))

	if !batchConfig.Concurrent || batchConfig.MaxWorkers <= 1 {
		return cg.generateSequential(batchConfig.Configs, batchConfig.OnError, batchConfig.OnProgress), nil
	}

	workers := batchConfig.MaxWorkers
	if workers <= 0 {
		workers = 4
	}

	jobs := make(chan cachedBatchJob, len(batchConfig.Configs))
	resultsChan := make(chan cachedBatchResult, len(batchConfig.Configs))

	var wg sync.WaitGroup
	cg.startWorkers(workers, jobs, resultsChan, &wg)

	go func() {
		for i, config := range batchConfig.Configs {
			jobs <- cachedBatchJob{index: i, config: config}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	collectCachedResults(resultsChan, results, batchConfig.OnError, batchConfig.OnProgress, len(batchConfig.Configs))

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
