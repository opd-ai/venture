package cache

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewSpriteCache(t *testing.T) {
	cache := NewSpriteCache(1024 * 1024) // 1MB
	if cache == nil {
		t.Fatal("NewSpriteCache returned nil")
	}
	if cache.MaxSize() != 1024*1024 {
		t.Errorf("MaxSize = %d, want %d", cache.MaxSize(), 1024*1024)
	}
	if cache.Count() != 0 {
		t.Errorf("Initial count = %d, want 0", cache.Count())
	}
}

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name  string
		seed  int64
		state string
		frame int
		want  CacheKey
	}{
		{"simple", 12345, "idle", 0, "12345:idle:0"},
		{"different seed", 67890, "idle", 0, "67890:idle:0"},
		{"different state", 12345, "walk", 0, "12345:walk:0"},
		{"different frame", 12345, "idle", 5, "12345:idle:5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateKey(tt.seed, tt.state, tt.frame)
			if got != tt.want {
				t.Errorf("GenerateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateCompositeKey(t *testing.T) {
	key1 := GenerateCompositeKey(12345, []string{"body", "head", "weapon"})
	key2 := GenerateCompositeKey(12345, []string{"body", "head", "weapon"})
	key3 := GenerateCompositeKey(12345, []string{"body", "head"})

	// Same inputs should produce same key
	if key1 != key2 {
		t.Error("Same inputs produced different keys")
	}

	// Different inputs should produce different keys
	if key1 == key3 {
		t.Error("Different inputs produced same key")
	}
}

func TestSpriteCache_PutAndGet(t *testing.T) {
	cache := NewSpriteCache(10 * 1024 * 1024) // 10MB
	img := ebiten.NewImage(32, 32)
	key := GenerateKey(12345, "idle", 0)

	// Initially should not be in cache
	if _, ok := cache.Get(key); ok {
		t.Error("Cache should be empty initially")
	}

	// Put image in cache
	cache.Put(key, img)

	// Should now be in cache
	gotImg, ok := cache.Get(key)
	if !ok {
		t.Fatal("Get failed after Put")
	}
	if gotImg != img {
		t.Error("Got different image from cache")
	}

	// Stats should reflect one hit
	stats := cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("Hits = %d, want 1", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Misses = %d, want 1", stats.Misses)
	}
}

func TestSpriteCache_LRUEviction(t *testing.T) {
	// Create small cache (enough for ~2 images of 32x32)
	imageSize := 32 * 32 * 4 // RGBA
	cache := NewSpriteCache(int64(imageSize * 2))

	img1 := ebiten.NewImage(32, 32)
	img2 := ebiten.NewImage(32, 32)
	img3 := ebiten.NewImage(32, 32)

	key1 := GenerateKey(1, "idle", 0)
	key2 := GenerateKey(2, "idle", 0)
	key3 := GenerateKey(3, "idle", 0)

	// Add first two images
	cache.Put(key1, img1)
	cache.Put(key2, img2)

	// Both should be in cache
	if _, ok := cache.Get(key1); !ok {
		t.Error("key1 should be in cache")
	}
	if _, ok := cache.Get(key2); !ok {
		t.Error("key2 should be in cache")
	}

	// Add third image, should evict key1 (least recently used)
	cache.Put(key3, img3)

	// key1 should have been evicted
	if _, ok := cache.Get(key1); ok {
		t.Error("key1 should have been evicted")
	}

	// key2 and key3 should still be in cache
	if _, ok := cache.Get(key2); !ok {
		t.Error("key2 should still be in cache")
	}
	if _, ok := cache.Get(key3); !ok {
		t.Error("key3 should still be in cache")
	}

	// Check eviction stats
	stats := cache.Stats()
	if stats.Evictions == 0 {
		t.Error("Expected at least one eviction")
	}
}

func TestSpriteCache_Clear(t *testing.T) {
	cache := NewSpriteCache(10 * 1024 * 1024)
	img := ebiten.NewImage(32, 32)

	// Add some entries
	for i := 0; i < 5; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		cache.Put(key, img)
	}

	if cache.Count() != 5 {
		t.Errorf("Count = %d, want 5", cache.Count())
	}

	// Clear cache
	cache.Clear()

	if cache.Count() != 0 {
		t.Errorf("After Clear, Count = %d, want 0", cache.Count())
	}
	if cache.Size() != 0 {
		t.Errorf("After Clear, Size = %d, want 0", cache.Size())
	}
}

func TestSpriteCache_Remove(t *testing.T) {
	cache := NewSpriteCache(10 * 1024 * 1024)
	img := ebiten.NewImage(32, 32)
	key := GenerateKey(12345, "idle", 0)

	// Add entry
	cache.Put(key, img)

	if !cache.Contains(key) {
		t.Error("Key should be in cache")
	}

	// Remove entry
	if !cache.Remove(key) {
		t.Error("Remove should return true for existing key")
	}

	if cache.Contains(key) {
		t.Error("Key should not be in cache after Remove")
	}

	// Removing again should return false
	if cache.Remove(key) {
		t.Error("Remove should return false for non-existent key")
	}
}

func TestSpriteCache_SetMaxSize(t *testing.T) {
	cache := NewSpriteCache(10 * 1024 * 1024)
	img := ebiten.NewImage(32, 32)

	// Add entries
	for i := 0; i < 10; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		cache.Put(key, img)
	}

	initialCount := cache.Count()
	if initialCount != 10 {
		t.Errorf("Initial count = %d, want 10", initialCount)
	}

	// Reduce max size to force evictions
	imageSize := 32 * 32 * 4
	cache.SetMaxSize(int64(imageSize * 3)) // Only room for 3 images

	// Should have evicted some entries
	if cache.Count() >= initialCount {
		t.Error("SetMaxSize should have triggered evictions")
	}

	// Size should be within new limit
	if cache.Size() > cache.MaxSize() {
		t.Errorf("Size %d exceeds MaxSize %d", cache.Size(), cache.MaxSize())
	}
}

func TestSpriteCache_Contains(t *testing.T) {
	cache := NewSpriteCache(10 * 1024 * 1024)
	img := ebiten.NewImage(32, 32)
	key := GenerateKey(12345, "idle", 0)

	if cache.Contains(key) {
		t.Error("Contains should return false for non-existent key")
	}

	cache.Put(key, img)

	if !cache.Contains(key) {
		t.Error("Contains should return true for existing key")
	}
}

func TestStatistics_HitRate(t *testing.T) {
	tests := []struct {
		name   string
		hits   uint64
		misses uint64
		want   float64
	}{
		{"no data", 0, 0, 0.0},
		{"all hits", 10, 0, 1.0},
		{"all misses", 0, 10, 0.0},
		{"50% hit rate", 5, 5, 0.5},
		{"75% hit rate", 75, 25, 0.75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := Statistics{
				Hits:   tt.hits,
				Misses: tt.misses,
			}
			got := stats.HitRate()
			if got != tt.want {
				t.Errorf("HitRate() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestSpriteCache_ConcurrentAccess(t *testing.T) {
	cache := NewSpriteCache(10 * 1024 * 1024)
	img := ebiten.NewImage(32, 32)

	// Test concurrent Put/Get operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := GenerateKey(int64(id), "idle", 0)
			cache.Put(key, img)
			_, _ = cache.Get(key)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Cache should have all entries
	if cache.Count() != 10 {
		t.Errorf("After concurrent access, Count = %d, want 10", cache.Count())
	}
}

// Benchmarks

func BenchmarkSpriteCache_Put(b *testing.B) {
	cache := NewSpriteCache(100 * 1024 * 1024) // 100MB
	img := ebiten.NewImage(32, 32)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := GenerateKey(int64(i%1000), "idle", i%10)
		cache.Put(key, img)
	}
}

func BenchmarkSpriteCache_Get_Hit(b *testing.B) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	img := ebiten.NewImage(32, 32)

	// Pre-populate cache
	keys := make([]CacheKey, 100)
	for i := 0; i < 100; i++ {
		keys[i] = GenerateKey(int64(i), "idle", 0)
		cache.Put(keys[i], img)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(keys[i%100])
	}
}

func BenchmarkSpriteCache_Get_Miss(b *testing.B) {
	cache := NewSpriteCache(100 * 1024 * 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		cache.Get(key)
	}
}

func BenchmarkGenerateKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateKey(12345, "idle", i%10)
	}
}

func BenchmarkGenerateCompositeKey(b *testing.B) {
	layers := []string{"body", "head", "weapon", "armor"}
	for i := 0; i < b.N; i++ {
		GenerateCompositeKey(12345, layers)
	}
}

// Phase 45 tests for updated constants and 64×64 default support

func TestPhase45_SpriteSizeConstants(t *testing.T) {
	// Verify sprite size constants are correct
	if BytesPerPixel != 4 {
		t.Errorf("BytesPerPixel = %d, want 4 (RGBA)", BytesPerPixel)
	}

	// Test primary constants (preferred naming convention)
	if Sprite32MemorySize != 32*32*4 {
		t.Errorf("Sprite32MemorySize = %d, want %d (4KB)", Sprite32MemorySize, 32*32*4)
	}

	if Sprite64MemorySize != 64*64*4 {
		t.Errorf("Sprite64MemorySize = %d, want %d (16KB)", Sprite64MemorySize, 64*64*4)
	}

	if Sprite128MemorySize != 128*128*4 {
		t.Errorf("Sprite128MemorySize = %d, want %d (64KB)", Sprite128MemorySize, 128*128*4)
	}

	// Test deprecated aliases still work
	if SpriteSize32 != Sprite32MemorySize {
		t.Errorf("SpriteSize32 (%d) != Sprite32MemorySize (%d)", SpriteSize32, Sprite32MemorySize)
	}
	if SpriteSize64 != Sprite64MemorySize {
		t.Errorf("SpriteSize64 (%d) != Sprite64MemorySize (%d)", SpriteSize64, Sprite64MemorySize)
	}
	if SpriteSize128 != Sprite128MemorySize {
		t.Errorf("SpriteSize128 (%d) != Sprite128MemorySize (%d)", SpriteSize128, Sprite128MemorySize)
	}
}

func TestPhase45_CacheSizeConstants(t *testing.T) {
	// Verify cache size constants are sensible
	if DefaultCacheSize != 16*1024*1024 {
		t.Errorf("DefaultCacheSize = %d, want 16MB", DefaultCacheSize)
	}

	if MaxCacheSize != 300*1024*1024 {
		t.Errorf("MaxCacheSize = %d, want 300MB", MaxCacheSize)
	}

	// Verify default can hold reasonable number of 64×64 sprites
	// DefaultCacheSize (16MB) / Sprite64MemorySize (16KB) = 1024 sprites
	sprites64InDefault := DefaultCacheSize / Sprite64MemorySize

	// Verify the calculated capacity matches documentation (~1024)
	if sprites64InDefault < 1000 || sprites64InDefault > 1100 {
		t.Errorf("Expected ~1024 sprites, got %d", sprites64InDefault)
	}

	// Verify max stays under memory target
	if MaxCacheSize > 300*1024*1024 {
		t.Errorf("MaxCacheSize = %d exceeds 300MB target", MaxCacheSize)
	}
}

func TestPhase45_CacheWith64x64Sprites(t *testing.T) {
	// Create cache sized for 64×64 sprites
	cache := NewSpriteCache(Sprite64MemorySize * 10) // Room for 10 sprites

	// Add 64×64 sprites
	for i := 0; i < 5; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		img := ebiten.NewImage(64, 64)
		cache.Put(key, img)
	}

	if cache.Count() != 5 {
		t.Errorf("Count = %d, want 5", cache.Count())
	}

	// Verify size calculation is correct (5 × 16KB = 80KB)
	expectedSize := int64(5 * Sprite64MemorySize)
	if cache.Size() != expectedSize {
		t.Errorf("Size = %d, want %d", cache.Size(), expectedSize)
	}

	// Add more sprites to trigger eviction
	for i := 5; i < 15; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		img := ebiten.NewImage(64, 64)
		cache.Put(key, img)
	}

	// Cache should have evicted old entries
	stats := cache.Stats()
	if stats.Evictions == 0 {
		t.Error("Expected evictions when cache is full")
	}

	// Size should not exceed max
	if cache.Size() > cache.MaxSize() {
		t.Errorf("Size %d exceeds MaxSize %d", cache.Size(), cache.MaxSize())
	}
}

func TestPhase45_DefaultCacheCapacity(t *testing.T) {
	// Create cache with default size
	cache := NewSpriteCache(DefaultCacheSize)

	if cache.MaxSize() != DefaultCacheSize {
		t.Errorf("MaxSize = %d, want DefaultCacheSize (%d)", cache.MaxSize(), DefaultCacheSize)
	}

	// Should be able to hold many 64×64 sprites
	for i := 0; i < 100; i++ {
		key := GenerateKey(int64(i), "walk", i%8)
		img := ebiten.NewImage(64, 64)
		cache.Put(key, img)
	}

	if cache.Count() != 100 {
		t.Errorf("Count = %d, want 100 (all should fit)", cache.Count())
	}

	// Total size should be 100 × 16KB = 1.6MB, well under 16MB limit
	expectedSize := int64(100 * Sprite64MemorySize)
	if cache.Size() != expectedSize {
		t.Errorf("Size = %d, want %d", cache.Size(), expectedSize)
	}
}
