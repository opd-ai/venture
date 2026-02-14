// Package engine provides the ManaRegenParticleSystem for visual mana regeneration feedback.
// This system connects ManaComponent with ParticleSystem to spawn genre-aware
// particle effects when entities actively regenerate mana.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ManaRegenParticleSystem spawns particle effects when entities regenerate mana.
// It monitors mana levels and provides visual feedback with genre-aware particles
// indicating active mana flow. Effects are triggered when mana is below max and
// regenerating, with intensity scaling based on regen rate.
type ManaRegenParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Effect thresholds and timing
	minRegenRate    float64 // Minimum regen rate to trigger particles
	pulseInterval   float64 // Seconds between particle pulses
	timeSinceEmit   float64 // Accumulator for pulse timing
	baseCount       int     // Base particle count per pulse
	effectDuration  float64 // How long particles live
	spreadFactor    float64 // Particle spread radius
	intensityScaler float64 // Scales particle count with regen rate

	// Track previous mana levels to detect active regeneration
	prevMana map[uint64]int
}

// NewManaRegenParticleSystem creates a new mana regeneration particle system.
func NewManaRegenParticleSystem(world *World, seed int64) *ManaRegenParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "mana_regen_particle")
		logEntry.Debug("mana regen particle system created")
	}

	return &ManaRegenParticleSystem{
		world:           world,
		seed:            seed,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logEntry,
		genreID:         "fantasy",
		minRegenRate:    1.0, // At least 1 mana/sec to show particles
		pulseInterval:   1.2, // Pulse every 1.2 seconds
		timeSinceEmit:   0.0,
		baseCount:       6,    // Base particle count
		effectDuration:  0.8,  // Particles live 0.8 seconds
		spreadFactor:    32.0, // Particles spread 32 pixels
		intensityScaler: 0.3,  // Scale factor for regen rate
		prevMana:        make(map[uint64]int),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ManaRegenParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ManaRegenParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update monitors mana regeneration and spawns visual feedback particles.
func (s *ManaRegenParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	s.timeSinceEmit += deltaTime
	if s.timeSinceEmit < s.pulseInterval {
		return
	}

	for _, entity := range entities {
		s.processEntity(entity)
	}

	// Reset timer after processing
	if s.timeSinceEmit >= s.pulseInterval {
		s.timeSinceEmit = 0
	}
}

// processEntity checks if entity is actively regenerating mana and spawns particles.
func (s *ManaRegenParticleSystem) processEntity(entity *Entity) {
	manaComp, ok := entity.GetComponent("mana")
	if !ok {
		// Clean up tracking for entities that lost mana component
		delete(s.prevMana, entity.ID)
		return
	}

	mana, ok := manaComp.(*ManaComponent)
	if !ok || mana.Regen < s.minRegenRate {
		return
	}

	// Only spawn particles if mana is below max and actively regenerating
	if mana.Current >= mana.Max {
		s.prevMana[entity.ID] = mana.Current
		return
	}

	// Check if mana actually increased since last check
	prevMana, hasPrev := s.prevMana[entity.ID]
	s.prevMana[entity.ID] = mana.Current

	// Skip if mana didn't increase (spell cast, damage, etc.)
	if hasPrev && mana.Current <= prevMana {
		return
	}

	// Get position for particle spawn
	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Spawn mana regen particles
	s.spawnManaRegenEffect(pos.X, pos.Y, entity.ID, mana.Regen)
}

// spawnManaRegenEffect creates genre-appropriate mana regeneration particles.
func (s *ManaRegenParticleSystem) spawnManaRegenEffect(x, y float64, entityID uint64, regenRate float64) {
	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(entityID)

	// Scale particle count with regen rate
	count := s.baseCount + int(regenRate*s.intensityScaler)
	if count > 15 {
		count = 15
	}

	config := particles.Config{
		Type:     s.getManaParticleType(),
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -40.0, // Particles float upward
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   map[string]interface{}{"mana_regen": true},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entityID,
			"regen_rate": regenRate,
			"count":      count,
		}).Debug("mana regen particles spawned")
	}
}

// getManaParticleType returns genre-appropriate mana particle type.
func (s *ManaRegenParticleSystem) getManaParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleMagic
	case "scifi":
		return particles.ParticleSpark
	case "horror":
		return particles.ParticleSmoke
	case "cyberpunk":
		return particles.ParticleSpark
	case "postapoc":
		return particles.ParticleDust
	default:
		return particles.ParticleSparkle
	}
}
