//go:build ignore

// Package engine provides EquipmentDurabilityParticleSystem for visual feedback
// when equipment degrades through damage states. This system connects equipment
// durability systems (terrain, weather) with ParticleSystem to spawn genre-aware
// particle effects when equipment transitions between Pristine/Worn/Damaged/Broken states.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// EquipmentDurabilityParticleSystem spawns particle effects when equipment degrades.
// It provides visual feedback for equipment state transitions with genre-aware colors:
//   - Worn: Yellow/amber sparks (minor wear warning)
//   - Damaged: Orange/red fragments (significant damage)
//   - Broken: Dark red/brown debris with smoke (critical failure)
type EquipmentDurabilityParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Particle configuration per damage state
	particleCounts map[sprites.DamageState]int
	spreadFactors  map[sprites.DamageState]float64

	// Callback tracking to avoid duplicate effects
	lastStateCache map[uint64]map[EquipmentSlot]sprites.DamageState
}

// NewEquipmentDurabilityParticleSystem creates a new equipment durability particle system.
func NewEquipmentDurabilityParticleSystem(world *World, seed int64) *EquipmentDurabilityParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "equipment_durability_particle")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("equipment durability particle system created")
		}
	}

	return &EquipmentDurabilityParticleSystem{
		world:  world,
		seed:   seed,
		rng:    rand.New(rand.NewSource(seed)),
		logger: logEntry,
		particleCounts: map[sprites.DamageState]int{
			sprites.DamageStateWorn:    8,  // Minor sparks
			sprites.DamageStateDamaged: 15, // More particles
			sprites.DamageStateBroken:  25, // Heavy debris
		},
		spreadFactors: map[sprites.DamageState]float64{
			sprites.DamageStateWorn:    20.0,
			sprites.DamageStateDamaged: 30.0,
			sprites.DamageStateBroken:  50.0,
		},
		lastStateCache: make(map[uint64]map[EquipmentSlot]sprites.DamageState, 64),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *EquipmentDurabilityParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *EquipmentDurabilityParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities with equipment and spawns particles on state transitions.
func (s *EquipmentDurabilityParticleSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity checks equipment state changes and spawns appropriate particles.
func (s *EquipmentDurabilityParticleSystem) processEntity(entity *Entity) {
	equipComp := s.getEquipmentComponent(entity)
	if equipComp == nil {
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	// Initialize cache for this entity if needed
	if _, ok := s.lastStateCache[entity.ID]; !ok {
		s.lastStateCache[entity.ID] = make(map[EquipmentSlot]sprites.DamageState)
	}

	// Check each equipped item for state transitions
	for slot, item := range equipComp.Slots {
		if item == nil || item.Stats.DurabilityMax == 0 {
			continue
		}

		currentState := sprites.GetDamageStateFromDurability(item.Stats.Durability, item.Stats.DurabilityMax)
		lastState, tracked := s.lastStateCache[entity.ID][slot]

		// Only spawn particles on degradation (not repair)
		if tracked && currentState > lastState && currentState != sprites.DamageStatePristine {
			s.spawnDegradationParticles(entity, pos.X, pos.Y, slot, currentState)
		}

		// Update cache
		s.lastStateCache[entity.ID][slot] = currentState
	}
}

// spawnDegradationParticles creates particle effects for equipment degradation.
func (s *EquipmentDurabilityParticleSystem) spawnDegradationParticles(entity *Entity, x, y float64, slot EquipmentSlot, state sprites.DamageState) {
	count := s.particleCounts[state]
	spread := s.spreadFactors[state]

	// Determine particle type and color based on damage state and genre
	particleType := s.getParticleType(state)
	effectSeed := s.seed + int64(entity.ID) + int64(slot)*100 + int64(state)*1000

	config := particles.Config{
		Type:     particleType,
		Count:    count,
		GenreID:  s.genreID,
		Seed:     effectSeed,
		Duration: s.getDuration(state),
		SpreadX:  spread,
		SpreadY:  spread,
		Gravity:  s.getGravity(state),
		MinSize:  s.getMinSize(state),
		MaxSize:  s.getMaxSize(state),
		Custom:   make(map[string]interface{}),
	}

	// Add custom properties for the renderer
	config.Custom["equipment_degradation"] = true
	config.Custom["damage_state"] = state.String()
	config.Custom["slot"] = string(slot)

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"slot":         slot,
			"damage_state": state.String(),
			"x":            x,
			"y":            y,
			"count":        count,
		}).Debug("equipment degradation particles spawned")
	}
}

// getParticleType returns the appropriate particle type for a damage state.
func (s *EquipmentDurabilityParticleSystem) getParticleType(state sprites.DamageState) particles.ParticleType {
	switch state {
	case sprites.DamageStateWorn:
		return particles.ParticleSparkle // Subtle sparks
	case sprites.DamageStateDamaged:
		return particles.ParticleDebris // Fragments
	case sprites.DamageStateBroken:
		return particles.ParticleSmoke // Heavy debris with smoke
	default:
		return particles.ParticleSparkle
	}
}

// getDuration returns particle lifetime based on damage state.
func (s *EquipmentDurabilityParticleSystem) getDuration(state sprites.DamageState) float64 {
	switch state {
	case sprites.DamageStateWorn:
		return 0.4 // Quick sparks
	case sprites.DamageStateDamaged:
		return 0.6 // Longer fragments
	case sprites.DamageStateBroken:
		return 0.9 // Lingering debris
	default:
		return 0.5
	}
}

// getGravity returns particle gravity based on damage state.
func (s *EquipmentDurabilityParticleSystem) getGravity(state sprites.DamageState) float64 {
	switch state {
	case sprites.DamageStateWorn:
		return -30.0 // Light upward float
	case sprites.DamageStateDamaged:
		return 50.0 // Fall with fragments
	case sprites.DamageStateBroken:
		return 80.0 // Heavy falling debris
	default:
		return 0.0
	}
}

// getMinSize returns minimum particle size based on damage state.
func (s *EquipmentDurabilityParticleSystem) getMinSize(state sprites.DamageState) float64 {
	switch state {
	case sprites.DamageStateWorn:
		return 1.5
	case sprites.DamageStateDamaged:
		return 2.0
	case sprites.DamageStateBroken:
		return 2.5
	default:
		return 2.0
	}
}

// getMaxSize returns maximum particle size based on damage state.
func (s *EquipmentDurabilityParticleSystem) getMaxSize(state sprites.DamageState) float64 {
	switch state {
	case sprites.DamageStateWorn:
		return 3.0
	case sprites.DamageStateDamaged:
		return 5.0
	case sprites.DamageStateBroken:
		return 7.0
	default:
		return 4.0
	}
}

// getEquipmentComponent retrieves the equipment component from an entity.
func (s *EquipmentDurabilityParticleSystem) getEquipmentComponent(entity *Entity) *EquipmentComponent {
	comp, ok := entity.GetComponent("equipment")
	if !ok || comp == nil {
		return nil
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return nil
	}
	return equipComp
}

// OnEquipmentDegraded is a callback that can be registered with durability systems.
// This provides an alternative integration path for direct callback-based triggering.
func (s *EquipmentDurabilityParticleSystem) OnEquipmentDegraded(entity *Entity, slot EquipmentSlot, oldDurability, newDurability, maxDurability int) {
	if s.particleSystem == nil || s.world == nil || entity == nil {
		return
	}

	pos := entity.GetPosition()
	if pos == nil {
		return
	}

	oldState := sprites.GetDamageStateFromDurability(oldDurability, maxDurability)
	newState := sprites.GetDamageStateFromDurability(newDurability, maxDurability)

	// Only spawn particles on degradation (state increase = worse condition)
	if newState > oldState && newState != sprites.DamageStatePristine {
		s.spawnDegradationParticles(entity, pos.X, pos.Y, slot, newState)
	}
}

// ClearEntityCache removes cached state for an entity (call on entity removal).
func (s *EquipmentDurabilityParticleSystem) ClearEntityCache(entityID uint64) {
	delete(s.lastStateCache, entityID)
}
