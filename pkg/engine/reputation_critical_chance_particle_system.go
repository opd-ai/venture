// Package engine provides the ReputationCriticalChanceParticleSystem for visual
// feedback when faction reputation boosts critical hit chance. This system spawns
// genre-aware glint particles around the player's weapon area when reputation
// grants a crit bonus.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ReputationCriticalChanceParticleSystem spawns particle effects when faction
// reputation grants a critical chance bonus. It produces a subtle glint effect
// near the player's hands, with genre-aware colors and intensity proportional
// to the bonus percentage.
type ReputationCriticalChanceParticleSystem struct {
	world    *World
	critSys  *ReputationCriticalChanceBonusSystem
	partSys  *ParticleSystem
	genreID  string
	seed     int64
	rng      *rand.Rand
	logger   *logrus.Entry

	particleCount int
	spreadFactor  float64
	minBonusPct   float64 // Minimum bonus % to trigger particles
	emitInterval  float64 // Seconds between particle bursts
	timeSinceEmit float64
}

// NewReputationCriticalChanceParticleSystem creates a new reputation critical chance particle system.
func NewReputationCriticalChanceParticleSystem(world *World, seed int64) *ReputationCriticalChanceParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_critical_chance_particle")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("reputation critical chance particle system created")
		}
	}

	return &ReputationCriticalChanceParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		particleCount: 4,
		spreadFactor:  15.0,
		minBonusPct:   1.0,
		emitInterval:  1.2,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ReputationCriticalChanceParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.partSys = ps
}

// SetCritSystem sets the reputation critical chance system reference.
func (s *ReputationCriticalChanceParticleSystem) SetCritSystem(cs *ReputationCriticalChanceBonusSystem) {
	s.critSys = cs
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ReputationCriticalChanceParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update spawns glint particles for entities with active crit bonuses.
func (s *ReputationCriticalChanceParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.partSys == nil || s.critSys == nil || s.world == nil {
		return
	}

	s.timeSinceEmit += deltaTime
	if s.timeSinceEmit < s.emitInterval {
		return
	}
	s.timeSinceEmit = 0

	for _, entity := range entities {
		if _, ok := entity.GetComponent("input"); !ok {
			continue
		}

		bonusPct := s.critSys.GetCritBonusPercent(entity.ID)
		if bonusPct < s.minBonusPct {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		s.spawnCritParticles(pos.X, pos.Y, bonusPct, entity.ID)
	}
}

// spawnCritParticles creates a glint effect near the entity's weapon hand.
func (s *ReputationCriticalChanceParticleSystem) spawnCritParticles(x, y, bonusPct float64, entityID uint64) {
	count := s.particleCount
	if bonusPct > 4.0 {
		count = int(float64(count) * 1.5)
	}
	if bonusPct > 6.0 {
		count = int(float64(count) * 1.5)
	}
	if count > 12 {
		count = 12
	}

	effectSeed := s.seed + int64(entityID) + int64(bonusPct*100)
	particleType := s.getParticleTypeForGenre()

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.3,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor * 0.6,
		Gravity:  -8.0,
		MinSize:  1.0,
		MaxSize:  3.0,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["reputation_crit"] = true
	config.Custom["bonus_percent"] = bonusPct

	// Offset slightly upward toward weapon area
	s.partSys.SpawnParticles(s.world, config, x+6, y-4)
}

// getParticleTypeForGenre returns the appropriate particle type for the genre.
func (s *ReputationCriticalChanceParticleSystem) getParticleTypeForGenre() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Divine blade glint
	case "scifi":
		return particles.ParticleSpark // Targeting laser flash
	case "horror":
		return particles.ParticleSmoke // Dark precision mist
	case "cyberpunk":
		return particles.ParticleSpark // Implant activation flicker
	case "postapoc":
		return particles.ParticleDust // Sharpened edge dust
	default:
		return particles.ParticleSparkle
	}
}
