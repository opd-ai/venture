// Package item provides benchmark tests for item generation.
// Target: <10ms per item batch generation.
package item

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// BenchmarkItemGenerator benchmarks item batch generation (default 10 items).
func BenchmarkItemGenerator(b *testing.B) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 10},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i)*1000, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkItemGeneratorSingleItem benchmarks single item generation.
func BenchmarkItemGeneratorSingleItem(b *testing.B) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 1},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i)*1000, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkItemGenerator100Items benchmarks large batch generation.
func BenchmarkItemGenerator100Items(b *testing.B) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 100},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i)*1000, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkItemGeneratorSciFi benchmarks sci-fi genre generation.
func BenchmarkItemGeneratorSciFi(b *testing.B) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "scifi",
		Custom:     map[string]interface{}{"count": 10},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i)*1000, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkItemGeneratorWeaponsOnly benchmarks weapon-only generation.
func BenchmarkItemGeneratorWeaponsOnly(b *testing.B) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"count": 10,
			"type":  "weapon",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i)*1000, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkItemGeneratorHighDepth benchmarks generation at high dungeon depth.
func BenchmarkItemGeneratorHighDepth(b *testing.B) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      50,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 10},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i)*1000, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkItemValidation benchmarks item validation.
func BenchmarkItemValidation(b *testing.B) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 10},
	}

	// Pre-generate items for validation benchmark
	result, err := gen.Generate(12345, params)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := gen.Validate(result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkItemGeneratorParallel benchmarks parallel item generation.
func BenchmarkItemGeneratorParallel(b *testing.B) {
	gen := NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 10},
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		seed := int64(0)
		for pb.Next() {
			_, err := gen.Generate(seed, params)
			if err != nil {
				b.Fatal(err)
			}
			seed++
		}
	})
}
