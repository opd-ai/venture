// Package quest provides benchmark tests for quest generation.
// Target: <10ms per quest batch generation.
package quest

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// BenchmarkQuestGenerator benchmarks quest batch generation (default 5 quests).
func BenchmarkQuestGenerator(b *testing.B) {
	gen := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 5},
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

// BenchmarkQuestGeneratorSingleQuest benchmarks single quest generation.
func BenchmarkQuestGeneratorSingleQuest(b *testing.B) {
	gen := NewQuestGenerator()
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

// BenchmarkQuestGenerator20Quests benchmarks large batch generation.
func BenchmarkQuestGenerator20Quests(b *testing.B) {
	gen := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 20},
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

// BenchmarkQuestGeneratorSciFi benchmarks sci-fi quest generation.
func BenchmarkQuestGeneratorSciFi(b *testing.B) {
	gen := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "scifi",
		Custom:     map[string]interface{}{"count": 5},
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

// BenchmarkQuestGeneratorHighDifficulty benchmarks high difficulty generation.
func BenchmarkQuestGeneratorHighDifficulty(b *testing.B) {
	gen := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.95,
		Depth:      50,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 5},
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

// BenchmarkQuestValidate benchmarks quest validation.
func BenchmarkQuestValidate(b *testing.B) {
	gen := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 5},
	}

	// Pre-generate quests for validation benchmark
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

// BenchmarkQuestGeneratorParallel benchmarks parallel quest generation.
func BenchmarkQuestGeneratorParallel(b *testing.B) {
	gen := NewQuestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 5},
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
