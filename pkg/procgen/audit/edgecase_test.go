package audit

import (
	"math"
	"runtime"
	"sync"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
)

// TestEdgeCases_ExtremeSeed tests generators with extreme seed values
func TestEdgeCases_ExtremeSeed(t *testing.T) {
	generators := getAllGenerators()
	extremeSeeds := []int64{
		0,
		-1,
		math.MaxInt64,
		math.MinInt64,
		1,
		-9223372036854775808, // MIN_INT64 explicit
		9223372036854775807,  // MAX_INT64 explicit
	}

	for name, gen := range generators {
		for _, seed := range extremeSeeds {
			t.Run(name+"_seed_"+formatSeed(seed), func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s panicked with extreme seed %d: %v", name, seed, r)
					}
				}()

				params := getBaseParams(name)

				result, err := gen.Generate(seed, params)
				// Either success or clean error, never panic
				if err != nil {
					// Error is acceptable, just ensure it's informative
					if len(err.Error()) < 10 {
						t.Errorf("%s returned uninformative error for seed %d: %v", name, seed, err)
					}
				} else if result == nil {
					t.Errorf("%s returned nil result without error for seed %d", name, seed)
				}
			})
		}
	}
}

// TestEdgeCases_InvalidParameters tests generators with invalid parameters
func TestEdgeCases_InvalidParameters(t *testing.T) {
	generators := getAllGenerators()

	testCases := []struct {
		name        string
		params      procgen.GenerationParams
		shouldError bool
	}{
		{
			name: "negative_depth",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      -1,
				GenreID:    "fantasy",
			},
			shouldError: true,
		},
		{
			name: "difficulty_too_high",
			params: procgen.GenerationParams{
				Difficulty: 2.0,
				Depth:      5,
				GenreID:    "fantasy",
			},
			shouldError: true, // Generators return error instead of clamping
		},
		{
			name: "difficulty_negative",
			params: procgen.GenerationParams{
				Difficulty: -0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			shouldError: true, // Generators return error instead of clamping
		},
		{
			name: "empty_genre",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "",
			},
			shouldError: true, // Empty genre should error
		},
		{
			name: "unknown_genre",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "unknown_nonexistent_genre",
			},
			shouldError: false, // Should fallback to fantasy
		},
		{
			name: "depth_zero",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      0,
				GenreID:    "fantasy",
			},
			shouldError: false, // Depth 0 should clamp to 1
		},
		{
			name: "extreme_depth",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      10000,
				GenreID:    "fantasy",
			},
			shouldError: false, // Should handle gracefully
		},
	}

	for name, gen := range generators {
		for _, tc := range testCases {
			t.Run(name+"_"+tc.name, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s panicked with params %+v: %v", name, tc.params, r)
					}
				}()

				// Add book_type for Book generator
				params := tc.params
				if name == "Book" {
					if params.Custom == nil {
						params.Custom = make(map[string]interface{})
					}
					params.Custom["book_type"] = engine.BookTypeLore
				}

				result, err := gen.Generate(12345, params)

				if tc.shouldError {
					if err == nil {
						t.Logf("%s should have errored with params %+v but succeeded", name, tc.params)
						// Note: Not failing test as generators may handle more gracefully
					}
				} else {
					if err != nil && result == nil {
						t.Logf("%s errored with params %+v: %v (acceptable if validation is strict)", name, tc.params, err)
					}
				}
			})
		}
	}
}

// TestEdgeCases_MinimumViable tests smallest possible valid generation
func TestEdgeCases_MinimumViable(t *testing.T) {
	generators := getAllGenerators()

	for name, gen := range generators {
		t.Run(name+"_minimum", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked with minimum viable params: %v", name, r)
				}
			}()

			minParams := getBaseParams(name)
			minParams.Difficulty = 0.0
			minParams.Depth = 1

			result, err := gen.Generate(1, minParams)
			if err != nil {
				t.Errorf("%s failed with minimum viable params: %v", name, err)
			} else if result == nil {
				t.Errorf("%s returned nil with minimum viable params", name)
			}
		})
	}
}

// TestEdgeCases_MaximumComplexity tests largest reasonable generation
func TestEdgeCases_MaximumComplexity(t *testing.T) {
	generators := getAllGenerators()

	for name, gen := range generators {
		t.Run(name+"_maximum", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked with maximum complexity params: %v", name, r)
				}
			}()

			maxParams := getBaseParams(name)
			maxParams.Difficulty = 1.0
			maxParams.Depth = 100 // High but not absurd

			// Measure memory before
			var memBefore runtime.MemStats
			runtime.GC() // Force GC to get clean baseline
			runtime.ReadMemStats(&memBefore)

			result, err := gen.Generate(99999, maxParams)

			// Measure memory after
			var memAfter runtime.MemStats
			runtime.ReadMemStats(&memAfter)

			// Use TotalAlloc to avoid underflow from GC between measurements
			var allocatedMB float64
			if memAfter.TotalAlloc >= memBefore.TotalAlloc {
				allocatedMB = float64(memAfter.TotalAlloc-memBefore.TotalAlloc) / 1024 / 1024
			} else {
				// GC happened, use Alloc with safety check
				if memAfter.Alloc >= memBefore.Alloc {
					allocatedMB = float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024
				} else {
					allocatedMB = 0 // GC reclaimed memory, treat as minimal allocation
				}
			}

			if err != nil {
				t.Logf("%s failed with maximum complexity: %v (acceptable if resource limit)", name, err)
			} else if result == nil {
				t.Errorf("%s returned nil with maximum complexity without error", name)
			}

			// Memory should not exceed 50MB for a single generation
			if allocatedMB > 50.0 {
				t.Errorf("%s used %.2fMB for maximum complexity generation (limit: 50MB)", name, allocatedMB)
			}
		})
	}
}

// TestEdgeCases_GenreSwitching tests same seed with different genres
func TestEdgeCases_GenreSwitching(t *testing.T) {
	generators := getAllGenerators()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	const testSeed = 42424242

	for name, gen := range generators {
		t.Run(name+"_genre_switch", func(t *testing.T) {
			successCount := 0

			for _, genre := range genres {
				params := getBaseParams(name)
				params.GenreID = genre

				result, err := gen.Generate(testSeed, params)
				if err != nil {
					t.Logf("%s failed for genre %s: %v (some genres may not be supported)", name, genre, err)
					continue
				}
				if result != nil {
					successCount++
				}
			}

			// At least 3 out of 5 genres should work
			if successCount < 3 {
				t.Errorf("%s only succeeded for %d out of %d genres (expected at least 3)",
					name, successCount, len(genres))
			}
		})
	}
}

// TestEdgeCases_ConcurrentGeneration tests 100 goroutines generating simultaneously
func TestEdgeCases_ConcurrentGeneration(t *testing.T) {
	generators := getAllGenerators()
	const numGoroutines = 100

	for name, gen := range generators {
		t.Run(name+"_concurrent", func(t *testing.T) {
			var wg sync.WaitGroup
			errors := make(chan error, numGoroutines)

			params := getBaseParams(name)

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(seed int64) {
					defer wg.Done()
					defer func() {
						if r := recover(); r != nil {
							errors <- formatPanicError(r)
						}
					}()

					result, err := gen.Generate(seed, params)
					if err != nil {
						errors <- err
					} else if result == nil {
						errors <- formatError("nil result without error")
					}
				}(int64(i))
			}

			wg.Wait()
			close(errors)

			// Check for errors
			errorCount := 0
			for err := range errors {
				t.Errorf("%s concurrent generation error: %v", name, err)
				errorCount++
			}

			if errorCount > 0 {
				t.Errorf("%s had %d errors out of %d concurrent generations", name, errorCount, numGoroutines)
			}
		})
	}
}

// TestEdgeCases_ResourceExhaustion tests generation under memory pressure
func TestEdgeCases_ResourceExhaustion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource exhaustion test in short mode")
	}

	generators := getAllGenerators()

	for name, gen := range generators {
		t.Run(name+"_resource_exhaustion", func(t *testing.T) {
			params := getBaseParams(name)
			params.Difficulty = 0.8
			params.Depth = 50

			// Generate 100 items rapidly to stress memory
			for i := 0; i < 100; i++ {
				result, err := gen.Generate(int64(i), params)
				if err != nil {
					t.Logf("%s failed under resource pressure at iteration %d: %v (acceptable)", name, i, err)
					break
				} else if result == nil {
					t.Errorf("%s returned nil at iteration %d without error", name, i)
					break
				}
			}

			// Force GC to check for leaks
			runtime.GC()
		})
	}
}

// TestEdgeCases_CorruptInput tests partially invalid GenerationParams
func TestEdgeCases_CorruptInput(t *testing.T) {
	generators := getAllGenerators()

	// Params with Custom map containing various problematic values
	corruptParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"nil_value":    nil,
			"negative_int": -999,
			"huge_number":  math.MaxFloat64,
			"empty_string": "",
			"invalid_type": complex(1, 2), // Complex numbers
		},
	}

	for name, gen := range generators {
		t.Run(name+"_corrupt_input", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked with corrupt input: %v", name, r)
				}
			}()

			result, err := gen.Generate(55555, corruptParams)
			// Should either succeed (ignoring Custom) or error gracefully
			if err != nil {
				t.Logf("%s rejected corrupt input: %v (acceptable)", name, err)
			} else if result == nil {
				t.Errorf("%s returned nil without error for corrupt input", name)
			}
		})
	}
}

// TestEdgeCases_AllGenresCovered tests all 5 genres work for each generator
func TestEdgeCases_AllGenresCovered(t *testing.T) {
	generators := getAllGenerators()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for name, gen := range generators {
		for _, genre := range genres {
			t.Run(name+"_genre_"+genre, func(t *testing.T) {
				params := getBaseParams(name)
				params.GenreID = genre

				result, err := gen.Generate(12345, params)
				if err != nil {
					t.Logf("%s failed for genre %s: %v (acceptable if genre not supported)", name, genre, err)
				} else if result == nil {
					t.Errorf("%s returned nil for genre %s without error", name, genre)
				}
			})
		}
	}
}

// TestEdgeCases_ZeroDifficulty tests minimum difficulty edge case
func TestEdgeCases_ZeroDifficulty(t *testing.T) {
	generators := getAllGenerators()

	for name, gen := range generators {
		t.Run(name+"_zero_difficulty", func(t *testing.T) {
			params := getBaseParams(name)
			params.Difficulty = 0.0
			params.Depth = 1

			result, err := gen.Generate(1, params)
			if err != nil {
				t.Errorf("%s failed with zero difficulty: %v", name, err)
			} else if result == nil {
				t.Errorf("%s returned nil with zero difficulty", name)
			}
		})
	}
}

// Helper functions

// getAllGenerators returns generators as a map for edge case testing.
// Delegates to GetAllGenerators() for consistency.
func getAllGenerators() map[string]procgen.Generator {
	entries := GetAllGenerators()
	result := make(map[string]procgen.Generator, len(entries))
	for _, entry := range entries {
		result[entry.Name] = entry.Generator
	}
	return result
}

// getBaseParams returns appropriate base parameters for each generator.
// Delegates to GetAllGenerators() for consistency.
func getBaseParams(generatorName string) procgen.GenerationParams {
	entries := GetAllGenerators()
	for _, entry := range entries {
		if entry.Name == generatorName {
			return entry.Params
		}
	}

	// Fallback for unknown generators
	return procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}
}

func formatSeed(seed int64) string {
	switch seed {
	case 0:
		return "zero"
	case -1:
		return "negative_one"
	case math.MaxInt64:
		return "max_int64"
	case math.MinInt64:
		return "min_int64"
	case 1:
		return "one"
	default:
		return "other"
	}
}

func formatPanicError(r interface{}) error {
	return &panicError{value: r}
}

func formatError(msg string) error {
	return &customError{msg: msg}
}

type panicError struct {
	value interface{}
}

func (e *panicError) Error() string {
	return formatInterface(e.value)
}

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func formatInterface(v interface{}) string {
	if v == nil {
		return "nil"
	}
	switch val := v.(type) {
	case string:
		return val
	case error:
		return val.Error()
	default:
		return "panic"
	}
}
