package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewPreGenerator(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	if pregen == nil {
		t.Fatal("NewPreGenerator returned nil")
	}

	if pregen.cache != cache {
		t.Error("PreGenerator cache not set correctly")
	}

	if len(pregen.queue) != 0 {
		t.Errorf("Expected empty queue, got %d items", len(pregen.queue))
	}
}

func TestQueue(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Queue a request
	key := GenerateKey(12345, "idle", 0)
	generator := func() (*ebiten.Image, error) {
		return ebiten.NewImage(32, 32), nil
	}

	pregen.Queue(key, generator)

	if pregen.QueueSize() != 1 {
		t.Errorf("Expected queue size 1, got %d", pregen.QueueSize())
	}

	stats := pregen.Stats()
	if stats.RequestsQueued != 1 {
		t.Errorf("RequestsQueued: got %d, want 1", stats.RequestsQueued)
	}
}

func TestQueue_AlreadyCached(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Pre-populate cache
	key := GenerateKey(12345, "idle", 0)
	img := ebiten.NewImage(32, 32)
	cache.Put(key, img)

	// Queue same sprite
	generator := func() (*ebiten.Image, error) {
		return ebiten.NewImage(32, 32), nil
	}

	pregen.Queue(key, generator)

	// Should not be queued (already in cache)
	if pregen.QueueSize() != 0 {
		t.Errorf("Expected queue size 0 (already cached), got %d", pregen.QueueSize())
	}

	stats := pregen.Stats()
	if stats.CacheHits != 1 {
		t.Errorf("CacheHits: got %d, want 1", stats.CacheHits)
	}
}

func TestQueueBatch(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Create batch requests
	requests := make([]PreGenRequest, 10)
	for i := 0; i < 10; i++ {
		requests[i] = PreGenRequest{
			Key: GenerateKey(int64(i), "idle", 0),
			Generator: func() (*ebiten.Image, error) {
				return ebiten.NewImage(32, 32), nil
			},
		}
	}

	pregen.QueueBatch(requests)

	if pregen.QueueSize() != 10 {
		t.Errorf("Expected queue size 10, got %d", pregen.QueueSize())
	}

	stats := pregen.Stats()
	if stats.RequestsQueued != 10 {
		t.Errorf("RequestsQueued: got %d, want 10", stats.RequestsQueued)
	}
}

func TestQueueBatch_PartialCached(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Pre-populate cache with some sprites
	cache.Put(GenerateKey(0, "idle", 0), ebiten.NewImage(32, 32))
	cache.Put(GenerateKey(1, "idle", 0), ebiten.NewImage(32, 32))

	// Create batch with 5 sprites (2 already cached)
	requests := make([]PreGenRequest, 5)
	for i := 0; i < 5; i++ {
		requests[i] = PreGenRequest{
			Key: GenerateKey(int64(i), "idle", 0),
			Generator: func() (*ebiten.Image, error) {
				return ebiten.NewImage(32, 32), nil
			},
		}
	}

	pregen.QueueBatch(requests)

	// Should only queue 3 (5 - 2 cached)
	if pregen.QueueSize() != 3 {
		t.Errorf("Expected queue size 3, got %d", pregen.QueueSize())
	}

	stats := pregen.Stats()
	if stats.CacheHits != 2 {
		t.Errorf("CacheHits: got %d, want 2", stats.CacheHits)
	}
}

func TestGenerate(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Queue some requests
	for i := 0; i < 5; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		generator := func() (*ebiten.Image, error) {
			return ebiten.NewImage(32, 32), nil
		}
		pregen.Queue(key, generator)
	}

	// Generate
	count := pregen.Generate()

	if count != 5 {
		t.Errorf("Generate() = %d, want 5", count)
	}

	// Queue should be empty
	if pregen.QueueSize() != 0 {
		t.Errorf("Expected empty queue after generate, got %d", pregen.QueueSize())
	}

	// Check stats
	stats := pregen.Stats()
	if stats.RequestsComplete != 5 {
		t.Errorf("RequestsComplete: got %d, want 5", stats.RequestsComplete)
	}

	// Check cache
	for i := 0; i < 5; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		if !cache.Contains(key) {
			t.Errorf("Expected sprite %d in cache", i)
		}
	}
}

func TestGenerate_WithFailures(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Queue requests with some failures
	for i := 0; i < 5; i++ {
		key := GenerateKey(int64(i), "idle", 0)

		var generator GeneratorFunc
		if i%2 == 0 {
			// Success
			generator = func() (*ebiten.Image, error) {
				return ebiten.NewImage(32, 32), nil
			}
		} else {
			// Failure
			generator = func() (*ebiten.Image, error) {
				return nil, errors.New("generation failed")
			}
		}

		pregen.Queue(key, generator)
	}

	// Generate
	count := pregen.Generate()

	// Should generate 3 (indices 0, 2, 4)
	if count != 3 {
		t.Errorf("Generate() = %d, want 3", count)
	}

	stats := pregen.Stats()
	if stats.RequestsComplete != 3 {
		t.Errorf("RequestsComplete: got %d, want 3", stats.RequestsComplete)
	}

	if stats.RequestsFailed != 2 {
		t.Errorf("RequestsFailed: got %d, want 2", stats.RequestsFailed)
	}
}

func TestGenerateAsync(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Queue requests
	for i := 0; i < 5; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		generator := func() (*ebiten.Image, error) {
			return ebiten.NewImage(32, 32), nil
		}
		pregen.Queue(key, generator)
	}

	// Generate async
	doneCh := make(chan int, 1)
	pregen.GenerateAsync(context.Background(), doneCh)

	// Wait for completion
	count := <-doneCh

	if count != 5 {
		t.Errorf("GenerateAsync() = %d, want 5", count)
	}
}

func TestClear(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Queue requests
	for i := 0; i < 5; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		generator := func() (*ebiten.Image, error) {
			return ebiten.NewImage(32, 32), nil
		}
		pregen.Queue(key, generator)
	}

	// Clear
	pregen.Clear()

	if pregen.QueueSize() != 0 {
		t.Errorf("Expected empty queue after clear, got %d", pregen.QueueSize())
	}
}

func TestHitRate(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Pre-populate cache with 3 sprites
	for i := 0; i < 3; i++ {
		cache.Put(GenerateKey(int64(i), "idle", 0), ebiten.NewImage(32, 32))
	}

	// Queue 10 requests (3 already cached)
	for i := 0; i < 10; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		generator := func() (*ebiten.Image, error) {
			return ebiten.NewImage(32, 32), nil
		}
		pregen.Queue(key, generator)
	}

	// Hit rate should be 3/7 (3 cached out of 7 new requests queued)
	hitRate := pregen.HitRate()
	expected := 3.0 / 7.0

	if hitRate != expected {
		t.Errorf("HitRate() = %f, want %f", hitRate, expected)
	}
}

func TestHitRate_Empty(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	hitRate := pregen.HitRate()
	if hitRate != 0.0 {
		t.Errorf("Expected 0.0 hit rate with no requests, got %f", hitRate)
	}
}

func TestSuccessRate(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Manually set stats
	pregen.mu.Lock()
	pregen.stats.RequestsComplete = 8
	pregen.stats.RequestsFailed = 2
	pregen.mu.Unlock()

	// Success rate should be 8/10 = 0.8
	successRate := pregen.SuccessRate()
	expected := 0.8

	if successRate != expected {
		t.Errorf("SuccessRate() = %f, want %f", successRate, expected)
	}
}

func TestSuccessRate_NoAttempts(t *testing.T) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	successRate := pregen.SuccessRate()
	if successRate != 1.0 {
		t.Errorf("Expected 1.0 success rate with no attempts, got %f", successRate)
	}
}

func BenchmarkQueue(b *testing.B) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	generator := func() (*ebiten.Image, error) {
		return ebiten.NewImage(32, 32), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		pregen.Queue(key, generator)
	}
}

func BenchmarkGenerate(b *testing.B) {
	cache := NewSpriteCache(100 * 1024 * 1024)
	pregen := NewPreGenerator(cache)

	// Pre-queue requests
	generator := func() (*ebiten.Image, error) {
		return ebiten.NewImage(32, 32), nil
	}

	for i := 0; i < b.N; i++ {
		key := GenerateKey(int64(i), "idle", 0)
		pregen.Queue(key, generator)
	}

	b.ResetTimer()
	pregen.Generate()
}
