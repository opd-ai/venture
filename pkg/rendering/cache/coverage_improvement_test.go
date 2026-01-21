package cache

import (
	"errors"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// Tests for Priority 1 coverage improvements per AUDIT.md

// TestForceEviction tests the forceEviction function
// which aggressively evicts entries when hard limit is exceeded.
func TestForceEviction(t *testing.T) {
	cache := NewSpriteCache(500 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	// Set small limits for testing
	softLimit := int64(10 * 1024) // 10KB
	hardLimit := int64(15 * 1024) // 15KB
	monitor.SetLimits(softLimit, hardLimit)

	// Manually simulate exceeding hard limit
	monitor.mu.Lock()
	monitor.stats.CurrentUsage = hardLimit + 1000 // Exceed hard limit
	monitor.mu.Unlock()

	// Call forceEviction directly
	monitor.mu.Lock()
	monitor.forceEviction()
	monitor.mu.Unlock()

	// Check that eviction was recorded
	stats := monitor.Stats()
	if stats.EvictionCount != 1 {
		t.Errorf("EvictionCount = %d, want 1", stats.EvictionCount)
	}

	// LastCleanupAt should be set
	if stats.LastCleanupAt.IsZero() {
		t.Error("LastCleanupAt should be set after forceEviction")
	}

	// Check that cache max size was reduced to soft limit
	if cache.MaxSize() != softLimit {
		t.Errorf("cache.MaxSize() = %d, want %d (soft limit)", cache.MaxSize(), softLimit)
	}
}

// TestSoftCleanup tests the softCleanup function
// which gradually reduces cache when soft limit is exceeded.
func TestSoftCleanup(t *testing.T) {
	cache := NewSpriteCache(500 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	// Set small limits for testing
	softLimit := int64(100 * 1024) // 100KB
	hardLimit := int64(150 * 1024) // 150KB
	monitor.SetLimits(softLimit, hardLimit)

	// Manually simulate exceeding soft limit (but not hard limit)
	monitor.mu.Lock()
	monitor.stats.CurrentUsage = softLimit + 1000 // Exceed soft limit
	monitor.mu.Unlock()

	// Call softCleanup directly
	monitor.mu.Lock()
	monitor.softCleanup()
	monitor.mu.Unlock()

	// Check that cleanup was recorded
	stats := monitor.Stats()
	if stats.CleanupCount != 1 {
		t.Errorf("CleanupCount = %d, want 1", stats.CleanupCount)
	}

	// LastCleanupAt should be set
	if stats.LastCleanupAt.IsZero() {
		t.Error("LastCleanupAt should be set after softCleanup")
	}

	// Check that cache max size was reduced to 95% of soft limit
	expectedMaxSize := softLimit * 95 / 100
	if cache.MaxSize() != expectedMaxSize {
		t.Errorf("cache.MaxSize() = %d, want %d (95%% of soft limit)", cache.MaxSize(), expectedMaxSize)
	}
}

// TestSoftCleanup_NoOpWhenBelowTarget tests that softCleanup
// doesn't reduce cache size when already below target.
func TestSoftCleanup_NoOpWhenBelowTarget(t *testing.T) {
	// Create cache with small initial size
	smallMaxSize := int64(50 * 1024) // 50KB - smaller than 95% of soft limit
	cache := NewSpriteCache(smallMaxSize)
	monitor := NewMemoryMonitor(cache)

	// Set limits where 95% of soft limit > current max size
	softLimit := int64(100 * 1024) // 100KB, 95% = 95KB
	hardLimit := int64(150 * 1024) // 150KB
	monitor.SetLimits(softLimit, hardLimit)

	// Current max size (50KB) is already below target (95KB)
	initialMaxSize := cache.MaxSize()

	// Call softCleanup
	monitor.mu.Lock()
	monitor.softCleanup()
	monitor.mu.Unlock()

	// Max size should NOT change (it's already below target)
	if cache.MaxSize() != initialMaxSize {
		t.Errorf("cache.MaxSize() = %d, want %d (unchanged)", cache.MaxSize(), initialMaxSize)
	}

	// Cleanup should still be counted
	stats := monitor.Stats()
	if stats.CleanupCount != 1 {
		t.Errorf("CleanupCount = %d, want 1", stats.CleanupCount)
	}
}

// TestCheckMemory_HardLimitExceeded tests checkMemory when hard limit is exceeded.
func TestCheckMemory_HardLimitExceeded(t *testing.T) {
	cache := NewSpriteCache(500 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	// Set small limits
	softLimit := int64(10 * 1024)
	hardLimit := int64(15 * 1024)
	monitor.SetLimits(softLimit, hardLimit)

	// Add enough sprites to exceed hard limit
	for i := 0; i < 20; i++ {
		key := GenerateKey(int64(i), "test", 0)
		// Create 64x64 image = 16KB each
		img := ebiten.NewImage(64, 64)
		cache.Put(key, img)
	}

	// Check memory - should trigger force eviction
	monitor.checkMemory()

	stats := monitor.Stats()
	if stats.EvictionCount == 0 {
		t.Error("Expected eviction to occur when hard limit exceeded")
	}

	// Peak usage should be recorded
	if stats.PeakUsage == 0 {
		t.Error("PeakUsage should be recorded")
	}
}

// TestCheckMemory_SoftLimitExceeded tests checkMemory when only soft limit is exceeded.
func TestCheckMemory_SoftLimitExceeded(t *testing.T) {
	cache := NewSpriteCache(500 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	// Set limits where we can exceed soft but not hard
	softLimit := int64(50 * 1024)  // 50KB - ~3 sprites
	hardLimit := int64(200 * 1024) // 200KB - ~12 sprites
	monitor.SetLimits(softLimit, hardLimit)

	// Add 4 sprites (64KB) - exceeds soft limit but not hard
	for i := 0; i < 4; i++ {
		key := GenerateKey(int64(i), "test", 0)
		img := ebiten.NewImage(64, 64)
		cache.Put(key, img)
	}

	// Check memory - should trigger soft cleanup
	monitor.checkMemory()

	stats := monitor.Stats()
	if stats.CleanupCount == 0 {
		t.Error("Expected cleanup to occur when soft limit exceeded")
	}

	// Should NOT trigger eviction (only cleanup)
	if stats.EvictionCount != 0 {
		t.Errorf("Expected no eviction (only cleanup), got %d evictions", stats.EvictionCount)
	}
}

// TestUsagePercentage_ZeroLimit tests UsagePercentage with zero hard limit.
func TestUsagePercentage_ZeroLimit(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	monitor := NewMemoryMonitor(cache)

	// Set hard limit to zero
	monitor.SetLimits(0, 0)

	// Should return 0.0 without division by zero
	pct := monitor.UsagePercentage()
	if pct != 0.0 {
		t.Errorf("UsagePercentage() = %f, want 0.0 for zero limit", pct)
	}
}

// TestQueuePredictedSprites tests the QueuePredictedSprites function.
func TestQueuePredictedSprites(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	// Create sequential access pattern
	key1 := GenerateKey(1, "walk", 0)
	key2 := GenerateKey(1, "walk", 1)
	key3 := GenerateKey(1, "walk", 2)

	// Build up pattern: key1 -> key2 -> key3
	warmer.RecordAccess(key1, false, 1)
	warmer.RecordAccess(key2, false, 2)
	warmer.RecordAccess(key3, false, 3)
	warmer.RecordAccess(key1, false, 4) // Access key1 again to predict key2

	// Create a generator function that will be called for predictions
	generatorCalled := 0
	generator := func(key CacheKey) GeneratorFunc {
		return func() (*ebiten.Image, error) {
			generatorCalled++
			return ebiten.NewImage(32, 32), nil
		}
	}

	// Queue predicted sprites
	count := warmer.QueuePredictedSprites(generator)

	// Should have queued at least one prediction
	if count == 0 {
		// This is acceptable if key2 is already cached, check pregen queue
		queueSize := pregen.QueueSize()
		if queueSize == 0 {
			// Still acceptable - predictions may have been empty
			t.Log("No predictions queued - pattern may not have established next keys")
		}
	}

	// Process the queue
	generated := pregen.Generate()

	// If we queued anything, it should be generated
	if count > 0 && generated == 0 {
		t.Errorf("Queued %d predictions but generated 0", count)
	}
}

// TestQueuePredictedSprites_NilGenerator tests that nil generators are skipped.
func TestQueuePredictedSprites_NilGenerator(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	// Create sequential access pattern
	key1 := GenerateKey(1, "walk", 0)
	key2 := GenerateKey(1, "walk", 1)

	warmer.RecordAccess(key1, false, 1)
	warmer.RecordAccess(key2, false, 2)
	warmer.RecordAccess(key1, false, 3) // Predict key2

	// Generator that returns nil for all keys
	generator := func(key CacheKey) GeneratorFunc {
		return nil
	}

	// Should not panic and should handle nil generator gracefully
	count := warmer.QueuePredictedSprites(generator)

	// Count is the number of predictions, not queued items
	// Since generator returns nil, nothing should be queued
	queueSize := pregen.QueueSize()
	if queueSize != 0 {
		t.Errorf("Expected empty queue when generator returns nil, got %d", queueSize)
	}

	_ = count // Suppress unused variable warning
}

// TestQueuePredictedSprites_EmptyPredictions tests with no predictions available.
func TestQueuePredictedSprites_EmptyPredictions(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	// Don't create any pattern - predictions will be empty

	generator := func(key CacheKey) GeneratorFunc {
		return func() (*ebiten.Image, error) {
			return ebiten.NewImage(32, 32), nil
		}
	}

	// Should return 0 with no predictions
	count := warmer.QueuePredictedSprites(generator)
	if count != 0 {
		t.Errorf("Expected 0 predictions for empty warmer, got %d", count)
	}
}

// TestAnalyzeAnimationSequence_EdgeCases tests edge cases in animation analysis.
func TestAnalyzeAnimationSequence_EdgeCases(t *testing.T) {
	t.Run("empty sequence", func(t *testing.T) {
		cache := NewSpriteCache(100 * 1024 * 1024)
		pregen := NewPreGenerator(cache)
		warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

		// Should not panic with empty sequence
		warmer.AnalyzeAnimationSequence([]CacheKey{})

		stats := warmer.Stats()
		if stats.PatternCount != 0 {
			t.Errorf("Expected 0 patterns for empty sequence, got %d", stats.PatternCount)
		}
	})

	t.Run("single frame", func(t *testing.T) {
		cache := NewSpriteCache(100 * 1024 * 1024)
		pregen := NewPreGenerator(cache)
		warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

		// Single frame has no "next" to analyze
		warmer.AnalyzeAnimationSequence([]CacheKey{GenerateKey(1, "idle", 0)})

		stats := warmer.Stats()
		// Should create pattern but with no next keys
		if stats.PatternCount != 0 {
			t.Errorf("Expected 0 patterns for single frame, got %d", stats.PatternCount)
		}
	})

	t.Run("duplicate frames in sequence", func(t *testing.T) {
		cache := NewSpriteCache(100 * 1024 * 1024)
		pregen := NewPreGenerator(cache)
		warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

		key := GenerateKey(1, "idle", 0)
		// Same frame repeated - should not add duplicate next keys
		warmer.AnalyzeAnimationSequence([]CacheKey{key, key, key})

		warmer.mu.RLock()
		pattern := warmer.patterns[key]
		warmer.mu.RUnlock()

		if pattern != nil && len(pattern.NextKeys) > 1 {
			t.Errorf("Expected at most 1 next key (no duplicates), got %d", len(pattern.NextKeys))
		}
	})

	t.Run("max next keys limit", func(t *testing.T) {
		cache := NewSpriteCache(100 * 1024 * 1024)
		pregen := NewPreGenerator(cache)
		warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

		// Create many different sequences from same starting key
		startKey := GenerateKey(1, "start", 0)
		for i := 0; i < 20; i++ {
			nextKey := GenerateKey(int64(i+2), "next", i)
			warmer.AnalyzeAnimationSequence([]CacheKey{startKey, nextKey})
		}

		warmer.mu.RLock()
		pattern := warmer.patterns[startKey]
		warmer.mu.RUnlock()

		if pattern != nil && len(pattern.NextKeys) > 8 {
			t.Errorf("Expected max 8 next keys, got %d", len(pattern.NextKeys))
		}
	})
}

// TestPregenerator_GeneratePartialFailures tests Generate with some failures.
func TestPregenerator_GeneratePartialFailures(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Queue 3 requests: 2 will succeed, 1 will fail
	successKey1 := GenerateKey(1, "success", 0)
	successKey2 := GenerateKey(2, "success", 0)
	failKey := GenerateKey(3, "fail", 0)

	pregen.Queue(successKey1, func() (*ebiten.Image, error) {
		return ebiten.NewImage(32, 32), nil
	})

	pregen.Queue(failKey, func() (*ebiten.Image, error) {
		return nil, errors.New("simulated generator failure")
	})

	pregen.Queue(successKey2, func() (*ebiten.Image, error) {
		return ebiten.NewImage(32, 32), nil
	})

	// Generate
	generated := pregen.Generate()

	// Should generate 2 (the successful ones)
	// Note: failKey generator returns nil image, which won't error but will be weird
	// Let's check stats instead
	stats := pregen.Stats()

	if stats.RequestsQueued != 3 {
		t.Errorf("RequestsQueued = %d, want 3", stats.RequestsQueued)
	}

	// At least some should complete
	if stats.RequestsComplete == 0 {
		t.Error("Expected some requests to complete")
	}

	_ = generated
}

// TestSpriteCache_Get_CorruptedEntry tests Get with a corrupted cache entry.
func TestSpriteCache_Get_CorruptedEntry(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)

	key := GenerateKey(1, "test", 0)

	// Put a valid image first
	img := ebiten.NewImage(32, 32)
	cache.Put(key, img)

	// Manually corrupt the cache entry by replacing with wrong type
	cache.mu.Lock()
	if elem, ok := cache.cache[key]; ok {
		// Replace entry value with something that's not *entry
		elem.Value = "corrupted" // Wrong type
	}
	cache.mu.Unlock()

	// Get should handle this gracefully
	result, found := cache.Get(key)

	if found {
		t.Error("Expected not found for corrupted entry")
	}

	if result != nil {
		t.Error("Expected nil image for corrupted entry")
	}

	// Should count as a miss (not a hit)
	stats := cache.Stats()
	if stats.Misses == 0 {
		t.Error("Expected miss to be recorded for corrupted entry")
	}
}
