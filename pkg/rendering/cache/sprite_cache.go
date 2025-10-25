package cache

import (
	"container/list"
	"fmt"
	"hash/fnv"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// CacheKey represents a unique identifier for a cached sprite.
type CacheKey string

// GenerateKey creates a cache key from seed, state, and frame information.
func GenerateKey(seed int64, state string, frame int) CacheKey {
	return CacheKey(fmt.Sprintf("%d:%s:%d", seed, state, frame))
}

// GenerateCompositeKey creates a cache key for composite sprites.
func GenerateCompositeKey(seed int64, layers []string) CacheKey {
	h := fnv.New64a()
	fmt.Fprintf(h, "%d", seed)
	for _, layer := range layers {
		fmt.Fprintf(h, ":%s", layer)
	}
	return CacheKey(fmt.Sprintf("composite:%x", h.Sum64()))
}

// entry represents a single cache entry with its metadata.
type entry struct {
	key   CacheKey
	image *ebiten.Image
	size  int64 // Estimated size in bytes
}

// Statistics holds cache performance metrics.
type Statistics struct {
	Hits       uint64
	Misses     uint64
	Evictions  uint64
	TotalSize  int64
	EntryCount int
}

// HitRate returns the cache hit rate as a value between 0.0 and 1.0.
func (s *Statistics) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0.0
	}
	return float64(s.Hits) / float64(total)
}

// SpriteCache implements an LRU cache for Ebiten images.
type SpriteCache struct {
	mu sync.RWMutex

	// Cache storage
	cache map[CacheKey]*list.Element
	lru   *list.List

	// Configuration
	maxSize int64 // Maximum cache size in bytes

	// Statistics
	stats Statistics
}

// NewSpriteCache creates a new sprite cache with the specified maximum size in bytes.
func NewSpriteCache(maxSize int64) *SpriteCache {
	return &SpriteCache{
		cache:   make(map[CacheKey]*list.Element),
		lru:     list.New(),
		maxSize: maxSize,
	}
}

// Get retrieves a sprite from the cache.
// Returns (image, true) if found, (nil, false) if not found.
func (c *SpriteCache) Get(key CacheKey) (*ebiten.Image, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		// Move to front (most recently used)
		c.lru.MoveToFront(elem)
		c.stats.Hits++
		return elem.Value.(*entry).image, true
	}

	c.stats.Misses++
	return nil, false
}

// Put adds a sprite to the cache.
// If the cache is full, it evicts the least recently used entries.
func (c *SpriteCache) Put(key CacheKey, img *ebiten.Image) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already in cache
	if elem, ok := c.cache[key]; ok {
		// Update existing entry and move to front
		c.lru.MoveToFront(elem)
		elem.Value.(*entry).image = img
		return
	}

	// Estimate size (width * height * 4 bytes per pixel for RGBA)
	bounds := img.Bounds()
	size := int64(bounds.Dx() * bounds.Dy() * 4)

	// Evict entries if necessary
	c.evictIfNeeded(size)

	// Add new entry
	e := &entry{
		key:   key,
		image: img,
		size:  size,
	}
	elem := c.lru.PushFront(e)
	c.cache[key] = elem
	c.stats.TotalSize += size
	c.stats.EntryCount++
}

// evictIfNeeded removes least recently used entries until there's enough space.
func (c *SpriteCache) evictIfNeeded(requiredSize int64) {
	for c.stats.TotalSize+requiredSize > c.maxSize && c.lru.Len() > 0 {
		// Remove least recently used (back of list)
		elem := c.lru.Back()
		if elem != nil {
			c.removeElement(elem)
		}
	}
}

// removeElement removes a specific element from the cache.
func (c *SpriteCache) removeElement(elem *list.Element) {
	c.lru.Remove(elem)
	e := elem.Value.(*entry)
	delete(c.cache, e.key)
	c.stats.TotalSize -= e.size
	c.stats.EntryCount--
	c.stats.Evictions++
}

// Clear removes all entries from the cache.
func (c *SpriteCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[CacheKey]*list.Element)
	c.lru = list.New()
	c.stats.TotalSize = 0
	c.stats.EntryCount = 0
}

// Stats returns a copy of the current cache statistics.
func (c *SpriteCache) Stats() Statistics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// Size returns the current cache size in bytes.
func (c *SpriteCache) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats.TotalSize
}

// Count returns the number of entries in the cache.
func (c *SpriteCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats.EntryCount
}

// MaxSize returns the maximum cache size in bytes.
func (c *SpriteCache) MaxSize() int64 {
	return c.maxSize
}

// SetMaxSize updates the maximum cache size and evicts entries if necessary.
func (c *SpriteCache) SetMaxSize(maxSize int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.maxSize = maxSize
	c.evictIfNeeded(0)
}

// Remove removes a specific entry from the cache.
func (c *SpriteCache) Remove(key CacheKey) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.removeElement(elem)
		return true
	}
	return false
}

// Contains checks if a key exists in the cache without affecting LRU order.
func (c *SpriteCache) Contains(key CacheKey) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.cache[key]
	return ok
}
