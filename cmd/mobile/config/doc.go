// Package config provides configuration utilities for mobile game initialization.
//
// This package handles environment-based configuration for the mobile client,
// including seed generation and genre selection. It provides a clean separation
// between configuration parsing and game logic.
//
// # Determinism Exception
//
// IMPORTANT: This package contains a documented exception to Coding Guideline #2
// (Deterministic Generation). GetSeedFromEnv() uses time.Now() as a fallback seed
// when VENTURE_SEED is unset. This is INTENTIONAL for mobile UX - casual players
// get a unique world each launch. For reproducible gameplay (bug reports, testing,
// multiplayer), always set VENTURE_SEED to a specific value.
//
// # Seed Configuration
//
// The world seed controls all procedural generation (terrain, items, enemies).
// Seeds can be specified via the VENTURE_SEED environment variable for
// deterministic/reproducible gameplay:
//
//	export VENTURE_SEED=12345
//
// If VENTURE_SEED is not set or invalid, a time-based seed is used. This is
// intentional for mobile UX - players get a different experience each launch
// unless they explicitly set a seed for testing or sharing worlds.
//
// # Genre Configuration
//
// The game genre controls theming for all procedural content. Valid genres:
//   - fantasy: Medieval fantasy with magic and dragons
//   - scifi: Space exploration and advanced technology
//   - horror: Dark atmosphere with survival elements
//   - cyberpunk: Neon-lit dystopian future
//   - postapoc: Post-apocalyptic wasteland survival
//
// Genre can be specified via VENTURE_GENRE:
//
//	export VENTURE_GENRE=cyberpunk
//
// If not set or invalid, a random genre is selected using the world seed's RNG.
//
// # Thread Safety
//
// All functions in this package are safe for concurrent use. They read
// environment variables atomically and do not maintain shared state.
//
// # Example Usage
//
//	logger := logrus.New()
//	seed := config.GetSeedFromEnv(logger)
//	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
//	rng := rand.New(rand.NewSource(seed))
//	genre := config.GetGenreFromEnv(genres, rng, logger)
package config
