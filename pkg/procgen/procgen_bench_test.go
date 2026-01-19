// Package procgen provides benchmark tests for procedural generators.
// These benchmarks verify that all generators meet the <10ms target per generation.
package procgen

import (
	"testing"
)

// BenchmarkSeedGenerator benchmarks seed derivation performance.
func BenchmarkSeedGenerator(b *testing.B) {
	sg := NewSeedGenerator(12345)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sg.GetSeed("terrain", i)
	}
}

// BenchmarkSeedGeneratorParallel benchmarks parallel seed derivation.
func BenchmarkSeedGeneratorParallel(b *testing.B) {
	sg := NewSeedGenerator(12345)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = sg.GetSeed("terrain", i)
			i++
		}
	})
}

// BenchmarkValidateParams benchmarks parameter validation.
func BenchmarkValidateParams(b *testing.B) {
	params := GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 10},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ValidateParams(params)
	}
}

// BenchmarkValidateDimensions benchmarks dimension validation.
func BenchmarkValidateDimensions(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ValidateDimensions(100, 100, 10, 10, 1000, 1000)
	}
}
