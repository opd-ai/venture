// Package engine provides the ReputationEquipmentDurabilityParticleSystem for
// visual feedback when faction reputation affects equipment durability. Spawns
// genre-aware particles: protective shimmer for positive reputation, corrosion
// wisps for hostile reputation.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ReputationEquipmentDurabilityParticleSystem spawns particle effects when
// faction reputation modifies equipment durability. Positive reputation
// produces a gentle protective glow around the player's armor, while negative
// reputation produces corrosion/decay wisps.
type ReputationEquipmentDurabilityParticleSystem struct {
	world      *World
	durabSys   *ReputationEquipmentDurabilitySystem
	partSys    *ParticleSystem
	genreID    string
	seed       int64
	rng        *rand.Rand
	logger     *logrus.Entry

	particleCount int
	spreadFactor  float64
	minModifierPct float64 // Minimum |modifier| % to trigger particles
	emitInterval   float64 // Seconds between particle bursts
	timeSinceEmit  float64
}

// NewReputationEquipmentDurabilityParticleSystem creates a new particle system
// for reputation equipment durability visual feedback.
func NewReputationEquipmentDurabilityParticleSystem(world *World, seed int64) *ReputationEquipmentDurabilityParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_equipment_durability_particle")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("reputation equipment durability particle system created")
		}
	}

	return &ReputationEquipmentDurabilityParticleSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		particleCount:  3,
		spreadFactor:   12.0,
		minModifierPct: 1.0,
		emitInterval:   1.5,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ReputationEquipmentDurabilityParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.partSys = ps
}

// SetDurabilitySystem sets the reputation equipment durability system reference.
func (s *ReputationEquipmentDurabilityParticleSystem) SetDurabilitySystem(ds *ReputationEquipmentDurabilitySystem) {
	s.durabSys = ds
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ReputationEquipmentDurabilityParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update spawns particles for entities with active durability modifiers.
func (s *ReputationEquipmentDurabilityParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.partSys == nil || s.durabSys == nil || s.world == nil {
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

		modPct := s.durabSys.GetDurabilityModifierPercent(entity.ID)
		if modPct > -s.minModifierPct && modPct < s.minModifierPct {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		if modPct > 0 {
			s.spawnProtectionParticles(pos.X, pos.Y, modPct, entity.ID)
		} else {
			s.spawnCorrosionParticles(pos.X, pos.Y, -modPct, entity.ID)
		}
	}
}

// spawnProtectionParticles creates a protective shimmer around the entity's armor.
func (s *ReputationEquipmentDurabilityParticleSystem) spawnProtectionParticles(x, y, modPct float64, entityID uint64) {
	count := s.particleCount
	if modPct > 10.0 {
		count = int(float64(count) * 1.5)
	}
	if count > 8 {
		count = 8
	}

	effectSeed := s.seed + int64(entityID) + int64(modPct*100)
	particleType := s.getProtectionParticleType()

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.5,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -5.0,
		MinSize:  1.0,
		MaxSize:  2.5,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["reputation_protection"] = true
	config.Custom["modifier_percent"] = modPct

	s.partSys.SpawnParticles(s.world, config, x, y-2)
}

// spawnCorrosionParticles creates decay wisps around the entity's equipment.
func (s *ReputationEquipmentDurabilityParticleSystem) spawnCorrosionParticles(x, y, modPct float64, entityID uint64) {
	count := s.particleCount
	if modPct > 10.0 {
		count = int(float64(count) * 1.5)
	}
	if count > 8 {
		count = 8
	}

	effectSeed := s.seed + int64(entityID) + int64(modPct*100) + 999
	particleType := s.getCorrosionParticleType()

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: 0.4,
		SpreadX:  s.spreadFactor * 0.8,
		SpreadY:  s.spreadFactor * 0.8,
		Gravity:  3.0, // Corrosion drips down
		MinSize:  1.0,
		MaxSize:  2.0,
		Custom:   make(map[string]interface{}),
	}

	config.Custom["reputation_corrosion"] = true
	config.Custom["modifier_percent"] = modPct

	s.partSys.SpawnParticles(s.world, config, x, y+4)
}

// getProtectionParticleType returns the particle type for protection effects.
func (s *ReputationEquipmentDurabilityParticleSystem) getProtectionParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Divine blessing shimmer
	case "scifi":
		return particles.ParticleSpark // Nano-repair field
	case "horror":
		return particles.ParticleSparkle // Faint ward glow
	case "cyberpunk":
		return particles.ParticleSpark // Maintenance drone sparks
	case "postapoc":
		return particles.ParticleDust // Careful repair dust
	default:
		return particles.ParticleSparkle
	}
}

// getCorrosionParticleType returns the particle type for corrosion/decay effects.
func (s *ReputationEquipmentDurabilityParticleSystem) getCorrosionParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSmoke // Curse smoke
	case "scifi":
		return particles.ParticleSpark // EMP interference
	case "horror":
		return particles.ParticleSmoke // Decay mist
	case "cyberpunk":
		return particles.ParticleSpark // Virus corruption
	case "postapoc":
		return particles.ParticleDust // Acid wind erosion
	default:
		return particles.ParticleSmoke
	}
}
