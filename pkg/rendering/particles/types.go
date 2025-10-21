package particles

import (
	"fmt"
	"image/color"
)

// ParticleType represents different types of particle effects.
type ParticleType int

const (
	// ParticleSpark represents quick, bright particles for impacts
	ParticleSpark ParticleType = iota
	// ParticleSmoke represents soft, fading smoke particles
	ParticleSmoke
	// ParticleMagic represents glowing magical particles
	ParticleMagic
	// ParticleFlame represents fire-like particles
	ParticleFlame
	// ParticleBlood represents blood splatter particles
	ParticleBlood
	// ParticleDust represents small dust particles
	ParticleDust
)

// String returns the string representation of a particle type.
func (p ParticleType) String() string {
	switch p {
	case ParticleSpark:
		return "spark"
	case ParticleSmoke:
		return "smoke"
	case ParticleMagic:
		return "magic"
	case ParticleFlame:
		return "flame"
	case ParticleBlood:
		return "blood"
	case ParticleDust:
		return "dust"
	default:
		return "unknown"
	}
}

// Config contains parameters for particle system generation.
type Config struct {
	// Type of particle effect
	Type ParticleType

	// Count is the number of particles to generate
	Count int

	// GenreID for color selection
	GenreID string

	// Seed for deterministic generation
	Seed int64

	// Duration in seconds that particles live
	Duration float64

	// SpreadX and SpreadY control initial velocity spread
	SpreadX float64
	SpreadY float64

	// Gravity affects particle motion (positive = downward)
	Gravity float64

	// Size range for particles (min and max)
	MinSize float64
	MaxSize float64

	// Custom parameters for specific particle types
	Custom map[string]interface{}
}

// DefaultConfig returns a default particle configuration.
func DefaultConfig() Config {
	return Config{
		Type:     ParticleSpark,
		Count:    20,
		GenreID:  "fantasy",
		Seed:     0,
		Duration: 1.0,
		SpreadX:  5.0,
		SpreadY:  5.0,
		Gravity:  0.0,
		MinSize:  1.0,
		MaxSize:  3.0,
		Custom:   make(map[string]interface{}),
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.Count <= 0 {
		return fmt.Errorf("count must be positive, got %d", c.Count)
	}
	if c.Count > 10000 {
		return fmt.Errorf("count too large (max 10000), got %d", c.Count)
	}
	if c.GenreID == "" {
		return fmt.Errorf("genreID cannot be empty")
	}
	if c.Duration <= 0 {
		return fmt.Errorf("duration must be positive, got %f", c.Duration)
	}
	if c.MinSize <= 0 {
		return fmt.Errorf("minSize must be positive, got %f", c.MinSize)
	}
	if c.MaxSize < c.MinSize {
		return fmt.Errorf("maxSize (%f) must be >= minSize (%f)", c.MaxSize, c.MinSize)
	}
	return nil
}

// Particle represents a single particle in a particle system.
type Particle struct {
	// Position
	X, Y float64

	// Velocity
	VX, VY float64

	// Color
	Color color.Color

	// Size in pixels
	Size float64

	// Life remaining (0.0 = dead, 1.0 = just spawned)
	Life float64

	// Initial life for fade calculations
	InitialLife float64

	// Rotation in radians
	Rotation float64

	// Rotation velocity
	RotationVel float64
}

// ParticleSystem represents a collection of particles.
type ParticleSystem struct {
	// Particles in the system
	Particles []Particle

	// Type of particle system
	Type ParticleType

	// Configuration used to generate this system
	Config Config

	// Time elapsed since creation
	ElapsedTime float64
}

// Update updates all particles in the system based on delta time.
func (ps *ParticleSystem) Update(deltaTime float64) {
	ps.ElapsedTime += deltaTime

	for i := range ps.Particles {
		p := &ps.Particles[i]

		// Update life
		p.Life -= deltaTime / p.InitialLife

		// Update position
		p.X += p.VX * deltaTime
		p.Y += p.VY * deltaTime

		// Apply gravity
		p.VY += ps.Config.Gravity * deltaTime

		// Update rotation
		p.Rotation += p.RotationVel * deltaTime
	}
}

// IsAlive returns true if any particles are still alive.
func (ps *ParticleSystem) IsAlive() bool {
	for i := range ps.Particles {
		if ps.Particles[i].Life > 0 {
			return true
		}
	}
	return false
}

// GetAliveParticles returns only the particles that are still alive.
func (ps *ParticleSystem) GetAliveParticles() []Particle {
	alive := make([]Particle, 0, len(ps.Particles))
	for i := range ps.Particles {
		if ps.Particles[i].Life > 0 {
			alive = append(alive, ps.Particles[i])
		}
	}
	return alive
}
