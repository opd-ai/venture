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
//   - Ember: Glowing fire embers that rise and fade
//   - Sparkle: Magical sparkles that orbit and trail
//   - SmokePlume: Billowing smoke clouds
//   - Debris: Bouncing debris chunks
//
// # Weather Systems (Phase 18.1)
//
// Advanced weather particle systems with environmental effects:
//   - Rain: Falling droplets with puddle accumulation
//   - Snow: Falling snowflakes with drift and settling
//   - Fog: Ambient fog with visibility reduction (30-50%)
//   - Dust: Swirling dust particles
//   - Ash: Falling ash particles
//   - NeonRain: Cyberpunk-style neon rain
//   - Smog: Industrial smog with visibility reduction
//   - Radiation: Radioactive particles (post-apocalyptic)
//   - Sandstorm: Intense sandstorm with severe visibility reduction
//   - BloodRain: Horror-themed blood rain with puddle formation
//
// Weather systems include:
//   - Puddle accumulation for rain and blood rain
//   - Snow depth tracking and wind drift
//   - Visibility modification (1.0 = normal, 0.0 = blind)
//   - Genre-specific weather types
//   - Support for 500-1000+ particles maintaining 60 FPS
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
// # Weather Usage
//
//	config := particles.WeatherConfig{
//	    Type:      particles.WeatherRain,
//	    Intensity: particles.IntensityHeavy,
//	    Width:     800,
//	    Height:    600,
//	    GenreID:   "fantasy",
//	    Seed:      12345,
//	    WindX:     10.0,
//	    WindY:     0.0,
//	}
//	ws, err := particles.GenerateWeather(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Simulate weather
//	for i := 0; i < 60; i++ {
//	    ws.Update(0.016) // 60 FPS
//	}
//
//	// Check environmental effects
//	visibility := ws.GetVisibilityModifier()
//	puddleLevel := ws.GetPuddleLevel(10, 5)
//	snowLevel := ws.GetSnowLevel(10, 5)
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
// times under 1ms for particle systems with up to 1000 particles. Weather systems
// can handle 500-1000+ particles while maintaining 60 FPS (16.67ms frame budget).
// Benchmark results (1000 particles):
//   - Generation: ~177µs
//   - Update: ~45µs per frame
//
// Use the Count parameter to control performance vs visual quality tradeoffs.
package particles
