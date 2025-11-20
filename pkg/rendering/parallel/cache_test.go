package parallel

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewThreadSafeCache(t *testing.T) {
	cache := NewThreadSafeCache()
	if cache == nil {
		t.Fatal("NewThreadSafeCache returned nil")
	}
	if cache.Size() != 0 {
		t.Errorf("New cache should be empty, got size %d", cache.Size())
	}
	if cache.HitRate() != 0.0 {
		t.Errorf("New cache should have 0%% hit rate, got %.2f%%", cache.HitRate())
	}
}

func TestCacheGetSet(t *testing.T) {
	cache := NewThreadSafeCache()

	// Set value
	cache.Set("key1", "value1")
	cache.Set("key2", 42)
	cache.Set("key3", struct{ X int }{X: 10})

	// Get existing values
	value, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}

	value, found = cache.Get("key2")
	if !found {
		t.Error("Expected to find key2")
	}
	if value != 42 {
		t.Errorf("Expected 42, got %v", value)
	}

	// Get non-existent value
	value, found = cache.Get("nonexistent")
	if found {
		t.Error("Should not find nonexistent key")
	}
	if value != nil {
		t.Errorf("Expected nil for missing key, got %v", value)
	}
}

func TestCacheDelete(t *testing.T) {
	cache := NewThreadSafeCache()

	cache.Set("key1", "value1")
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}

	cache.Delete("key1")
	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after delete, got %d", cache.Size())
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("Key should not be found after delete")
	}

	// Deleting non-existent key should be safe
	cache.Delete("nonexistent")
}

func TestCacheClear(t *testing.T) {
	cache := NewThreadSafeCache()

	// Add multiple entries
	for i := 0; i < 100; i++ {
		cache.Set(string(rune(i)), i)
	}

	if cache.Size() != 100 {
		t.Errorf("Expected size 100, got %d", cache.Size())
	}

	// Clear cache
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	stats := cache.GetStats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Stats should be reset after clear")
	}
}

func TestCacheHitRate(t *testing.T) {
	cache := NewThreadSafeCache()

	cache.Set("key1", "value1")

	// 5 hits
	for i := 0; i < 5; i++ {
		cache.Get("key1")
	}

	// 2 misses
	cache.Get("key2")
	cache.Get("key3")

	// Expected hit rate: 5 / (5 + 2) = 71.43%
	hitRate := cache.HitRate()
	expected := (5.0 / 7.0) * 100.0
	if hitRate < expected-0.1 || hitRate > expected+0.1 {
		t.Errorf("Expected hit rate ~%.2f%%, got %.2f%%", expected, hitRate)
	}

	stats := cache.GetStats()
	if stats.Hits != 5 {
		t.Errorf("Expected 5 hits, got %d", stats.Hits)
	}
	if stats.Misses != 2 {
		t.Errorf("Expected 2 misses, got %d", stats.Misses)
	}
}

func TestCacheGetOrCompute(t *testing.T) {
	cache := NewThreadSafeCache()

	// First call should compute
	computeCalls := 0
	compute := func() interface{} {
		computeCalls++
		return "computed-value"
	}

	value := cache.GetOrCompute("key1", compute)
	if value != "computed-value" {
		t.Errorf("Expected 'computed-value', got %v", value)
	}
	if computeCalls != 1 {
		t.Errorf("Expected 1 compute call, got %d", computeCalls)
	}

	// Second call should use cached value
	value = cache.GetOrCompute("key1", compute)
	if value != "computed-value" {
		t.Errorf("Expected cached 'computed-value', got %v", value)
	}
	if computeCalls != 1 {
		t.Errorf("Expected still 1 compute call (cached), got %d", computeCalls)
	}

	// Different key should compute again
	value = cache.GetOrCompute("key2", compute)
	if computeCalls != 2 {
		t.Errorf("Expected 2 compute calls for different key, got %d", computeCalls)
	}
}

func TestCacheKeys(t *testing.T) {
	cache := NewThreadSafeCache()

	// Empty cache
	keys := cache.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}

	// Add entries
	expectedKeys := map[string]bool{
		"key1": true,
		"key2": true,
		"key3": true,
	}
	for key := range expectedKeys {
		cache.Set(key, "value")
	}

	keys = cache.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Verify all expected keys are present
	for _, key := range keys {
		if !expectedKeys[key] {
			t.Errorf("Unexpected key: %s", key)
		}
	}
}

func TestCacheContainsKey(t *testing.T) {
	cache := NewThreadSafeCache()

	cache.Set("exists", "value")

	if !cache.ContainsKey("exists") {
		t.Error("Should contain 'exists' key")
	}

	if cache.ContainsKey("notexists") {
		t.Error("Should not contain 'notexists' key")
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	cache := NewThreadSafeCache()

	// Concurrent writers
	var wg sync.WaitGroup
	goroutineCount := 100
	itemsPerGoroutine := 100

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				key := string(rune(id*itemsPerGoroutine + j))
				cache.Set(key, id*itemsPerGoroutine+j)
			}
		}(i)
	}

	wg.Wait()

	expectedSize := goroutineCount * itemsPerGoroutine
	if cache.Size() != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, cache.Size())
	}
}

func TestCacheConcurrentReaders(t *testing.T) {
	cache := NewThreadSafeCache()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Set(string(rune(i)), i)
	}

	// Concurrent readers
	var wg sync.WaitGroup
	readerCount := 50
	readsPerReader := 1000

	var successReads int64

	for i := 0; i < readerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < readsPerReader; j++ {
				key := string(rune(j % 1000))
				if _, found := cache.Get(key); found {
					atomic.AddInt64(&successReads, 1)
				}
			}
		}()
	}

	wg.Wait()

	expectedReads := int64(readerCount * readsPerReader)
	if successReads != expectedReads {
		t.Errorf("Expected %d successful reads, got %d", expectedReads, successReads)
	}
}

func TestCacheConcurrentGetOrCompute(t *testing.T) {
	cache := NewThreadSafeCache()

	var computeCalls int64
	compute := func() interface{} {
		atomic.AddInt64(&computeCalls, 1)
		time.Sleep(1 * time.Millisecond) // Simulate expensive computation
		return "expensive-value"
	}

	// Multiple goroutines try to get the same key
	var wg sync.WaitGroup
	goroutineCount := 100

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value := cache.GetOrCompute("shared-key", compute)
			if value != "expensive-value" {
				t.Errorf("Expected 'expensive-value', got %v", value)
			}
		}()
	}

	wg.Wait()

	// Compute should be called at most a few times (due to race conditions)
	// but far fewer than goroutineCount
	if computeCalls > 10 {
		t.Errorf("Expected ~1-10 compute calls, got %d (too many redundant computations)", computeCalls)
	}
}

// Benchmark cache get performance
func BenchmarkCacheGet(b *testing.B) {
	cache := NewThreadSafeCache()
	cache.Set("key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

// Benchmark cache set performance
func BenchmarkCacheSet(b *testing.B) {
	cache := NewThreadSafeCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", i)
	}
}

// Benchmark concurrent reads
func BenchmarkCacheConcurrentReads(b *testing.B) {
	cache := NewThreadSafeCache()
	for i := 0; i < 1000; i++ {
		cache.Set(string(rune(i)), i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Get(string(rune(i % 1000)))
			i++
		}
	})
}

// Benchmark GetOrCompute
func BenchmarkCacheGetOrCompute(b *testing.B) {
	cache := NewThreadSafeCache()
	compute := func() interface{} {
		return "computed"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetOrCompute("key", compute)
	}
}
