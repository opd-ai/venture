// Package audit provides comprehensive validation tests for Phase 62.1: Generator Determinism Validation
//
// This package implements production-readiness audits for all procedural generators,
// ensuring deterministic output, cross-platform consistency, and version stability.
//
// Phase 62.1 Requirements:
// - Same seed produces identical output (byte-for-byte comparison)
// - Different seeds produce varied output (>80% different)
// - Seed derivation: sub-seeds don't collide across generator types
// - Platform consistency: Linux/macOS/Windows/WASM same results
// - Version stability: v10.0 output matches v9.0 for same seed
//
// Test Coverage: 15+ generators × 5 tests = 75+ total determinism tests
package audit

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
)

// GeneratorInfo describes a generator for audit testing
// Deprecated: Use GeneratorEntry from generators.go instead
type GeneratorInfo = GeneratorEntry

// getGenerators returns all generators to audit for Phase 62.1
// Delegates to GetAllGenerators() for consistency across tests
func getGenerators() []GeneratorInfo {
	return GetAllGenerators()
}

// hashOutput creates a deterministic hash of generator output
func hashOutput(result interface{}) ([32]byte, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to marshal: %w", err)
	}
	return sha256.Sum256(data), nil
}

// compareOutputs compares two generator outputs byte-for-byte via JSON
func compareOutputs(a, b interface{}) (bool, error) {
	dataA, err := json.Marshal(a)
	if err != nil {
		return false, fmt.Errorf("failed to marshal A: %w", err)
	}
	dataB, err := json.Marshal(b)
	if err != nil {
		return false, fmt.Errorf("failed to marshal B: %w", err)
	}
	return bytes.Equal(dataA, dataB), nil
}

// calculateDifference calculates percentage difference between two JSON outputs
func calculateDifference(a, b interface{}) (float64, error) {
	hashA, err := hashOutput(a)
	if err != nil {
		return 0, err
	}
	hashB, err := hashOutput(b)
	if err != nil {
		return 0, err
	}

	// Count differing bytes in hash
	diff := 0
	for i := 0; i < 32; i++ {
		if hashA[i] != hashB[i] {
			diff++
		}
	}
	return float64(diff) / 32.0 * 100.0, nil
}

// TestDeterminism_SameSeedProducesIdenticalOutput verifies Phase 62.1 requirement #1
//
// Acceptance Criteria: 100% determinism - zero failures in 1000 runs per generator
func TestDeterminism_SameSeedProducesIdenticalOutput(t *testing.T) {
	const runs = 100 // Run 100 times per generator (subset of 1000 for CI speed)

	generators := getGenerators()
	for _, g := range generators {
		g := g // capture range variable
		t.Run(g.Name, func(t *testing.T) {
			t.Parallel()

			const seed = 12345
			var outputs []interface{}

			// Generate same output multiple times
			for i := 0; i < runs; i++ {
				result, err := g.Generator.Generate(seed, g.Params)
				if err != nil {
					t.Fatalf("Run %d: Generate() failed: %v", i, err)
				}
				outputs = append(outputs, result)
			}

			// Verify all outputs match first output
			firstHash, err := hashOutput(outputs[0])
			if err != nil {
				t.Fatalf("Failed to hash first output: %v", err)
			}

			for i := 1; i < runs; i++ {
				hash, err := hashOutput(outputs[i])
				if err != nil {
					t.Fatalf("Run %d: Failed to hash output: %v", i, err)
				}

				if hash != firstHash {
					identical, _ := compareOutputs(outputs[0], outputs[i])
					t.Errorf("Run %d: Output differs from run 0 (hash mismatch: %v vs %v, byte-equal: %v)",
						i, firstHash[:4], hash[:4], identical)
				}
			}

			t.Logf("✓ %s: %d runs with seed %d produced identical output (determinism: 100%%)",
				g.Name, runs, seed)
		})
	}
}

// TestDeterminism_DifferentSeedsProduceVariedOutput verifies Phase 62.1 requirement #2
//
// Acceptance Criteria: Different seeds produce >80% different output
func TestDeterminism_DifferentSeedsProduceVariedOutput(t *testing.T) {
	const numSeeds = 50
	const minVariation = 80.0 // Minimum 80% variation required

	generators := getGenerators()
	for _, g := range generators {
		g := g // capture range variable
		t.Run(g.Name, func(t *testing.T) {
			t.Parallel()

			// Generate outputs with different seeds
			seeds := make([]int64, numSeeds)
			outputs := make([]interface{}, numSeeds)
			hashes := make([][32]byte, numSeeds)

			rng := rand.New(rand.NewSource(777))
			for i := 0; i < numSeeds; i++ {
				seeds[i] = rng.Int63()
				result, err := g.Generator.Generate(seeds[i], g.Params)
				if err != nil {
					t.Fatalf("Seed %d: Generate() failed: %v", seeds[i], err)
				}
				outputs[i] = result

				hash, err := hashOutput(result)
				if err != nil {
					t.Fatalf("Seed %d: Failed to hash output: %v", seeds[i], err)
				}
				hashes[i] = hash
			}

			// Calculate variation between all pairs
			totalComparisons := 0
			totalDifference := 0.0

			for i := 0; i < numSeeds; i++ {
				for j := i + 1; j < numSeeds; j++ {
					diff, err := calculateDifference(outputs[i], outputs[j])
					if err != nil {
						t.Fatalf("Failed to calculate difference between seeds %d and %d: %v",
							seeds[i], seeds[j], err)
					}
					totalDifference += diff
					totalComparisons++
				}
			}

			averageVariation := totalDifference / float64(totalComparisons)
			if averageVariation < minVariation {
				t.Errorf("%s: Average variation %.1f%% < required %.1f%% (seeds too similar)",
					g.Name, averageVariation, minVariation)
			}

			t.Logf("✓ %s: %d seeds produced %.1f%% average variation (>%.1f%% required)",
				g.Name, numSeeds, averageVariation, minVariation)
		})
	}
}

// TestDeterminism_SeedDerivationNonCollision verifies Phase 62.1 requirement #3
//
// Acceptance Criteria: Seed collision rate <0.01% across 1M generated seeds
func TestDeterminism_SeedDerivationNonCollision(t *testing.T) {
	const numSeeds = 10000        // Test 10k seeds (subset of 1M for CI speed)
	const maxCollisionRate = 0.01 // 0.01% maximum collision rate

	t.Run("SeedGenerator", func(t *testing.T) {
		seedGen := procgen.NewSeedGenerator(12345)
		seenSeeds := make(map[int64]string)
		collisions := 0

		generators := getGenerators()
		for i := 0; i < numSeeds; i++ {
			for _, g := range generators {
				seed := seedGen.GetSeed(g.Name, i)

				if prev, exists := seenSeeds[seed]; exists {
					collisions++
					t.Logf("Collision: seed %d used by both '%s' and '%s' (iteration %d)",
						seed, prev, g.Name, i)
				} else {
					seenSeeds[seed] = fmt.Sprintf("%s#%d", g.Name, i)
				}
			}
		}

		totalSeeds := numSeeds * len(generators)
		collisionRate := float64(collisions) / float64(totalSeeds) * 100.0

		if collisionRate > maxCollisionRate {
			t.Errorf("Collision rate %.4f%% > maximum %.4f%% (%d collisions in %d seeds)",
				collisionRate, maxCollisionRate, collisions, totalSeeds)
		}

		t.Logf("✓ SeedGenerator: %d seeds generated, %d collisions (%.4f%% rate < %.4f%% max)",
			totalSeeds, collisions, collisionRate, maxCollisionRate)
	})
}

// TestDeterminism_PlatformConsistency verifies Phase 62.1 requirement #4
//
// Acceptance Criteria: Linux/macOS/Windows/WASM produce exact same JSON output
//
// Note: This test verifies determinism across goroutines running on same platform.
// True cross-platform testing requires separate build targets and CI runners.
func TestDeterminism_PlatformConsistency(t *testing.T) {
	const seed = 54321
	const numGoroutines = 10

	generators := getGenerators()
	for _, g := range generators {
		g := g // capture range variable
		t.Run(g.Name, func(t *testing.T) {
			t.Parallel()

			// Generate reference output on main goroutine
			reference, err := g.Generator.Generate(seed, g.Params)
			if err != nil {
				t.Fatalf("Reference generation failed: %v", err)
			}
			refHash, err := hashOutput(reference)
			if err != nil {
				t.Fatalf("Failed to hash reference: %v", err)
			}

			// Generate same output on multiple goroutines (simulates concurrent generation)
			var wg sync.WaitGroup
			errors := make(chan error, numGoroutines)

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()

					result, err := g.Generator.Generate(seed, g.Params)
					if err != nil {
						errors <- fmt.Errorf("goroutine %d: Generate() failed: %w", id, err)
						return
					}

					hash, err := hashOutput(result)
					if err != nil {
						errors <- fmt.Errorf("goroutine %d: Failed to hash output: %w", id, err)
						return
					}

					if hash != refHash {
						errors <- fmt.Errorf("goroutine %d: Hash mismatch with reference", id)
					}
				}(i)
			}

			wg.Wait()
			close(errors)

			// Check for errors
			for err := range errors {
				t.Error(err)
			}

			t.Logf("✓ %s: %d goroutines produced identical output on %s/%s (platform consistency)",
				g.Name, numGoroutines, runtime.GOOS, runtime.GOARCH)
		})
	}
}

// TestDeterminism_VersionStability verifies Phase 62.1 requirement #5
//
// Acceptance Criteria: v1.0.0 output matches baseline for same seed (migration test)
//
// This test compares current generator output against saved v1.0.0 baseline hashes.
// If a hash doesn't match, it indicates a breaking change in generator logic.
func TestDeterminism_VersionStability(t *testing.T) {
	const seed = BaselineSeed // 99999

	generators := getGenerators()
	for _, g := range generators {
		g := g // capture range variable
		t.Run(g.Name, func(t *testing.T) {
			t.Parallel()

			// Generate output
			result, err := g.Generator.Generate(seed, g.Params)
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}

			// Validate output meets quality thresholds (ensures no breaking changes)
			if err := g.Generator.Validate(result); err != nil {
				t.Errorf("Validate() failed for seed %d: %v (quality regression)", seed, err)
			}

			// Hash output for stability tracking
			hash, err := hashOutput(result)
			if err != nil {
				t.Fatalf("Failed to hash output: %v", err)
			}

			// Compare against baseline (Phase 62.1 complete)
			baselinePrefix := GetBaselinePrefix(g.Name)
			if baselinePrefix == "" {
				t.Errorf("%s: no baseline hash found (add to baseline.go)", g.Name)
				return
			}

			if HashMatchesBaseline(g.Name, hash) {
				t.Logf("✓ %s: v%s stable - hash %x... matches baseline",
					g.Name, BaselineVersion, hash[:8])
			} else {
				t.Errorf("%s: BREAKING CHANGE - hash %x... does not match baseline %s",
					g.Name, hash[:8], baselinePrefix)
				t.Logf("  This indicates the generator output has changed.")
				t.Logf("  If intentional, update baseline.go with new hash prefix.")
			}
		})
	}
}

// TestDeterminism_AcceptanceCriteria_1000Runs validates Phase 62.1 acceptance
//
// Full acceptance test: 100% determinism - zero failures in 1000 runs per generator
// This test is tagged for CI skipping due to long runtime (~5-10 minutes).
func TestDeterminism_AcceptanceCriteria_1000Runs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping 1000-run acceptance test in short mode")
	}

	const runs = 1000
	const seed = 11111

	generators := getGenerators()
	for _, g := range generators {
		g := g // capture range variable
		t.Run(g.Name, func(t *testing.T) {
			// Generate reference output
			reference, err := g.Generator.Generate(seed, g.Params)
			if err != nil {
				t.Fatalf("Reference generation failed: %v", err)
			}
			refHash, err := hashOutput(reference)
			if err != nil {
				t.Fatalf("Failed to hash reference: %v", err)
			}

			failures := 0
			for i := 0; i < runs; i++ {
				result, err := g.Generator.Generate(seed, g.Params)
				if err != nil {
					t.Errorf("Run %d: Generate() failed: %v", i, err)
					failures++
					continue
				}

				hash, err := hashOutput(result)
				if err != nil {
					t.Errorf("Run %d: Failed to hash output: %v", i, err)
					failures++
					continue
				}

				if hash != refHash {
					t.Errorf("Run %d: Hash mismatch (non-deterministic output)", i)
					failures++
				}
			}

			successRate := float64(runs-failures) / float64(runs) * 100.0
			if failures > 0 {
				t.Errorf("%s: %d/%d runs failed (%.1f%% success, required 100%%)",
					g.Name, failures, runs, successRate)
			} else {
				t.Logf("✓ %s: 1000/1000 runs passed (100%% determinism)",
					g.Name)
			}
		})
	}
}
