// Package engine provides the DeathParticleSystem for visual death feedback.
// This system connects CombatSystem with ParticleSystem to spawn genre-aware
// particle effects when entities die, enhancing combat visual feedback.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// DeathParticleSystem spawns particle effects when entities die.
// It connects CombatSystem and ParticleSystem to provide visual feedback
// with genre-aware particle colors and death effects (smoke, debris, etc).
type DeathParticleSystem struct {
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

// NewDeathParticleSystem creates a new death particle system.
func NewDeathParticleSystem(world *World, seed int64) *DeathParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "death_particle")
		logEntry.Debug("death particle system created")
	}

	return &DeathParticleSystem{
		world:             world,
		genreID:           "fantasy",
		seed:              seed,
		rng:               rand.New(rand.NewSource(seed)),
		logger:            logEntry,
		baseParticleCount: 25,
		spreadFactor:      150.0,
		effectDuration:    0.8,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *DeathParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *DeathParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (no-op for this callback-driven system).
func (s *DeathParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnDeath, no per-frame processing needed
}

// OnDeath is called when an entity dies to spawn death particle effects.
// This method should be registered as a callback with the CombatSystem.
func (s *DeathParticleSystem) OnDeath(entity *Entity) {
	if s.particleSystem == nil || s.world == nil || entity == nil {
		return
	}

	// Get entity position for particle spawn
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Spawn death particles at entity location
	s.spawnDeathParticles(pos.X, pos.Y, entity)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"x":         pos.X,
			"y":         pos.Y,
		}).Debug("death particles spawned")
	}
}

// spawnDeathParticles creates the death particle effect.
func (s *DeathParticleSystem) spawnDeathParticles(x, y float64, entity *Entity) {
	count := s.baseParticleCount

	// Use deterministic seed offset for this specific death
	effectSeed := s.seed + int64(x*1000) + int64(y*1000) + int64(entity.ID)

	// Choose primary particle type based on genre
	primaryType := s.getPrimaryParticleType()
	secondaryType := s.getSecondaryParticleType()

	// Create primary death effect (smoke/debris based on genre)
	primaryConfig := particles.Config{
		Type:     primaryType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  50.0, // Slight downward for realism
		MinSize:  4.0,
		MaxSize:  10.0,
		Custom:   make(map[string]interface{}),
	}
	primaryConfig.Custom["death_effect"] = true
	primaryConfig.Custom["entity_id"] = entity.ID

	s.particleSystem.SpawnParticles(s.world, primaryConfig, x, y)

	// Add secondary burst for more visual impact
	secondaryConfig := particles.Config{
		Type:     secondaryType,
		Count:    count / 2,
		GenreID:  s.genreID,
		Seed:     effectSeed + 1,
		Duration: s.effectDuration * 0.6,
		SpreadX:  s.spreadFactor * 0.8,
		SpreadY:  s.spreadFactor * 0.8,
		Gravity:  80.0,
		MinSize:  2.0,
		MaxSize:  6.0,
		Custom:   make(map[string]interface{}),
	}
	secondaryConfig.Custom["death_effect"] = true

	s.particleSystem.SpawnParticles(s.world, secondaryConfig, x, y)
}

// getPrimaryParticleType returns the primary particle type based on genre.
func (s *DeathParticleSystem) getPrimaryParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSmoke
	case "scifi":
		return particles.ParticleSpark
	case "horror":
		return particles.ParticleBlood
	case "cyberpunk":
		return particles.ParticleEmber
	case "postapoc":
		return particles.ParticleDebris
	default:
		return particles.ParticleSmokePlume
	}
}

// getSecondaryParticleType returns the secondary particle type based on genre.
func (s *DeathParticleSystem) getSecondaryParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleDust
	case "scifi":
		return particles.ParticleSmoke
	case "horror":
		return particles.ParticleSmoke
	case "cyberpunk":
		return particles.ParticleSpark
	case "postapoc":
		return particles.ParticleDust
	default:
		return particles.ParticleDust
	}
}

// SpawnDeathEffect allows external systems to trigger death particles directly.
// This is useful for environmental deaths or scripted deaths.
func (s *DeathParticleSystem) SpawnDeathEffect(x, y float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}
	// Create a temporary entity reference for seed calculation
	tempEntity := &Entity{ID: uint64(x*1000 + y*1000)}
	s.spawnDeathParticles(x, y, tempEntity)
}
