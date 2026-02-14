// Package engine provides the ReputationSpellDamageBonusParticleSystem for visual
// feedback when faction reputation enhances the player's spell damage. This system
// spawns genre-aware arcane particles around the player, giving visible confirmation
// that faction standing is boosting MagicPower.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ReputationSpellDamageBonusParticleSystem spawns particle effects when faction
// reputation provides a spell damage bonus. It emits arcane motes orbiting the
// player, with genre-aware colors and intensity proportional to the bonus.
type ReputationSpellDamageBonusParticleSystem struct {
	world             *World
	particleSystem    *ParticleSystem
	spellDamageSystem *ReputationSpellDamageBonusSystem
	genreID           string
	seed              int64
	rng               *rand.Rand
	logger            *logrus.Entry

	// Particle configuration
	particleCount int
	spreadFactor  float64
	minBonus      float64 // Minimum bonus to trigger particles

	// Throttle particle spawning
	spawnInterval  float64
	timeSinceSpawn float64
}

// NewReputationSpellDamageBonusParticleSystem creates a new reputation spell damage particle system.
func NewReputationSpellDamageBonusParticleSystem(world *World, seed int64) *ReputationSpellDamageBonusParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_spell_damage_bonus_particle")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("reputation spell damage bonus particle system created")
		}
	}

	return &ReputationSpellDamageBonusParticleSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		particleCount:  5,
		spreadFactor:   24.0,
		minBonus:       0.5,
		spawnInterval:  1.2, // Spawn particles every 1.2 seconds
		timeSinceSpawn: 0,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ReputationSpellDamageBonusParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetSpellDamageSystem sets the reputation spell damage system for bonus queries.
func (s *ReputationSpellDamageBonusParticleSystem) SetSpellDamageSystem(sds *ReputationSpellDamageBonusSystem) {
	s.spellDamageSystem = sds
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ReputationSpellDamageBonusParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update spawns arcane particles on player entities that have an active
// reputation spell damage bonus.
func (s *ReputationSpellDamageBonusParticleSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceSpawn += deltaTime
	if s.timeSinceSpawn < s.spawnInterval {
		return
	}
	s.timeSinceSpawn = 0

	if s.particleSystem == nil || s.spellDamageSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		if _, ok := entity.GetComponent("input"); !ok {
			continue
		}

		bonus := s.spellDamageSystem.GetBonus(entity.ID)
		if bonus < s.minBonus {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		s.spawnArcaneParticles(pos.X, pos.Y, bonus)

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"bonus":     bonus,
				"x":         pos.X,
				"y":         pos.Y,
			}).Debug("reputation spell damage particles spawned")
		}
	}
}

// spawnArcaneParticles creates arcane mote particles around the player.
func (s *ReputationSpellDamageBonusParticleSystem) spawnArcaneParticles(x, y, bonus float64) {
	// Scale particle count with bonus magnitude
	count := s.particleCount
	if bonus >= 5.0 {
		count = int(float64(count) * 1.4)
	}
	if bonus >= 10.0 {
		count = int(float64(count) * 1.6)
	}
	if count > 14 {
		count = 14
	}

	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(bonus*100)
	particleType := s.getParticleTypeForGenre()

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.9,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -40.0, // Gentle upward drift for magical feel
		MinSize:  1.5,
		MaxSize:  4.0,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["reputation_spell_damage"] = true
	config.Custom["bonus_amount"] = bonus

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getParticleTypeForGenre returns the appropriate particle type for the genre.
func (s *ReputationSpellDamageBonusParticleSystem) getParticleTypeForGenre() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Arcane motes
	case "scifi":
		return particles.ParticleSpark // Energy amplification pulses
	case "horror":
		return particles.ParticleSmoke // Dark eldritch wisps
	case "cyberpunk":
		return particles.ParticleSpark // Overclocked spell circuits
	case "postapoc":
		return particles.ParticleDust // Residual magical dust
	default:
		return particles.ParticleSparkle
	}
}
