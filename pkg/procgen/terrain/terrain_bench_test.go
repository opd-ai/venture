// Package terrain provides benchmark tests for terrain generation.
// Target: <10ms for small/medium terrains, <50ms for large composite terrains.
package terrain

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

// BenchmarkBSPGen benchmarks BSP dungeon generation.
func BenchmarkBSPGen(b *testing.B) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
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

// BenchmarkBSPGenLarge benchmarks large BSP generation.
func BenchmarkBSPGenLarge(b *testing.B) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  200,
			"height": 200,
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

// BenchmarkCellularGen benchmarks cellular automata cave generation.
func BenchmarkCellularGen(b *testing.B) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
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

// BenchmarkCellularGenLarge benchmarks large cellular cave generation.
func BenchmarkCellularGenLarge(b *testing.B) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  200,
			"height": 200,
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

// BenchmarkMazeGen benchmarks maze generation.
func BenchmarkMazeGen(b *testing.B) {
	gen := NewMazeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
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

// BenchmarkMazeGenLarge benchmarks large maze generation.
func BenchmarkMazeGenLarge(b *testing.B) {
	gen := NewMazeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  200,
			"height": 200,
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

// BenchmarkForestGen benchmarks forest terrain generation.
func BenchmarkForestGen(b *testing.B) {
	gen := NewForestGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
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

// BenchmarkCityGen benchmarks city terrain generation.
func BenchmarkCityGen(b *testing.B) {
	gen := NewCityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  100,
			"height": 100,
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

// BenchmarkCompositeGen benchmarks composite multi-biome generation.
func BenchmarkCompositeGen(b *testing.B) {
	gen := NewCompositeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":      80,
			"height":     50,
			"biomeCount": 3,
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

// BenchmarkCompositeGenLarge benchmarks large composite generation.
func BenchmarkCompositeGenLarge(b *testing.B) {
	gen := NewCompositeGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":      200,
			"height":     200,
			"biomeCount": 4,
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

// BenchmarkBSPValidate benchmarks BSP terrain validation.
func BenchmarkBSPValidate(b *testing.B) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	// Pre-generate terrain for validation benchmark
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

// BenchmarkBSPGenParallel benchmarks parallel BSP generation.
func BenchmarkBSPGenParallel(b *testing.B) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
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
