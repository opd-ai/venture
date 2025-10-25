package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/tiles"
)

func TestTileCache_Get(t *testing.T) {
	cache := NewTileCache(10)

	key := TileCacheKey{
		TileType: tiles.TileFloor,
		GenreID:  "fantasy",
		Seed:     12345,
		Variant:  0.5,
		Width:    32,
		Height:   32,
	}

	// First access should be a miss
	img1, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if img1 == nil {
		t.Fatal("Expected non-nil image")
	}

	hits, misses := cache.Stats()
	if hits != 0 {
		t.Errorf("Expected 0 hits, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("Expected 1 miss, got %d", misses)
	}

	// Second access should be a hit
	img2, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if img2 != img1 {
		t.Error("Expected same image instance from cache")
	}

	hits, misses = cache.Stats()
	if hits != 1 {
		t.Errorf("Expected 1 hit, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("Expected 1 miss, got %d", misses)
	}
}

func TestTileCache_Eviction(t *testing.T) {
	cache := NewTileCache(3) // Small cache for testing eviction

	keys := []TileCacheKey{
		{TileType: tiles.TileFloor, GenreID: "fantasy", Seed: 1, Variant: 0.5, Width: 32, Height: 32},
		{TileType: tiles.TileWall, GenreID: "fantasy", Seed: 2, Variant: 0.5, Width: 32, Height: 32},
		{TileType: tiles.TileDoor, GenreID: "fantasy", Seed: 3, Variant: 0.5, Width: 32, Height: 32},
		{TileType: tiles.TileCorridor, GenreID: "fantasy", Seed: 4, Variant: 0.5, Width: 32, Height: 32},
	}

	// Fill cache to capacity
	for i := 0; i < 3; i++ {
		_, err := cache.Get(keys[i])
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
	}

	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Size())
	}

	// Add one more - should evict oldest
	_, err := cache.Get(keys[3])
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if cache.Size() != 3 {
		t.Errorf("Expected cache size to remain 3, got %d", cache.Size())
	}

	// First key should have been evicted
	cache.Get(keys[0])
	_, misses := cache.Stats()
	// 3 misses from initial fill + 1 from keys[3] + 1 from keys[0] being evicted = 5
	if misses != 5 {
		t.Errorf("Expected 5 total misses, got %d", misses)
	}
}

func TestTileCache_Clear(t *testing.T) {
	cache := NewTileCache(10)

	key := TileCacheKey{
		TileType: tiles.TileFloor,
		GenreID:  "fantasy",
		Seed:     12345,
		Variant:  0.5,
		Width:    32,
		Height:   32,
	}

	// Add entry
	_, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Size())
	}

	// Clear cache
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}

	hits, misses := cache.Stats()
	if hits != 0 || misses != 0 {
		t.Errorf("Expected stats to reset, got hits=%d, misses=%d", hits, misses)
	}
}

func TestTileCache_HitRate(t *testing.T) {
	cache := NewTileCache(10)

	key := TileCacheKey{
		TileType: tiles.TileFloor,
		GenreID:  "fantasy",
		Seed:     12345,
		Variant:  0.5,
		Width:    32,
		Height:   32,
	}

	// First access - miss
	cache.Get(key)
	if hitRate := cache.HitRate(); hitRate != 0.0 {
		t.Errorf("Expected hit rate 0%%, got %.1f%%", hitRate)
	}

	// Three more hits
	cache.Get(key)
	cache.Get(key)
	cache.Get(key)

	hitRate := cache.HitRate()
	expected := 75.0 // 3 hits out of 4 total accesses
	if hitRate != expected {
		t.Errorf("Expected hit rate %.1f%%, got %.1f%%", expected, hitRate)
	}
}

func TestTileCache_DifferentKeys(t *testing.T) {
	cache := NewTileCache(10)

	keys := []TileCacheKey{
		{TileType: tiles.TileFloor, GenreID: "fantasy", Seed: 1, Variant: 0.5, Width: 32, Height: 32},
		{TileType: tiles.TileFloor, GenreID: "scifi", Seed: 1, Variant: 0.5, Width: 32, Height: 32},
		{TileType: tiles.TileFloor, GenreID: "fantasy", Seed: 2, Variant: 0.5, Width: 32, Height: 32},
		{TileType: tiles.TileWall, GenreID: "fantasy", Seed: 1, Variant: 0.5, Width: 32, Height: 32},
	}

	// Each key should generate different tiles
	images := make(map[string]interface{})
	for i, key := range keys {
		img, err := cache.Get(key)
		if err != nil {
			t.Fatalf("Get failed for key %d: %v", i, err)
		}
		keyStr := key.String()
		if _, exists := images[keyStr]; exists {
			t.Errorf("Duplicate image for key %s", keyStr)
		}
		images[keyStr] = img
	}

	if cache.Size() != 4 {
		t.Errorf("Expected cache size 4, got %d", cache.Size())
	}
}

func TestTileCacheKey_String(t *testing.T) {
	key := TileCacheKey{
		TileType: tiles.TileFloor,
		GenreID:  "fantasy",
		Seed:     12345,
		Variant:  0.5,
		Width:    32,
		Height:   32,
	}

	str := key.String()
	expected := "floor-fantasy-12345-0.50-32x32"
	if str != expected {
		t.Errorf("Expected key string %q, got %q", expected, str)
	}
}

func BenchmarkTileCache_Get(b *testing.B) {
	cache := NewTileCache(100)

	key := TileCacheKey{
		TileType: tiles.TileFloor,
		GenreID:  "fantasy",
		Seed:     12345,
		Variant:  0.5,
		Width:    32,
		Height:   32,
	}

	// Pre-warm cache
	cache.Get(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(key)
	}
}

func BenchmarkTileCache_GetMixed(b *testing.B) {
	cache := NewTileCache(100)

	keys := []TileCacheKey{
		{TileType: tiles.TileFloor, GenreID: "fantasy", Seed: 1, Variant: 0.5, Width: 32, Height: 32},
		{TileType: tiles.TileWall, GenreID: "fantasy", Seed: 2, Variant: 0.5, Width: 32, Height: 32},
		{TileType: tiles.TileDoor, GenreID: "scifi", Seed: 3, Variant: 0.5, Width: 32, Height: 32},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%len(keys)]
		_, _ = cache.Get(key)
	}
}
