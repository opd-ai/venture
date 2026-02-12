// Package engine provides the CriticalHitParticleSystem for visual critical hit feedback.
// This system connects CombatSystem with ParticleSystem to spawn genre-aware particle
// effects when critical hits occur, enhancing combat visual feedback.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// CriticalHitParticleSystem spawns particle effects for critical hits.
// It connects CombatSystem and ParticleSystem to provide visual feedback
// with genre-aware particle colors and behaviors.
type CriticalHitParticleSystem struct {
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

// NewCriticalHitParticleSystem creates a new critical hit particle system.
func NewCriticalHitParticleSystem(world *World, seed int64) *CriticalHitParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "critical_hit_particle")
		logEntry.Debug("critical hit particle system created")
	}

	return &CriticalHitParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		particleCount: 20,    // Default particle count for crit effects
		spreadFactor:  180.0, // Default spread for crit particles
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *CriticalHitParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *CriticalHitParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *CriticalHitParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnCriticalHit, no per-frame processing needed
}

// OnCriticalHit is called when a critical hit occurs to spawn particle effects.
// This method should be registered as a callback with the CombatSystem.
func (s *CriticalHitParticleSystem) OnCriticalHit(attacker, target *Entity, damage float64) {
	if s.particleSystem == nil || s.world == nil || target == nil {
		return
	}

	// Get target position for particle spawn
	pos := target.GetPosition()
	if pos == nil {
		return
	}

	// Spawn critical hit particles at target location
	s.spawnCritParticles(pos.X, pos.Y, damage)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id": target.ID,
			"damage":    damage,
			"x":         pos.X,
			"y":         pos.Y,
		}).Debug("critical hit particles spawned")
	}
}

// spawnCritParticles creates the critical hit particle effect.
func (s *CriticalHitParticleSystem) spawnCritParticles(x, y, damage float64) {
	// Scale particle count based on damage (more particles for bigger crits)
	count := s.particleCount
	if damage > 50 {
		count = int(float64(count) * 1.5)
	}
	if damage > 100 {
		count = int(float64(count) * 2.0)
	}
	// Cap at reasonable maximum
	if count > 60 {
		count = 60
	}

	// Use deterministic seed offset for this specific crit
	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(damage*10)

	// Create sparkle config for critical hit effect
	config := particles.Config{
		Type:     particles.ParticleSparkle,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.6,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -80.0, // Float upward slightly
		MinSize:  3.0,
		MaxSize:  7.0,
		Custom:   make(map[string]interface{}),
	}

	// Mark as critical hit for potential special rendering
	config.Custom["crit_effect"] = true

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// SpawnCritEffect allows external systems to trigger crit particles directly.
// This is useful for projectile crits or spell crits.
func (s *CriticalHitParticleSystem) SpawnCritEffect(x, y, damage float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.spawnCritParticles(x, y, damage)
}
