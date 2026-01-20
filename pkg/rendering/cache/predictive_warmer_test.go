package cache

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewPredictiveCacheWarmer(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	config := DefaultWarmerConfig()

	warmer := NewPredictiveCacheWarmer(cache, pregen, config)

	if warmer == nil {
		t.Fatal("NewPredictiveCacheWarmer returned nil")
	}

	if warmer.windowSize != config.WindowSize {
		t.Errorf("windowSize = %d, want %d", warmer.windowSize, config.WindowSize)
	}

	if warmer.hotThreshold != config.HotThreshold {
		t.Errorf("hotThreshold = %d, want %d", warmer.hotThreshold, config.HotThreshold)
	}
}

func TestNewPredictiveCacheWarmer_Defaults(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Test with zero values (should use defaults)
	warmer := NewPredictiveCacheWarmer(cache, pregen, PredictiveWarmerConfig{})

	if warmer.windowSize != 1000 {
		t.Errorf("windowSize = %d, want 1000 (default)", warmer.windowSize)
	}

	if warmer.hotThreshold != 5 {
		t.Errorf("hotThreshold = %d, want 5 (default)", warmer.hotThreshold)
	}

	if warmer.maxPredictions != 50 {
		t.Errorf("maxPredictions = %d, want 50 (default)", warmer.maxPredictions)
	}
}

func TestDefaultWarmerConfig(t *testing.T) {
	config := DefaultWarmerConfig()

	if config.WindowSize != 1000 {
		t.Errorf("WindowSize = %d, want 1000", config.WindowSize)
	}

	if config.HotThreshold != 5 {
		t.Errorf("HotThreshold = %d, want 5", config.HotThreshold)
	}

	if config.MaxPredictions != 50 {
		t.Errorf("MaxPredictions = %d, want 50", config.MaxPredictions)
	}
}

func TestRecordAccess(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	key := GenerateKey(12345, "idle", 0)
	warmer.RecordAccess(key, true, 1)

	stats := warmer.Stats()
	if stats.AccessLogSize != 1 {
		t.Errorf("AccessLogSize = %d, want 1", stats.AccessLogSize)
	}

	if stats.PatternCount != 1 {
		t.Errorf("PatternCount = %d, want 1", stats.PatternCount)
	}
}

func TestRecordAccess_WindowTrim(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	config := PredictiveWarmerConfig{
		WindowSize:     10,
		HotThreshold:   2,
		MaxPredictions: 5,
	}
	warmer := NewPredictiveCacheWarmer(cache, pregen, config)

	// Record more accesses than window size
	for i := 0; i < 15; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		warmer.RecordAccess(key, true, int64(i))
	}

	stats := warmer.Stats()
	// Window should be trimmed to 10
	if stats.AccessLogSize != 10 {
		t.Errorf("AccessLogSize = %d, want 10 (window size)", stats.AccessLogSize)
	}
}

func TestRecordAccess_SequentialPattern(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	key1 := GenerateKey(1, "idle", 0)
	key2 := GenerateKey(1, "idle", 1)
	key3 := GenerateKey(1, "idle", 2)

	// Simulate sequential frame access
	warmer.RecordAccess(key1, false, 1)
	warmer.RecordAccess(key2, false, 2)
	warmer.RecordAccess(key3, false, 3)

	// key1 should have key2 as a predicted next key
	warmer.mu.RLock()
	pattern1 := warmer.patterns[key1]
	warmer.mu.RUnlock()

	if pattern1 == nil {
		t.Fatal("pattern for key1 not found")
	}

	found := false
	for _, k := range pattern1.NextKeys {
		if k == key2 {
			found = true
			break
		}
	}

	if !found {
		t.Error("key2 not found in key1's NextKeys")
	}
}

func TestGetHotSprites(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	config := PredictiveWarmerConfig{
		WindowSize:     1000,
		HotThreshold:   3, // Access 3 times to be "hot"
		MaxPredictions: 50,
	}
	warmer := NewPredictiveCacheWarmer(cache, pregen, config)

	hotKey := GenerateKey(1, "hot", 0)
	coldKey := GenerateKey(2, "cold", 0)

	// Access hot key 5 times
	for i := 0; i < 5; i++ {
		warmer.RecordAccess(hotKey, true, int64(i))
	}

	// Access cold key 1 time
	warmer.RecordAccess(coldKey, true, 5)

	hotSprites := warmer.GetHotSprites()

	// Should contain hotKey but not coldKey
	foundHot := false
	foundCold := false
	for _, k := range hotSprites {
		if k == hotKey {
			foundHot = true
		}
		if k == coldKey {
			foundCold = true
		}
	}

	if !foundHot {
		t.Error("hotKey not found in hot sprites")
	}

	if foundCold {
		t.Error("coldKey should not be in hot sprites")
	}
}

func TestGetHotSprites_Empty(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	hotSprites := warmer.GetHotSprites()
	if len(hotSprites) != 0 {
		t.Errorf("Expected empty hot sprites, got %d", len(hotSprites))
	}
}

func TestPredictNext(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	key1 := GenerateKey(1, "walk", 0)
	key2 := GenerateKey(1, "walk", 1)

	// Create sequential pattern
	warmer.RecordAccess(key1, false, 1)
	warmer.RecordAccess(key2, false, 2)

	// Now access key1 again
	warmer.RecordAccess(key1, true, 3)

	// Should predict key2
	predictions := warmer.PredictNext()

	found := false
	for _, k := range predictions {
		if k == key2 {
			found = true
			break
		}
	}

	if !found {
		t.Error("key2 not predicted after key1")
	}
}

func TestPredictNext_AlreadyCached(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	key1 := GenerateKey(1, "walk", 0)
	key2 := GenerateKey(1, "walk", 1)

	// Create sequential pattern
	warmer.RecordAccess(key1, false, 1)
	warmer.RecordAccess(key2, false, 2)

	// Put key2 in cache with a real image
	img := ebiten.NewImage(32, 32)
	cache.Put(key2, img)

	// Access key1
	warmer.RecordAccess(key1, true, 3)

	// Should NOT predict key2 (already cached)
	predictions := warmer.PredictNext()

	for _, k := range predictions {
		if k == key2 {
			t.Error("key2 should not be predicted (already cached)")
		}
	}
}

func TestPredictNext_Empty(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	predictions := warmer.PredictNext()
	if predictions != nil {
		t.Errorf("Expected nil predictions for empty warmer, got %d", len(predictions))
	}
}

func TestWarmerStats(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	config := PredictiveWarmerConfig{
		WindowSize:     1000,
		HotThreshold:   3,
		MaxPredictions: 50,
	}
	warmer := NewPredictiveCacheWarmer(cache, pregen, config)

	// Record mixed hits and misses
	for i := 0; i < 10; i++ {
		key := GenerateKey(int64(i%3), "state", 0)
		hit := i%2 == 0 // 5 hits, 5 misses
		warmer.RecordAccess(key, hit, int64(i))
	}

	stats := warmer.Stats()

	if stats.AccessLogSize != 10 {
		t.Errorf("AccessLogSize = %d, want 10", stats.AccessLogSize)
	}

	if stats.WindowHitRate != 0.5 {
		t.Errorf("WindowHitRate = %f, want 0.5", stats.WindowHitRate)
	}

	if stats.WindowMissRate != 0.5 {
		t.Errorf("WindowMissRate = %f, want 0.5", stats.WindowMissRate)
	}
}

func TestWarmerStats_EmptyWindow(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	stats := warmer.Stats()

	if stats.WindowHitRate != 0.0 {
		t.Errorf("WindowHitRate = %f, want 0.0 for empty window", stats.WindowHitRate)
	}

	if stats.WindowMissRate != 0.0 {
		t.Errorf("WindowMissRate = %f, want 0.0 for empty window", stats.WindowMissRate)
	}
}

func TestReset(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	// Add some data
	for i := 0; i < 10; i++ {
		key := GenerateKey(int64(i), "state", 0)
		warmer.RecordAccess(key, true, int64(i))
	}

	// Verify data exists
	stats := warmer.Stats()
	if stats.AccessLogSize == 0 {
		t.Fatal("Expected non-empty access log before reset")
	}

	// Reset
	warmer.Reset()

	// Verify data cleared
	stats = warmer.Stats()
	if stats.AccessLogSize != 0 {
		t.Errorf("AccessLogSize = %d after reset, want 0", stats.AccessLogSize)
	}

	if stats.PatternCount != 0 {
		t.Errorf("PatternCount = %d after reset, want 0", stats.PatternCount)
	}
}

func TestAnalyzeAnimationSequence(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	// Create animation frame keys
	frameKeys := []CacheKey{
		GenerateKey(1, "walk", 0),
		GenerateKey(1, "walk", 1),
		GenerateKey(1, "walk", 2),
		GenerateKey(1, "walk", 3),
	}

	warmer.AnalyzeAnimationSequence(frameKeys)

	// Check that patterns were registered
	warmer.mu.RLock()
	pattern := warmer.patterns[frameKeys[0]]
	warmer.mu.RUnlock()

	if pattern == nil {
		t.Fatal("pattern for frame 0 not found")
	}

	found := false
	for _, k := range pattern.NextKeys {
		if k == frameKeys[1] {
			found = true
			break
		}
	}

	if !found {
		t.Error("frame 1 not found in frame 0's NextKeys")
	}
}

func TestPredictAnimationFrames(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	// Create animation frame keys
	frameKeys := []CacheKey{
		GenerateKey(1, "walk", 0),
		GenerateKey(1, "walk", 1),
		GenerateKey(1, "walk", 2),
		GenerateKey(1, "walk", 3),
	}

	warmer.AnalyzeAnimationSequence(frameKeys)

	// Predict remaining frames from frame 0
	predicted := warmer.PredictAnimationFrames(frameKeys[0])

	if len(predicted) < 2 {
		t.Errorf("Expected at least 2 predicted frames, got %d", len(predicted))
	}

	// First predicted should be frame 1
	if len(predicted) > 0 && predicted[0] != frameKeys[1] {
		t.Errorf("First predicted = %s, want %s", predicted[0], frameKeys[1])
	}
}

func TestPredictAnimationFrames_NoCycle(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	// Create a looping animation
	key1 := GenerateKey(1, "loop", 0)
	key2 := GenerateKey(1, "loop", 1)

	// Manually create cyclic pattern
	warmer.mu.Lock()
	warmer.patterns[key1] = &AccessPattern{
		Key:      key1,
		NextKeys: []CacheKey{key2},
	}
	warmer.patterns[key2] = &AccessPattern{
		Key:      key2,
		NextKeys: []CacheKey{key1}, // Cycle back
	}
	warmer.mu.Unlock()

	// Should not infinite loop
	predicted := warmer.PredictAnimationFrames(key1)

	// Should stop when cycle detected (max 16 frames)
	if len(predicted) > 16 {
		t.Errorf("Predicted %d frames, expected max 16 due to cycle detection", len(predicted))
	}
}

func BenchmarkRecordAccess(b *testing.B) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := GenerateKey(int64(i%100), "state", i%8)
		warmer.RecordAccess(key, i%2 == 0, int64(i))
	}
}

func BenchmarkGetHotSprites(b *testing.B) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	// Pre-populate with patterns
	for i := 0; i < 100; i++ {
		key := GenerateKey(int64(i%10), "state", 0)
		warmer.RecordAccess(key, true, int64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = warmer.GetHotSprites()
	}
}

func BenchmarkPredictNext(b *testing.B) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	// Pre-populate with sequential patterns
	for i := 0; i < 100; i++ {
		key := GenerateKey(int64(i), "walk", i%8)
		warmer.RecordAccess(key, false, int64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = warmer.PredictNext()
	}
}

func BenchmarkAnalyzeAnimationSequence(b *testing.B) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)
	warmer := NewPredictiveCacheWarmer(cache, pregen, DefaultWarmerConfig())

	frameKeys := make([]CacheKey, 8)
	for i := 0; i < 8; i++ {
		frameKeys[i] = GenerateKey(1, "walk", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		warmer.AnalyzeAnimationSequence(frameKeys)
	}
}
