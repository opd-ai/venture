// Package engine provides the ReputationEquipmentDurabilitySystem which bridges
// faction reputation with equipment durability preservation. Players with high
// faction standing get reduced equipment degradation rates from allied faction
// blessings/maintenance, while hostile reputation can accelerate wear.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ReputationEquipmentDurabilitySystem connects faction reputation with equipment
// durability. It applies a periodic durability protection or penalty modifier
// based on the player's highest and lowest faction reputations.
//
// Integration Points:
// - Reads from: FactionSystem (player reputation with factions)
// - Reads from: EquipmentComponent (equipped item durability)
// - Modifies: EquipmentComponent item durability via periodic adjustment
// - Triggers: EquipmentVisualComponent dirty flag on state change
//
// Protection Tiers (based on highest allied faction reputation):
// - Hostile (-100 to -50): +15% degradation (cursed/sabotaged gear)
// - Suspicious (-49 to 0): No effect
// - Neutral (1 to 50): -5% degradation (minor upkeep)
// - Friendly (51 to 75): -12% degradation (faction maintenance)
// - Honored (76 to 100): -20% degradation (master-crafted care)
//
// Genre Modifiers:
// - Fantasy: 1.0 (divine blessing preserves equipment)
// - Scifi: 1.2 (nano-repair contracts from allied corps)
// - Horror: 0.6 (cursed realms resist all protection)
// - Cyberpunk: 1.1 (faction mechanics keep gear running)
// - PostApoc: 0.8 (salvage skills only go so far)
type ReputationEquipmentDurabilitySystem struct {
	world         *World
	rng           *rand.Rand
	factionSystem *FactionSystem
	logger        *logrus.Entry
	genreID       string

	// updateInterval controls how frequently durability modifiers are applied (seconds)
	updateInterval float64
	timeSinceCheck float64

	// appliedModifiers caches per-entity modifier to detect changes
	appliedModifiers map[uint64]float64

	// genreMultipliers scales faction durability protection by genre
	genreMultipliers map[string]float64
}

// NewReputationEquipmentDurabilitySystem creates a new reputation equipment durability system.
func NewReputationEquipmentDurabilitySystem(world *World, seed int64) *ReputationEquipmentDurabilitySystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_equipment_durability")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("ReputationEquipmentDurabilitySystem created")
		}
	}

	return &ReputationEquipmentDurabilitySystem{
		world:            world,
		rng:              rand.New(rand.NewSource(seed)),
		updateInterval:   2.0, // Check every 2 seconds
		appliedModifiers: make(map[uint64]float64),
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,
			"scifi":     1.2,
			"horror":    0.6,
			"cyberpunk": 1.1,
			"postapoc":  0.8,
		},
	}
}

// SetFactionSystem sets the faction system reference for reputation lookups.
func (s *ReputationEquipmentDurabilitySystem) SetFactionSystem(fs *FactionSystem) {
	s.factionSystem = fs
}

// SetGenre sets the genre for genre-aware durability scaling.
func (s *ReputationEquipmentDurabilitySystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for reputation equipment durability")
	}
}

// Update applies durability protection/penalty to player entities based on faction reputation.
func (s *ReputationEquipmentDurabilitySystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	if s.factionSystem == nil {
		return
	}

	modifier := s.calculateDurabilityModifier()

	for _, entity := range entities {
		s.processEntity(entity, modifier)
	}
}

// processEntity applies reputation-based durability restoration to a single entity.
func (s *ReputationEquipmentDurabilitySystem) processEntity(entity *Entity, modifier float64) {
	// Only affect player entities
	if _, ok := entity.GetComponent("input"); !ok {
		return
	}

	equipComp := s.getEquipmentComponent(entity)
	if equipComp == nil {
		return
	}

	prevModifier := s.appliedModifiers[entity.ID]
	s.appliedModifiers[entity.ID] = modifier

	// If modifier is positive (protection from high rep), slowly restore durability
	// If modifier is negative (penalty from hostile rep), slowly degrade durability
	if modifier == 0 {
		return
	}

	visualDirty := false
	for _, item := range equipComp.Slots {
		if item == nil || item.Stats.DurabilityMax <= 0 {
			continue
		}

		oldDurability := item.Stats.Durability

		if modifier > 0 {
			// Positive modifier: restore a small amount of durability
			restore := int(modifier * float64(item.Stats.DurabilityMax) * 0.01)
			if restore < 1 {
				restore = 1
			}
			item.Stats.Durability += restore
			if item.Stats.Durability > item.Stats.DurabilityMax {
				item.Stats.Durability = item.Stats.DurabilityMax
			}
		} else {
			// Negative modifier: degrade durability
			damage := int(-modifier * float64(item.Stats.DurabilityMax) * 0.005)
			if damage < 1 {
				damage = 1
			}
			item.Stats.Durability -= damage
			if item.Stats.Durability < 0 {
				item.Stats.Durability = 0
			}
		}

		if s.damageStateChanged(oldDurability, item.Stats.Durability, item.Stats.DurabilityMax) {
			visualDirty = true
		}
	}

	if visualDirty {
		s.markVisualDirty(entity)
		equipComp.StatsDirty = true
	}

	if modifier != prevModifier && s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"modifier":  modifier,
			"genre":     s.genreID,
		}).Debug("reputation equipment durability modifier applied")
	}
}

// calculateDurabilityModifier computes the durability modifier from faction reputation.
// Positive values = protection (slow restoration), negative = penalty (faster degradation).
func (s *ReputationEquipmentDurabilitySystem) calculateDurabilityModifier() float64 {
	if s.factionSystem == nil {
		return 0.0
	}

	bestRep := 0
	worstRep := 0
	for factionID := range s.factionSystem.Factions {
		rep := s.factionSystem.GetPlayerReputation(factionID)
		if rep > bestRep {
			bestRep = rep
		}
		if rep < worstRep {
			worstRep = rep
		}
	}

	baseModifier := s.modifierForReputation(bestRep, worstRep)
	if baseModifier == 0.0 {
		return 0.0
	}

	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	return baseModifier * genreMult
}

// modifierForReputation returns the base durability modifier for given reputation values.
func (s *ReputationEquipmentDurabilitySystem) modifierForReputation(bestRep, worstRep int) float64 {
	// High ally reputation provides protection
	if bestRep > 75 {
		return 0.20 // Honored: 20% protection
	}
	if bestRep > 50 {
		return 0.12 // Friendly: 12% protection
	}
	if bestRep > 0 {
		return 0.05 // Neutral-positive: 5% protection
	}

	// Hostile reputation causes degradation
	if worstRep < -50 {
		return -0.15 // Hostile: 15% penalty
	}

	return 0.0
}

// GetDurabilityModifier returns the current modifier for an entity (positive=protection, negative=penalty).
func (s *ReputationEquipmentDurabilitySystem) GetDurabilityModifier(entityID uint64) float64 {
	return s.appliedModifiers[entityID]
}

// GetDurabilityModifierPercent returns the modifier as a percentage.
func (s *ReputationEquipmentDurabilitySystem) GetDurabilityModifierPercent(entityID uint64) float64 {
	return s.appliedModifiers[entityID] * 100.0
}

// GetGenreMultiplier returns the genre-specific durability modifier multiplier.
func (s *ReputationEquipmentDurabilitySystem) GetGenreMultiplier() float64 {
	mult := s.genreMultipliers[s.genreID]
	if mult == 0 {
		return 1.0
	}
	return mult
}

// getEquipmentComponent retrieves the equipment component from an entity.
func (s *ReputationEquipmentDurabilitySystem) getEquipmentComponent(entity *Entity) *EquipmentComponent {
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

// damageStateChanged checks if the durability crossed a damage state threshold.
func (s *ReputationEquipmentDurabilitySystem) damageStateChanged(oldDur, newDur, maxDur int) bool {
	if maxDur == 0 {
		return false
	}
	oldPct := float64(oldDur) / float64(maxDur)
	newPct := float64(newDur) / float64(maxDur)
	thresholds := []float64{0.75, 0.50, 0.25}
	for _, t := range thresholds {
		if (oldPct > t && newPct <= t) || (oldPct <= t && newPct > t) {
			return true
		}
	}
	return false
}

// markVisualDirty marks the equipment visual component as dirty for regeneration.
func (s *ReputationEquipmentDurabilitySystem) markVisualDirty(entity *Entity) {
	comp, ok := entity.GetComponent("equipment_visual")
	if !ok || comp == nil {
		return
	}
	visualComp, ok := comp.(*EquipmentVisualComponent)
	if !ok {
		return
	}
	visualComp.MarkDirty()
}
