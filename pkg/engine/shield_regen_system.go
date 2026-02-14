// Package engine provides the ShieldRegenSystem for passive shield regeneration.
// This system connects ShieldComponent with ParticleSystem to provide visual feedback
// when shields passively regenerate over time, with genre-aware particle effects.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// ShieldRegenSystem manages passive shield regeneration and visual feedback.
// It monitors entities with ShieldComponent and regenerates shield amount
// over time when below maximum, spawning genre-aware particles to indicate
// active shield recovery.
type ShieldRegenSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Regeneration configuration
	baseRegenRate  float64 // Base shield regen per second
	regenDelay     float64 // Delay after damage before regen starts
	pulseInterval  float64 // Seconds between particle pulses
	timeSinceEmit  float64 // Accumulator for pulse timing
	baseCount      int     // Base particle count per pulse
	effectDuration float64 // How long particles live
	spreadFactor   float64 // Particle spread radius

	// Track damage timestamps for regen delay
	lastDamageTime map[uint64]float64
	gameTime       float64
}

// NewShieldRegenSystem creates a new shield regeneration system.
func NewShieldRegenSystem(world *World, seed int64) *ShieldRegenSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "shield_regen")
		logEntry.Debug("shield regen system created")
	}

	return &ShieldRegenSystem{
		world:          world,
		seed:           seed,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		genreID:        "fantasy",
		baseRegenRate:  5.0, // 5 shield per second base regen
		regenDelay:     3.0, // 3 second delay after taking damage
		pulseInterval:  0.8, // Pulse every 0.8 seconds
		timeSinceEmit:  0.0,
		baseCount:      5,    // Base particle count
		effectDuration: 0.6,  // Particles live 0.6 seconds
		spreadFactor:   40.0, // Particles spread 40 pixels
		lastDamageTime: make(map[uint64]float64),
		gameTime:       0.0,
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *ShieldRegenSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *ShieldRegenSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetBaseRegenRate sets the base shield regeneration rate per second.
func (s *ShieldRegenSystem) SetBaseRegenRate(rate float64) {
	if rate < 0 {
		rate = 0
	}
	s.baseRegenRate = rate
}

// SetRegenDelay sets the delay after damage before regeneration starts.
func (s *ShieldRegenSystem) SetRegenDelay(delay float64) {
	if delay < 0 {
		delay = 0
	}
	s.regenDelay = delay
}

// Update processes shield regeneration and spawns visual feedback particles.
func (s *ShieldRegenSystem) Update(entities []*Entity, deltaTime float64) {
	s.gameTime += deltaTime
	s.timeSinceEmit += deltaTime

	shouldSpawnParticles := s.timeSinceEmit >= s.pulseInterval

	for _, entity := range entities {
		s.processEntity(entity, deltaTime, shouldSpawnParticles)
	}

	// Reset particle timer after processing
	if shouldSpawnParticles {
		s.timeSinceEmit = 0
	}
}

// processEntity handles shield regeneration for a single entity.
func (s *ShieldRegenSystem) processEntity(entity *Entity, deltaTime float64, shouldSpawnParticles bool) {
	shieldComp, ok := entity.GetComponent("shield")
	if !ok {
		// Clean up tracking for entities that lost shield component
		delete(s.lastDamageTime, entity.ID)
		return
	}

	shield, ok := shieldComp.(*ShieldComponent)
	if !ok {
		return
	}

	// Check if shield is at max or inactive
	if shield.Amount >= shield.MaxAmount || shield.Duration <= 0 {
		return
	}

	// Check regen delay (if recently damaged, skip regen)
	lastDamage, hasDamage := s.lastDamageTime[entity.ID]
	if hasDamage && (s.gameTime-lastDamage) < s.regenDelay {
		return
	}

	// Calculate regen amount
	regenAmount := s.baseRegenRate * deltaTime

	// Apply regeneration
	oldAmount := shield.Amount
	shield.Amount += regenAmount
	if shield.Amount > shield.MaxAmount {
		shield.Amount = shield.MaxAmount
	}

	// Only spawn particles if we actually regenerated and it's time
	actualRegen := shield.Amount - oldAmount
	if actualRegen > 0 && shouldSpawnParticles && s.particleSystem != nil {
		pos := entity.GetPosition()
		if pos != nil {
			s.spawnRegenParticles(pos.X, pos.Y, entity.ID, actualRegen)
		}
	}

	if s.logger != nil && actualRegen > 0.1 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"regen":      actualRegen,
			"shield":     shield.Amount,
			"max_shield": shield.MaxAmount,
		}).Debug("shield regenerated")
	}
}

// OnShieldDamaged should be called when a shield takes damage to reset regen delay.
// This method should be registered as a callback with the CombatSystem.
func (s *ShieldRegenSystem) OnShieldDamaged(target *Entity, damageAmount float64) {
	if target == nil || damageAmount <= 0 {
		return
	}
	s.lastDamageTime[target.ID] = s.gameTime
}

// spawnRegenParticles creates genre-appropriate shield regeneration particles.
func (s *ShieldRegenSystem) spawnRegenParticles(x, y float64, entityID uint64, regenAmount float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	effectSeed := s.seed + int64(x*100) + int64(y*100) + int64(entityID)

	// Scale particle count slightly with regen amount
	count := s.baseCount
	if regenAmount > 2.0 {
		count = int(float64(count) * 1.3)
	}
	if count > 10 {
		count = 10
	}

	config := particles.Config{
		Type:     s.getShieldParticleType(),
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.effectDuration,
		SpreadX:  s.spreadFactor,
		SpreadY:  s.spreadFactor,
		Gravity:  -25.0, // Gentle upward drift
		MinSize:  2.0,
		MaxSize:  4.0,
		Custom:   map[string]interface{}{"shield_regen": true},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entityID,
			"regen":     regenAmount,
			"count":     count,
		}).Debug("shield regen particles spawned")
	}
}

// getShieldParticleType returns genre-appropriate shield particle type.
func (s *ShieldRegenSystem) getShieldParticleType() particles.ParticleType {
	switch s.genreID {
	case "fantasy":
		return particles.ParticleSparkle // Magical barrier sparkles
	case "scifi":
		return particles.ParticleSpark // Energy shield sparks
	case "horror":
		return particles.ParticleSmoke // Dark protective mist
	case "cyberpunk":
		return particles.ParticleSpark // Holographic shield flicker
	case "postapoc":
		return particles.ParticleDust // Scavenged barrier shimmer
	default:
		return particles.ParticleSparkle
	}
}
