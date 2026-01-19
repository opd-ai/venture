// Package magic provides benchmark tests for spell generation.
// Target: <10ms per spell batch generation.
package magic

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// BenchmarkSpellGenerator benchmarks spell batch generation (default 10 spells).
func BenchmarkSpellGenerator(b *testing.B) {
	gen := NewSpellGenerator()
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

// BenchmarkSpellGeneratorSingleSpell benchmarks single spell generation.
func BenchmarkSpellGeneratorSingleSpell(b *testing.B) {
	gen := NewSpellGenerator()
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

// BenchmarkSpellGenerator50Spells benchmarks large batch generation.
func BenchmarkSpellGenerator50Spells(b *testing.B) {
	gen := NewSpellGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 50},
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

// BenchmarkSpellGeneratorSciFi benchmarks sci-fi spell generation.
func BenchmarkSpellGeneratorSciFi(b *testing.B) {
	gen := NewSpellGenerator()
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

// BenchmarkSpellGeneratorHorror benchmarks horror spell generation.
func BenchmarkSpellGeneratorHorror(b *testing.B) {
	gen := NewSpellGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "horror",
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

// BenchmarkSpellGeneratorHighDifficulty benchmarks high difficulty generation.
func BenchmarkSpellGeneratorHighDifficulty(b *testing.B) {
	gen := NewSpellGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.95,
		Depth:      30,
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

// BenchmarkSpellValidation benchmarks spell validation.
func BenchmarkSpellValidation(b *testing.B) {
	gen := NewSpellGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 10},
	}

	// Pre-generate spells for validation benchmark
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

// BenchmarkSpellGeneratorParallel benchmarks parallel spell generation.
func BenchmarkSpellGeneratorParallel(b *testing.B) {
	gen := NewSpellGenerator()
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
