// Package engine provides the CompanionManaRegenSystem for boosting
// an owner's mana regeneration based on companion bonding level and type.
// This connects CompanionComponent (loyalty, type, perks) with the owner's
// ManaComponent to create synergy between pet loyalty and magical sustain.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CompanionManaRegenSystem boosts owner mana regeneration based on their
// companions' bonding level, type, and unlocked perks.
// Magical companions (Spirit, Elemental) provide stronger mana bonuses.
type CompanionManaRegenSystem struct {
	world   *World
	genreID string
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry

	// Bonus configuration
	baseBonusPerLoyalty float64 // Base mana regen bonus per loyalty point (0-100)
	perkManaBonus       float64 // Additional bonus from relevant perks

	// Companion type multipliers (magical companions give more mana)
	typeMultipliers map[CompanionType]float64

	// Genre-specific multipliers
	genreMultipliers map[string]float64

	// Cache original mana regen per owner (ownerID -> original regen rate)
	originalRegen map[uint64]float64

	// Cache computed bonuses per owner (ownerID -> bonus multiplier)
	bonusMultiplier map[uint64]float64

	// Update throttle
	updateAccum float64
	updateRate  float64 // Update every N seconds
}

// NewCompanionManaRegenSystem creates a new companion mana regen system.
func NewCompanionManaRegenSystem(world *World, seed int64) *CompanionManaRegenSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "companion_mana_regen")
		logEntry.Debug("companion mana regen system created")
	}

	return &CompanionManaRegenSystem{
		world:               world,
		seed:                seed,
		rng:                 rand.New(rand.NewSource(seed)),
		logger:              logEntry,
		baseBonusPerLoyalty: 0.002, // 0.2% per loyalty point (max 20% at 100 loyalty)
		perkManaBonus:       0.05,  // +5% from relevant perks
		typeMultipliers: map[CompanionType]float64{
			CompanionTypePet:       0.6, // Regular pets give less mana bonus
			CompanionTypeSummon:    1.0, // Summoned creatures are balanced
			CompanionTypeHireling:  0.4, // Hirelings aren't magical
			CompanionTypeElemental: 1.5, // Elementals channel arcane energy
			CompanionTypeUndead:    0.8, // Undead drain magic slightly
			CompanionTypeRobot:     0.3, // Robots don't enhance mana
			CompanionTypeSpirit:    1.8, // Spirits are most magical
			CompanionTypeInsect:    0.5, // Insects have weak magical bonds
		},
		genreMultipliers: map[string]float64{
			"fantasy":   1.3, // Strong magical world
			"scifi":     0.5, // Tech-focused, less magic
			"horror":    1.1, // Dark magical energies
			"cyberpunk": 0.4, // Minimal magic
			"postapoc":  0.7, // Remnant magic
		},
		originalRegen:   make(map[uint64]float64),
		bonusMultiplier: make(map[uint64]float64),
		updateAccum:     0,
		updateRate:      0.5, // Update every 0.5 seconds
	}
}

// SetGenre sets the genre ID for genre-aware mana regen bonuses.
func (s *CompanionManaRegenSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes all companions and modifies owner mana regeneration.
func (s *CompanionManaRegenSystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Throttle updates for performance
	s.updateAccum += deltaTime
	if s.updateAccum < s.updateRate {
		return
	}
	s.updateAccum = 0

	// Track which owners have companions this cycle
	activeOwners := make(map[uint64]float64)

	// Process all companions and accumulate bonuses per owner
	for _, entity := range entities {
		compComp, ok := entity.GetComponent("companion")
		if !ok {
			continue
		}
		companionComp, ok := compComp.(*CompanionComponent)
		if !ok {
			continue
		}

		ownerID := companionComp.OwnerID
		if ownerID == 0 {
			continue
		}

		// Calculate bonus from this companion
		bonus := s.calculateBonus(companionComp)

		// Accumulate bonuses (multiple companions stack)
		activeOwners[ownerID] += bonus
	}

	// Apply bonuses to owner mana components
	s.applyManaRegenBonuses(entities, activeOwners)

	// Update cached bonuses
	s.bonusMultiplier = activeOwners
}

// calculateBonus computes mana regen bonus from a single companion.
func (s *CompanionManaRegenSystem) calculateBonus(comp *CompanionComponent) float64 {
	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	typeMult := s.typeMultipliers[comp.CompanionType]
	if typeMult == 0 {
		typeMult = 1.0
	}

	// Base bonus scales with loyalty (0-100)
	loyalty := comp.Loyalty
	if loyalty < 0 {
		loyalty = 0
	}
	if loyalty > 100 {
		loyalty = 100
	}

	bonus := loyalty * s.baseBonusPerLoyalty * genreMult * typeMult

	// Perk bonuses - certain perks enhance mana connection
	if comp.HasPerk(PerkFasterLearning) {
		// Faster Learning indicates stronger magical bond
		bonus += s.perkManaBonus * genreMult * typeMult
	}
	if comp.HasPerk(PerkSharedExperience) {
		// Shared Experience indicates deeper connection
		bonus += s.perkManaBonus * 0.5 * genreMult * typeMult
	}

	// Bonding level bonus: each perk adds small stacking bonus
	perkCount := len(comp.BondingPerks)
	if perkCount > 0 {
		stackBonus := float64(perkCount) * 0.008 * genreMult * typeMult
		bonus += stackBonus
	}

	return bonus
}

// applyManaRegenBonuses applies calculated bonuses to owner entities.
func (s *CompanionManaRegenSystem) applyManaRegenBonuses(entities []*Entity, activeOwners map[uint64]float64) {
	for _, entity := range entities {
		// Check if entity is an owner with active bonus
		bonus, hasBonus := activeOwners[entity.ID]

		manaComp, ok := entity.GetComponent("mana")
		if !ok {
			continue
		}
		mana, ok := manaComp.(*ManaComponent)
		if !ok {
			continue
		}

		if hasBonus && bonus > 0 {
			// Store original regen if not already stored
			if _, exists := s.originalRegen[entity.ID]; !exists {
				s.originalRegen[entity.ID] = mana.Regen
			}

			// Apply bonus multiplier
			originalRegen := s.originalRegen[entity.ID]
			newRegen := originalRegen * (1.0 + bonus)
			mana.Regen = newRegen

			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id":      entity.ID,
					"original_regen": originalRegen,
					"bonus":          bonus,
					"new_regen":      newRegen,
				}).Debug("mana regen boosted by companion")
			}
		} else {
			// Restore original regen if bonus removed
			if originalRegen, exists := s.originalRegen[entity.ID]; exists {
				mana.Regen = originalRegen
				delete(s.originalRegen, entity.ID)

				if s.logger != nil {
					s.logger.WithFields(logrus.Fields{
						"entity_id":      entity.ID,
						"restored_regen": originalRegen,
					}).Debug("mana regen restored")
				}
			}
		}
	}
}

// GetBonusMultiplier returns the current mana regen bonus multiplier for an owner.
// Returns 0 if owner has no active companion bonus.
func (s *CompanionManaRegenSystem) GetBonusMultiplier(ownerID uint64) float64 {
	return s.bonusMultiplier[ownerID]
}

// HasActiveBonus returns whether an owner has any active mana regen bonus.
func (s *CompanionManaRegenSystem) HasActiveBonus(ownerID uint64) bool {
	bonus := s.bonusMultiplier[ownerID]
	return bonus > 0
}

// GetActiveBonusCount returns the number of owners with active mana bonuses.
func (s *CompanionManaRegenSystem) GetActiveBonusCount() int {
	return len(s.bonusMultiplier)
}

// GetOriginalRegen returns the stored original regen rate for an entity.
// Returns 0 if no original rate is stored (entity has no companion bonus).
func (s *CompanionManaRegenSystem) GetOriginalRegen(entityID uint64) float64 {
	return s.originalRegen[entityID]
}
