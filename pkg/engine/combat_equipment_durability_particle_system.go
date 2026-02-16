// Package engine provides CombatEquipmentDurabilityParticleSystem for visual feedback
// when equipment degrades from combat damage. This system connects CombatSystem damage
// events with ParticleSystem to spawn genre-aware particle effects when armor absorbs
// damage and transitions between Pristine/Worn/Damaged/Broken states.
//
// Unlike TerrainEquipmentDurabilitySystem and WeatherEquipmentDurabilitySystem which
// handle environmental wear, this system handles direct combat impact damage to armor.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// CombatEquipmentDurabilityParticleSystem spawns particle effects when armor takes
// combat damage and degrades. It provides visual feedback for damage absorption with
// genre-aware particle colors and intensities:
//   - Light damage: Subtle sparks on armor surface
//   - Heavy damage: Metal fragments and debris
//   - Critical hits: Intense sparks with armor crack effects
type CombatEquipmentDurabilityParticleSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry

	// Configuration
	baseDurabilityLossPerDamage float64 // Durability points lost per 100 damage
	critMultiplier              float64 // Extra durability loss on critical hits

	// Particle configuration
	particleCountBase  int     // Base particles for minor damage
	particleCountScale float64 // Scale factor for damage intensity
	spreadBase         float64 // Base particle spread radius

	// State tracking to detect damage state transitions
	lastStateCache map[uint64]map[EquipmentSlot]sprites.DamageState
}

// NewCombatEquipmentDurabilityParticleSystem creates a new combat equipment durability particle system.
func NewCombatEquipmentDurabilityParticleSystem(world *World, seed int64) *CombatEquipmentDurabilityParticleSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "combat_equipment_durability_particle")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("combat equipment durability particle system created")
		}
	}

	return &CombatEquipmentDurabilityParticleSystem{
		world:                       world,
		seed:                        seed,
		rng:                         rand.New(rand.NewSource(seed)),
		logger:                      logEntry,
		baseDurabilityLossPerDamage: 2.0,  // 2 durability per 100 damage
		critMultiplier:              1.5,  // 50% extra durability loss on crits
		particleCountBase:           6,    // Base 6 particles
		particleCountScale:          0.1,  // +1 particle per 10 damage
		spreadBase:                  15.0, // 15px spread radius
		lastStateCache:              make(map[uint64]map[EquipmentSlot]sprites.DamageState, 64),
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *CombatEquipmentDurabilityParticleSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.Debug("particle system linked")
	}
}

// SetGenre sets the genre ID for genre-aware particle colors.
func (s *CombatEquipmentDurabilityParticleSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities (callback-driven, minimal per-frame work).
func (s *CombatEquipmentDurabilityParticleSystem) Update(entities []*Entity, deltaTime float64) {
	// This system is callback-driven via OnDamageTaken, no per-frame processing needed
}

// OnDamageTaken is called when an entity takes combat damage.
// Register this with CombatSystem.AddDamageCallback().
func (s *CombatEquipmentDurabilityParticleSystem) OnDamageTaken(attacker, target *Entity, damage float64) {
	if target == nil || damage <= 0 {
		return
	}

	// Get equipment component
	equipComp := s.getEquipmentComponent(target)
	if equipComp == nil {
		return
	}

	// Get position for particle effects
	x, y := s.getEntityPosition(target)

	// Process armor slots that absorb damage
	armorSlots := []EquipmentSlot{SlotChest, SlotHead, SlotBoots, SlotGloves, SlotLegs}
	equippedCount := 0
	for _, slot := range armorSlots {
		if equipComp.GetEquipped(slot) != nil {
			equippedCount++
		}
	}

	if equippedCount == 0 {
		return
	}

	// Distribute durability loss across equipped armor
	durabilityLoss := (damage / 100.0) * s.baseDurabilityLossPerDamage
	lossPerPiece := durabilityLoss / float64(equippedCount)

	// Apply durability loss and check for state transitions
	for _, slot := range armorSlots {
		item := equipComp.GetEquipped(slot)
		if item == nil {
			continue
		}

		// Initialize state cache for this entity if needed
		if _, ok := s.lastStateCache[target.ID]; !ok {
			s.lastStateCache[target.ID] = make(map[EquipmentSlot]sprites.DamageState)
		}

		// Get previous and current damage states
		oldState := sprites.GetDamageStateFromDurability(item.Stats.Durability, item.Stats.DurabilityMax)
		if _, tracked := s.lastStateCache[target.ID][slot]; !tracked {
			s.lastStateCache[target.ID][slot] = oldState
		}

		// Apply durability loss (use ceiling to ensure at least 1 point lost when there is damage)
		oldDurability := item.Stats.Durability
		item.Stats.Durability -= int(math.Ceil(lossPerPiece))
		if item.Stats.Durability < 0 {
			item.Stats.Durability = 0
		}

		// Check for state transition
		newState := sprites.GetDamageStateFromDurability(item.Stats.Durability, item.Stats.DurabilityMax)

		// Spawn particles based on damage intensity
		if s.particleSystem != nil {
			// Always spawn minor impact particles for damage absorption
			if damage >= 10 {
				s.spawnImpactParticles(target, x, y, damage)
			}

			// Spawn additional particles on state transition
			if newState > oldState && newState != sprites.DamageStatePristine {
				s.spawnDegradationParticles(target, x, y, slot, newState)
			}
		}

		// Update cache
		s.lastStateCache[target.ID][slot] = newState

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel && oldDurability != item.Stats.Durability {
			s.logger.WithFields(logrus.Fields{
				"entity_id":       target.ID,
				"slot":            slot,
				"old_durability":  oldDurability,
				"new_durability":  item.Stats.Durability,
				"damage_received": damage,
			}).Debug("combat equipment durability reduced")
		}
	}
}

// spawnImpactParticles spawns particles for damage impact on armor.
func (s *CombatEquipmentDurabilityParticleSystem) spawnImpactParticles(entity *Entity, x, y, damage float64) {
	// Calculate particle count based on damage
	count := s.particleCountBase + int(damage*s.particleCountScale)
	if count > 20 {
		count = 20 // Cap maximum particles
	}

	// Genre-aware color selection
	primaryColor, secondaryColor := s.getImpactColors()

	config := particles.Config{
		Type:     particles.ParticleSparkle,
		Count:    count,
		Duration: 0.3,
		MinSize:  1.0,
		MaxSize:  2.5,
		SpreadX:  s.spreadBase,
		SpreadY:  s.spreadBase,
		Gravity:  -20.0, // Upward float for impact sparks
		Custom:   map[string]interface{}{"primary_color": primaryColor, "secondary_color": secondaryColor},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// spawnDegradationParticles spawns particles for equipment state transitions.
func (s *CombatEquipmentDurabilityParticleSystem) spawnDegradationParticles(entity *Entity, x, y float64, slot EquipmentSlot, state sprites.DamageState) {
	count := s.getParticleCount(state)
	spread := s.getSpread(state)
	primaryColor, secondaryColor := s.getDegradationColors(state)

	config := particles.Config{
		Type:     s.getParticleType(state),
		Count:    count,
		Duration: s.getDuration(state),
		MinSize:  s.getMinSize(state),
		MaxSize:  s.getMaxSize(state),
		SpreadX:  spread,
		SpreadY:  spread,
		Gravity:  s.getGravity(state),
		Custom:   map[string]interface{}{"primary_color": primaryColor, "secondary_color": secondaryColor},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"slot":         slot,
			"damage_state": state.String(),
			"count":        count,
		}).Debug("combat equipment degradation particles spawned")
	}
}

// getImpactColors returns genre-aware colors for impact sparks.
func (s *CombatEquipmentDurabilityParticleSystem) getImpactColors() (string, string) {
	switch s.genreID {
	case "fantasy":
		return "silver", "white"
	case "sci-fi":
		return "cyan", "blue"
	case "horror":
		return "gray", "darkgray"
	case "cyberpunk":
		return "orange", "yellow"
	case "post-apocalyptic":
		return "rust", "brown"
	default:
		return "silver", "white"
	}
}

// getDegradationColors returns genre-aware colors for degradation effects.
func (s *CombatEquipmentDurabilityParticleSystem) getDegradationColors(state sprites.DamageState) (string, string) {
	switch state {
	case sprites.DamageStateWorn:
		return "yellow", "orange"
	case sprites.DamageStateDamaged:
		return "orange", "red"
	case sprites.DamageStateBroken:
		return "darkred", "brown"
	default:
		return "gray", "white"
	}
}

// getParticleType returns the appropriate particle type for a damage state.
func (s *CombatEquipmentDurabilityParticleSystem) getParticleType(state sprites.DamageState) particles.ParticleType {
	switch state {
	case sprites.DamageStateWorn:
		return particles.ParticleSparkle
	case sprites.DamageStateDamaged:
		return particles.ParticleDebris
	case sprites.DamageStateBroken:
		return particles.ParticleSmoke
	default:
		return particles.ParticleSparkle
	}
}

// getParticleCount returns particle count for a damage state.
func (s *CombatEquipmentDurabilityParticleSystem) getParticleCount(state sprites.DamageState) int {
	switch state {
	case sprites.DamageStateWorn:
		return 10
	case sprites.DamageStateDamaged:
		return 18
	case sprites.DamageStateBroken:
		return 30
	default:
		return 8
	}
}

// getDuration returns particle lifetime based on damage state.
func (s *CombatEquipmentDurabilityParticleSystem) getDuration(state sprites.DamageState) float64 {
	switch state {
	case sprites.DamageStateWorn:
		return 0.4
	case sprites.DamageStateDamaged:
		return 0.6
	case sprites.DamageStateBroken:
		return 0.9
	default:
		return 0.5
	}
}

// getGravity returns particle gravity based on damage state.
func (s *CombatEquipmentDurabilityParticleSystem) getGravity(state sprites.DamageState) float64 {
	switch state {
	case sprites.DamageStateWorn:
		return -25.0 // Upward float for sparks
	case sprites.DamageStateDamaged:
		return 40.0 // Falling fragments
	case sprites.DamageStateBroken:
		return 70.0 // Heavy debris
	default:
		return 0.0
	}
}

// getMinSize returns minimum particle size for a damage state.
func (s *CombatEquipmentDurabilityParticleSystem) getMinSize(state sprites.DamageState) float64 {
	switch state {
	case sprites.DamageStateWorn:
		return 1.0
	case sprites.DamageStateDamaged:
		return 1.5
	case sprites.DamageStateBroken:
		return 2.0
	default:
		return 1.0
	}
}

// getMaxSize returns maximum particle size for a damage state.
func (s *CombatEquipmentDurabilityParticleSystem) getMaxSize(state sprites.DamageState) float64 {
	switch state {
	case sprites.DamageStateWorn:
		return 2.5
	case sprites.DamageStateDamaged:
		return 4.0
	case sprites.DamageStateBroken:
		return 5.5
	default:
		return 2.0
	}
}

// getSpread returns particle spread radius for a damage state.
func (s *CombatEquipmentDurabilityParticleSystem) getSpread(state sprites.DamageState) float64 {
	switch state {
	case sprites.DamageStateWorn:
		return 20.0
	case sprites.DamageStateDamaged:
		return 30.0
	case sprites.DamageStateBroken:
		return 45.0
	default:
		return 15.0
	}
}

// Helper methods

func (s *CombatEquipmentDurabilityParticleSystem) getEquipmentComponent(entity *Entity) *EquipmentComponent {
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

func (s *CombatEquipmentDurabilityParticleSystem) getEntityPosition(entity *Entity) (float64, float64) {
	if pos := entity.GetPosition(); pos != nil {
		return pos.X, pos.Y
	}
	return 0, 0
}
