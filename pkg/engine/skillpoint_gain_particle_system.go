// Package engine provides the SkillPointGainParticleSystem for visual skill point
// gain feedback. This system connects progression events with ParticleSystem to
// spawn genre-aware particle effects when players earn skill points, enhancing
// progression visual feedback.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// SkillPointGainCallback is called when skill points are awarded.
// It receives the entity and the number of skill points gained.
type SkillPointGainCallback func(entity *Entity, pointsGained int)

// SkillPointGainParticleSystem spawns particle effects when entities gain skill points.
// It provides visual feedback with genre-aware particle colors and knowledge-themed effects.
type SkillPointGainParticleSystem struct {
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

// NewSkillPointGainParticleSystem creates a new skill point gain particle system.
func NewSkillPointGainParticleSystem(world *World, seed int64) *SkillPointGainParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "skillpoint_gain_particle")
		logEntry.Debug("skill point gain particle system created")
	}

	return &SkillPointGainParticleSystem{
		world:             world,
		genreID:           "fantasy",
		seed:              seed,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		baseParticleCount: 12, // Subtle - fewer than level-up
		spreadFactor:      80.0,
		effectDuration:    0.8,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *SkillPointGainParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *SkillPointGainParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *SkillPointGainParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnSkillPointGain, no per-frame processing needed
}

// OnSkillPointGain is called when an entity gains skill points to spawn effect particles.
// This method should be registered as a callback with progression-related systems.
func (s *SkillPointGainParticleSystem) OnSkillPointGain(entity *Entity, pointsGained int) {
	if s.particleSystem == nil || s.world == nil || entity == nil || pointsGained <= 0 {
		return
	}

	// Get entity position for particle spawn
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Spawn skill point gain particles at entity location
	s.spawnSkillPointParticles(pos.X, pos.Y, pointsGained)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entity.ID,
			"points_gained": pointsGained,
			"x":             pos.X,
			"y":             pos.Y,
		}).Debug("skill point gain particles spawned")
	}
}

// spawnSkillPointParticles creates the skill point gain particle effect.
func (s *SkillPointGainParticleSystem) spawnSkillPointParticles(x, y float64, points int) {
	// Scale particle count based on points gained
	count := s.baseParticleCount * points
	if count > 50 {
		count = 50 // Cap at reasonable maximum
	}

	// Use deterministic seed offset for this specific effect
	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(points*50)

	// Get genre-aware particle types
	primaryType, secondaryType := s.getParticleTypes()

	// Create primary knowledge/magic effect - particles spiral upward
	primaryConfig := particles.Config{
		Type:     primaryType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor * 0.5, // More vertical spread
		Gravity:  -100.0,               // Float upward for "gaining knowledge" feel
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}
	primaryConfig.Custom["skillpoint_effect"] = true
	primaryConfig.Custom["points"] = points

	s.particleSystem.SpawnParticles(s.world, primaryConfig, x, y)

	// Add secondary glow effect for higher point gains
	if points >= 2 {
		secondaryConfig := particles.Config{
			Type:     secondaryType,
			Count:    count / 3,
			GenreID:  s.genreID,
			Seed:     effectSeed + 1,
			Duration: s.effectDuration * 0.6,
			SpreadX:  s.spreadFactor * 0.8,
			SpreadY:  s.spreadFactor * 0.8,
			Gravity:  -50.0,
			MinSize:  3.0,
			MaxSize:  6.0,
			Custom:   make(map[string]interface{}),
		}
		secondaryConfig.Custom["skillpoint_effect"] = true

		s.particleSystem.SpawnParticles(s.world, secondaryConfig, x, y)
	}
}

// getParticleTypes returns genre-aware primary and secondary particle types.
func (s *SkillPointGainParticleSystem) getParticleTypes() (particles.ParticleType, particles.ParticleType) {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleMagic, particles.ParticleSparkle
	case "scifi":
		return particles.ParticleSpark, particles.ParticleMagic
	case "horror":
		return particles.ParticleSmoke, particles.ParticleEmber
	case "cyberpunk":
		return particles.ParticleSpark, particles.ParticleSparkle
	case "postapoc":
		return particles.ParticleDust, particles.ParticleEmber
	default:
		return particles.ParticleMagic, particles.ParticleSparkle
	}
}

// SpawnSkillPointEffect allows external systems to trigger skill point particles directly.
// Useful for quest rewards, book reading, or other skill point sources.
func (s *SkillPointGainParticleSystem) SpawnSkillPointEffect(x, y float64, points int) {
	if s.particleSystem == nil || s.world == nil || points <= 0 {
		return
	}
	s.spawnSkillPointParticles(x, y, points)
}
