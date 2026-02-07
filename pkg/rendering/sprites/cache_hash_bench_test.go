package sprites

import (
	"testing"
)

// BenchmarkHashConfig_NoCustom benchmarks hash generation without Custom params.
func BenchmarkHashConfig_NoCustom(b *testing.B) {
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
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = cache.hashConfig(config)
	}
}

// BenchmarkHashConfig_WithCustomSmall benchmarks hash with 4 Custom params.
func BenchmarkHashConfig_WithCustomSmall(b *testing.B) {
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
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = cache.hashConfig(config)
	}
}

// BenchmarkHashConfig_WithCustomLarge benchmarks hash with 10 Custom params.
func BenchmarkHashConfig_WithCustomLarge(b *testing.B) {
	cache := NewCache(100)
	config := Config{
		Type:    0,
		Seed:    12345,
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"equipped":   true,
			"direction":  2,
			"level":      5,
			"rarity":     "epic",
			"armor":      "plate",
			"weapon":     "sword",
			"shield":     true,
			"health":     100,
			"mana":       50,
			"experience": int64(1234567),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = cache.hashConfig(config)
	}
}

// BenchmarkHashConfig_CustomMixedTypes benchmarks hash with various value types.
func BenchmarkHashConfig_CustomMixedTypes(b *testing.B) {
	cache := NewCache(100)
	config := Config{
		Type:    0,
		Seed:    12345,
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"name":       "PlayerCharacter",
			"equipped":   true,
			"direction":  2,
			"level":      5,
			"health":     100.5,
			"critChance": float32(0.25),
			"xp":         int64(999999),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = cache.hashConfig(config)
	}
}
