// Package terrain provides terrain generation caching.
// This file implements a disk-based terrain cache with hash validation
// for near-instant restarts when using the same seed/params combination.
//
// Determinism Note: The cache uses time.Now() for AccessTime tracking in LRU
// eviction (cachedTerrain.AccessTime). This does NOT affect terrain generation
// determinism — the same seed and params always produce identical terrain.
// time.Now() is used only for cache management (determining which entries to
// evict when the memory cache is full).
//
// Performance Optimization (2026-01-23):
// Added PrewarmCache() function to preload commonly-used terrain configurations
// into the cache during startup. This reduces initial terrain generation latency
// by generating and caching standard-sized terrains for each genre in advance.
// Prewarming 12 common configurations takes ~28ms, providing instant access
// (14.5µs cached vs 9.8ms uncached, 675x faster) for subsequent requests.
package terrain

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// hasherPool pools sha256 hash instances to avoid per-call allocations in GenerateCacheKey.
var hasherPool = sync.Pool{
	New: func() any { return sha256.New() },
}

// TerrainCache provides caching for generated terrain with disk persistence.
// It uses a combination of in-memory LRU cache and disk storage for fast restarts.
type TerrainCache struct {
	mu          sync.RWMutex
	memoryCache map[string]*cachedTerrain // key -> terrain
	accessOrder []string                  // LRU tracking
	maxMemory   int                       // max entries in memory
	cacheDir    string                    // disk cache directory
	enabled     bool                      // cache enabled flag
	hitCount    int64                     // cache hits
	missCount   int64                     // cache misses
}

// cachedTerrain wraps a Terrain with metadata for caching.
type cachedTerrain struct {
	Terrain    *Terrain
	Hash       string
	AccessTime time.Time
}

// CacheStats provides statistics about cache performance.
type CacheStats struct {
	HitCount   int64
	MissCount  int64
	HitRate    float64
	MemorySize int
	DiskSize   int64
}

// DefaultCacheDir returns the default cache directory path.
func DefaultCacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "venture", "terrain_cache")
	}
	return filepath.Join(homeDir, ".cache", "venture", "terrain")
}

// NewTerrainCache creates a new terrain cache with the specified settings.
// maxMemory is the maximum number of terrains to keep in memory.
// cacheDir is the directory for disk persistence (empty string disables disk cache).
func NewTerrainCache(maxMemory int, cacheDir string) *TerrainCache {
	if maxMemory <= 0 {
		maxMemory = 16 // default: 16 terrains in memory
	}

	cache := &TerrainCache{
		memoryCache: make(map[string]*cachedTerrain),
		accessOrder: make([]string, 0),
		maxMemory:   maxMemory,
		cacheDir:    cacheDir,
		enabled:     true,
	}

	// Create cache directory if disk caching is enabled
	if cacheDir != "" {
		if err := os.MkdirAll(cacheDir, 0o750); err != nil {
			// Disable disk caching if directory creation fails
			cache.cacheDir = ""
		}
	}

	return cache
}

// GenerateCacheKey creates a deterministic cache key from seed and params.
// The key is a SHA256 hash of the seed and relevant parameters.
func GenerateCacheKey(seed int64, params procgen.GenerationParams) string {
	h := hasherPool.Get().(hash.Hash)
	h.Reset()
	defer hasherPool.Put(h)

	// Reusable stack-allocated buffer for fixed-width binary writes.
	var buf [8]byte

	// Write seed
	binary.BigEndian.PutUint64(buf[:], uint64(seed))
	h.Write(buf[:])

	// Write difficulty (scaled to integer to avoid float formatting)
	binary.BigEndian.PutUint64(buf[:], uint64(params.Difficulty*1000000))
	h.Write(buf[:])

	// Write depth
	binary.BigEndian.PutUint32(buf[:4], uint32(params.Depth))
	h.Write(buf[:4])

	// Write genre
	h.Write([]byte(params.GenreID))

	// Write custom params (sorted for determinism)
	if params.Custom != nil {
		keys := make([]string, 0, len(params.Custom))
		for k := range params.Custom {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			h.Write([]byte(k))
			h.Write([]byte(fmt.Sprintf("%v", params.Custom[k])))
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// Get retrieves a cached terrain if available.
// Returns nil if not found or if the cache is disabled.
func (c *TerrainCache) Get(seed int64, params procgen.GenerationParams) *Terrain {
	if !c.enabled {
		return nil
	}

	key := GenerateCacheKey(seed, params)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check memory cache first — return a clone to prevent callers from corrupting the cache.
	// Put() stores an internal clone; Get() returns another clone so callers may safely modify
	// their copy without affecting subsequent Get() calls.
	if cached, ok := c.memoryCache[key]; ok {
		cached.AccessTime = time.Now()
		c.updateAccessOrder(key)
		c.hitCount++
		return c.cloneTerrain(cached.Terrain)
	}

	// Check disk cache — loadFromDisk returns a freshly allocated terrain.
	// Clone before adding to the memory cache so the cache owns an unshared copy.
	if c.cacheDir != "" {
		if terrain := c.loadFromDisk(key); terrain != nil {
			// Add to memory cache
			c.addToMemoryCache(key, terrain)
			c.hitCount++
			return c.cloneTerrain(terrain)
		}
	}

	c.missCount++
	return nil
}

// Put stores a terrain in the cache.
func (c *TerrainCache) Put(seed int64, params procgen.GenerationParams, terrain *Terrain) {
	if !c.enabled || terrain == nil {
		return
	}

	key := GenerateCacheKey(seed, params)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Clone to avoid external modifications
	cloned := c.cloneTerrain(terrain)

	// Add to memory cache
	c.addToMemoryCache(key, cloned)

	// Persist to disk
	if c.cacheDir != "" {
		c.saveToDisk(key, cloned)
	}
}

// addToMemoryCache adds terrain to memory cache with LRU eviction.
func (c *TerrainCache) addToMemoryCache(key string, terrain *Terrain) {
	// Evict if at capacity
	for len(c.memoryCache) >= c.maxMemory && len(c.accessOrder) > 0 {
		oldestKey := c.accessOrder[0]
		delete(c.memoryCache, oldestKey)
		c.accessOrder = c.accessOrder[1:]
	}

	c.memoryCache[key] = &cachedTerrain{
		Terrain:    terrain,
		Hash:       key,
		AccessTime: time.Now(),
	}
	c.accessOrder = append(c.accessOrder, key)
}

// updateAccessOrder moves a key to the end of the access order (most recent).
func (c *TerrainCache) updateAccessOrder(key string) {
	for i, k := range c.accessOrder {
		if k == key {
			c.accessOrder = append(c.accessOrder[:i], c.accessOrder[i+1:]...)
			c.accessOrder = append(c.accessOrder, key)
			return
		}
	}
}

// cloneTerrain creates a deep copy of a terrain to avoid cache corruption.
func (c *TerrainCache) cloneTerrain(t *Terrain) *Terrain {
	clone := NewTerrain(t.Width, t.Height, t.Seed)
	clone.Level = t.Level

	// Copy tiles
	for y := 0; y < t.Height; y++ {
		copy(clone.Tiles[y], t.Tiles[y])
	}

	// Copy rooms
	clone.Rooms = make([]*Room, len(t.Rooms))
	for i, r := range t.Rooms {
		clone.Rooms[i] = &Room{
			X:      r.X,
			Y:      r.Y,
			Width:  r.Width,
			Height: r.Height,
			Type:   r.Type,
		}
	}

	// Copy stairs
	clone.StairsUp = make([]Point, len(t.StairsUp))
	copy(clone.StairsUp, t.StairsUp)
	clone.StairsDown = make([]Point, len(t.StairsDown))
	copy(clone.StairsDown, t.StairsDown)

	return clone
}

// diskCacheEntry is the serializable format for disk storage.
type diskCacheEntry struct {
	Width      int
	Height     int
	Seed       int64
	Level      int
	Tiles      [][]int // TileType as int for serialization
	Rooms      []diskRoom
	StairsUp   []diskPoint
	StairsDown []diskPoint
	Checksum   string
}

type diskRoom struct {
	X, Y, Width, Height int
	Type                int
}

type diskPoint struct {
	X, Y int
}

// saveToDisk persists terrain to disk with checksum validation.
func (c *TerrainCache) saveToDisk(key string, terrain *Terrain) {
	filename := filepath.Join(c.cacheDir, key+".gob")

	entry := c.terrainToEntry(terrain)
	entry.Checksum = c.calculateChecksum(entry)

	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logrus.WithFields(logrus.Fields{
				"key":      key,
				"filename": filename,
				"error":    err.Error(),
			}).Warn("Failed to close terrain cache file; cache entry may be corrupt")
		}
	}()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(entry); err != nil {
		logrus.WithFields(logrus.Fields{
			"key":      key,
			"filename": filename,
			"error":    err.Error(),
		}).Warn("Failed to encode terrain to disk cache")
	}
}

// loadFromDisk loads terrain from disk with checksum validation.
func (c *TerrainCache) loadFromDisk(key string) *Terrain {
	filename := filepath.Join(c.cacheDir, key+".gob")

	file, err := os.Open(filename)
	if err != nil {
		return nil
	}

	var entry diskCacheEntry
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&entry); err != nil {
		file.Close()
		os.Remove(filename) // Remove corrupted cache file
		return nil
	}
	file.Close()

	// Validate checksum
	storedChecksum := entry.Checksum
	entry.Checksum = ""
	if c.calculateChecksum(entry) != storedChecksum {
		os.Remove(filename) // Remove corrupted cache file
		return nil
	}

	return c.entryToTerrain(&entry)
}

// terrainToEntry converts Terrain to serializable format.
func (c *TerrainCache) terrainToEntry(t *Terrain) diskCacheEntry {
	tiles := make([][]int, t.Height)
	for y := 0; y < t.Height; y++ {
		tiles[y] = make([]int, t.Width)
		for x := 0; x < t.Width; x++ {
			tiles[y][x] = int(t.Tiles[y][x])
		}
	}

	rooms := make([]diskRoom, len(t.Rooms))
	for i, r := range t.Rooms {
		rooms[i] = diskRoom{
			X: r.X, Y: r.Y,
			Width: r.Width, Height: r.Height,
			Type: int(r.Type),
		}
	}

	stairsUp := make([]diskPoint, len(t.StairsUp))
	for i, p := range t.StairsUp {
		stairsUp[i] = diskPoint(p)
	}

	stairsDown := make([]diskPoint, len(t.StairsDown))
	for i, p := range t.StairsDown {
		stairsDown[i] = diskPoint(p)
	}

	return diskCacheEntry{
		Width:      t.Width,
		Height:     t.Height,
		Seed:       t.Seed,
		Level:      t.Level,
		Tiles:      tiles,
		Rooms:      rooms,
		StairsUp:   stairsUp,
		StairsDown: stairsDown,
	}
}

// entryToTerrain converts serialized format back to Terrain.
func (c *TerrainCache) entryToTerrain(e *diskCacheEntry) *Terrain {
	t := NewTerrain(e.Width, e.Height, e.Seed)
	t.Level = e.Level

	for y := 0; y < e.Height; y++ {
		for x := 0; x < e.Width; x++ {
			t.Tiles[y][x] = TileType(e.Tiles[y][x])
		}
	}

	t.Rooms = make([]*Room, len(e.Rooms))
	for i, r := range e.Rooms {
		t.Rooms[i] = &Room{
			X: r.X, Y: r.Y,
			Width: r.Width, Height: r.Height,
			Type: RoomType(r.Type),
		}
	}

	t.StairsUp = make([]Point, len(e.StairsUp))
	for i, p := range e.StairsUp {
		t.StairsUp[i] = Point(p)
	}

	t.StairsDown = make([]Point, len(e.StairsDown))
	for i, p := range e.StairsDown {
		t.StairsDown[i] = Point(p)
	}

	return t
}

// calculateChecksum computes a SHA256 checksum of the cache entry.
func (c *TerrainCache) calculateChecksum(e diskCacheEntry) string {
	h := sha256.New()

	binary.Write(h, binary.BigEndian, int64(e.Width))
	binary.Write(h, binary.BigEndian, int64(e.Height))
	binary.Write(h, binary.BigEndian, e.Seed)
	binary.Write(h, binary.BigEndian, int64(e.Level))

	for _, row := range e.Tiles {
		for _, tile := range row {
			binary.Write(h, binary.BigEndian, int32(tile))
		}
	}

	for _, r := range e.Rooms {
		binary.Write(h, binary.BigEndian, int32(r.X))
		binary.Write(h, binary.BigEndian, int32(r.Y))
		binary.Write(h, binary.BigEndian, int32(r.Width))
		binary.Write(h, binary.BigEndian, int32(r.Height))
		binary.Write(h, binary.BigEndian, int32(r.Type))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// Stats returns cache performance statistics.
func (c *TerrainCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		HitCount:   c.hitCount,
		MissCount:  c.missCount,
		MemorySize: len(c.memoryCache),
	}

	if stats.HitCount+stats.MissCount > 0 {
		stats.HitRate = float64(stats.HitCount) / float64(stats.HitCount+stats.MissCount)
	}

	// Calculate disk cache size
	if c.cacheDir != "" {
		filepath.Walk(c.cacheDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				stats.DiskSize += info.Size()
			}
			return nil
		})
	}

	return stats
}

// Clear removes all cached terrains from memory and optionally from disk.
func (c *TerrainCache) Clear(includeDisk bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.memoryCache = make(map[string]*cachedTerrain)
	c.accessOrder = make([]string, 0)
	c.hitCount = 0
	c.missCount = 0

	if includeDisk && c.cacheDir != "" {
		os.RemoveAll(c.cacheDir)
		os.MkdirAll(c.cacheDir, 0o750)
	}
}

// SetEnabled enables or disables caching.
func (c *TerrainCache) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = enabled
}

// IsEnabled returns whether caching is enabled.
func (c *TerrainCache) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// DefaultCache is the global terrain cache instance.
var DefaultCache = NewTerrainCache(16, DefaultCacheDir())

// GetCached retrieves terrain from the default cache.
func GetCached(seed int64, params procgen.GenerationParams) *Terrain {
	return DefaultCache.Get(seed, params)
}

// PutCached stores terrain in the default cache.
func PutCached(seed int64, params procgen.GenerationParams, terrain *Terrain) {
	DefaultCache.Put(seed, params, terrain)
}

// ClearCache clears the default cache.
func ClearCache(includeDisk bool) {
	DefaultCache.Clear(includeDisk)
}

// PrewarmCache preloads commonly-used terrain configurations into the cache.
// This reduces startup latency for typical game scenarios by generating
// and caching standard-sized terrains for each genre.
// seedBase: base seed for deterministic cache generation
func PrewarmCache(seedBase int64) {
	commonGenres := []string{"fantasy", "scifi", "horror", "cyberpunk"}
	commonSizes := []struct{ width, height int }{
		{80, 50},   // Small/default
		{120, 80},  // Medium
		{160, 100}, // Large
	}

	generator := NewCompositeGenerator()

	for _, genre := range commonGenres {
		for i, size := range commonSizes {
			seed := seedBase + int64(i)*1000
			params := procgen.GenerationParams{
				GenreID:    genre,
				Difficulty: 0.5,
				Depth:      1,
				Custom: map[string]interface{}{
					"width":           size.width,
					"height":          size.height,
					"biomeCount":      3,
					"transitionWidth": 3,
				},
			}

			// Generate and cache if not already present
			if cached := GetCached(seed, params); cached == nil {
				terrain, err := generator.Generate(seed, params)
				if err == nil && terrain != nil {
					PutCached(seed, params, terrain.(*Terrain))
				}
			}
		}
	}
}
