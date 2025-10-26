//go:build !test
// +build !test

package sprites

import (
	"sync"
	"testing"
)

func TestNewCache(t *testing.T) {
	tests := []struct {
		name             string
		capacity         int
		expectedCapacity int
	}{
		{"positive capacity", 50, 50},
		{"zero capacity uses default", 0, 100},
		{"negative capacity uses default", -10, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewCache(tt.capacity)
			if cache.Capacity() != tt.expectedCapacity {
				t.Errorf("Capacity() = %d, want %d", cache.Capacity(), tt.expectedCapacity)
			}
			if cache.Size() != 0 {
				t.Errorf("Size() = %d, want 0", cache.Size())
			}
		})
	}
}

func TestCache_Stats(t *testing.T) {
	cache := NewCache(10)

	stats := cache.Stats()
	if stats.Hits != 0 {
		t.Errorf("Initial hits = %d, want 0", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("Initial misses = %d, want 0", stats.Misses)
	}
	if stats.HitRate != 0.0 {
		t.Errorf("Initial hit rate = %f, want 0.0", stats.HitRate)
	}
	if stats.Size != 0 {
		t.Errorf("Initial size = %d, want 0", stats.Size)
	}
	if stats.Capacity != 10 {
		t.Errorf("Capacity = %d, want 10", stats.Capacity)
	}
}

func TestCache_HashConfig(t *testing.T) {
	cache := NewCache(10)

	config1 := Config{
		Type:       0, // TypeHumanoid
		Width:      32,
		Height:     32,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Variation:  1,
	}

	config2 := Config{
		Type:       0, // TypeHumanoid
		Width:      32,
		Height:     32,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Variation:  1,
	}

	config3 := Config{
		Type:       0, // TypeHumanoid
		Width:      32,
		Height:     32,
		Seed:       54321, // Different seed
		GenreID:    "fantasy",
		Complexity: 0.5,
		Variation:  1,
	}

	hash1 := cache.hashConfig(config1)
	hash2 := cache.hashConfig(config2)
	hash3 := cache.hashConfig(config3)

	// Same config should produce same hash
	if hash1 != hash2 {
		t.Errorf("Identical configs produced different hashes: %d vs %d", hash1, hash2)
	}

	// Different config should produce different hash
	if hash1 == hash3 {
		t.Errorf("Different configs produced same hash: %d", hash1)
	}
}

func TestCache_HashConfigWithCustom(t *testing.T) {
	cache := NewCache(10)

	config1 := Config{
		Type:    0, // TypeHumanoid
		Seed:    12345,
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"equipped":  true,
			"direction": 0,
		},
	}

	config2 := Config{
		Type:    0, // TypeHumanoid
		Seed:    12345,
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"equipped":  false, // Different value
			"direction": 0,
		},
	}

	hash1 := cache.hashConfig(config1)
	hash2 := cache.hashConfig(config2)

	// Different custom params should produce different hashes
	if hash1 == hash2 {
		t.Errorf("Different custom params produced same hash: %d", hash1)
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(10)

	// Simulate some activity
	cache.misses = 5
	cache.hits = 3

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Size after Clear() = %d, want 0", cache.Size())
	}

	stats := cache.Stats()
	if stats.Hits != 0 {
		t.Errorf("Hits after Clear() = %d, want 0", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("Misses after Clear() = %d, want 0", stats.Misses)
	}
}

func TestCache_SetCapacity(t *testing.T) {
	cache := NewCache(10)

	// Test increasing capacity
	cache.SetCapacity(20)
	if cache.Capacity() != 20 {
		t.Errorf("Capacity after SetCapacity(20) = %d, want 20", cache.Capacity())
	}

	// Test decreasing capacity
	cache.SetCapacity(5)
	if cache.Capacity() != 5 {
		t.Errorf("Capacity after SetCapacity(5) = %d, want 5", cache.Capacity())
	}

	// Test invalid capacity (should be ignored)
	cache.SetCapacity(0)
	if cache.Capacity() != 5 {
		t.Errorf("Capacity after SetCapacity(0) = %d, want 5 (unchanged)", cache.Capacity())
	}

	cache.SetCapacity(-10)
	if cache.Capacity() != 5 {
		t.Errorf("Capacity after SetCapacity(-10) = %d, want 5 (unchanged)", cache.Capacity())
	}
}

func TestCacheStats_HitRate(t *testing.T) {
	tests := []struct {
		name     string
		hits     uint64
		misses   uint64
		expected float64
	}{
		{"no hits or misses", 0, 0, 0.0},
		{"only hits", 10, 0, 1.0},
		{"only misses", 0, 10, 0.0},
		{"50% hit rate", 5, 5, 0.5},
		{"75% hit rate", 15, 5, 0.75},
		{"25% hit rate", 5, 15, 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewCache(10)
			cache.hits = tt.hits
			cache.misses = tt.misses

			stats := cache.Stats()
			if stats.HitRate != tt.expected {
				t.Errorf("HitRate = %f, want %f", stats.HitRate, tt.expected)
			}
		})
	}
}

func TestCachedGenerator_New(t *testing.T) {
	cg := NewCachedGenerator(50)

	if cg == nil {
		t.Fatal("NewCachedGenerator returned nil")
	}

	if cg.generator == nil {
		t.Error("generator is nil")
	}

	if cg.cache == nil {
		t.Error("cache is nil")
	}

	if cg.cache.Capacity() != 50 {
		t.Errorf("cache capacity = %d, want 50", cg.cache.Capacity())
	}

	if !cg.enabled {
		t.Error("cache should be enabled by default")
	}
}

func TestCachedGenerator_SetCacheEnabled(t *testing.T) {
	cg := NewCachedGenerator(10)

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

func TestCachedGenerator_Cache(t *testing.T) {
	cg := NewCachedGenerator(10)

	cache := cg.Cache()
	if cache == nil {
		t.Error("Cache() returned nil")
	}

	if cache.Capacity() != 10 {
		t.Errorf("Cache capacity = %d, want 10", cache.Capacity())
	}
}

func TestCachedGenerator_ClearCache(t *testing.T) {
	cg := NewCachedGenerator(10)

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

func TestCachedGenerator_Stats(t *testing.T) {
	cg := NewCachedGenerator(10)

	stats := cg.Stats()

	if stats.Capacity != 10 {
		t.Errorf("Stats.Capacity = %d, want 10", stats.Capacity)
	}

	if stats.Size != 0 {
		t.Errorf("Stats.Size = %d, want 0", stats.Size)
	}
}

func TestBatchConfig_Struct(t *testing.T) {
	// Test that BatchConfig struct fields are accessible
	bc := BatchConfig{
		Configs:    []Config{{Type: 0, Seed: 123}}, // TypeHumanoid
		Concurrent: true,
		MaxWorkers: 4,
		OnProgress: func(completed, total int) {},
		OnError:    func(index int, err error) {},
	}

	if len(bc.Configs) != 1 {
		t.Errorf("Configs length = %d, want 1", len(bc.Configs))
	}

	if !bc.Concurrent {
		t.Error("Concurrent should be true")
	}

	if bc.MaxWorkers != 4 {
		t.Errorf("MaxWorkers = %d, want 4", bc.MaxWorkers)
	}

	if bc.OnProgress == nil {
		t.Error("OnProgress should not be nil")
	}

	if bc.OnError == nil {
		t.Error("OnError should not be nil")
	}
}

func TestBatchConfig_EmptyConfigs(t *testing.T) {
	bc := BatchConfig{
		Configs: []Config{},
	}

	if len(bc.Configs) != 0 {
		t.Errorf("Empty Configs length = %d, want 0", len(bc.Configs))
	}
}

func TestCache_Concurrency(t *testing.T) {
	cache := NewCache(100)

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	// Concurrent reads and writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				config := Config{
					Type:   0, // TypeHumanoid
					Seed:   int64(id*1000 + j),
					Width:  32,
					Height: 32,
				}

				// Simulate Get operation
				_ = cache.Get(config)

				// Check stats (read operation)
				_ = cache.Stats()

				// Check size (read operation)
				_ = cache.Size()
			}
		}(i)
	}

	wg.Wait()

	// Verify cache is in valid state
	stats := cache.Stats()
	if stats.Misses != uint64(numGoroutines*operationsPerGoroutine) {
		t.Errorf("Expected %d misses, got %d", numGoroutines*operationsPerGoroutine, stats.Misses)
	}

	if stats.Hits != 0 {
		t.Errorf("Expected 0 hits, got %d", stats.Hits)
	}
}

func TestCacheEntry_Structure(t *testing.T) {
	// Test that cacheEntry struct fields are accessible in package
	entry := &cacheEntry{
		key:     12345,
		sprite:  nil,
		element: nil,
	}

	if entry.key != 12345 {
		t.Errorf("key = %d, want 12345", entry.key)
	}

	if entry.sprite != nil {
		t.Error("sprite should be nil")
	}

	if entry.element != nil {
		t.Error("element should be nil")
	}
}

func TestCacheStats_Fields(t *testing.T) {
	stats := CacheStats{
		Hits:     100,
		Misses:   50,
		Size:     75,
		Capacity: 100,
		HitRate:  0.667,
	}

	if stats.Hits != 100 {
		t.Errorf("Hits = %d, want 100", stats.Hits)
	}

	if stats.Misses != 50 {
		t.Errorf("Misses = %d, want 50", stats.Misses)
	}

	if stats.Size != 75 {
		t.Errorf("Size = %d, want 75", stats.Size)
	}

	if stats.Capacity != 100 {
		t.Errorf("Capacity = %d, want 100", stats.Capacity)
	}

	if stats.HitRate < 0.666 || stats.HitRate > 0.668 {
		t.Errorf("HitRate = %f, want ~0.667", stats.HitRate)
	}
}

func TestHashConfig_Deterministic(t *testing.T) {
	cache := NewCache(10)

	config := Config{
		Type:       1, // TypeMonster
		Width:      64,
		Height:     64,
		Seed:       99999,
		GenreID:    "sci-fi",
		Complexity: 0.75,
		Variation:  3,
	}

	// Hash same config multiple times
	hashes := make([]uint64, 10)
	for i := 0; i < 10; i++ {
		hashes[i] = cache.hashConfig(config)
	}

	// All hashes should be identical
	firstHash := hashes[0]
	for i, hash := range hashes {
		if hash != firstHash {
			t.Errorf("Hash %d = %d, want %d (non-deterministic)", i, hash, firstHash)
		}
	}
}

func TestHashConfig_FieldSensitivity(t *testing.T) {
	cache := NewCache(10)

	baseConfig := Config{
		Type:       0, // TypeHumanoid
		Width:      32,
		Height:     32,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Variation:  1,
	}

	baseHash := cache.hashConfig(baseConfig)

	tests := []struct {
		name   string
		modify func(*Config)
	}{
		{"different Type", func(c *Config) { c.Type = 1 }}, // TypeMonster
		{"different Width", func(c *Config) { c.Width = 64 }},
		{"different Height", func(c *Config) { c.Height = 64 }},
		{"different Seed", func(c *Config) { c.Seed = 54321 }},
		{"different GenreID", func(c *Config) { c.GenreID = "sci-fi" }},
		{"different Complexity", func(c *Config) { c.Complexity = 0.75 }},
		{"different Variation", func(c *Config) { c.Variation = 2 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modified := baseConfig
			tt.modify(&modified)

			modifiedHash := cache.hashConfig(modified)
			if modifiedHash == baseHash {
				t.Errorf("Modified config produced same hash as base: %d", baseHash)
			}
		})
	}
}
