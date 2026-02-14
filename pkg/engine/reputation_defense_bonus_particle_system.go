// Package engine provides the ReputationDefenseBonusParticleSystem for visual feedback
// when faction reputation reduces incoming damage. This system connects
// ReputationDefenseBonusSystem with ParticleSystem to spawn genre-aware particle
// effects when a player's allied-faction reputation absorbs damage, giving players
// visual confirmation that their faction standing is providing protection.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ReputationDefenseBonusParticleSystem spawns particle effects when faction
// reputation reduces incoming damage. It provides a visible shield-like
// particle burst at the defender's position, with genre-aware colors and
// intensity proportional to the damage absorbed.
type ReputationDefenseBonusParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration
	particleCount int
	spreadFactor  float64
	minAbsorbed   float64 // Minimum damage absorbed to trigger particles
}

// NewReputationDefenseBonusParticleSystem creates a new reputation defense bonus particle system.
func NewReputationDefenseBonusParticleSystem(world *World, seed int64) *ReputationDefenseBonusParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_defense_bonus_particle")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("reputation defense bonus particle system created")
		}
	}

	return &ReputationDefenseBonusParticleSystem{
		world:         world,
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		logger:        logEntry,
		particleCount: 8,
		spreadFactor:  35.0,
		minAbsorbed:   1.0,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ReputationDefenseBonusParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ReputationDefenseBonusParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *ReputationDefenseBonusParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// Callback-driven via OnReputationDefense; no per-frame work.
}

// OnReputationDefense is called when faction reputation reduces damage.
//
// Parameters:
//   - defender: entity that received damage reduction
//   - attacker: entity whose damage was reduced
//   - originalDamage: damage before reputation reduction
//   - reducedDamage: damage after reputation reduction
//   - bonusPercent: the defense bonus percentage that was applied (0.0-1.0)
func (s *ReputationDefenseBonusParticleSystem) OnReputationDefense(
	defender, attacker *Entity,
	originalDamage, reducedDamage, bonusPercent float64,
) {
	if s.particleSystem == nil || s.world == nil || defender == nil {
		return
	}

	absorbed := originalDamage - reducedDamage
	if absorbed < s.minAbsorbed {
		return
	}

	pos := defender.GetPosition()
	if pos == nil {
		return
	}

	s.spawnDefenseParticles(pos.X, pos.Y, absorbed, bonusPercent)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		attackerID := uint64(0)
		if attacker != nil {
			attackerID = attacker.ID
		}
		s.logger.WithFields(logrus.Fields{
			"defender_id":     defender.ID,
			"attacker_id":     attackerID,
			"original_damage": originalDamage,
			"reduced_damage":  reducedDamage,
			"absorbed":        absorbed,
			"bonus_percent":   bonusPercent,
			"x":               pos.X,
			"y":               pos.Y,
		}).Debug("reputation defense bonus particles spawned")
	}
}

// spawnDefenseParticles creates a shield-like particle burst at the defender.
func (s *ReputationDefenseBonusParticleSystem) spawnDefenseParticles(
	x, y, absorbed, bonusPercent float64,
) {
	// Scale count with absorbed damage
	count := s.particleCount
	if absorbed > 10 {
		count = int(float64(count) * 1.3)
	}
	if absorbed > 25 {
		count = int(float64(count) * 1.5)
	}
	if count > 20 {
		count = 20
	}

	// Deterministic seed offset
	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(bonusPercent*10000)

	particleType := s.getParticleTypeForGenre()

	// Wider horizontal spread, less vertical for a shield-wall feel
	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.45,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor * 0.5,
		Gravity:  -15.0, // Gentle upward drift
		MinSize:  2.0,
		MaxSize:  5.0,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["reputation_defense"] = true
	config.Custom["bonus_percent"] = bonusPercent
	config.Custom["damage_absorbed"] = absorbed

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getParticleTypeForGenre returns the appropriate particle type for the genre.
func (s *ReputationDefenseBonusParticleSystem) getParticleTypeForGenre() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Golden blessing shield
	case "scifi":
		return particles.ParticleSpark // Energy barrier flicker
	case "horror":
		return particles.ParticleSmoke // Dark ward aura
	case "cyberpunk":
		return particles.ParticleSpark // Hardlight corporate shield
	case "postapoc":
		return particles.ParticleDust // Scrap shield deflection
	default:
		return particles.ParticleSparkle
	}
}

// SpawnDefenseEffect allows external systems to trigger reputation defense particles.
func (s *ReputationDefenseBonusParticleSystem) SpawnDefenseEffect(x, y float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	s.spawnDefenseParticles(x, y, 10.0, 0.05)
}
