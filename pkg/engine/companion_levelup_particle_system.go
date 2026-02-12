// Package engine provides the CompanionLevelUpParticleSystem for visual feedback
// when companions level up. This connects the CompanionProgressionSystem with
// the ParticleSystem to spawn genre-aware celebration particles.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// CompanionLevelUpCallback is called when a companion levels up.
type CompanionLevelUpCallback func(entity *Entity, newLevel int)

// CompanionLevelUpParticleSystem spawns particle effects when companions level up.
// It provides visual feedback distinct from player level-ups (smaller, different colors).
type CompanionLevelUpParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration (smaller than player level-ups)
	baseParticleCount int
	spreadFactor      float64
	effectDuration    float64
}

// NewCompanionLevelUpParticleSystem creates a new companion level-up particle system.
func NewCompanionLevelUpParticleSystem(world *World, seed int64) *CompanionLevelUpParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "companion_levelup_particle")
		logEntry.Debug("companion level-up particle system created")
	}

	return &CompanionLevelUpParticleSystem{
		world:             world,
		seed:              seed,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		baseParticleCount: 15, // Smaller than player (30)
		spreadFactor:      100.0,
		effectDuration:    0.8,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *CompanionLevelUpParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *CompanionLevelUpParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *CompanionLevelUpParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// Callback-driven via OnCompanionLevelUp, no per-frame processing needed
}

// OnCompanionLevelUp is called when a companion levels up to spawn particles.
// Register this with CompanionProgressionSystem.AddLevelUpCallback().
func (s *CompanionLevelUpParticleSystem) OnCompanionLevelUp(entity *Entity, newLevel int) {
	if s.particleSystem == nil || s.world == nil || entity == nil {
		return
	}

	// Verify it's a companion
	if !entity.HasComponent("companion") {
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	s.spawnCompanionLevelUpParticles(pos.X, pos.Y, newLevel)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"new_level": newLevel,
			"x":         pos.X,
			"y":         pos.Y,
		}).Debug("companion level-up particles spawned")
	}
}

// spawnCompanionLevelUpParticles creates companion-specific level-up effects.
// These are smaller and use different particle types than player level-ups.
func (s *CompanionLevelUpParticleSystem) spawnCompanionLevelUpParticles(x, y float64, level int) {
	count := s.baseParticleCount
	// Scale slightly for milestone levels
	if level%5 == 0 {
		count = int(float64(count) * 1.3)
	}
	if count > 50 {
		count = 50
	}

	// Deterministic seed for this specific effect
	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(level*100) + 7777

	// Heart/sparkle particles for companion bond celebration
	sparkleConfig := particles.Config{
		Type:     particles.ParticleSparkle,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -80.0, // Gentle float upward
		MinSize:  3.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}
	sparkleConfig.Custom["companion_levelup"] = true
	sparkleConfig.Custom["level"] = level

	s.particleSystem.SpawnParticles(s.world, sparkleConfig, x, y)

	// Add small dust burst at feet
	dustConfig := particles.Config{
		Type:     particles.ParticleDust,
		Count:    count / 2,
		GenreID:  s.genreID,
		Seed:     effectSeed + 1,
		Duration: s.effectDuration * 0.6,
		SpreadX:  s.spreadFactor * 0.8,
		SpreadY:  s.spreadFactor * 0.5,
		Gravity:  20.0, // Settle downward
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   make(map[string]interface{}),
	}
	dustConfig.Custom["companion_levelup"] = true

	s.particleSystem.SpawnParticles(s.world, dustConfig, x, y)
}
