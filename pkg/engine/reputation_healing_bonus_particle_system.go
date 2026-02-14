// Package engine provides the ReputationHealingBonusParticleSystem for visual feedback
// when faction reputation passively heals the player. This system spawns genre-aware
// healing particles at the player's position, giving visible confirmation that
// faction standing is providing regeneration.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ReputationHealingBonusParticleSystem spawns particle effects when faction
// reputation provides passive health regeneration. It emits a gentle upward
// healing glow at the player's position, with genre-aware colors and
// intensity proportional to the regen rate.
type ReputationHealingBonusParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	healingSystem  *ReputationHealingBonusSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	particleCount int
	spreadFactor  float64
	minRegenRate  float64 // Minimum regen rate to trigger particles

	// Throttle particle spawning
	spawnInterval  float64
	timeSinceSpawn float64
}

// NewReputationHealingBonusParticleSystem creates a new reputation healing particle system.
func NewReputationHealingBonusParticleSystem(world *World, seed int64) *ReputationHealingBonusParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_healing_bonus_particle")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("reputation healing bonus particle system created")
		}
	}

	return &ReputationHealingBonusParticleSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		particleCount:  4,
		spreadFactor:   20.0,
		minRegenRate:   0.1,
		spawnInterval:  1.0, // Spawn particles every 1 second
		timeSinceSpawn: 0,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ReputationHealingBonusParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetHealingSystem sets the reputation healing system for regen rate queries.
func (s *ReputationHealingBonusParticleSystem) SetHealingSystem(hs *ReputationHealingBonusSystem) {
	s.healingSystem = hs
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ReputationHealingBonusParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update spawns healing particles on player entities that are actively regenerating
// from faction reputation bonuses.
func (s *ReputationHealingBonusParticleSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceSpawn += deltaTime
	if s.timeSinceSpawn < s.spawnInterval {
		return
	}
	s.timeSinceSpawn = 0

	if s.particleSystem == nil || s.healingSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		if _, ok := entity.GetComponent("input"); !ok {
			continue
		}

		regenRate := s.healingSystem.GetRegenRate(entity.ID)
		if regenRate < s.minRegenRate {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		s.spawnHealingParticles(pos.X, pos.Y, regenRate)

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"regen_rate": regenRate,
				"x":          pos.X,
				"y":          pos.Y,
			}).Debug("reputation healing particles spawned")
		}
	}
}

// spawnHealingParticles creates gentle upward healing particles.
func (s *ReputationHealingBonusParticleSystem) spawnHealingParticles(x, y, regenRate float64) {
	// Scale particle count with regen rate
	count := s.particleCount
	if regenRate >= 1.0 {
		count = int(float64(count) * 1.5)
	}
	if regenRate >= 2.0 {
		count = int(float64(count) * 1.5)
	}
	if count > 12 {
		count = 12
	}

	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(regenRate*100)
	particleType := s.getParticleTypeForGenre()

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.8,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor * 0.8,
		Gravity:  -25.0, // Upward drift for healing feel
		MinSize:  1.5,
		MaxSize:  3.5,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["reputation_healing"] = true
	config.Custom["regen_rate"] = regenRate

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getParticleTypeForGenre returns the appropriate particle type for the genre.
func (s *ReputationHealingBonusParticleSystem) getParticleTypeForGenre() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Golden divine healing motes
	case "scifi":
		return particles.ParticleSpark // Nanite repair particles
	case "horror":
		return particles.ParticleSmoke // Dark vitality wisps
	case "cyberpunk":
		return particles.ParticleSpark // Bio-enhancer pulses
	case "postapoc":
		return particles.ParticleDust // Herbal remedy particles
	default:
		return particles.ParticleSparkle
	}
}
