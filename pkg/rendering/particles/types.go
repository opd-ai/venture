// Package particles provides particle type definitions.
// This file defines particle data structures, emitter types, and
// behavior parameters used by the particle generator.
package particles

import (
	"fmt"
	"image/color"
)

// ParticleType represents different types of particle effects.
type ParticleType int

const (
	// ParticleSpark represents quick, bright particles for impacts
	ParticleSpark ParticleType = iota + 1
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
	// ParticleEmber represents glowing fire embers that rise and fade
	ParticleEmber
	// ParticleSparkle represents magical sparkles that orbit and trail
	ParticleSparkle
	// ParticleSmokePlume represents billowing smoke clouds
	ParticleSmokePlume
	// ParticleDebris represents bouncing debris chunks
	ParticleDebris
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
	case ParticleEmber:
		return "ember"
	case ParticleSparkle:
		return "sparkle"
	case ParticleSmokePlume:
		return "smoke_plume"
	case ParticleDebris:
		return "debris"
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
	// Phase 45: Scaled 2× (4-16px range for 64×64 tiles)
	MinSize float64
	MaxSize float64

	// ZLayer defines the rendering z-order (Phase 45)
	// Ground-level particles (dust, debris) use ZLayerGround
	// Particles above entities (magic, sparkles) use ZLayerAbove
	ZLayer ZLayer

	// Custom parameters for specific particle types
	Custom map[string]interface{}
}

// ZLayer defines the rendering z-order for particles.
// Lower values render behind higher values.
type ZLayer int

const (
	// ZLayerGround is for ground-level particles (dust, debris, blood splatters)
	ZLayerGround ZLayer = 0
	// ZLayerEntity is the default layer, at entity level
	ZLayerEntity ZLayer = 1
	// ZLayerAbove is for particles above entities (magic, sparkles)
	ZLayerAbove ZLayer = 2
	// ZLayerSky is for high-altitude particles (embers, smoke plumes)
	ZLayerSky ZLayer = 3
)

// DefaultConfig returns a default particle configuration.
// Phase 45: Particle sizes scaled 2× for 64×64 tile compatibility.
func DefaultConfig() Config {
	return Config{
		Type:     ParticleSpark,
		Count:    20,
		GenreID:  "fantasy",
		Seed:     0,
		Duration: 1.0,
		SpreadX:  10.0, // Scaled 2× for larger viewport
		SpreadY:  10.0, // Scaled 2× for larger viewport
		Gravity:  0.0,
		MinSize:  4.0,  // 4-16px range per PLAN
		MaxSize:  16.0, // 4-16px range per PLAN
		ZLayer:   ZLayerEntity,
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

	// Color stored as RGBA to avoid conversion overhead
	// Performance: Eliminates color.RGBAModel.Convert() allocations per particle
	Color color.RGBA

	// Size in pixels
	// Phase 45: Scaled 2× for 64×64 tiles
	Size float64

	// InitialSize stores the starting size for top-down scaling (Phase 45)
	// Rising particles shrink as they rise; falling particles grow before impact
	InitialSize float64

	// Life remaining (0.0 = dead, 1.0 = just spawned)
	Life float64

	// Initial life for fade calculations
	InitialLife float64

	// Rotation in radians
	Rotation float64

	// Rotation velocity
	RotationVel float64

	// ZLayer defines the rendering z-order (Phase 45)
	// Higher values render on top of lower values
	ZLayer ZLayer

	// Behavior flags (new for Phase 14.3)
	Behavior ParticleBehavior

	// Physics configuration (new for Phase 14.3)
	Physics PhysicsConfig
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

		// Apply physics behaviors (new for Phase 14.3)
		if p.Behavior != BehaviorNone {
			ApplyPhysics(p, p.Behavior, p.Physics, deltaTime)
		} else {
			// Legacy behavior: basic gravity from config
			p.VY += ps.Config.Gravity * deltaTime
		}

		// Update position
		p.X += p.VX * deltaTime
		p.Y += p.VY * deltaTime

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
// Note: This method allocates a new slice. For hot paths, use VisitAliveParticles instead.
func (ps *ParticleSystem) GetAliveParticles() []Particle {
	// Performance: O(N) with allocation. Not suitable for hot paths (e.g., per-frame rendering).
	alive := make([]Particle, 0, len(ps.Particles))
	for i := range ps.Particles {
		if ps.Particles[i].Life > 0 {
			alive = append(alive, ps.Particles[i])
		}
	}
	return alive
}

// VisitAliveParticles calls the visitor function for each alive particle.
// This is a zero-allocation alternative to GetAliveParticles for hot paths.
// The visitor receives a pointer to the particle; do not store it.
func (ps *ParticleSystem) VisitAliveParticles(visitor func(p *Particle)) {
	for i := range ps.Particles {
		if ps.Particles[i].Life > 0 {
			visitor(&ps.Particles[i])
		}
	}
}
