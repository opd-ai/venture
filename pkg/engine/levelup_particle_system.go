// Package engine provides the LevelUpParticleSystem for visual level-up feedback.
// This system connects ProgressionSystem with ParticleSystem to spawn genre-aware
// particle effects when entities level up, enhancing progression visual feedback.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// LevelUpParticleSystem spawns particle effects when entities level up.
// It connects ProgressionSystem and ParticleSystem to provide visual feedback
// with genre-aware particle colors and celebration effects.
type LevelUpParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	baseParticleCount int
	spreadFactor      float64
	effectDuration    float64
}

// NewLevelUpParticleSystem creates a new level-up particle system.
func NewLevelUpParticleSystem(world *World, seed int64) *LevelUpParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "levelup_particle")
		logEntry.Debug("level-up particle system created")
	}

	return &LevelUpParticleSystem{
		world:             world,
		genreID:           "fantasy",
		seed:              seed,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		baseParticleCount: 30,
		spreadFactor:      200.0,
		effectDuration:    1.0,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *LevelUpParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *LevelUpParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *LevelUpParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnLevelUp, no per-frame processing needed
}

// OnLevelUp is called when an entity levels up to spawn celebration particles.
// This method should be registered as a callback with the ProgressionSystem.
func (s *LevelUpParticleSystem) OnLevelUp(entity *Entity, newLevel int) {
	if s.particleSystem == nil || s.world == nil || entity == nil {
		return
	}

	// Get entity position for particle spawn
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Spawn level-up particles at entity location
	s.spawnLevelUpParticles(pos.X, pos.Y, newLevel)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"new_level": newLevel,
			"x":         pos.X,
			"y":         pos.Y,
		}).Debug("level-up particles spawned")
	}
}

// spawnLevelUpParticles creates the level-up celebration particle effect.
func (s *LevelUpParticleSystem) spawnLevelUpParticles(x, y float64, level int) {
	// Scale particle count based on level milestone
	count := s.baseParticleCount
	if level%5 == 0 {
		count = int(float64(count) * 1.5) // 50% more for every 5 levels
	}
	if level%10 == 0 {
		count = int(float64(count) * 2.0) // Double for every 10 levels
	}
	// Cap at reasonable maximum
	if count > 100 {
		count = 100
	}

	// Use deterministic seed offset for this specific level-up
	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(level*100)

	// Create sparkle config for level-up effect (rising celebration particles)
	sparkleConfig := particles.Config{
		Type:     particles.ParticleSparkle,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -120.0, // Float upward for celebration effect
		MinSize:  4.0,
		MaxSize:  8.0,
		Custom:   make(map[string]interface{}),
	}
	sparkleConfig.Custom["levelup_effect"] = true
	sparkleConfig.Custom["level"] = level

	s.particleSystem.SpawnParticles(s.world, sparkleConfig, x, y)

	// Add secondary magic burst for extra visual impact
	magicConfig := particles.Config{
		Type:     particles.ParticleMagic,
		Count:    count / 2,
		GenreID:  s.genreID,
		Seed:     effectSeed + 1,
		Duration: s.effectDuration * 0.8,
		SpreadX:  s.spreadFactor * 1.2,
		SpreadY:  s.spreadFactor * 1.2,
		Gravity:  -60.0,
		MinSize:  3.0,
		MaxSize:  6.0,
		Custom:   make(map[string]interface{}),
	}
	magicConfig.Custom["levelup_effect"] = true

	s.particleSystem.SpawnParticles(s.world, magicConfig, x, y)
}

// SpawnLevelUpEffect allows external systems to trigger level-up particles directly.
// This is useful for prestige level-ups or other progression milestones.
func (s *LevelUpParticleSystem) SpawnLevelUpEffect(x, y float64, level int) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.spawnLevelUpParticles(x, y, level)
}
