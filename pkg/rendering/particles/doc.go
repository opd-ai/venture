// Package particles provides procedural particle effect generation for the Venture game.
//
// This package generates visual particle effects using mathematical algorithms and
// genre-based color palettes. All particle generation is deterministic based on seed
// values, ensuring reproducibility and network synchronization in multiplayer scenarios.
//
// # Particle Types
//
// The package supports several types of particle effects:
//   - Spark: Quick, bright particles for impacts and explosions
//   - Smoke: Soft, fading particles for atmospheric effects
//   - Magic: Glowing particles for spell effects
//   - Flame: Fire-like particles with color gradients
//   - Blood: Splatter particles for combat effects
//   - Dust: Small particles for environmental effects
//
// # Basic Usage
//
//	gen := particles.NewGenerator()
//	config := particles.Config{
//	    Type:       particles.ParticleSpark,
//	    Count:      50,
//	    GenreID:    "fantasy",
//	    Seed:       12345,
//	    Duration:   1.0,
//	    SpreadX:    10.0,
//	    SpreadY:    10.0,
//	}
//	system, err := gen.Generate(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Determinism
//
// All particle generation uses seed-based RNG to ensure the same configuration
// always produces identical particle patterns. This is critical for:
//   - Multiplayer synchronization
//   - Replay systems
//   - Testing and debugging
//
// # Performance
//
// Particle generation is optimized for runtime creation with typical generation
// times under 1ms for particle systems with up to 1000 particles. Use the Count
// parameter to control performance vs visual quality tradeoffs.
package particles
