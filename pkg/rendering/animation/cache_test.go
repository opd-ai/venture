package animation

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewAnimationCache(t *testing.T) {
	cache := NewAnimationCache(10*1024*1024, 100)

	if cache == nil {
		t.Fatal("NewAnimationCache returned nil")
	}
	if cache.maxSize != 10*1024*1024 {
		t.Errorf("maxSize = %v, want %v", cache.maxSize, 10*1024*1024)
	}
	if cache.maxEntries != 100 {
		t.Errorf("maxEntries = %v, want 100", cache.maxEntries)
	}
}

func TestNewAnimationCacheDefaults(t *testing.T) {
	cache := NewAnimationCache(0, 0)

	if cache.maxSize != 50*1024*1024 {
		t.Errorf("default maxSize = %v, want %v", cache.maxSize, 50*1024*1024)
	}
	if cache.maxEntries != 1000 {
		t.Errorf("default maxEntries = %v, want 1000", cache.maxEntries)
	}
}

func TestCacheKeyString(t *testing.T) {
	key := CacheKey{
		Seed:       12345,
		State:      "walk",
		Direction:  Dir8North,
		FrameIndex: 3,
	}

	str := key.String()
	expected := "12345:walk:north:3"
	if str != expected {
		t.Errorf("CacheKey.String() = %v, want %v", str, expected)
	}
}

func TestCachePutGet(t *testing.T) {
	cache := NewAnimationCache(10*1024*1024, 100)

	// Create test image
	img := ebiten.NewImage(32, 32)

	key := CacheKey{
		Seed:       12345,
		State:      "idle",
		Direction:  Dir8South,
		FrameIndex: 0,
	}

	// Put in cache
	cache.Put(key, img)

	// Get from cache
	retrieved, found := cache.Get(key)
	if !found {
		t.Fatal("Cache miss on item just inserted")
	}
	if retrieved != img {
		t.Error("Retrieved image does not match inserted image")
	}

	// Verify stats
	stats := cache.GetStats()
	if stats.Hits != 1 {
		t.Errorf("Hits = %v, want 1", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("Misses = %v, want 0", stats.Misses)
	}
	if stats.EntryCount != 1 {
		t.Errorf("EntryCount = %v, want 1", stats.EntryCount)
	}
}

func TestCacheMiss(t *testing.T) {
	cache := NewAnimationCache(10*1024*1024, 100)

	key := CacheKey{
		Seed:       12345,
		State:      "walk",
		Direction:  Dir8East,
		FrameIndex: 0,
	}

	// Get non-existent key
	_, found := cache.Get(key)
	if found {
		t.Error("Cache hit on non-existent key")
	}

	// Verify stats
	stats := cache.GetStats()
	if stats.Misses != 1 {
		t.Errorf("Misses = %v, want 1", stats.Misses)
	}
	if stats.Hits != 0 {
		t.Errorf("Hits = %v, want 0", stats.Hits)
	}
}

func TestCacheLRUEviction(t *testing.T) {
	// Small cache that can only hold 2 entries
	cache := NewAnimationCache(10*1024, 2)

	img1 := ebiten.NewImage(32, 32)
	img2 := ebiten.NewImage(32, 32)
	img3 := ebiten.NewImage(32, 32)

	key1 := CacheKey{Seed: 1, State: "idle", Direction: Dir8South, FrameIndex: 0}
	key2 := CacheKey{Seed: 2, State: "idle", Direction: Dir8South, FrameIndex: 0}
	key3 := CacheKey{Seed: 3, State: "idle", Direction: Dir8South, FrameIndex: 0}

	// Add first two
	cache.Put(key1, img1)
	cache.Put(key2, img2)

	// Both should be cached
	if _, found := cache.Get(key1); !found {
		t.Error("key1 should be cached")
	}
	if _, found := cache.Get(key2); !found {
		t.Error("key2 should be cached")
	}

	// Add third, should evict key1 (least recently used)
	cache.Put(key3, img3)

	// key1 should be evicted
	if _, found := cache.Get(key1); found {
		t.Error("key1 should have been evicted")
	}

	// key2 and key3 should be cached
	if _, found := cache.Get(key2); !found {
		t.Error("key2 should still be cached")
	}
	if _, found := cache.Get(key3); !found {
		t.Error("key3 should be cached")
	}

	// Verify eviction stats
	stats := cache.GetStats()
	if stats.Evictions < 1 {
		t.Errorf("Expected at least 1 eviction, got %v", stats.Evictions)
	}
}

func TestCacheClear(t *testing.T) {
	cache := NewAnimationCache(10*1024*1024, 100)

	img := ebiten.NewImage(32, 32)
	key := CacheKey{Seed: 12345, State: "walk", Direction: Dir8North, FrameIndex: 0}

	cache.Put(key, img)

	// Verify cached
	if cache.Count() != 1 {
		t.Error("Cache should have 1 entry")
	}

	// Clear
	cache.Clear()

	// Verify empty
	if cache.Count() != 0 {
		t.Error("Cache should be empty after Clear()")
	}
	if _, found := cache.Get(key); found {
		t.Error("Cache should not contain key after Clear()")
	}
}

func TestCacheHitRate(t *testing.T) {
	cache := NewAnimationCache(10*1024*1024, 100)

	img := ebiten.NewImage(32, 32)
	key := CacheKey{Seed: 12345, State: "idle", Direction: Dir8South, FrameIndex: 0}

	cache.Put(key, img)

	// 3 hits
	cache.Get(key)
	cache.Get(key)
	cache.Get(key)

	// 1 miss
	cache.Get(CacheKey{Seed: 99999, State: "walk", Direction: Dir8East, FrameIndex: 0})

	stats := cache.GetStats()
	hitRate := stats.HitRate()

	// Hit rate should be 75% (3 hits / 4 total)
	expected := 75.0
	if hitRate < expected-0.1 || hitRate > expected+0.1 {
		t.Errorf("HitRate() = %v, want ~%v", hitRate, expected)
	}
}

func TestCacheTrimToSize(t *testing.T) {
	cache := NewAnimationCache(10*1024*1024, 100)

	// Add several entries
	for i := 0; i < 10; i++ {
		img := ebiten.NewImage(64, 64)
		key := CacheKey{Seed: int64(i), State: "idle", Direction: Dir8South, FrameIndex: 0}
		cache.Put(key, img)
	}

	initialSize := cache.Size()
	initialCount := cache.Count()

	// Trim to half size
	targetSize := initialSize / 2
	evicted := cache.TrimToSize(targetSize)

	if evicted == 0 {
		t.Error("TrimToSize should have evicted entries")
	}
	if cache.Size() > targetSize {
		t.Errorf("Cache size %v exceeds target %v after trim", cache.Size(), targetSize)
	}
	if cache.Count() >= initialCount {
		t.Error("Entry count should decrease after trim")
	}
}

func TestCacheTrimToCount(t *testing.T) {
	cache := NewAnimationCache(10*1024*1024, 100)

	// Add 10 entries
	for i := 0; i < 10; i++ {
		img := ebiten.NewImage(32, 32)
		key := CacheKey{Seed: int64(i), State: "walk", Direction: Dir8East, FrameIndex: 0}
		cache.Put(key, img)
	}

	// Trim to 5 entries
	evicted := cache.TrimToCount(5)

	if evicted != 5 {
		t.Errorf("TrimToCount evicted %v entries, want 5", evicted)
	}
	if cache.Count() != 5 {
		t.Errorf("Cache count = %v, want 5", cache.Count())
	}
}

func TestEstimateFrameSize(t *testing.T) {
	tests := []struct {
		width  int
		height int
		want   int64
	}{
		{32, 32, 32 * 32 * 4},
		{64, 64, 64 * 64 * 4},
		{128, 128, 128 * 128 * 4},
	}

	for _, tt := range tests {
		got := EstimateFrameSize(tt.width, tt.height)
		if got != tt.want {
			t.Errorf("EstimateFrameSize(%v, %v) = %v, want %v", tt.width, tt.height, got, tt.want)
		}
	}
}

func TestCommonAnimationStates(t *testing.T) {
	states := CommonAnimationStates()

	if len(states) == 0 {
		t.Error("CommonAnimationStates returned empty slice")
	}

	// Should include essential states
	essential := []string{"idle", "walk", "run", "attack"}
	for _, es := range essential {
		found := false
		for _, s := range states {
			if s == es {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CommonAnimationStates missing essential state: %v", es)
		}
	}
}

func TestCommonDirections(t *testing.T) {
	dirs := CommonDirections()

	if len(dirs) != 4 {
		t.Errorf("CommonDirections returned %v directions, want 4", len(dirs))
	}

	// Should be primary directions only (no diagonals)
	for _, dir := range dirs {
		if dir.IsDiagonal() {
			t.Errorf("CommonDirections should not include diagonal: %v", dir)
		}
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cache := NewAnimationCache(10*1024*1024, 1000)
	img := ebiten.NewImage(64, 64)
	key := CacheKey{Seed: 12345, State: "walk", Direction: Dir8North, FrameIndex: 0}
	cache.Put(key, img)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(key)
	}
}

func BenchmarkCachePut(b *testing.B) {
	cache := NewAnimationCache(100*1024*1024, 10000)
	img := ebiten.NewImage(64, 64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := CacheKey{Seed: int64(i), State: "idle", Direction: Dir8South, FrameIndex: i % 8}
		cache.Put(key, img)
	}
}

func BenchmarkCacheEviction(b *testing.B) {
	// Small cache that forces frequent evictions
	cache := NewAnimationCache(64*64*4*5, 5) // 5 entries max
	img := ebiten.NewImage(64, 64)

	// Pre-fill cache
	for i := 0; i < 5; i++ {
		key := CacheKey{Seed: int64(i), State: "walk", Direction: Dir8North, FrameIndex: 0}
		cache.Put(key, img)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := CacheKey{Seed: int64(i + 1000), State: "walk", Direction: Dir8South, FrameIndex: i % 8}
		cache.Put(key, img)
	}
}
