package terrain

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name        string
		seed1       int64
		params1     procgen.GenerationParams
		seed2       int64
		params2     procgen.GenerationParams
		shouldMatch bool
	}{
		{
			name:  "same seed and params match",
			seed1: 12345,
			params1: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			seed2: 12345,
			params2: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			shouldMatch: true,
		},
		{
			name:  "different seeds don't match",
			seed1: 12345,
			params1: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			seed2: 54321,
			params2: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			shouldMatch: false,
		},
		{
			name:  "different difficulty don't match",
			seed1: 12345,
			params1: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			seed2: 12345,
			params2: procgen.GenerationParams{
				Difficulty: 0.8,
				Depth:      1,
				GenreID:    "fantasy",
			},
			shouldMatch: false,
		},
		{
			name:  "different genre don't match",
			seed1: 12345,
			params1: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
			},
			seed2: 12345,
			params2: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "scifi",
			},
			shouldMatch: false,
		},
		{
			name:  "same custom params match",
			seed1: 12345,
			params1: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{"width": 80, "height": 50},
			},
			seed2: 12345,
			params2: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{"width": 80, "height": 50},
			},
			shouldMatch: true,
		},
		{
			name:  "different custom params don't match",
			seed1: 12345,
			params1: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{"width": 80},
			},
			seed2: 12345,
			params2: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      1,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{"width": 100},
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := GenerateCacheKey(tt.seed1, tt.params1)
			key2 := GenerateCacheKey(tt.seed2, tt.params2)

			if (key1 == key2) != tt.shouldMatch {
				t.Errorf("keys match=%v, want match=%v (key1=%s, key2=%s)",
					key1 == key2, tt.shouldMatch, key1, key2)
			}
		})
	}
}

func TestTerrainCache_MemoryOnly(t *testing.T) {
	cache := NewTerrainCache(4, "") // No disk cache

	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
	}

	// Cache miss
	result := cache.Get(seed, params)
	if result != nil {
		t.Error("expected cache miss, got hit")
	}

	// Create terrain
	terrain := NewTerrain(20, 15, seed)
	terrain.SetTile(5, 5, TileFloor)
	terrain.Rooms = []*Room{{X: 2, Y: 2, Width: 10, Height: 8, Type: RoomNormal}}

	// Store in cache
	cache.Put(seed, params, terrain)

	// Cache hit
	result = cache.Get(seed, params)
	if result == nil {
		t.Fatal("expected cache hit, got miss")
	}

	// Verify data
	if result.Width != 20 || result.Height != 15 {
		t.Errorf("wrong dimensions: %dx%d", result.Width, result.Height)
	}
	if result.GetTile(5, 5) != TileFloor {
		t.Error("tile not preserved")
	}
	if len(result.Rooms) != 1 {
		t.Errorf("expected 1 room, got %d", len(result.Rooms))
	}

	// Verify it's a clone (not same pointer)
	if result == terrain {
		t.Error("cache returned same pointer instead of clone")
	}
}

func TestTerrainCache_DiskPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewTerrainCache(4, tmpDir)

	seed := int64(99999)
	params := procgen.GenerationParams{
		Difficulty: 0.3,
		Depth:      2,
		GenreID:    "horror",
	}

	// Create terrain with stairs
	terrain := NewTerrain(30, 25, seed)
	terrain.SetTile(10, 10, TileFloor)
	terrain.AddStairs(15, 15, true)
	terrain.AddStairs(5, 5, false)
	terrain.Level = 2

	// Store in cache
	cache.Put(seed, params, terrain)

	// Verify disk file exists
	key := GenerateCacheKey(seed, params)
	filename := filepath.Join(tmpDir, key+".gob")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("disk cache file not created")
	}

	// Clear memory cache and reload from disk
	cache.Clear(false)
	result := cache.Get(seed, params)
	if result == nil {
		t.Fatal("expected disk cache hit")
	}

	// Verify data
	if result.Level != 2 {
		t.Errorf("level not preserved: got %d", result.Level)
	}
	if len(result.StairsUp) != 1 {
		t.Errorf("stairs up not preserved: got %d", len(result.StairsUp))
	}
	if len(result.StairsDown) != 1 {
		t.Errorf("stairs down not preserved: got %d", len(result.StairsDown))
	}
}

func TestTerrainCache_LRUEviction(t *testing.T) {
	cache := NewTerrainCache(3, "") // Max 3 entries

	params := procgen.GenerationParams{GenreID: "fantasy"}

	// Add 3 terrains
	for i := int64(1); i <= 3; i++ {
		terrain := NewTerrain(10, 10, i)
		cache.Put(i, params, terrain)
	}

	// Verify all 3 are cached
	for i := int64(1); i <= 3; i++ {
		if cache.Get(i, params) == nil {
			t.Errorf("terrain %d should be cached", i)
		}
	}

	// Add 4th terrain (should evict oldest)
	terrain := NewTerrain(10, 10, 4)
	cache.Put(4, params, terrain)

	// First terrain should be evicted
	if cache.Get(1, params) != nil {
		t.Error("terrain 1 should have been evicted")
	}

	// 2, 3, 4 should still be cached
	for i := int64(2); i <= 4; i++ {
		if cache.Get(i, params) == nil {
			t.Errorf("terrain %d should be cached", i)
		}
	}
}

func TestTerrainCache_Stats(t *testing.T) {
	cache := NewTerrainCache(4, "")

	params := procgen.GenerationParams{GenreID: "fantasy"}

	// Initial stats
	stats := cache.Stats()
	if stats.HitCount != 0 || stats.MissCount != 0 {
		t.Error("initial stats should be zero")
	}

	// Miss
	cache.Get(1, params)
	stats = cache.Stats()
	if stats.MissCount != 1 {
		t.Errorf("expected 1 miss, got %d", stats.MissCount)
	}

	// Add and hit
	cache.Put(1, params, NewTerrain(10, 10, 1))
	cache.Get(1, params)
	stats = cache.Stats()
	if stats.HitCount != 1 {
		t.Errorf("expected 1 hit, got %d", stats.HitCount)
	}

	// Verify hit rate
	expectedRate := 1.0 / 2.0 // 1 hit, 1 miss
	if stats.HitRate != expectedRate {
		t.Errorf("expected hit rate %f, got %f", expectedRate, stats.HitRate)
	}
}

func TestTerrainCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewTerrainCache(4, tmpDir)

	params := procgen.GenerationParams{GenreID: "fantasy"}

	// Add terrain
	cache.Put(1, params, NewTerrain(10, 10, 1))

	// Clear memory only
	cache.Clear(false)
	stats := cache.Stats()
	if stats.MemorySize != 0 {
		t.Error("memory cache should be empty")
	}

	// Disk cache should still work (reload on get)
	result := cache.Get(1, params)
	if result == nil {
		t.Error("disk cache should still have terrain")
	}

	// Clear including disk
	cache.Clear(true)

	// Create new cache to verify disk is cleared
	cache2 := NewTerrainCache(4, tmpDir)
	result = cache2.Get(1, params)
	if result != nil {
		t.Error("disk cache should be cleared")
	}
}

func TestTerrainCache_EnableDisable(t *testing.T) {
	cache := NewTerrainCache(4, "")

	params := procgen.GenerationParams{GenreID: "fantasy"}

	// Add terrain while enabled
	cache.Put(1, params, NewTerrain(10, 10, 1))

	// Disable cache
	cache.SetEnabled(false)
	if cache.IsEnabled() {
		t.Error("cache should be disabled")
	}

	// Get should return nil when disabled
	if cache.Get(1, params) != nil {
		t.Error("disabled cache should return nil")
	}

	// Put should do nothing when disabled
	cache.Put(2, params, NewTerrain(10, 10, 2))

	// Re-enable
	cache.SetEnabled(true)

	// Original terrain should still be there
	if cache.Get(1, params) == nil {
		t.Error("original terrain should still be cached")
	}

	// New terrain should not be there (wasn't stored when disabled)
	if cache.Get(2, params) != nil {
		t.Error("terrain 2 should not be cached")
	}
}

func TestTerrainCache_Concurrency(t *testing.T) {
	cache := NewTerrainCache(16, "")

	params := procgen.GenerationParams{GenreID: "fantasy"}

	var wg sync.WaitGroup
	const goroutines = 10
	const operations = 100

	// Concurrent reads and writes
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				seed := int64(id*operations + j)

				// Write
				cache.Put(seed, params, NewTerrain(10, 10, seed))

				// Read
				cache.Get(seed, params)
			}
		}(i)
	}

	wg.Wait()

	// Verify no panics occurred and stats are reasonable
	stats := cache.Stats()
	if stats.HitCount+stats.MissCount == 0 {
		t.Error("expected some cache operations")
	}
}

func TestTerrainCache_CorruptedDiskCache(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewTerrainCache(4, tmpDir)

	params := procgen.GenerationParams{GenreID: "fantasy"}
	seed := int64(12345)

	// Add terrain
	cache.Put(seed, params, NewTerrain(10, 10, seed))

	// Corrupt the file
	key := GenerateCacheKey(seed, params)
	filename := filepath.Join(tmpDir, key+".gob")
	os.WriteFile(filename, []byte("corrupted data"), 0o644)

	// Clear memory cache
	cache.Clear(false)

	// Get should return nil (corrupted file detected via checksum)
	result := cache.Get(seed, params)
	if result != nil {
		t.Error("corrupted cache should return nil")
	}

	// Corrupted file should be removed
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Error("corrupted file should be removed")
	}
}

func TestTerrainCache_ClonePreservesData(t *testing.T) {
	cache := NewTerrainCache(4, "")

	params := procgen.GenerationParams{GenreID: "fantasy"}
	seed := int64(12345)

	// Create complex terrain
	terrain := NewTerrain(20, 20, seed)
	terrain.Level = 5
	terrain.SetTile(5, 5, TileFloor)
	terrain.SetTile(10, 10, TileDoor)
	terrain.Rooms = []*Room{
		{X: 1, Y: 1, Width: 8, Height: 8, Type: RoomTreasure},
		{X: 10, Y: 10, Width: 5, Height: 5, Type: RoomBoss},
	}
	terrain.AddStairs(3, 3, true)
	terrain.AddStairs(15, 15, false)

	// Store and retrieve
	cache.Put(seed, params, terrain)
	result := cache.Get(seed, params)

	// Modify original
	terrain.SetTile(5, 5, TileWall)
	terrain.Level = 99

	// Clone should be unaffected
	if result.GetTile(5, 5) != TileFloor {
		t.Error("clone was affected by original modification")
	}
	if result.Level != 5 {
		t.Error("clone level was affected")
	}

	// Modify clone
	result.SetTile(10, 10, TileWall)

	// Get again and verify new copy is unaffected
	result2 := cache.Get(seed, params)
	if result2.GetTile(10, 10) != TileDoor {
		t.Error("cache was corrupted by clone modification")
	}
}

func TestDefaultCacheDir(t *testing.T) {
	dir := DefaultCacheDir()
	if dir == "" {
		t.Error("default cache dir should not be empty")
	}
}

func TestGlobalCacheFunctions(t *testing.T) {
	// Reset default cache for testing
	DefaultCache = NewTerrainCache(4, "")

	params := procgen.GenerationParams{GenreID: "fantasy"}
	seed := int64(777)

	// Miss
	if GetCached(seed, params) != nil {
		t.Error("expected cache miss")
	}

	// Put and get
	PutCached(seed, params, NewTerrain(10, 10, seed))
	if GetCached(seed, params) == nil {
		t.Error("expected cache hit")
	}

	// Clear
	ClearCache(false)
	if GetCached(seed, params) != nil {
		t.Error("cache should be cleared")
	}
}

func BenchmarkTerrainCache_Get(b *testing.B) {
	cache := NewTerrainCache(100, "")
	params := procgen.GenerationParams{GenreID: "fantasy"}

	// Pre-populate cache
	for i := int64(0); i < 50; i++ {
		cache.Put(i, params, NewTerrain(80, 50, i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		seed := int64(i % 50)
		cache.Get(seed, params)
	}
}

func BenchmarkTerrainCache_Put(b *testing.B) {
	cache := NewTerrainCache(1000, "")
	params := procgen.GenerationParams{GenreID: "fantasy"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		seed := int64(i)
		cache.Put(seed, params, NewTerrain(80, 50, seed))
	}
}

func BenchmarkGenerateCacheKey(b *testing.B) {
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"width": 80, "height": 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateCacheKey(int64(i), params)
	}
}
