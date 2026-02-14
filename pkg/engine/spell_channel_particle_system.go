// Package engine provides the SpellChannelParticleSystem for visual spell channeling feedback.
// This system connects SpellSlotComponent's casting state with ParticleSystem to spawn
// genre-aware particles that swirl around the caster while a spell is being channeled.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// SpellChannelParticleSystem spawns particle effects around entities channeling spells.
// It monitors SpellSlotComponent.IsCasting() and provides visual feedback with
// element-colored particles based on the spell being cast.
type SpellChannelParticleSystem struct {
	world           *World
	particleSys     *ParticleSystem
	rng             *rand.Rand
	genreID         string
	logger          *logrus.Entry
	spawnCooldowns  map[uint64]float64 // Tracks cooldown per entity to prevent particle spam
	spawnInterval   float64            // Time between particle spawns (seconds)
	particleCount   int                // Particles per spawn
	baseParticleMap map[uint64]bool    // Tracks if we've announced channel start
}

// NewSpellChannelParticleSystem creates a new spell channeling particle system.
func NewSpellChannelParticleSystem(world *World, seed int64) *SpellChannelParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "spell_channel_particle")
		logEntry.Debug("spell channel particle system created")
	}

	return &SpellChannelParticleSystem{
		world:           world,
		rng:             rand.New(rand.NewSource(seed)),
		genreID:         "fantasy",
		logger:          logEntry,
		spawnCooldowns:  make(map[uint64]float64),
		spawnInterval:   0.15, // Spawn particles every 150ms during channel
		particleCount:   6,    // 6 particles per spawn for visible effect
		baseParticleMap: make(map[uint64]bool),
	}
}

// SetParticleSystem sets the particle system to use for spawning effects.
func (s *SpellChannelParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSys = ps
}

// SetGenre sets the genre for particle color selection.
func (s *SpellChannelParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update monitors entities for spell channeling and spawns visual feedback particles.
func (s *SpellChannelParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSys == nil || s.world == nil {
		return
	}

	// Update cooldowns
	for id := range s.spawnCooldowns {
		s.spawnCooldowns[id] -= deltaTime
		if s.spawnCooldowns[id] < 0 {
			s.spawnCooldowns[id] = 0
		}
	}

	for _, entity := range entities {
		s.processEntity(entity, deltaTime)
	}

	// Clean up entities that are no longer channeling
	s.cleanupStaleEntries(entities)
}

// processEntity checks if an entity is channeling a spell and spawns particles.
func (s *SpellChannelParticleSystem) processEntity(entity *Entity, deltaTime float64) {
	// Get spell slots component
	slotComp, hasSlots := entity.GetComponent("spell_slots")
	if !hasSlots {
		return
	}

	slots, ok := slotComp.(*SpellSlotComponent)
	if !ok {
		return
	}

	// Check if entity is currently casting
	if !slots.IsCasting() {
		// Clear channel state when casting ends
		if s.baseParticleMap[entity.ID] {
			delete(s.baseParticleMap, entity.ID)
			delete(s.spawnCooldowns, entity.ID)
		}
		return
	}

	// Get the spell being cast
	spell := slots.GetSlot(slots.Casting)
	if spell == nil {
		return
	}

	// Get position for particle spawning
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Log channel start once
	if !s.baseParticleMap[entity.ID] {
		s.baseParticleMap[entity.ID] = true
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":  entity.ID,
				"spell_name": spell.Name,
				"element":    spell.Element.String(),
			}).Debug("spell channeling started")
		}
	}

	// Check spawn cooldown
	if s.spawnCooldowns[entity.ID] > 0 {
		return
	}

	// Spawn channeling particles
	s.spawnChannelEffect(pos.X, pos.Y, entity.ID, spell, slots.CastingBar)
	s.spawnCooldowns[entity.ID] = s.spawnInterval
}

// spawnChannelEffect creates element-appropriate channeling particles around the caster.
func (s *SpellChannelParticleSystem) spawnChannelEffect(x, y float64, entityID uint64, spell *magic.Spell, progress float64) {
	particleType := s.getChannelParticleType(spell.Element)

	// Particle count scales with casting progress (more intense as spell builds)
	count := s.particleCount + int(float64(s.particleCount)*progress)

	// Spread increases as cast progresses (energy building up)
	baseSpread := 30.0
	spreadBonus := 20.0 * progress
	spread := baseSpread + spreadBonus

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     int64(entityID) + s.rng.Int63(),
		Duration: 0.4,
		SpreadX:  spread,
		SpreadY:  spread,
		Gravity:  -60.0, // Particles rise during channeling
		MinSize:  3.0,
		MaxSize:  5.0 + float64(2.0*progress), // Larger particles as cast progresses
		ZLayer:   particles.ZLayerAbove,
		Custom:   map[string]interface{}{"color": spell.Element.String()},
	}

	s.particleSys.SpawnParticles(s.world, config, x, y)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entityID,
			"spell_name":   spell.Name,
			"progress":     progress,
			"particle_cnt": count,
		}).Debug("channel particles spawned")
	}
}

// getChannelParticleType returns the appropriate particle type for the spell element.
func (s *SpellChannelParticleSystem) getChannelParticleType(element magic.ElementType) particles.ParticleType {
	switch element {
	case magic.ElementFire:
		return particles.ParticleEmber
	case magic.ElementIce:
		return particles.ParticleSparkle
	case magic.ElementLightning:
		return particles.ParticleSpark
	case magic.ElementEarth:
		return particles.ParticleDust
	case magic.ElementWind:
		return particles.ParticleDust
	case magic.ElementLight:
		return particles.ParticleSparkle
	case magic.ElementDark:
		return particles.ParticleSmoke
	case magic.ElementArcane:
		return particles.ParticleMagic
	default:
		return particles.ParticleMagic
	}
}

// cleanupStaleEntries removes tracking data for entities no longer in the world.
func (s *SpellChannelParticleSystem) cleanupStaleEntries(entities []*Entity) {
	// Build set of current entity IDs
	currentIDs := make(map[uint64]bool, len(entities))
	for _, e := range entities {
		currentIDs[e.ID] = true
	}

	// Remove stale entries
	for id := range s.baseParticleMap {
		if !currentIDs[id] {
			delete(s.baseParticleMap, id)
			delete(s.spawnCooldowns, id)
		}
	}
}
