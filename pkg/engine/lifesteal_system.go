// Package engine provides the LifestealSystem for combat health restoration.
// This system connects CombatSystem with HealthComponent to heal attackers
// based on damage dealt when they have a positive Lifesteal stat.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// LifestealSystem heals attackers for a percentage of damage dealt.
// It connects CombatSystem damage events with HealthComponent healing
// and spawns genre-aware healing particle effects for visual feedback.
type LifestealSystem struct {
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

// NewLifestealSystem creates a new lifesteal system.
func NewLifestealSystem(world *World, seed int64) *LifestealSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "lifesteal")
		logEntry.Debug("lifesteal system created")
	}

	return &LifestealSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		particleCount: 8,    // Modest particle count for heal effect
		spreadFactor:  60.0, // Tighter spread than combat effects
	}
}

// SetParticleSystem sets the particle system for spawning heal effects.
func (s *LifestealSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *LifestealSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *LifestealSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnDamageDealt, no per-frame processing needed
}

// OnDamageDealt is called when damage is dealt to apply lifesteal healing.
// This method should be registered as a callback with the CombatSystem.
func (s *LifestealSystem) OnDamageDealt(attacker, target *Entity, damage float64) {
	if attacker == nil || damage <= 0 {
		return
	}

	// Get attacker's stats to check for lifesteal
	statsComp, hasStats := attacker.GetComponent("stats")
	if !hasStats || statsComp == nil {
		return
	}

	stats, ok := statsComp.(*StatsComponent)
	if !ok || stats.Lifesteal <= 0 {
		return
	}

	// Get attacker's health component for healing
	healthComp, hasHealth := attacker.GetComponent("health")
	if !hasHealth || healthComp == nil {
		return
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	// Calculate heal amount: damage * lifesteal percentage
	healAmount := damage * stats.Lifesteal

	// Cap healing at 25% of max health per hit to prevent excessive sustain
	maxHealPerHit := health.Max * 0.25
	if healAmount > maxHealPerHit {
		healAmount = maxHealPerHit
	}

	// Apply healing
	oldHealth := health.Current
	health.Heal(healAmount)
	actualHeal := health.Current - oldHealth

	// Spawn heal particles if meaningful healing occurred
	if actualHeal > 0 {
		s.spawnHealParticles(attacker, actualHeal)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"attacker_id":   attacker.ID,
			"damage":        damage,
			"lifesteal":     stats.Lifesteal,
			"heal_amount":   healAmount,
			"actual_healed": actualHeal,
		}).Debug("lifesteal healing applied")
	}
}

// spawnHealParticles creates healing particle effects at attacker location.
func (s *LifestealSystem) spawnHealParticles(attacker *Entity, healAmount float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	pos := attacker.GetPosition()
	if pos == nil {
		return
	}

	// Scale particle count based on heal amount
	count := s.particleCount
	if healAmount > 20 {
		count = int(float64(count) * 1.5)
	}
	if healAmount > 50 {
		count = int(float64(count) * 2.0)
	}
	if count > 24 {
		count = 24
	}

	// Use deterministic seed offset for this heal event
	effectSeed := s.seed + int64(pos.X*1000) + int64(pos.Y*1000) + int64(healAmount*10)

	// Create heal particle config - rising magical particles
	config := particles.Config{
		Type:     particles.ParticleMagic,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.8,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -100.0, // Float upward for healing effect
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	// Mark as lifesteal for potential special rendering
	config.Custom["lifesteal_effect"] = true

	s.particleSystem.SpawnParticles(s.world, config, pos.X, pos.Y)
}

// GetLifestealAmount calculates how much an entity would heal from damage.
// Useful for UI display or combat preview.
func (s *LifestealSystem) GetLifestealAmount(attacker *Entity, damage float64) float64 {
	if attacker == nil || damage <= 0 {
		return 0
	}

	statsComp, hasStats := attacker.GetComponent("stats")
	if !hasStats || statsComp == nil {
		return 0
	}

	stats, ok := statsComp.(*StatsComponent)
	if !ok || stats.Lifesteal <= 0 {
		return 0
	}

	return damage * stats.Lifesteal
}
