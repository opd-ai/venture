package sprites

import (
	"testing"
)

// BenchmarkCache_Get tests cache lookup performance.
func BenchmarkCache_Get(b *testing.B) {
	cache := NewCache(100)
	config := Config{
		Type:   0,
		Seed:   12345,
		Width:  32,
		Height: 32,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Get(config)
	}
}

// BenchmarkCache_HashConfig tests hash generation performance.
func BenchmarkCache_HashConfig(b *testing.B) {
	cache := NewCache(100)
	config := Config{
		Type:       0,
		Seed:       12345,
		Width:      32,
		Height:     32,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Variation:  1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.hashConfig(config)
	}
}

// BenchmarkCache_HashConfigWithCustom tests hash generation with custom params.
func BenchmarkCache_HashConfigWithCustom(b *testing.B) {
	cache := NewCache(100)
	config := Config{
		Type:    0,
		Seed:    12345,
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"equipped":  true,
			"direction": 2,
			"level":     5,
			"rarity":    "epic",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.hashConfig(config)
	}
}

// BenchmarkCache_Stats tests stats gathering performance.
func BenchmarkCache_Stats(b *testing.B) {
	cache := NewCache(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Stats()
	}
}

// BenchmarkCache_Concurrent tests concurrent access performance.
func BenchmarkCache_Concurrent(b *testing.B) {
	cache := NewCache(100)
	config := Config{
		Type:   0,
		Seed:   12345,
		Width:  32,
		Height: 32,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cache.Get(config)
		}
	})
}

// BenchmarkCachedGenerator_Generate benchmarks sprite generation without cache hits.
// This represents worst-case performance (all cache misses).
func BenchmarkCachedGenerator_Generate_CacheMiss(b *testing.B) {
	cg := NewCachedGenerator(100)
	cg.SetCacheEnabled(false) // Disable cache to always generate

	config := Config{
		Type:    0,
		Seed:    12345,
		Width:   32,
		Height:  32,
		GenreID: "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Seed = int64(i) // Different seed each time
		_, _ = cg.Generate(config)
	}
}

// BenchmarkHashConfig_FieldVariations benchmarks hash sensitivity to field changes.
func BenchmarkHashConfig_FieldVariations(b *testing.B) {
	cache := NewCache(100)

	baseConfig := Config{
		Type:       0,
		Width:      32,
		Height:     32,
		Seed:       12345,
		GenreID:    "fantasy",
		Complexity: 0.5,
		Variation:  1,
	}

	b.Run("SameSeed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = cache.hashConfig(baseConfig)
		}
	})

	b.Run("DifferentSeeds", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config := baseConfig
			config.Seed = int64(i)
			_ = cache.hashConfig(config)
		}
	})

	b.Run("DifferentGenres", func(b *testing.B) {
		genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apoc"}
		for i := 0; i < b.N; i++ {
			config := baseConfig
			config.GenreID = genres[i%len(genres)]
			_ = cache.hashConfig(config)
		}
	})
}

// BenchmarkCache_Eviction benchmarks LRU eviction performance.
func BenchmarkCache_Eviction(b *testing.B) {
	cache := NewCache(50) // Small capacity to force evictions

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := Config{
			Type:   0,
			Seed:   int64(i),
			Width:  32,
			Height: 32,
		}
		// This will trigger evictions once cache is full
		_ = cache.Get(config)
	}
}

// BenchmarkCacheStats_HitRate benchmarks hit rate calculation.
func BenchmarkCacheStats_HitRate(b *testing.B) {
	cache := NewCache(100)
	cache.hits = 1000
	cache.misses = 500

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Stats()
	}
}
