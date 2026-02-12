// Package engine provides the DamageResistanceParticleSystem for visual resistance feedback.
// This system connects CombatSystem with ParticleSystem to spawn genre-aware particle
// effects when damage is significantly reduced by resistances, giving players visual
// feedback that their resistances are effective.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// DamageResistanceParticleSystem spawns particle effects when damage is resisted.
// It connects CombatSystem and ParticleSystem to provide visual feedback
// when entities resist significant damage, with genre-aware particle colors.
type DamageResistanceParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	particleCount    int
	spreadFactor     float64
	resistThreshold  float64 // Minimum resistance % to trigger particles (0.0-1.0)
	minDamageReduced float64 // Minimum damage reduced to trigger particles
}

// NewDamageResistanceParticleSystem creates a new damage resistance particle system.
func NewDamageResistanceParticleSystem(world *World, seed int64) *DamageResistanceParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "damage_resistance_particle")
		logEntry.Debug("damage resistance particle system created")
	}

	return &DamageResistanceParticleSystem{
		world:            world,
		seed:             seed,
		rng:              rand.New(rand.NewSource(seed)),
		logger:           logEntry,
		particleCount:    12,    // Moderate particle count for resist effects
		spreadFactor:     100.0, // Spread radius for resist particles
		resistThreshold:  0.25,  // Trigger at 25%+ resistance
		minDamageReduced: 5.0,   // Minimum 5 damage must be reduced
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *DamageResistanceParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *DamageResistanceParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetResistThreshold sets the minimum resistance percentage to trigger particles.
// Value should be between 0.0 and 1.0 (e.g., 0.25 = 25% resistance).
func (s *DamageResistanceParticleSystem) SetResistThreshold(threshold float64) {
	if threshold < 0.0 {
		threshold = 0.0
	}
	if threshold > 1.0 {
		threshold = 1.0
	}
	s.resistThreshold = threshold
}

// Update processes entities (no-op for this callback-driven system).
func (s *DamageResistanceParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnDamageResisted, no per-frame processing needed
}

// OnDamageResisted is called when damage is reduced by resistances to spawn particles.
// This method should be registered as a callback with the CombatSystem.
//
// Parameters:
//   - target: The entity that resisted the damage
//   - damageType: The type of damage that was resisted
//   - originalDamage: Damage before resistance reduction
//   - finalDamage: Damage after resistance reduction
//   - resistance: The resistance value that was applied (0.0-1.0)
func (s *DamageResistanceParticleSystem) OnDamageResisted(
	target *Entity,
	damageType combat.DamageType,
	originalDamage, finalDamage, resistance float64,
) {
	if s.particleSystem == nil || s.world == nil || target == nil {
		return
	}

	// Check if resistance meets threshold
	if resistance < s.resistThreshold {
		return
	}

	// Check if enough damage was actually reduced
	damageReduced := originalDamage - finalDamage
	if damageReduced < s.minDamageReduced {
		return
	}

	// Get target position for particle spawn
	pos := target.GetPosition()
	if pos == nil {
		return
	}

	// Spawn resistance particles at target location
	s.spawnResistParticles(pos.X, pos.Y, damageType, resistance, damageReduced)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id":      target.ID,
			"damage_type":    damageType.String(),
			"original":       originalDamage,
			"final":          finalDamage,
			"resistance":     resistance,
			"damage_reduced": damageReduced,
			"x":              pos.X,
			"y":              pos.Y,
		}).Debug("damage resistance particles spawned")
	}
}

// spawnResistParticles creates the resistance particle effect based on damage type.
func (s *DamageResistanceParticleSystem) spawnResistParticles(
	x, y float64,
	damageType combat.DamageType,
	resistance, damageReduced float64,
) {
	// Scale particle count based on resistance strength and damage reduced
	count := s.particleCount
	if resistance > 0.5 {
		count = int(float64(count) * 1.3)
	}
	if resistance > 0.75 {
		count = int(float64(count) * 1.5)
	}
	if damageReduced > 20 {
		count = int(float64(count) * 1.2)
	}
	// Cap at reasonable maximum
	if count > 30 {
		count = 30
	}

	// Use deterministic seed offset based on position and damage type
	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(damageType)*100

	// Select particle type based on damage type
	particleType := s.getParticleTypeForDamage(damageType)

	// Create particle config
	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.5, // Quick fade for resist effect
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -40.0, // Gentle upward drift
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	// Mark resistance type for color selection
	config.Custom["resist_type"] = damageType.String()
	config.Custom["resist_effect"] = true

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getParticleTypeForDamage returns the appropriate particle type for the damage type.
func (s *DamageResistanceParticleSystem) getParticleTypeForDamage(damageType combat.DamageType) particles.ParticleType {
	switch damageType {
	case combat.DamageFire:
		return particles.ParticleEmber // Glowing embers for fire resist
	case combat.DamageIce:
		return particles.ParticleSparkle // Crystalline sparkles for ice resist
	case combat.DamageLightning:
		return particles.ParticleSpark // Electric sparks for lightning resist
	case combat.DamagePoison:
		return particles.ParticleSmoke // Dissipating poison clouds
	case combat.DamageMagical:
		return particles.ParticleMagic // Magical glow for magic resist
	case combat.DamagePhysical:
		return particles.ParticleDust // Dust puffs for physical resist
	default:
		return particles.ParticleSpark
	}
}

// SpawnResistEffect allows external systems to trigger resistance particles directly.
// Useful for spell resistance effects or shield blocks.
func (s *DamageResistanceParticleSystem) SpawnResistEffect(
	x, y float64,
	damageType combat.DamageType,
	resistance float64,
) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.spawnResistParticles(x, y, damageType, resistance, s.minDamageReduced)
}
