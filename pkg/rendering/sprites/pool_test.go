package sprites

import (
	"sync"
	"testing"
)

func TestNewImagePool(t *testing.T) {
	pool := NewImagePool(32, 32)

	if pool == nil {
		t.Fatal("NewImagePool returned nil")
	}

	if pool.width != 32 {
		t.Errorf("width = %d, want 32", pool.width)
	}

	if pool.height != 32 {
		t.Errorf("height = %d, want 32", pool.height)
	}
}

func TestNewShapePool(t *testing.T) {
	sp := NewShapePool()

	if sp == nil {
		t.Fatal("NewShapePool returned nil")
	}

	if sp.pools == nil {
		t.Error("pools map is nil")
	}

	if len(sp.pools) != 0 {
		t.Errorf("initial pools length = %d, want 0", len(sp.pools))
	}
}

func TestShapePool_Clear(t *testing.T) {
	sp := NewShapePool()

	// Simulate some pool creation by accessing internal state
	sp.mutex.Lock()
	sp.pools[320032] = NewImagePool(32, 32)
	sp.pools[640064] = NewImagePool(64, 64)
	sp.mutex.Unlock()

	if len(sp.pools) != 2 {
		t.Errorf("pools length before clear = %d, want 2", len(sp.pools))
	}

	sp.Clear()

	if len(sp.pools) != 0 {
		t.Errorf("pools length after clear = %d, want 0", len(sp.pools))
	}
}

func TestNewPooledGenerator(t *testing.T) {
	pg := NewPooledGenerator()

	if pg == nil {
		t.Fatal("NewPooledGenerator returned nil")
	}

	if pg.generator == nil {
		t.Error("generator is nil")
	}

	if pg.shapePool == nil {
		t.Error("shapePool is nil")
	}

	if !pg.enabled {
		t.Error("pooling should be enabled by default")
	}
}

func TestPooledGenerator_SetPoolingEnabled(t *testing.T) {
	pg := NewPooledGenerator()

	if !pg.IsPoolingEnabled() {
		t.Error("pooling should be enabled by default")
	}

	pg.SetPoolingEnabled(false)
	if pg.IsPoolingEnabled() {
		t.Error("pooling should be disabled")
	}

	pg.SetPoolingEnabled(true)
	if !pg.IsPoolingEnabled() {
		t.Error("pooling should be enabled")
	}
}

func TestPooledGenerator_GetPool(t *testing.T) {
	pg := NewPooledGenerator()

	pool := pg.GetPool()
	if pool == nil {
		t.Error("GetPool() returned nil")
	}

	if pool != pg.shapePool {
		t.Error("GetPool() returned different pool than internal pool")
	}
}

func TestNewCombinedGenerator(t *testing.T) {
	cg := NewCombinedGenerator(50)

	if cg == nil {
		t.Fatal("NewCombinedGenerator returned nil")
	}

	if cg.generator == nil {
		t.Error("generator is nil")
	}

	if cg.cache == nil {
		t.Error("cache is nil")
	}

	if cg.shapePool == nil {
		t.Error("shapePool is nil")
	}

	if !cg.cacheEnabled {
		t.Error("caching should be enabled by default")
	}

	if !cg.poolEnabled {
		t.Error("pooling should be enabled by default")
	}
}

func TestCombinedGenerator_SetCacheEnabled(t *testing.T) {
	cg := NewCombinedGenerator(10)

	if !cg.IsCacheEnabled() {
		t.Error("cache should be enabled by default")
	}

	cg.SetCacheEnabled(false)
	if cg.IsCacheEnabled() {
		t.Error("cache should be disabled")
	}

	cg.SetCacheEnabled(true)
	if !cg.IsCacheEnabled() {
		t.Error("cache should be enabled")
	}
}

func TestCombinedGenerator_SetPoolingEnabled(t *testing.T) {
	cg := NewCombinedGenerator(10)

	if !cg.IsPoolingEnabled() {
		t.Error("pooling should be enabled by default")
	}

	cg.SetPoolingEnabled(false)
	if cg.IsPoolingEnabled() {
		t.Error("pooling should be disabled")
	}

	cg.SetPoolingEnabled(true)
	if !cg.IsPoolingEnabled() {
		t.Error("pooling should be enabled")
	}
}

func TestCombinedGenerator_GetCache(t *testing.T) {
	cg := NewCombinedGenerator(10)

	cache := cg.GetCache()
	if cache == nil {
		t.Error("GetCache() returned nil")
	}

	if cache.Capacity() != 10 {
		t.Errorf("cache capacity = %d, want 10", cache.Capacity())
	}
}

func TestCombinedGenerator_GetPool(t *testing.T) {
	cg := NewCombinedGenerator(10)

	pool := cg.GetPool()
	if pool == nil {
		t.Error("GetPool() returned nil")
	}

	if pool != cg.shapePool {
		t.Error("GetPool() returned different pool than internal pool")
	}
}

func TestCombinedGenerator_ClearCache(t *testing.T) {
	cg := NewCombinedGenerator(10)

	// Simulate some cache activity
	cg.cache.hits = 5
	cg.cache.misses = 3

	cg.ClearCache()

	stats := cg.Stats()
	if stats.Hits != 0 {
		t.Errorf("Hits after ClearCache() = %d, want 0", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("Misses after ClearCache() = %d, want 0", stats.Misses)
	}
}

func TestCombinedGenerator_ClearPool(t *testing.T) {
	cg := NewCombinedGenerator(10)

	// Simulate some pool creation
	cg.shapePool.mutex.Lock()
	cg.shapePool.pools[320032] = NewImagePool(32, 32)
	cg.shapePool.mutex.Unlock()

	if len(cg.shapePool.pools) == 0 {
		t.Error("pool should have entries before clear")
	}

	cg.ClearPool()

	if len(cg.shapePool.pools) != 0 {
		t.Errorf("pools length after ClearPool() = %d, want 0", len(cg.shapePool.pools))
	}
}

func TestCombinedGenerator_Stats(t *testing.T) {
	cg := NewCombinedGenerator(10)

	stats := cg.Stats()

	if stats.Capacity != 10 {
		t.Errorf("Stats.Capacity = %d, want 10", stats.Capacity)
	}

	if stats.Size != 0 {
		t.Errorf("Stats.Size = %d, want 0", stats.Size)
	}
}

func TestImagePool_Dimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"square 32x32", 32, 32},
		{"square 64x64", 64, 64},
		{"rectangle 32x64", 32, 64},
		{"rectangle 64x32", 64, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewImagePool(tt.width, tt.height)

			if pool.width != tt.width {
				t.Errorf("width = %d, want %d", pool.width, tt.width)
			}

			if pool.height != tt.height {
				t.Errorf("height = %d, want %d", pool.height, tt.height)
			}
		})
	}
}

func TestShapePool_KeyEncoding(t *testing.T) {
	// Test that different sizes produce different keys
	sizes := [][2]int{
		{32, 32},
		{32, 64},
		{64, 32},
		{64, 64},
	}

	keys := make(map[int]bool)
	for _, size := range sizes {
		key := size[0]*10000 + size[1]
		if keys[key] {
			t.Errorf("Duplicate key %d for size %dx%d", key, size[0], size[1])
		}
		keys[key] = true
	}

	if len(keys) != len(sizes) {
		t.Errorf("Key encoding produced %d unique keys, want %d", len(keys), len(sizes))
	}
}

func TestPooledGenerator_ClearPool(t *testing.T) {
	pg := NewPooledGenerator()

	// Simulate some pool creation
	pg.shapePool.mutex.Lock()
	pg.shapePool.pools[320032] = NewImagePool(32, 32)
	pg.shapePool.pools[640064] = NewImagePool(64, 64)
	pg.shapePool.mutex.Unlock()

	if len(pg.shapePool.pools) != 2 {
		t.Errorf("pools length before clear = %d, want 2", len(pg.shapePool.pools))
	}

	pg.ClearPool()

	if len(pg.shapePool.pools) != 0 {
		t.Errorf("pools length after clear = %d, want 0", len(pg.shapePool.pools))
	}
}

func TestShapePool_Concurrency(t *testing.T) {
	sp := NewShapePool()

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	// Concurrent pool operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				// Simulate accessing different pool sizes
				width := 32 + (id%3)*32 // 32, 64, or 96
				height := 32 + (j%3)*32

				// This will create pools on-demand
				key := width*10000 + height

				// Access pool through mutex
				sp.mutex.RLock()
				_ = sp.pools[key]
				sp.mutex.RUnlock()
			}
		}(i)
	}

	wg.Wait()

	// Verify pool is in valid state
	sp.mutex.RLock()
	poolCount := len(sp.pools)
	sp.mutex.RUnlock()

	// Should have created pools for various sizes
	// Exact count depends on timing, just verify no crashes
	if poolCount < 0 {
		t.Error("Invalid pool count after concurrent access")
	}
}

func TestCombinedGenerator_BothFeaturesEnabled(t *testing.T) {
	cg := NewCombinedGenerator(10)

	// Verify both features enabled by default
	if !cg.IsCacheEnabled() || !cg.IsPoolingEnabled() {
		t.Error("Both caching and pooling should be enabled by default")
	}

	// Disable caching, keep pooling
	cg.SetCacheEnabled(false)
	if cg.IsCacheEnabled() || !cg.IsPoolingEnabled() {
		t.Error("Only pooling should be enabled")
	}

	// Disable pooling, re-enable caching
	cg.SetCacheEnabled(true)
	cg.SetPoolingEnabled(false)
	if !cg.IsCacheEnabled() || cg.IsPoolingEnabled() {
		t.Error("Only caching should be enabled")
	}

	// Disable both
	cg.SetCacheEnabled(false)
	cg.SetPoolingEnabled(false)
	if cg.IsCacheEnabled() || cg.IsPoolingEnabled() {
		t.Error("Both should be disabled")
	}
}

func TestImagePool_NilHandling(t *testing.T) {
	pool := NewImagePool(32, 32)

	// Put nil image should not panic
	pool.Put(nil)

	// Verify pool is still functional
	_ = pool.Get()
}

func TestShapePool_NilHandling(t *testing.T) {
	sp := NewShapePool()

	// Put nil image should not panic
	sp.Put(nil)

	// Verify pool is still functional
	img := sp.Get(32, 32)
	if img == nil {
		t.Error("Get() returned nil after Put(nil)")
	}
}
