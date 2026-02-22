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
// # Advanced Particle Physics (Phase 18.2)
//
// Four specialized physics simulation types for realistic motion:
//
// 1. Fluid Simulation (SPH):
//   - Smoothed Particle Hydrodynamics for water and blood
//   - Density and pressure calculation
//   - Surface tension and cohesion effects
//   - Viscosity control (0.0-1.0)
//   - Performance: 349µs per frame for 200 particles
//
// 2. Fire Propagation:
//   - Heat transfer between particles
//   - Ignition thresholds and fuel consumption
//   - Buoyancy forces (hot particles rise)
//   - Ember spawning
//   - Performance: 129µs per frame for 200 particles
//
// 3. Smoke Billowing:
//   - Turbulence using noise functions
//   - Rising motion with configurable speed
//   - Size expansion over time
//   - Dissipation effects
//   - Performance: 9.3µs per frame for 200 particles
//
// 4. Debris Collision:
//   - Particle-particle collision detection
//   - Spatial hashing for efficient neighbor queries
//   - Restitution and friction coefficients
//   - Angular velocity and rotation
//   - Performance: 96µs per frame for 200 particles
//
// Combined performance: All 4 simulations run in <5% of frame time (60 FPS).
//
// # Environmental Ambience (Phase 18.3)
//
// Ambient particle effects for environmental atmosphere with 10 distinct environment types:
//   - Dungeon: Dust motes and moisture droplets (slow drifting)
//   - Cave: Mineral dust and water drips (occasional falling drops)
//   - Forest: Falling leaves and pollen (gentle swaying motion)
//   - Desert: Sand particles and heat shimmer (wind-blown, gusting)
//   - Snow: Drifting snow and ice crystals (gentle falling with drift)
//   - Swamp: Mist and fireflies (slow floating with vertical oscillation)
//   - Lava: Ash and embers (rising with turbulence)
//   - City: Paper debris and smoke (tumbling, rising)
//   - Laboratory: Energy particles and sparks (smooth orbital motion)
//   - Ruins: Dust and falling debris (floating dust, occasional falling)
//
// Ambient systems provide:
//   - 50-100 particles per environment (configurable density 0.0-1.0)
//   - Area-specific effects (dripping water in caves, leaves in forests)
//   - Biome-specific particles (forest leaves, desert sand)
//   - Subtle background movement for atmosphere
//   - Environment-specific behaviors (drift, sway, gust, rise, tumble, orbit)
//   - Automatic particle respawning for infinite ambience
//   - Performance: <2% frame time overhead (~333µs at 60 FPS)
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
//	    // Handle error appropriately in production (log.Fatal is for examples only)
//	    logrus.WithError(err).Error("Failed to generate particle system")
//	    return
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
//	    logrus.WithError(err).Error("Failed to generate weather")
//	    return
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
// # Advanced Physics Usage
//
//	// Create physics particles
//	particles := make([]PhysicsParticle, 200)
//	// ... initialize particles ...
//
//	// Fluid simulation
//	sphConfig := DefaultSPHConfig()
//	UpdateSPH(particles, sphConfig, 0.016)
//
//	// Fire propagation
//	fireConfig := DefaultFireConfig()
//	rng := rand.New(rand.NewSource(12345))
//	UpdateFire(particles, fireConfig, 0.016, rng)
//
//	// Smoke turbulence
//	smokeConfig := DefaultSmokeConfig()
//	UpdateSmoke(particles, smokeConfig, 0.016, time)
//
//	// Debris collisions
//	debrisConfig := DefaultDebrisConfig()
//	UpdateDebris(particles, debrisConfig, 0.016, groundY)
//
// # Ambient Particle Usage
//
//	config := particles.AmbienceConfig{
//	    Type:    particles.EnvironmentForest,
//	    Width:   800,
//	    Height:  600,
//	    GenreID: "fantasy",
//	    Seed:    12345,
//	    Density: 0.5, // 0.0-1.0 scale
//	}
//	ambience, err := particles.GenerateAmbience(config)
//	if err != nil {
//	    logrus.WithError(err).Error("Failed to generate ambience")
//	    return
//	}
//
//	// Update ambience particles each frame
//	for i := 0; i < 60; i++ {
//	    ambience.Update(0.016) // 60 FPS
//	}
//
//	// Access particle data for rendering
//	for _, p := range ambience.Particles {
//	    // Render particle at (p.X, p.Y) with color p.Color and size p.Size
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
// times under 1ms for particle systems with up to 1000 particles. Weather systems
// can handle 500-1000+ particles while maintaining 60 FPS (16.67ms frame budget).
//
// Benchmark results (Phase 18.2, 200 particles):
//   - SPH Fluid: 349µs per frame
//   - Fire Propagation: 129µs per frame
//   - Smoke Turbulence: 9.3µs per frame
//   - Debris Collision: 96µs per frame
//   - Total: 583µs (3.5% of 60 FPS frame budget)
//
// Use the Count parameter to control performance vs visual quality tradeoffs.
package particles
