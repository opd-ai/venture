package animation

import (
	"container/list"
	"fmt"
	"strconv"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// CacheKey uniquely identifies an animation sequence.
type CacheKey struct {
	Seed       int64
	State      string
	Direction  Direction8
	FrameIndex int
}

// String returns the string representation of the cache key.
// Optimized to use strconv instead of fmt.Sprintf to reduce allocations.
func (k CacheKey) String() string {
	// Estimate buffer size: seed ~20 chars, state ~10 chars, direction ~10 chars, frame ~5 chars, 3 colons
	// Pre-allocate buffer to avoid intermediate allocations
	buf := make([]byte, 0, 48+len(k.State))
	buf = strconv.AppendInt(buf, k.Seed, 10)
	buf = append(buf, ':')
	buf = append(buf, k.State...)
	buf = append(buf, ':')
	buf = append(buf, k.Direction.String()...)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(k.FrameIndex), 10)
	return string(buf)
}

// cacheEntry represents a single cached animation frame.
type cacheEntry struct {
	key   CacheKey
	frame *ebiten.Image
	size  int64 // Estimated size in bytes (width * height * 4 bytes per pixel)
}

// CacheStats holds cache performance statistics.
type CacheStats struct {
	Hits       uint64
	Misses     uint64
	Evictions  uint64
	TotalSize  int64
	EntryCount int
	MaxSize    int64
}

// HitRate returns the cache hit rate as a percentage (0.0 to 100.0).
func (s *CacheStats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0.0
	}
	return (float64(s.Hits) / float64(total)) * 100.0
}

// AnimationCache implements an LRU cache for animation frames.
// Thread-safe for concurrent access.
type AnimationCache struct {
	mu sync.RWMutex

	// Cache storage
	cache map[string]*list.Element // map[key.String()]*list.Element
	lru   *list.List               // LRU list of *cacheEntry

	// Configuration
	maxSize      int64 // Maximum cache size in bytes
	maxEntries   int   // Maximum number of entries (for ≥85% hit rate target)
	currentSize  int64 // Current cache size in bytes
	currentCount int   // Current number of entries

	// Statistics
	stats CacheStats
}

// NewAnimationCache creates a new animation cache.
// maxSizeBytes: maximum cache size in bytes (default: 50MB for animation frames)
// maxEntries: maximum number of cached frames (default: 1000)
func NewAnimationCache(maxSizeBytes int64, maxEntries int) *AnimationCache {
	if maxSizeBytes <= 0 {
		maxSizeBytes = 50 * 1024 * 1024 // 50MB default
	}
	if maxEntries <= 0 {
		maxEntries = 1000 // Default for ~85% hit rate
	}

	return &AnimationCache{
		cache:      make(map[string]*list.Element),
		lru:        list.New(),
		maxSize:    maxSizeBytes,
		maxEntries: maxEntries,
		stats: CacheStats{
			MaxSize: maxSizeBytes,
		},
	}
}

// Get retrieves an animation frame from the cache.
// Returns (frame, true) if found, (nil, false) if not found.
func (c *AnimationCache) Get(key CacheKey) (*ebiten.Image, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keyStr := key.String()
	elem, ok := c.cache[keyStr]
	if !ok {
		c.stats.Misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.lru.MoveToFront(elem)
	c.stats.Hits++

	entry := elem.Value.(*cacheEntry)
	return entry.frame, true
}

// Put stores an animation frame in the cache.
// Evicts least recently used entries if cache is full.
func (c *AnimationCache) Put(key CacheKey, frame *ebiten.Image) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keyStr := key.String()

	// Check if already cached
	if elem, exists := c.cache[keyStr]; exists {
		// Update existing entry and move to front
		c.lru.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		entry.frame = frame
		return
	}

	// Calculate frame size (width * height * 4 bytes per RGBA pixel)
	bounds := frame.Bounds()
	frameSize := int64(bounds.Dx() * bounds.Dy() * 4)

	// Evict entries if necessary to make room
	c.evictIfNeeded(frameSize)

	// Add new entry
	entry := &cacheEntry{
		key:   key,
		frame: frame,
		size:  frameSize,
	}
	elem := c.lru.PushFront(entry)
	c.cache[keyStr] = elem
	c.currentSize += frameSize
	c.currentCount++
	c.stats.TotalSize = c.currentSize
	c.stats.EntryCount = c.currentCount
}

// evictIfNeeded evicts least recently used entries until there's room for newSize bytes.
func (c *AnimationCache) evictIfNeeded(newSize int64) {
	// Evict by size constraint
	for c.currentSize+newSize > c.maxSize && c.lru.Len() > 0 {
		c.evictOldest()
	}

	// Evict by entry count constraint
	for c.currentCount >= c.maxEntries && c.lru.Len() > 0 {
		c.evictOldest()
	}
}

// evictOldest removes the least recently used entry from the cache.
func (c *AnimationCache) evictOldest() {
	elem := c.lru.Back()
	if elem == nil {
		return
	}

	c.lru.Remove(elem)
	entry := elem.Value.(*cacheEntry)
	delete(c.cache, entry.key.String())
	c.currentSize -= entry.size
	c.currentCount--
	c.stats.Evictions++
	c.stats.TotalSize = c.currentSize
	c.stats.EntryCount = c.currentCount
}

// Clear removes all entries from the cache.
func (c *AnimationCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*list.Element)
	c.lru = list.New()
	c.currentSize = 0
	c.currentCount = 0
	c.stats.TotalSize = 0
	c.stats.EntryCount = 0
}

// GetStats returns a copy of current cache statistics.
func (c *AnimationCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// Size returns the current cache size in bytes.
func (c *AnimationCache) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentSize
}

// Count returns the current number of cached entries.
func (c *AnimationCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentCount
}

// Prewarm pre-loads common animation sequences into the cache.
// This is useful for pre-computing player and common enemy animations.
func (c *AnimationCache) Prewarm(sequences []PrewarmSequence, generator func(CacheKey) (*ebiten.Image, error)) error {
	for _, seq := range sequences {
		for frameIdx := 0; frameIdx < seq.FrameCount; frameIdx++ {
			key := CacheKey{
				Seed:       seq.Seed,
				State:      seq.State,
				Direction:  seq.Direction,
				FrameIndex: frameIdx,
			}

			// Check if already cached
			if _, found := c.Get(key); found {
				continue
			}

			// Generate and cache
			frame, err := generator(key)
			if err != nil {
				return fmt.Errorf("failed to generate frame for prewarm: %w", err)
			}
			c.Put(key, frame)
		}
	}
	return nil
}

// PrewarmSequence defines an animation sequence to pre-load.
type PrewarmSequence struct {
	Seed       int64
	State      string
	Direction  Direction8
	FrameCount int
}

// CommonAnimationStates returns the most frequently used animation states
// for pre-warming the cache.
func CommonAnimationStates() []string {
	return []string{
		"idle",
		"walk",
		"run",
		"attack",
	}
}

// CommonDirections returns the primary directions for pre-warming.
// For performance, we prewarm only primary directions (not diagonals)
// since diagonals can be interpolated or generated on-demand.
func CommonDirections() []Direction8 {
	return []Direction8{
		Dir8North,
		Dir8East,
		Dir8South,
		Dir8West,
	}
}

// EstimateFrameSize estimates the memory size of a frame in bytes.
func EstimateFrameSize(width, height int) int64 {
	return int64(width * height * 4) // RGBA = 4 bytes per pixel
}

// TrimToSize reduces cache size to target bytes by evicting oldest entries.
// Useful for dynamic memory management.
func (c *AnimationCache) TrimToSize(targetSize int64) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	evicted := 0
	for c.currentSize > targetSize && c.lru.Len() > 0 {
		c.evictOldest()
		evicted++
	}
	return evicted
}

// TrimToCount reduces cache to target number of entries.
func (c *AnimationCache) TrimToCount(targetCount int) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	evicted := 0
	for c.currentCount > targetCount && c.lru.Len() > 0 {
		c.evictOldest()
		evicted++
	}
	return evicted
}
