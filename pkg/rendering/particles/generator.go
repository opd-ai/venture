// Package particles provides procedural particle effect generation.
// This file implements particle generators for visual effects like fire,
// smoke, magic, and explosions.
package particles

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// Generator creates procedural particle systems.
type Generator struct {
	paletteGen *palette.Generator
	logger     *logrus.Entry
}

// NewGenerator creates a new particle system generator.
func NewGenerator() *Generator {
	return NewGeneratorWithLogger(nil)
}

// NewGeneratorWithLogger creates a new particle system generator with a logger.
func NewGeneratorWithLogger(logger *logrus.Logger) *Generator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"generator": "particle",
		})
	}
	return &Generator{
		paletteGen: palette.NewGenerator(),
		logger:     logEntry,
	}
}

// Generate creates a particle system from the given configuration.
func (g *Generator) Generate(config Config) (*ParticleSystem, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"type":    config.Type,
			"genreID": config.GenreID,
			"seed":    config.Seed,
			"count":   config.Count,
		}).Debug("generating particle system")
	}

	if err := config.Validate(); err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("invalid particle config")
		}
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create RNG from seed
	rng := rand.New(rand.NewSource(config.Seed))

	// Generate color palette for genre
	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		if g.logger != nil {
			g.logger.WithError(err).Error("palette generation failed")
		}
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	// Create particle system from pool with pre-allocated particles
	// Note: NewParticleSystem expects particles to be passed in, but we
	// need to generate them. Create temporary slice, then pass to pooled system.
	particles := make([]Particle, config.Count)

	// Temporarily create system for generation (will be replaced with pooled version)
	system := &ParticleSystem{
		Particles:   particles,
		Type:        config.Type,
		Config:      config,
		ElapsedTime: 0,
	}

	// Generate particles based on type
	switch config.Type {
	case ParticleSpark:
		g.generateSparks(system, pal, rng, config)
	case ParticleSmoke:
		g.generateSmoke(system, pal, rng, config)
	case ParticleMagic:
		g.generateMagic(system, pal, rng, config)
	case ParticleFlame:
		g.generateFlame(system, pal, rng, config)
	case ParticleBlood:
		g.generateBlood(system, pal, rng, config)
	case ParticleDust:
		g.generateDust(system, pal, rng, config)
	case ParticleEmber:
		g.generateEmbers(system, pal, rng, config)
	case ParticleSparkle:
		g.generateSparkles(system, pal, rng, config)
	case ParticleSmokePlume:
		g.generateSmokePlume(system, pal, rng, config)
	case ParticleDebris:
		g.generateDebris(system, pal, rng, config)
	default:
		err := fmt.Errorf("unknown particle type: %d", config.Type)
		if g.logger != nil {
			g.logger.WithError(err).WithField("type", config.Type).Error("unknown particle type")
		}
		return nil, err
	}

	if g.logger != nil {
		g.logger.WithFields(logrus.Fields{
			"type":  config.Type,
			"count": config.Count,
		}).Info("particle system generated")
	}

	// Use pooled particle system instead of direct allocation
	// This transfers particles to a pooled system, reducing GC pressure
	pooledSystem := NewParticleSystem(system.Particles, config.Type, config)

	return pooledSystem, nil
}

// generateSparks creates bright, quick-moving spark particles.
// Phase 45: Sparks render at entity level with scaled sizes.
func (g *Generator) generateSparks(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	colors := []color.RGBA{
		{255, 255, 200, 255}, // Bright yellow
		{255, 200, 100, 255}, // Orange
		{255, 255, 255, 255}, // White
	}

	for i := range system.Particles {
		angle := rng.Float64() * 2 * math.Pi
		speed := rng.Float64() * config.SpreadX
		size := config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize)

		system.Particles[i] = Particle{
			X:           0,
			Y:           0,
			VX:          math.Cos(angle) * speed,
			VY:          math.Sin(angle) * speed,
			Color:       colors[rng.Intn(len(colors))],
			Size:        size,
			InitialSize: size,
			Life:        1.0,
			InitialLife: config.Duration * (0.5 + rng.Float64()*0.5),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 5,
			ZLayer:      ZLayerEntity,
		}
	}
}

// generateSmoke creates soft, slowly rising smoke particles.
// Phase 45: Smoke rises with shrink-on-rise behavior, rendering above entities.
func (g *Generator) generateSmoke(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	smokeColors := []color.RGBA{
		{100, 100, 100, 200},
		{120, 120, 120, 180},
		{80, 80, 80, 220},
	}

	for i := range system.Particles {
		angle := (rng.Float64()*2 - 1) * math.Pi / 4 // Upward cone
		speed := rng.Float64() * config.SpreadY * 0.5
		size := config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize)

		physics := PhysicsConfig{
			Gravity:       config.Gravity,
			AirResistance: 0.1, // Light air resistance for smoke
		}

		system.Particles[i] = Particle{
			X:           (rng.Float64()*2 - 1) * 4, // Scaled 2× spread
			Y:           0,
			VX:          math.Cos(angle-math.Pi/2) * speed,
			VY:          math.Sin(angle-math.Pi/2) * speed,
			Color:       smokeColors[rng.Intn(len(smokeColors))],
			Size:        size,
			InitialSize: size,
			Life:        1.0,
			InitialLife: config.Duration * (0.8 + rng.Float64()*0.4),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 2,
			ZLayer:      ZLayerAbove,
			Behavior:    BehaviorRising | BehaviorAirResistance | BehaviorShrinkOnRise,
			Physics:     physics,
		}
	}
}

// generateMagic creates glowing magical particles.
// Phase 45: Magic particles render above entities with orbital movement.
func (g *Generator) generateMagic(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Use palette colors for magic
	magicColors := []color.RGBA{
		color.RGBAModel.Convert(pal.Primary).(color.RGBA),
		color.RGBAModel.Convert(pal.Secondary).(color.RGBA),
		color.RGBAModel.Convert(pal.Colors[rng.Intn(len(pal.Colors)/2)]).(color.RGBA),
	}

	for i := range system.Particles {
		angle := rng.Float64() * 2 * math.Pi
		speed := rng.Float64() * config.SpreadX * 0.7
		size := config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize)

		system.Particles[i] = Particle{
			X:           0,
			Y:           0,
			VX:          math.Cos(angle) * speed,
			VY:          math.Sin(angle) * speed,
			Color:       magicColors[rng.Intn(len(magicColors))],
			Size:        size,
			InitialSize: size,
			Life:        1.0,
			InitialLife: config.Duration * (0.7 + rng.Float64()*0.6),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 3,
			ZLayer:      ZLayerAbove,
		}
	}
}

// generateFlame creates fire-like particles.
// Phase 45: Flames render above entities, shrinking as they rise.
func (g *Generator) generateFlame(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	flameColors := []color.RGBA{
		{255, 100, 0, 255},   // Orange
		{255, 200, 0, 255},   // Yellow
		{200, 50, 0, 255},    // Red
		{255, 255, 100, 255}, // Bright yellow
	}

	for i := range system.Particles {
		// Flames rise upward with some spread
		angle := -math.Pi/2 + (rng.Float64()*2-1)*math.Pi/6
		speed := config.SpreadY * (0.5 + rng.Float64()*0.5)
		size := config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize)

		physics := PhysicsConfig{
			Gravity: config.Gravity,
		}

		system.Particles[i] = Particle{
			X:           (rng.Float64()*2 - 1) * 6, // Scaled 2× spread
			Y:           0,
			VX:          math.Cos(angle) * speed,
			VY:          math.Sin(angle) * speed,
			Color:       flameColors[rng.Intn(len(flameColors))],
			Size:        size,
			InitialSize: size,
			Life:        1.0,
			InitialLife: config.Duration * (0.3 + rng.Float64()*0.4),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 4,
			ZLayer:      ZLayerAbove,
			Behavior:    BehaviorRising | BehaviorShrinkOnRise,
			Physics:     physics,
		}
	}
}

// generateBlood creates blood splatter particles.
// Phase 45: Blood splatters render at ground level.
func (g *Generator) generateBlood(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	bloodColors := []color.RGBA{
		{150, 0, 0, 255},
		{180, 20, 20, 255},
		{120, 0, 0, 255},
	}

	for i := range system.Particles {
		angle := rng.Float64() * 2 * math.Pi
		speed := rng.Float64() * config.SpreadX * 0.8
		size := config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize)

		physics := PhysicsConfig{
			Gravity:       config.Gravity,
			BounceDamping: 0.5, // Moderate bounce for blood droplets
		}

		system.Particles[i] = Particle{
			X:           0,
			Y:           0,
			VX:          math.Cos(angle) * speed,
			VY:          math.Sin(angle) * speed,
			Color:       bloodColors[rng.Intn(len(bloodColors))],
			Size:        size,
			InitialSize: size,
			Life:        1.0,
			InitialLife: config.Duration * (0.6 + rng.Float64()*0.4),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 2,
			ZLayer:      ZLayerGround,
			Behavior:    BehaviorGravity | BehaviorGrowOnFall,
			Physics:     physics,
		}
	}
}

// generateDust creates small dust particles.
// Phase 45: Dust renders at ground level with scaled sizes.
func (g *Generator) generateDust(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	dustColors := []color.RGBA{
		{160, 140, 120, 180},
		{140, 120, 100, 160},
		{180, 160, 140, 170},
	}

	for i := range system.Particles {
		angle := rng.Float64() * 2 * math.Pi
		speed := rng.Float64() * config.SpreadX * 0.3
		size := config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize)*0.5

		system.Particles[i] = Particle{
			X:           (rng.Float64()*2 - 1) * 10, // Scaled 2× spread
			Y:           (rng.Float64()*2 - 1) * 10, // Scaled 2× spread
			VX:          math.Cos(angle) * speed,
			VY:          math.Sin(angle)*speed - 2, // Slight upward drift, scaled
			Color:       dustColors[rng.Intn(len(dustColors))],
			Size:        size,
			InitialSize: size,
			Life:        1.0,
			InitialLife: config.Duration * (0.8 + rng.Float64()*0.4),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 1,
			ZLayer:      ZLayerGround,
		}
	}
}

// generateEmbers creates glowing fire embers that rise and fade.
// Phase 45: Embers use rising behavior with shrink-on-rise, rendering at sky level.
func (g *Generator) generateEmbers(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	emberColors := []color.RGBA{
		{255, 100, 0, 255},   // Bright orange
		{255, 150, 50, 255},  // Light orange
		{200, 50, 0, 255},    // Dark red
		{255, 200, 100, 255}, // Yellow-orange
	}

	for i := range system.Particles {
		// Embers rise upward with some horizontal drift
		angle := -math.Pi/2 + (rng.Float64()*2-1)*math.Pi/8
		speed := config.SpreadY * (0.3 + rng.Float64()*0.4)
		size := config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize)

		// Physics config for rising + air resistance
		physics := PhysicsConfig{
			Gravity:       config.Gravity,
			AirResistance: 0.3, // Moderate air resistance
		}

		system.Particles[i] = Particle{
			X:           (rng.Float64()*2 - 1) * 4, // Scaled 2× spread
			Y:           0,
			VX:          math.Cos(angle) * speed * 0.5,
			VY:          math.Sin(angle) * speed,
			Color:       emberColors[rng.Intn(len(emberColors))],
			Size:        size,
			InitialSize: size,
			Life:        1.0,
			InitialLife: config.Duration * (0.6 + rng.Float64()*0.8),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 3,
			ZLayer:      ZLayerSky,
			Behavior:    BehaviorRising | BehaviorAirResistance | BehaviorShrinkOnRise,
			Physics:     physics,
		}
	}
}

// generateSparkles creates magical sparkles that orbit and trail.
// Phase 45: Sparkles render above entities with orbital movement.
func (g *Generator) generateSparkles(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Use palette colors for genre-specific sparkles
	sparkleColors := []color.RGBA{
		color.RGBAModel.Convert(pal.Primary).(color.RGBA),
		color.RGBAModel.Convert(pal.Secondary).(color.RGBA),
		color.RGBAModel.Convert(pal.Colors[rng.Intn(len(pal.Colors)/2)]).(color.RGBA),
		{255, 255, 255, 255}, // White sparkle
	}

	for i := range system.Particles {
		// Random starting position in a circle (scaled 2× for larger viewport)
		angle := rng.Float64() * 2 * math.Pi
		radius := rng.Float64() * 60.0 // Scaled 2×

		// Orbital physics
		physics := PhysicsConfig{
			AttractorX:  0,
			AttractorY:  0,
			OrbitRadius: 80.0, // Scaled 2×
			OrbitSpeed:  2.0 + rng.Float64()*2.0,
		}

		size := config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize)*0.6

		system.Particles[i] = Particle{
			X:           math.Cos(angle) * radius,
			Y:           math.Sin(angle) * radius,
			VX:          -math.Sin(angle) * physics.OrbitSpeed * radius,
			VY:          math.Cos(angle) * physics.OrbitSpeed * radius,
			Color:       sparkleColors[rng.Intn(len(sparkleColors))],
			Size:        size,
			InitialSize: size,
			Life:        1.0,
			InitialLife: config.Duration * (0.8 + rng.Float64()*0.4),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 4,
			ZLayer:      ZLayerAbove,
			Behavior:    BehaviorOrbit,
			Physics:     physics,
		}
	}
}

// generateSmokePlume creates billowing smoke clouds.
// Phase 45: Smoke plumes use rising + air resistance with shrink-on-rise, rendering at sky level.
func (g *Generator) generateSmokePlume(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	plumeColors := []color.RGBA{
		{80, 80, 80, 180},
		{100, 100, 100, 160},
		{120, 120, 120, 140},
		{60, 60, 60, 200},
	}

	for i := range system.Particles {
		// Plumes billow upward and outward
		angle := -math.Pi/2 + (rng.Float64()*2-1)*math.Pi/3
		speed := config.SpreadY * (0.4 + rng.Float64()*0.6)
		size := config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize)*1.5 // Larger

		physics := PhysicsConfig{
			Gravity:       -50.0, // Slight upward force
			AirResistance: 0.4,   // High air resistance for slow billowing
		}

		system.Particles[i] = Particle{
			X:           (rng.Float64()*2 - 1) * 10, // Scaled 2×
			Y:           0,
			VX:          math.Cos(angle) * speed * 0.7,
			VY:          math.Sin(angle) * speed,
			Color:       plumeColors[rng.Intn(len(plumeColors))],
			Size:        size,
			InitialSize: size,
			Life:        1.0,
			InitialLife: config.Duration * (1.0 + rng.Float64()*0.5), // Longer life
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 1.5,
			ZLayer:      ZLayerSky,
			Behavior:    BehaviorRising | BehaviorAirResistance | BehaviorShrinkOnRise,
			Physics:     physics,
		}
	}
}

// generateDebris creates bouncing debris chunks.
// Phase 45: Debris uses gravity + bouncing with grow-on-fall, rendering at ground level.
func (g *Generator) generateDebris(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	debrisColors := []color.RGBA{
		{100, 80, 60, 255},   // Brown
		{120, 120, 120, 255}, // Gray
		{80, 60, 40, 255},    // Dark brown
		{140, 120, 100, 255}, // Light brown
	}

	for i := range system.Particles {
		// Debris shoots outward in all directions
		angle := rng.Float64() * 2 * math.Pi
		speed := config.SpreadX * (0.6 + rng.Float64()*0.8)

		// Ground level from config (if specified)
		groundY := 0.0
		if val, ok := config.Custom["groundY"]; ok {
			if gy, ok := val.(float64); ok {
				groundY = gy
			}
		}

		physics := PhysicsConfig{
			Gravity:       config.Gravity,
			GroundY:       groundY,
			BounceDamping: 0.4 + rng.Float64()*0.3, // Variable bounce
			AirResistance: 0.1,                     // Light air resistance
		}

		size := config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize)*1.2

		system.Particles[i] = Particle{
			X:           0,
			Y:           -rng.Float64() * 6, // Start slightly above origin, scaled 2×
			VX:          math.Cos(angle) * speed,
			VY:          math.Sin(angle) * speed,
			Color:       debrisColors[rng.Intn(len(debrisColors))],
			Size:        size,
			InitialSize: size,
			Life:        1.0,
			InitialLife: config.Duration * (0.8 + rng.Float64()*0.4),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 6, // Fast rotation
			ZLayer:      ZLayerGround,
			Behavior:    BehaviorGravity | BehaviorBounce | BehaviorAirResistance | BehaviorGrowOnFall,
			Physics:     physics,
		}
	}
}

// Validate implements the procgen.Generator interface.
func (g *Generator) Validate(result interface{}) error {
	system, ok := result.(*ParticleSystem)
	if !ok {
		return fmt.Errorf("result is not a *ParticleSystem")
	}

	if system == nil {
		return fmt.Errorf("particle system is nil")
	}

	if len(system.Particles) == 0 {
		return fmt.Errorf("particle system has no particles")
	}

	// Validate that particles have reasonable values
	for i, p := range system.Particles {
		if p.Size <= 0 {
			return fmt.Errorf("particle %d has invalid size: %f", i, p.Size)
		}
		if p.InitialLife <= 0 {
			return fmt.Errorf("particle %d has invalid initial life: %f", i, p.InitialLife)
		}
		if p.Life < 0 || p.Life > 1 {
			return fmt.Errorf("particle %d has invalid life: %f", i, p.Life)
		}
	}

	return nil
}
