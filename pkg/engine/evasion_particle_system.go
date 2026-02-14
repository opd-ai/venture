// Package engine provides the EvasionParticleSystem for visual dodge feedback.
// This system connects CombatSystem evasion events with ParticleSystem to spawn
// genre-aware particle effects when entities dodge attacks, giving players
// immediate visual feedback that their evasion stats are working.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// EvasionParticleSystem spawns particle effects when attacks are evaded.
// It connects CombatSystem evasion checks with ParticleSystem to provide
// visual feedback for successful dodges, with genre-aware particle colors.
type EvasionParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	particleCount int
	spreadFactor  float64
}

// NewEvasionParticleSystem creates a new evasion particle system.
func NewEvasionParticleSystem(world *World, seed int64) *EvasionParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "evasion_particle")
		logEntry.Debug("evasion particle system created")
	}

	return &EvasionParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		particleCount: 6,    // Light particle count for quick dodge effect
		spreadFactor:  50.0, // Spread radius for dodge particles
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *EvasionParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *EvasionParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *EvasionParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnEvasion, no per-frame processing needed
}

// OnEvasion is called when an attack is evaded to spawn particles.
// This method should be registered as a callback with the CombatSystem.
//
// Parameters:
//   - attacker: The entity that attacked
//   - target: The entity that evaded the attack
//   - evasionChance: The evasion chance that resulted in the dodge
func (s *EvasionParticleSystem) OnEvasion(
	attacker, target *Entity,
	evasionChance float64,
) {
	if s.particleSystem == nil || s.world == nil || target == nil {
		return
	}

	// Get target position for particle spawn
	pos := target.GetPosition()
	if pos == nil {
		return
	}

	// Spawn evasion particles at target location
	s.spawnEvasionParticles(pos.X, pos.Y, evasionChance)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"attacker_id":    attacker.ID,
			"target_id":      target.ID,
			"evasion_chance": evasionChance,
			"x":              pos.X,
			"y":              pos.Y,
		}).Debug("evasion particles spawned")
	}
}

// spawnEvasionParticles creates the evasion particle effect.
func (s *EvasionParticleSystem) spawnEvasionParticles(
	x, y, evasionChance float64,
) {
	// Scale particle count slightly with high evasion
	count := s.particleCount
	if evasionChance > 0.3 {
		count = int(float64(count) * 1.2)
	}
	if evasionChance > 0.5 {
		count = int(float64(count) * 1.3)
	}
	// Cap at reasonable maximum
	if count > 12 {
		count = 12
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
		Duration: 0.35, // Quick burst for dodge effect
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -50.0, // Upward drift to suggest agility
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   make(map[string]interface{}),
	}

	// Mark as evasion effect for color selection
	config.Custom["evasion_effect"] = true
	config.Custom["evasion_chance"] = evasionChance

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getParticleTypeForGenre returns the appropriate particle type for the genre.
func (s *EvasionParticleSystem) getParticleTypeForGenre() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Magical agility
	case "scifi":
		return particles.ParticleSpark // Energy displacement
	case "horror":
		return particles.ParticleSmoke // Shadowy dodge
	case "cyberpunk":
		return particles.ParticleSpark // Holographic afterimage
	case "postapoc":
		return particles.ParticleDust // Quick movement dust
	default:
		return particles.ParticleSparkle
	}
}

// SpawnEvasionEffect allows external systems to trigger evasion particles directly.
// Useful for abilities that grant temporary dodge visualization.
func (s *EvasionParticleSystem) SpawnEvasionEffect(x, y float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.spawnEvasionParticles(x, y, 0.2) // Default moderate evasion
}
