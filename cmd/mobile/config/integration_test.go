package config

import (
	"math/rand"
	"os"
	"testing"
)

// TestIntegration_SeedAndGenreFromEnv tests the full configuration flow
func TestIntegration_SeedAndGenreFromEnv(t *testing.T) {
	// Setup: Clear environment
	os.Unsetenv("VENTURE_SEED")
	os.Unsetenv("VENTURE_GENRE")

	// Test 1: Default behavior (time-based seed, random genre)
	t.Run("default_configuration", func(t *testing.T) {
		seed1 := GetSeedFromEnv(nil)
		genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
		rng := rand.New(rand.NewSource(seed1))
		genre1 := GetGenreFromEnv(genres, rng, nil)

		// Verify seed is positive (time-based)
		if seed1 <= 0 {
			t.Errorf("expected positive seed, got %d", seed1)
		}

		// Verify genre is valid
		validGenre := false
		for _, g := range genres {
			if genre1 == g {
				validGenre = true
				break
			}
		}
		if !validGenre {
			t.Errorf("genre %q not in valid list %v", genre1, genres)
		}
	})

	// Test 2: Explicit configuration
	t.Run("explicit_configuration", func(t *testing.T) {
		os.Setenv("VENTURE_SEED", "99999")
		os.Setenv("VENTURE_GENRE", "horror")
		defer os.Unsetenv("VENTURE_SEED")
		defer os.Unsetenv("VENTURE_GENRE")

		seed := GetSeedFromEnv(nil)
		if seed != 99999 {
			t.Errorf("expected seed 99999, got %d", seed)
		}

		genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
		rng := rand.New(rand.NewSource(seed))
		genre := GetGenreFromEnv(genres, rng, nil)

		if genre != "horror" {
			t.Errorf("expected genre horror, got %q", genre)
		}
	})

	// Test 3: Partial configuration (seed only)
	t.Run("seed_only_configuration", func(t *testing.T) {
		os.Setenv("VENTURE_SEED", "55555")
		os.Unsetenv("VENTURE_GENRE")
		defer os.Unsetenv("VENTURE_SEED")

		seed := GetSeedFromEnv(nil)
		if seed != 55555 {
			t.Errorf("expected seed 55555, got %d", seed)
		}

		genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
		rng := rand.New(rand.NewSource(seed))
		genre := GetGenreFromEnv(genres, rng, nil)

		// Genre should be deterministic based on seed
		rng2 := rand.New(rand.NewSource(seed))
		genre2 := GetGenreFromEnv(genres, rng2, nil)

		if genre != genre2 {
			t.Errorf("genre not deterministic: %q != %q", genre, genre2)
		}
	})

	// Test 4: Partial configuration (genre only)
	t.Run("genre_only_configuration", func(t *testing.T) {
		os.Unsetenv("VENTURE_SEED")
		os.Setenv("VENTURE_GENRE", "cyberpunk")
		defer os.Unsetenv("VENTURE_GENRE")

		seed := GetSeedFromEnv(nil)
		if seed <= 0 {
			t.Errorf("expected positive time-based seed, got %d", seed)
		}

		genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
		rng := rand.New(rand.NewSource(seed))
		genre := GetGenreFromEnv(genres, rng, nil)

		if genre != "cyberpunk" {
			t.Errorf("expected genre cyberpunk, got %q", genre)
		}
	})

	// Test 5: Invalid configuration fallback
	t.Run("invalid_configuration_fallback", func(t *testing.T) {
		os.Setenv("VENTURE_SEED", "invalid-seed")
		os.Setenv("VENTURE_GENRE", "invalid-genre")
		defer os.Unsetenv("VENTURE_SEED")
		defer os.Unsetenv("VENTURE_GENRE")

		// Should fall back to time-based seed
		seed := GetSeedFromEnv(nil)
		if seed <= 0 {
			t.Errorf("expected positive fallback seed, got %d", seed)
		}

		// Should fall back to random genre
		genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
		rng := rand.New(rand.NewSource(seed))
		genre := GetGenreFromEnv(genres, rng, nil)

		validGenre := false
		for _, g := range genres {
			if genre == g {
				validGenre = true
				break
			}
		}
		if !validGenre {
			t.Errorf("fallback genre %q not in valid list %v", genre, genres)
		}
	})
}

// TestIntegration_Determinism verifies reproducible world generation with same seed
func TestIntegration_Determinism(t *testing.T) {
	testSeed := int64(12345)
	os.Setenv("VENTURE_SEED", "12345")
	os.Setenv("VENTURE_GENRE", "fantasy")
	defer os.Unsetenv("VENTURE_SEED")
	defer os.Unsetenv("VENTURE_GENRE")

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	// First run
	seed1 := GetSeedFromEnv(nil)
	rng1 := rand.New(rand.NewSource(seed1))
	genre1 := GetGenreFromEnv(genres, rng1, nil)

	// Second run (should be identical)
	seed2 := GetSeedFromEnv(nil)
	rng2 := rand.New(rand.NewSource(seed2))
	genre2 := GetGenreFromEnv(genres, rng2, nil)

	if seed1 != seed2 {
		t.Errorf("seeds not identical: %d != %d", seed1, seed2)
	}

	if seed1 != testSeed {
		t.Errorf("seed not as expected: %d != %d", seed1, testSeed)
	}

	if genre1 != genre2 {
		t.Errorf("genres not identical: %q != %q", genre1, genre2)
	}

	if genre1 != "fantasy" {
		t.Errorf("genre not as expected: %q != %q", genre1, "fantasy")
	}

	// Third run with same RNG state should produce same random values
	rng3 := rand.New(rand.NewSource(testSeed))
	val1 := rng3.Intn(100)

	rng4 := rand.New(rand.NewSource(testSeed))
	val2 := rng4.Intn(100)

	if val1 != val2 {
		t.Errorf("RNG not deterministic: %d != %d", val1, val2)
	}
}
