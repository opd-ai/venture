// Package engine provides tile caching for efficient rendering.
// This file implements TileCache which caches rendered tiles to avoid
// redundant procedural generation and improve performance.
package engine

import (
	"container/list"
	"fmt"
	"image"
	"sync"

	"github.com/opd-ai/venture/pkg/rendering/tiles"
)

// TileCacheKey uniquely identifies a cached tile.
type TileCacheKey struct {
	TileType tiles.TileType
	GenreID  string
	Seed     int64
	Variant  float64
	Width    int
	Height   int
}

// String returns a string representation of the cache key.
func (k TileCacheKey) String() string {
	return fmt.Sprintf("%s-%s-%d-%.2f-%dx%d", k.TileType, k.GenreID, k.Seed, k.Variant, k.Width, k.Height)
}

// TileCacheEntry holds a cached tile image and its access information.
type tileCacheEntry struct {
	key   TileCacheKey
	image *image.RGBA
	elem  *list.Element // For LRU tracking
}

// TileCache provides LRU caching for procedurally generated tiles.
type TileCache struct {
	mu      sync.RWMutex
	maxSize int
	cache   map[string]*tileCacheEntry
	lruList *list.List
	gen     *tiles.Generator
	hits    uint64
	misses  uint64
}

// NewTileCache creates a new tile cache with the specified maximum size.
func NewTileCache(maxSize int) *TileCache {
	return &TileCache{
		maxSize: maxSize,
		cache:   make(map[string]*tileCacheEntry),
		lruList: list.New(),
		gen:     tiles.NewGenerator(),
	}
}

// Get retrieves a tile from the cache or generates it if not present.
func (c *TileCache) Get(key TileCacheKey) (*image.RGBA, error) {
	keyStr := key.String()

	// Try to get from cache (read lock)
	c.mu.RLock()
	if entry, ok := c.cache[keyStr]; ok {
		c.mu.RUnlock()

		// Move to front of LRU list (write lock)
		c.mu.Lock()
		c.lruList.MoveToFront(entry.elem)
		c.hits++
		c.mu.Unlock()

		return entry.image, nil
	}
	c.mu.RUnlock()

	// Cache miss - generate tile
	c.mu.Lock()
	defer c.mu.Unlock()

	c.misses++

	// Double-check after acquiring write lock (another goroutine might have generated it)
	if entry, ok := c.cache[keyStr]; ok {
		c.lruList.MoveToFront(entry.elem)
		return entry.image, nil
	}

	// Generate the tile
	config := tiles.Config{
		Type:    key.TileType,
		Width:   key.Width,
		Height:  key.Height,
		GenreID: key.GenreID,
		Seed:    key.Seed,
		Variant: key.Variant,
	}

	img, err := c.gen.Generate(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tile: %w", err)
	}

	// Evict LRU entry if cache is full
	if c.lruList.Len() >= c.maxSize {
		c.evictOldest()
	}

	// Add to cache
	elem := c.lruList.PushFront(keyStr)
	c.cache[keyStr] = &tileCacheEntry{
		key:   key,
		image: img,
		elem:  elem,
	}

	return img, nil
}

// evictOldest removes the least recently used entry from the cache.
// Must be called with write lock held.
func (c *TileCache) evictOldest() {
	elem := c.lruList.Back()
	if elem == nil {
		return
	}

	c.lruList.Remove(elem)
	keyStr := elem.Value.(string)
	delete(c.cache, keyStr)
}

// Clear removes all entries from the cache.
func (c *TileCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*tileCacheEntry)
	c.lruList.Init()
	c.hits = 0
	c.misses = 0
}

// Size returns the current number of cached tiles.
func (c *TileCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Stats returns cache hit/miss statistics.
func (c *TileCache) Stats() (hits, misses uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses
}

// HitRate returns the cache hit rate as a percentage.
func (c *TileCache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	if total == 0 {
		return 0.0
	}
	return float64(c.hits) / float64(total) * 100.0
}
