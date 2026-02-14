// Package engine provides the ShieldAbsorbParticleSystem for visual shield feedback.
// This system connects ShieldComponent damage absorption with ParticleSystem to spawn
// genre-aware particle effects when shields block incoming damage, giving players
// immediate visual feedback that their shields are active and protecting them.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ShieldAbsorbParticleSystem spawns particle effects when shields absorb damage.
// It connects ShieldComponent and ParticleSystem to provide visual feedback
// when entities' shields absorb damage, with genre-aware particle colors.
type ShieldAbsorbParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	particleCount   int
	spreadFactor    float64
	minAbsorbAmount float64 // Minimum damage absorbed to trigger particles
}

// NewShieldAbsorbParticleSystem creates a new shield absorb particle system.
func NewShieldAbsorbParticleSystem(world *World, seed int64) *ShieldAbsorbParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "shield_absorb_particle")
		logEntry.Debug("shield absorb particle system created")
	}

	return &ShieldAbsorbParticleSystem{
		world:           world,
		seed:            seed,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logEntry,
		particleCount:   8,    // Moderate particle count for shield effects
		spreadFactor:    80.0, // Spread radius for shield particles
		minAbsorbAmount: 1.0,  // Minimum 1 damage absorbed to trigger
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ShieldAbsorbParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ShieldAbsorbParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetMinAbsorbAmount sets the minimum damage absorbed to trigger particles.
func (s *ShieldAbsorbParticleSystem) SetMinAbsorbAmount(amount float64) {
	if amount < 0.0 {
		amount = 0.0
	}
	s.minAbsorbAmount = amount
}

// Update processes entities (no-op for this callback-driven system).
func (s *ShieldAbsorbParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnShieldAbsorb, no per-frame processing needed
}

// OnShieldAbsorb is called when a shield absorbs damage to spawn particles.
// This method should be registered as a callback with the CombatSystem.
//
// Parameters:
//   - target: The entity whose shield absorbed damage
//   - absorbed: Amount of damage absorbed by the shield
//   - remaining: Remaining shield amount after absorption
func (s *ShieldAbsorbParticleSystem) OnShieldAbsorb(
	target *Entity,
	absorbed, remaining float64,
) {
	if s.particleSystem == nil || s.world == nil || target == nil {
		return
	}

	// Check if enough damage was absorbed
	if absorbed < s.minAbsorbAmount {
		return
	}

	// Get target position for particle spawn
	pos := target.GetPosition()
	if pos == nil {
		return
	}

	// Spawn shield particles at target location
	s.spawnShieldParticles(pos.X, pos.Y, absorbed, remaining)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id": target.ID,
			"absorbed":  absorbed,
			"remaining": remaining,
			"x":         pos.X,
			"y":         pos.Y,
		}).Debug("shield absorb particles spawned")
	}
}

// spawnShieldParticles creates the shield particle effect based on absorption.
func (s *ShieldAbsorbParticleSystem) spawnShieldParticles(
	x, y, absorbed, remaining float64,
) {
	// Scale particle count based on absorbed damage
	count := s.particleCount
	if absorbed > 20 {
		count = int(float64(count) * 1.3)
	}
	if absorbed > 50 {
		count = int(float64(count) * 1.5)
	}
	// Cap at reasonable maximum
	if count > 20 {
		count = 20
	}

	// Use deterministic seed offset based on position
	effectSeed := s.seed + int64(x*1000) + int64(y*1000)

	// Select particle type based on genre
	particleType := s.getParticleTypeForGenre()

	// Create particle config
	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.4, // Quick flash for shield effect
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -30.0, // Gentle upward drift
		MinSize:  3.0,
		MaxSize:  6.0,
		Custom:   make(map[string]interface{}),
	}

	// Mark as shield effect for color selection
	config.Custom["shield_effect"] = true
	config.Custom["absorbed_damage"] = absorbed

	// Shield breaking visual when low
	if remaining < 10 {
		config.Custom["shield_breaking"] = true
		config.Count = int(float64(config.Count) * 1.5)
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getParticleTypeForGenre returns the appropriate particle type for the genre.
func (s *ShieldAbsorbParticleSystem) getParticleTypeForGenre() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Magical sparkles
	case "scifi":
		return particles.ParticleSpark // Energy field sparks
	case "horror":
		return particles.ParticleSmoke // Dark wispy protection
	case "cyberpunk":
		return particles.ParticleSpark // Holographic deflection
	case "postapoc":
		return particles.ParticleDust // Scavenged barrier dust
	default:
		return particles.ParticleSparkle
	}
}

// SpawnShieldEffect allows external systems to trigger shield particles directly.
// Useful for spell shields or other protective effects.
func (s *ShieldAbsorbParticleSystem) SpawnShieldEffect(x, y, absorbed float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.spawnShieldParticles(x, y, absorbed, 100.0)
}
