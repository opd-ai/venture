package particles

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// Generator creates procedural particle systems.
type Generator struct {
	paletteGen *palette.Generator
}

// NewGenerator creates a new particle system generator.
func NewGenerator() *Generator {
	return &Generator{
		paletteGen: palette.NewGenerator(),
	}
}

// Generate creates a particle system from the given configuration.
func (g *Generator) Generate(config Config) (*ParticleSystem, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create RNG from seed
	rng := rand.New(rand.NewSource(config.Seed))

	// Generate color palette for genre
	pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	// Create particle system
	system := &ParticleSystem{
		Particles:   make([]Particle, config.Count),
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
	default:
		return nil, fmt.Errorf("unknown particle type: %d", config.Type)
	}

	return system, nil
}

// generateSparks creates bright, quick-moving spark particles.
func (g *Generator) generateSparks(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	colors := []color.Color{
		color.RGBA{255, 255, 200, 255}, // Bright yellow
		color.RGBA{255, 200, 100, 255}, // Orange
		color.RGBA{255, 255, 255, 255}, // White
	}

	for i := range system.Particles {
		angle := rng.Float64() * 2 * math.Pi
		speed := rng.Float64() * config.SpreadX

		system.Particles[i] = Particle{
			X:           0,
			Y:           0,
			VX:          math.Cos(angle) * speed,
			VY:          math.Sin(angle) * speed,
			Color:       colors[rng.Intn(len(colors))],
			Size:        config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize),
			Life:        1.0,
			InitialLife: config.Duration * (0.5 + rng.Float64()*0.5),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 5,
		}
	}
}

// generateSmoke creates soft, slowly rising smoke particles.
func (g *Generator) generateSmoke(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	smokeColors := []color.Color{
		color.RGBA{100, 100, 100, 200},
		color.RGBA{120, 120, 120, 180},
		color.RGBA{80, 80, 80, 220},
	}

	for i := range system.Particles {
		angle := (rng.Float64()*2 - 1) * math.Pi / 4 // Upward cone
		speed := rng.Float64() * config.SpreadY * 0.5

		system.Particles[i] = Particle{
			X:           (rng.Float64()*2 - 1) * 2,
			Y:           0,
			VX:          math.Cos(angle-math.Pi/2) * speed,
			VY:          math.Sin(angle-math.Pi/2) * speed,
			Color:       smokeColors[rng.Intn(len(smokeColors))],
			Size:        config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize),
			Life:        1.0,
			InitialLife: config.Duration * (0.8 + rng.Float64()*0.4),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 2,
		}
	}
}

// generateMagic creates glowing magical particles.
func (g *Generator) generateMagic(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	// Use palette colors for magic
	magicColors := []color.Color{
		pal.Primary,
		pal.Secondary,
		pal.Colors[rng.Intn(len(pal.Colors)/2)],
	}

	for i := range system.Particles {
		angle := rng.Float64() * 2 * math.Pi
		speed := rng.Float64() * config.SpreadX * 0.7

		system.Particles[i] = Particle{
			X:           0,
			Y:           0,
			VX:          math.Cos(angle) * speed,
			VY:          math.Sin(angle) * speed,
			Color:       magicColors[rng.Intn(len(magicColors))],
			Size:        config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize),
			Life:        1.0,
			InitialLife: config.Duration * (0.7 + rng.Float64()*0.6),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 3,
		}
	}
}

// generateFlame creates fire-like particles.
func (g *Generator) generateFlame(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	flameColors := []color.Color{
		color.RGBA{255, 100, 0, 255},   // Orange
		color.RGBA{255, 200, 0, 255},   // Yellow
		color.RGBA{200, 50, 0, 255},    // Red
		color.RGBA{255, 255, 100, 255}, // Bright yellow
	}

	for i := range system.Particles {
		// Flames rise upward with some spread
		angle := -math.Pi/2 + (rng.Float64()*2-1)*math.Pi/6
		speed := config.SpreadY * (0.5 + rng.Float64()*0.5)

		system.Particles[i] = Particle{
			X:           (rng.Float64()*2 - 1) * 3,
			Y:           0,
			VX:          math.Cos(angle) * speed,
			VY:          math.Sin(angle) * speed,
			Color:       flameColors[rng.Intn(len(flameColors))],
			Size:        config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize),
			Life:        1.0,
			InitialLife: config.Duration * (0.3 + rng.Float64()*0.4),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 4,
		}
	}
}

// generateBlood creates blood splatter particles.
func (g *Generator) generateBlood(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	bloodColors := []color.Color{
		color.RGBA{150, 0, 0, 255},
		color.RGBA{180, 20, 20, 255},
		color.RGBA{120, 0, 0, 255},
	}

	for i := range system.Particles {
		angle := rng.Float64() * 2 * math.Pi
		speed := rng.Float64() * config.SpreadX * 0.8

		system.Particles[i] = Particle{
			X:           0,
			Y:           0,
			VX:          math.Cos(angle) * speed,
			VY:          math.Sin(angle) * speed,
			Color:       bloodColors[rng.Intn(len(bloodColors))],
			Size:        config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize),
			Life:        1.0,
			InitialLife: config.Duration * (0.6 + rng.Float64()*0.4),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 2,
		}
	}
}

// generateDust creates small dust particles.
func (g *Generator) generateDust(system *ParticleSystem, pal *palette.Palette, rng *rand.Rand, config Config) {
	dustColors := []color.Color{
		color.RGBA{160, 140, 120, 180},
		color.RGBA{140, 120, 100, 160},
		color.RGBA{180, 160, 140, 170},
	}

	for i := range system.Particles {
		angle := rng.Float64() * 2 * math.Pi
		speed := rng.Float64() * config.SpreadX * 0.3

		system.Particles[i] = Particle{
			X:           (rng.Float64()*2 - 1) * 5,
			Y:           (rng.Float64()*2 - 1) * 5,
			VX:          math.Cos(angle) * speed,
			VY:          math.Sin(angle)*speed - 1, // Slight upward drift
			Color:       dustColors[rng.Intn(len(dustColors))],
			Size:        config.MinSize + rng.Float64()*(config.MaxSize-config.MinSize)*0.5,
			Life:        1.0,
			InitialLife: config.Duration * (0.8 + rng.Float64()*0.4),
			Rotation:    rng.Float64() * 2 * math.Pi,
			RotationVel: (rng.Float64()*2 - 1) * 1,
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
