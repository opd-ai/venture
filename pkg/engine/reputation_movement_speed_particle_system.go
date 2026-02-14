// Package engine provides the ReputationMovementSpeedParticleSystem for visual feedback
// when faction reputation boosts movement speed. This system spawns genre-aware
// wind-trail particles at the player's feet when reputation grants a speed bonus.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ReputationMovementSpeedParticleSystem spawns particle effects when faction
// reputation grants a movement speed bonus. It produces a subtle wind-trail
// effect at the player's position, with genre-aware colors and intensity
// proportional to the bonus percentage.
type ReputationMovementSpeedParticleSystem struct {
	world       *World
	speedSystem *ReputationMovementSpeedSystem
	particleSys *ParticleSystem
	genreID     string
	seed        int64
	rng         *rand.Rand
	logger      *logrus.Entry

	// Particle configuration
	particleCount  int
	spreadFactor   float64
	minBonusPct    float64 // Minimum bonus % to trigger particles
	emitInterval   float64 // Seconds between particle bursts
	timeSinceEmit  float64
	lastBonusCache map[uint64]float64
}

// NewReputationMovementSpeedParticleSystem creates a new reputation movement speed particle system.
func NewReputationMovementSpeedParticleSystem(world *World, seed int64) *ReputationMovementSpeedParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_movement_speed_particle")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("reputation movement speed particle system created")
		}
	}

	return &ReputationMovementSpeedParticleSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		particleCount:  5,
		spreadFactor:   20.0,
		minBonusPct:    1.0,
		emitInterval:   0.8,
		lastBonusCache: make(map[uint64]float64),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ReputationMovementSpeedParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSys = ps
}

// SetSpeedSystem sets the reputation movement speed system reference.
func (s *ReputationMovementSpeedParticleSystem) SetSpeedSystem(ss *ReputationMovementSpeedSystem) {
	s.speedSystem = ss
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ReputationMovementSpeedParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update spawns wind-trail particles for entities with active speed bonuses.
func (s *ReputationMovementSpeedParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSys == nil || s.speedSystem == nil || s.world == nil {
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

		bonus := s.speedSystem.GetSpeedBonus(entity.ID)
		if bonus < s.minBonusPct {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		s.spawnSpeedParticles(pos.X, pos.Y, bonus, entity.ID)
	}
}

// spawnSpeedParticles creates a wind-trail effect at the entity's feet.
func (s *ReputationMovementSpeedParticleSystem) spawnSpeedParticles(x, y, bonusPct float64, entityID uint64) {
	count := s.particleCount
	if bonusPct > 5.0 {
		count = int(float64(count) * 1.4)
	}
	if bonusPct > 8.0 {
		count = int(float64(count) * 1.6)
	}
	if count > 16 {
		count = 16
	}

	effectSeed := s.seed + int64(entityID) + int64(bonusPct*100)
	particleType := s.getParticleTypeForGenre()

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.35,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor * 0.3,
		Gravity:  -10.0,
		MinSize:  1.5,
		MaxSize:  3.5,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["reputation_speed"] = true
	config.Custom["bonus_percent"] = bonusPct

	s.particleSys.SpawnParticles(s.world, config, x, y+12)
}

// getParticleTypeForGenre returns the appropriate particle type for the genre.
func (s *ReputationMovementSpeedParticleSystem) getParticleTypeForGenre() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Blessed wind
	case "scifi":
		return particles.ParticleSpark // Thruster glow
	case "horror":
		return particles.ParticleSmoke // Dark mist trail
	case "cyberpunk":
		return particles.ParticleSpark // Neon speed lines
	case "postapoc":
		return particles.ParticleDust // Kicked-up debris
	default:
		return particles.ParticleDust
	}
}
