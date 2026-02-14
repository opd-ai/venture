// Package engine provides the StatusEffectDamageParticleSystem for visual DOT feedback.
// This system connects StatusEffectSystem with ParticleSystem to spawn genre-aware
// particle effects when damage-over-time effects tick, enhancing combat visual feedback.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// StatusEffectDamageParticleSystem spawns particle effects when status effects tick.
// It provides visual feedback for burning (flames), poison (drips), and regeneration
// (healing sparkles) with genre-aware particle colors.
type StatusEffectDamageParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration per effect type
	burningParticleCount int
	poisonParticleCount  int
	regenParticleCount   int
	spreadFactor         float64
}

// NewStatusEffectDamageParticleSystem creates a new status effect damage particle system.
func NewStatusEffectDamageParticleSystem(world *World, seed int64) *StatusEffectDamageParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "status_effect_damage_particle")
		logEntry.Debug("status effect damage particle system created")
	}

	return &StatusEffectDamageParticleSystem{
		world:                world,
		genreID:              "fantasy",
		seed:                 seed,
		rng:                  rand.New(rand.NewSource(seed)),
		logger:               logEntry,
		burningParticleCount: 8,
		poisonParticleCount:  6,
		regenParticleCount:   10,
		spreadFactor:         40.0,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *StatusEffectDamageParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *StatusEffectDamageParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *StatusEffectDamageParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnStatusEffectTick, no per-frame processing needed
}

// OnStatusEffectTick is called when a status effect deals damage or healing.
// This method should be registered as a callback with the StatusEffectSystem.
func (s *StatusEffectDamageParticleSystem) OnStatusEffectTick(entity *Entity, effectType string, magnitude float64) {
	if s.particleSystem == nil || s.world == nil || entity == nil {
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	switch effectType {
	case "burning":
		s.spawnBurningParticles(pos.X, pos.Y, magnitude)
	case "poisoned":
		s.spawnPoisonParticles(pos.X, pos.Y, magnitude)
	case "regeneration":
		s.spawnRegenParticles(pos.X, pos.Y, magnitude)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"effect_type": effectType,
			"magnitude":   magnitude,
			"x":           pos.X,
			"y":           pos.Y,
		}).Debug("status effect damage particles spawned")
	}
}

// spawnBurningParticles creates flame particles for burning damage.
func (s *StatusEffectDamageParticleSystem) spawnBurningParticles(x, y, damage float64) {
	count := s.burningParticleCount
	if damage > 10 {
		count = int(float64(count) * 1.5)
	}
	if count > 20 {
		count = 20
	}

	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(damage*10)

	config := particles.Config{
		Type:     particles.ParticleFlame,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.5,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor * 0.5,
		Gravity:  -60.0, // Flames rise
		MinSize:  3.0,
		MaxSize:  6.0,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"burning_tick": true},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// spawnPoisonParticles creates drip particles for poison damage.
func (s *StatusEffectDamageParticleSystem) spawnPoisonParticles(x, y, damage float64) {
	count := s.poisonParticleCount
	if damage > 10 {
		count = int(float64(count) * 1.3)
	}
	if count > 15 {
		count = 15
	}

	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(damage*10) + 1000

	config := particles.Config{
		Type:     particles.ParticleMagic, // Magic particles with green tint
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.6,
		SpreadX:  s.spreadFactor * 0.8,
		SpreadY:  s.spreadFactor * 0.5,
		Gravity:  40.0, // Drips fall
		MinSize:  2.0,
		MaxSize:  5.0,
		ZLayer:   particles.ZLayerEntity,
		Custom:   map[string]interface{}{"poison_tick": true, "color_override": "green"},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// spawnRegenParticles creates sparkle particles for healing over time.
func (s *StatusEffectDamageParticleSystem) spawnRegenParticles(x, y, healing float64) {
	count := s.regenParticleCount
	if healing > 10 {
		count = int(float64(count) * 1.5)
	}
	if count > 25 {
		count = 25
	}

	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(healing*10) + 2000

	config := particles.Config{
		Type:     particles.ParticleSparkle,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.7,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -30.0, // Sparkles float up gently
		MinSize:  2.0,
		MaxSize:  4.0,
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"regen_tick": true, "color_override": "gold"},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}
