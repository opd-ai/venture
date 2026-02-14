// Package engine provides the CompanionSpellAmplificationSystem for boosting
// an owner's spell damage and healing based on companion bonding level and perks.
// This connects CompanionComponent (bonding level, loyalty, perks) with the owner's
// spell effectiveness to create synergy between pet loyalty and magical power.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CompanionSpellAmplificationSystem boosts owner spell damage and healing
// based on their companions' bonding level and unlocked perks.
// Higher loyalty companions provide more spell amplification to their owner.
type CompanionSpellAmplificationSystem struct {
	world   *World
	genreID string
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry

	// Bonus configuration
	baseDamageBonusPerLoyalty  float64 // Base damage bonus per loyalty point (0-100)
	baseHealingBonusPerLoyalty float64 // Base healing bonus per loyalty point (0-100)
	perkDamageBonus            float64 // Additional damage bonus from PerkExtraDamage
	perkHealingBonus           float64 // Additional healing bonus from PerkExtraHealth

	// Genre-specific multipliers
	genreMultipliers map[string]float64

	// Cache spell bonuses per owner (ownerID -> damage multiplier)
	damageBonus  map[uint64]float64
	healingBonus map[uint64]float64

	// Update throttle
	updateAccum float64
	updateRate  float64 // Update every N seconds
}

// NewCompanionSpellAmplificationSystem creates a new companion spell amplification system.
func NewCompanionSpellAmplificationSystem(world *World, seed int64) *CompanionSpellAmplificationSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "companion_spell_amplification")
		logEntry.Debug("companion spell amplification system created")
	}

	return &CompanionSpellAmplificationSystem{
		world:                      world,
		seed:                       seed,
		rng:                        rand.New(rand.NewSource(seed)),
		logger:                     logEntry,
		baseDamageBonusPerLoyalty:  0.001,  // 0.1% per loyalty point (max 10% at 100 loyalty)
		baseHealingBonusPerLoyalty: 0.0008, // 0.08% per loyalty point (max 8% at 100 loyalty)
		perkDamageBonus:            0.05,   // +5% from PerkExtraDamage
		perkHealingBonus:           0.04,   // +4% from PerkExtraHealth
		genreMultipliers: map[string]float64{
			"fantasy":   1.2, // Strong magical bonds
			"scifi":     0.7, // Tech companions, less magic
			"horror":    1.0, // Balanced eldritch bonds
			"cyberpunk": 0.6, // Cyber pets, minimal magic
			"postapoc":  0.8, // Feral companions, some magic
		},
		damageBonus:  make(map[uint64]float64),
		healingBonus: make(map[uint64]float64),
		updateAccum:  0,
		updateRate:   0.5, // Update every 0.5 seconds
	}
}

// SetGenre sets the genre ID for genre-aware spell amplification.
func (s *CompanionSpellAmplificationSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes all companions and calculates spell amplification for their owners.
func (s *CompanionSpellAmplificationSystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Throttle updates for performance
	s.updateAccum += deltaTime
	if s.updateAccum < s.updateRate {
		return
	}
	s.updateAccum = 0

	// Reset bonuses each update cycle
	newDamageBonus := make(map[uint64]float64)
	newHealingBonus := make(map[uint64]float64)

	// Process all companions
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

		// Calculate bonuses from this companion
		dmgBonus, healBonus := s.calculateBonuses(companionComp)

		// Accumulate bonuses (multiple companions stack)
		newDamageBonus[ownerID] += dmgBonus
		newHealingBonus[ownerID] += healBonus
	}

	// Update cached bonuses
	s.damageBonus = newDamageBonus
	s.healingBonus = newHealingBonus

	// Log significant changes
	if s.logger != nil && len(newDamageBonus) > 0 {
		s.logger.WithFields(logrus.Fields{
			"active_owners": len(newDamageBonus),
		}).Debug("companion spell amplification updated")
	}
}

// calculateBonuses computes spell bonuses from a single companion.
func (s *CompanionSpellAmplificationSystem) calculateBonuses(comp *CompanionComponent) (damage, healing float64) {
	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	// Base bonus scales with loyalty (0-100)
	loyalty := comp.Loyalty
	if loyalty < 0 {
		loyalty = 0
	}
	if loyalty > 100 {
		loyalty = 100
	}

	damage = loyalty * s.baseDamageBonusPerLoyalty * genreMult
	healing = loyalty * s.baseHealingBonusPerLoyalty * genreMult

	// Perk bonuses
	if comp.HasPerk(PerkExtraDamage) {
		damage += s.perkDamageBonus * genreMult
	}
	if comp.HasPerk(PerkExtraHealth) {
		healing += s.perkHealingBonus * genreMult
	}

	// Bonding level bonus: each bonding perk adds a small stacking bonus
	perkCount := len(comp.BondingPerks)
	if perkCount > 0 {
		stackBonus := float64(perkCount) * 0.01 * genreMult // +1% per perk
		damage += stackBonus
		healing += stackBonus * 0.8 // Healing bonus is slightly less
	}

	return damage, healing
}

// GetDamageMultiplier returns the spell damage multiplier for an owner entity.
// Called by SpellCastingSystem when calculating spell damage.
func (s *CompanionSpellAmplificationSystem) GetDamageMultiplier(ownerID uint64) float64 {
	bonus, exists := s.damageBonus[ownerID]
	if !exists {
		return 1.0 // No bonus
	}
	return 1.0 + bonus
}

// GetHealingMultiplier returns the spell healing multiplier for an owner entity.
// Called by SpellCastingSystem when calculating spell healing.
func (s *CompanionSpellAmplificationSystem) GetHealingMultiplier(ownerID uint64) float64 {
	bonus, exists := s.healingBonus[ownerID]
	if !exists {
		return 1.0 // No bonus
	}
	return 1.0 + bonus
}

// GetBonusForOwner returns the current bonus values for an owner (for UI display).
func (s *CompanionSpellAmplificationSystem) GetBonusForOwner(ownerID uint64) (damageBonus, healingBonus float64) {
	return s.damageBonus[ownerID], s.healingBonus[ownerID]
}

// HasActiveBonus returns whether an owner has any active spell amplification bonus.
func (s *CompanionSpellAmplificationSystem) HasActiveBonus(ownerID uint64) bool {
	dmg := s.damageBonus[ownerID]
	heal := s.healingBonus[ownerID]
	return dmg > 0 || heal > 0
}

// GetActiveBonusCount returns the number of owners with active spell bonuses.
func (s *CompanionSpellAmplificationSystem) GetActiveBonusCount() int {
	return len(s.damageBonus)
}
