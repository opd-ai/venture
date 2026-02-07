package config

import (
	"math/rand"
	"os"
	"testing"
)

func TestGetSeedFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		wantDefault bool
		checkSeed   func(int64) bool
	}{
		{
			name:        "valid seed from environment",
			envValue:    "12345",
			wantDefault: false,
			checkSeed:   func(s int64) bool { return s == 12345 },
		},
		{
			name:        "negative seed from environment",
			envValue:    "-9876",
			wantDefault: false,
			checkSeed:   func(s int64) bool { return s == -9876 },
		},
		{
			name:        "zero seed from environment",
			envValue:    "0",
			wantDefault: false,
			checkSeed:   func(s int64) bool { return s == 0 },
		},
		{
			name:        "large seed from environment",
			envValue:    "9223372036854775807", // max int64
			wantDefault: false,
			checkSeed:   func(s int64) bool { return s == 9223372036854775807 },
		},
		{
			name:        "invalid seed - not a number",
			envValue:    "not-a-number",
			wantDefault: true,
			checkSeed:   func(s int64) bool { return s > 0 }, // Should be time-based
		},
		{
			name:        "invalid seed - overflow",
			envValue:    "99999999999999999999999",
			wantDefault: true,
			checkSeed:   func(s int64) bool { return s > 0 }, // Should be time-based
		},
		{
			name:        "empty environment variable",
			envValue:    "",
			wantDefault: true,
			checkSeed:   func(s int64) bool { return s > 0 }, // Should be time-based
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("VENTURE_SEED", tt.envValue)
				defer os.Unsetenv("VENTURE_SEED")
			} else {
				os.Unsetenv("VENTURE_SEED")
			}

			seed := GetSeedFromEnv(nil)

			if !tt.checkSeed(seed) {
				t.Errorf("getSeedFromEnv() = %d, seed check failed", seed)
			}

			// Verify determinism for valid seeds (non-default case)
			if !tt.wantDefault {
				seed2 := GetSeedFromEnv(nil)
				if seed != seed2 {
					t.Errorf("getSeedFromEnv() not deterministic: %d != %d", seed, seed2)
				}
			}
		})
	}
}

func TestGetGenreFromEnv(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	rng := rand.New(rand.NewSource(12345))

	tests := []struct {
		name       string
		envValue   string
		wantGenre  string
		wantRandom bool
	}{
		{
			name:       "valid genre fantasy",
			envValue:   "fantasy",
			wantGenre:  "fantasy",
			wantRandom: false,
		},
		{
			name:       "valid genre scifi",
			envValue:   "scifi",
			wantGenre:  "scifi",
			wantRandom: false,
		},
		{
			name:       "valid genre horror",
			envValue:   "horror",
			wantGenre:  "horror",
			wantRandom: false,
		},
		{
			name:       "valid genre cyberpunk",
			envValue:   "cyberpunk",
			wantGenre:  "cyberpunk",
			wantRandom: false,
		},
		{
			name:       "valid genre postapoc",
			envValue:   "postapoc",
			wantGenre:  "postapoc",
			wantRandom: false,
		},
		{
			name:       "invalid genre",
			envValue:   "invalid-genre",
			wantRandom: true,
		},
		{
			name:       "empty environment variable",
			envValue:   "",
			wantRandom: true,
		},
		{
			name:       "case sensitive - Fantasy",
			envValue:   "Fantasy",
			wantRandom: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("VENTURE_GENRE", tt.envValue)
				defer os.Unsetenv("VENTURE_GENRE")
			} else {
				os.Unsetenv("VENTURE_GENRE")
			}

			// Reset RNG for deterministic random selection
			rng = rand.New(rand.NewSource(12345))
			genre := GetGenreFromEnv(genres, rng, nil)

			if tt.wantRandom {
				// Verify it's a valid genre from the list
				found := false
				for _, validGenre := range genres {
					if genre == validGenre {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("getGenreFromEnv() = %q, not in valid genres %v", genre, genres)
				}
			} else {
				if genre != tt.wantGenre {
					t.Errorf("getGenreFromEnv() = %q, want %q", genre, tt.wantGenre)
				}
			}
		})
	}
}

func TestGetGenreFromEnv_Determinism(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	// Test deterministic random selection with same seed
	rng1 := rand.New(rand.NewSource(12345))
	rng2 := rand.New(rand.NewSource(12345))

	os.Unsetenv("VENTURE_GENRE")
	genre1 := GetGenreFromEnv(genres, rng1, nil)
	genre2 := GetGenreFromEnv(genres, rng2, nil)

	if genre1 != genre2 {
		t.Errorf("getGenreFromEnv() not deterministic with same seed: %q != %q", genre1, genre2)
	}
}

func TestGetGenreFromEnv_AllGenresValid(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	// Verify all genres in the list are accepted
	for _, validGenre := range genres {
		t.Run(validGenre, func(t *testing.T) {
			os.Setenv("VENTURE_GENRE", validGenre)
			defer os.Unsetenv("VENTURE_GENRE")

			rng := rand.New(rand.NewSource(12345))
			genre := GetGenreFromEnv(genres, rng, nil)

			if genre != validGenre {
				t.Errorf("getGenreFromEnv() = %q, want %q", genre, validGenre)
			}
		})
	}
}

func BenchmarkGetSeedFromEnv_Valid(b *testing.B) {
	os.Setenv("VENTURE_SEED", "12345")
	defer os.Unsetenv("VENTURE_SEED")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetSeedFromEnv(nil)
	}
}

func BenchmarkGetSeedFromEnv_TimeBased(b *testing.B) {
	os.Unsetenv("VENTURE_SEED")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetSeedFromEnv(nil)
	}
}

func BenchmarkGetGenreFromEnv_Valid(b *testing.B) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	os.Setenv("VENTURE_GENRE", "fantasy")
	defer os.Unsetenv("VENTURE_GENRE")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewSource(12345))
		_ = GetGenreFromEnv(genres, rng, nil)
	}
}

func BenchmarkGetGenreFromEnv_Random(b *testing.B) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	os.Unsetenv("VENTURE_GENRE")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewSource(12345))
		_ = GetGenreFromEnv(genres, rng, nil)
	}
}
