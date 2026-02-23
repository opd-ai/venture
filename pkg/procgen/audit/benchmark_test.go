package audit

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// BenchmarkHashOutput benchmarks the JSON marshal + SHA256 hash operation
// used for determinism validation during CI runs.
func BenchmarkHashOutput(b *testing.B) {
	// Generate a sample entity to hash
	gen := entity.NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}
	result, err := gen.Generate(12345, params)
	if err != nil {
		b.Fatalf("Failed to generate entity: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hashOutput(result)
		if err != nil {
			b.Fatalf("hashOutput failed: %v", err)
		}
	}
}

// BenchmarkHashOutput_Item benchmarks hash output for items (smaller payload).
func BenchmarkHashOutput_Item(b *testing.B) {
	gen := item.NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}
	result, err := gen.Generate(12345, params)
	if err != nil {
		b.Fatalf("Failed to generate item: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hashOutput(result)
		if err != nil {
			b.Fatalf("hashOutput failed: %v", err)
		}
	}
}

// BenchmarkCompareOutputs benchmarks the JSON marshal + bytes.Equal comparison
// used to verify generator determinism.
func BenchmarkCompareOutputs(b *testing.B) {
	gen := entity.NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}
	result1, err := gen.Generate(12345, params)
	if err != nil {
		b.Fatalf("Failed to generate entity 1: %v", err)
	}
	result2, err := gen.Generate(12345, params)
	if err != nil {
		b.Fatalf("Failed to generate entity 2: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compareOutputs(result1, result2)
		if err != nil {
			b.Fatalf("compareOutputs failed: %v", err)
		}
	}
}

// BenchmarkCompareOutputs_Different benchmarks comparison of different outputs.
func BenchmarkCompareOutputs_Different(b *testing.B) {
	gen := entity.NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}
	result1, err := gen.Generate(12345, params)
	if err != nil {
		b.Fatalf("Failed to generate entity 1: %v", err)
	}
	result2, err := gen.Generate(54321, params)
	if err != nil {
		b.Fatalf("Failed to generate entity 2: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compareOutputs(result1, result2)
		if err != nil {
			b.Fatalf("compareOutputs failed: %v", err)
		}
	}
}

// BenchmarkGetBaselinePrefix benchmarks the baseline prefix map lookup.
func BenchmarkGetBaselinePrefix(b *testing.B) {
	generators := []string{
		"EntityGenerator",
		"ItemGenerator",
		"MagicGenerator",
		"QuestGenerator",
		"TerrainGenerator",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, gen := range generators {
			_ = GetBaselinePrefix(gen)
		}
	}
}

// BenchmarkGetBaselinePrefix_Miss benchmarks map lookup for non-existent keys.
func BenchmarkGetBaselinePrefix_Miss(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetBaselinePrefix("NonExistentGenerator")
	}
}

// BenchmarkHashMatchesBaseline benchmarks the full baseline comparison operation.
func BenchmarkHashMatchesBaseline(b *testing.B) {
	gen := entity.NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}
	result, err := gen.Generate(BaselineSeed, params)
	if err != nil {
		b.Fatalf("Failed to generate entity: %v", err)
	}
	hash, err := hashOutput(result)
	if err != nil {
		b.Fatalf("Failed to hash output: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HashMatchesBaseline("EntityGenerator", hash)
	}
}
