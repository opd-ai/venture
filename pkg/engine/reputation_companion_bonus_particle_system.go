// Package engine provides the ReputationCompanionBonusParticleSystem for visual feedback
// when faction reputation boosts companion stats. This system spawns genre-aware
// aura particles around companions whose stats are buffed by owner reputation.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ReputationCompanionBonusParticleSystem spawns particle effects on companions
// that are receiving stat bonuses from their owner's faction reputation.
// It emits an aura glow around buffed companions with genre-aware colors.
type ReputationCompanionBonusParticleSystem struct {
	world       *World
	particleSys *ParticleSystem
	bonusSys    *ReputationCompanionBonusSystem
	genreID     string
	seed        int64
	rng         *rand.Rand
	logger      *logrus.Entry

	// Particle configuration
	particleCount int
	spreadFactor  float64

	// Throttle particle spawning
	spawnInterval  float64
	timeSinceSpawn float64
}

// NewReputationCompanionBonusParticleSystem creates a new reputation companion bonus particle system.
func NewReputationCompanionBonusParticleSystem(world *World, seed int64) *ReputationCompanionBonusParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_companion_bonus_particle")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("reputation companion bonus particle system created")
		}
	}

	return &ReputationCompanionBonusParticleSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		particleCount:  3,
		spreadFactor:   16.0,
		spawnInterval:  1.5,
		timeSinceSpawn: 0,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ReputationCompanionBonusParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSys = ps
}

// SetBonusSystem sets the reputation companion bonus system for bonus queries.
func (s *ReputationCompanionBonusParticleSystem) SetBonusSystem(bs *ReputationCompanionBonusSystem) {
	s.bonusSys = bs
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ReputationCompanionBonusParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update spawns aura particles on companion entities with active reputation bonuses.
func (s *ReputationCompanionBonusParticleSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceSpawn += deltaTime
	if s.timeSinceSpawn < s.spawnInterval {
		return
	}
	s.timeSinceSpawn = 0

	if s.particleSys == nil || s.bonusSys == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		if _, ok := entity.GetComponent("companion"); !ok {
			continue
		}

		if !s.bonusSys.HasActiveBonus(entity.ID) {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		attack, _, _ := s.bonusSys.GetCompanionBonus(entity.ID)
		s.spawnAuraParticles(pos.X, pos.Y, attack)

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"attack_mult": attack,
				"x":           pos.X,
				"y":           pos.Y,
			}).Debug("reputation companion aura particles spawned")
		}
	}
}

// spawnAuraParticles creates gentle aura particles around the companion.
func (s *ReputationCompanionBonusParticleSystem) spawnAuraParticles(x, y, attackMult float64) {
	count := s.particleCount
	if attackMult >= 1.10 {
		count = int(float64(count) * 1.5)
	}
	if attackMult >= 1.15 {
		count = int(float64(count) * 1.5)
	}
	if count > 10 {
		count = 10
	}

	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(attackMult*100)
	particleType := s.getParticleTypeForGenre()

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 1.0,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -15.0,
		MinSize:  1.0,
		MaxSize:  2.5,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["reputation_companion_aura"] = true
	config.Custom["attack_mult"] = attackMult

	s.particleSys.SpawnParticles(s.world, config, x, y)
}

// getParticleTypeForGenre returns the appropriate particle type for the genre.
func (s *ReputationCompanionBonusParticleSystem) getParticleTypeForGenre() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Golden blessing motes
	case "scifi":
		return particles.ParticleSpark // Data-link sync pulses
	case "horror":
		return particles.ParticleSmoke // Dark bond wisps
	case "cyberpunk":
		return particles.ParticleSpark // Mod-enhancement sparks
	case "postapoc":
		return particles.ParticleDust // Loyalty bond dust
	default:
		return particles.ParticleSparkle
	}
}
