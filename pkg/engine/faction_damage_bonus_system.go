// Package engine provides the FactionDamageBonusSystem which bridges faction
// reputation with outgoing damage modifiers in combat.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// FactionDamageBonusSystem connects faction reputation with combat damage.
// When the player attacks an enemy of a faction they have good standing with,
// they deal bonus damage based on their reputation level with that faction.
//
// Integration Points:
// - Reads from: FactionComponent (player faction reputation)
// - Reads from: FactionSystem (faction relationships - who is enemy of whom)
// - Modifies: StatsComponent.Attack (temporarily for damage calculation)
//
// Damage Bonus Tiers (based on reputation with factions who consider target enemy):
// - Hostile (-100 to -50): No bonus (they're your enemies too)
// - Suspicious (-49 to 0): No bonus
// - Neutral (1 to 50): +3% damage bonus
// - Friendly (51 to 100): +5% to +15% damage bonus (scales with reputation)
//
// Genre Modifiers:
// - Fantasy: Base multipliers (honor and allegiance matter)
// - Scifi: +10% faction bonuses (corporate loyalty rewarded)
// - Horror: -20% faction bonuses (survival trumps politics)
// - Cyberpunk: +15% faction bonuses (gang loyalty is everything)
// - PostApoc: -10% faction bonuses (alliances are fragile)
type FactionDamageBonusSystem struct {
	world         *World
	rng           *rand.Rand
	factionSystem *FactionSystem
	logger        *logrus.Entry
	genreID       string

	// damageMultipliers caches damage bonus multipliers per entity per frame.
	// Maps attackerID -> target factionID -> multiplier
	damageMultipliers map[uint64]map[string]float64

	// genreMultipliers scales faction damage bonuses by genre
	genreMultipliers map[string]float64
}

// NewFactionDamageBonusSystem creates a new faction damage bonus system.
func NewFactionDamageBonusSystem(world *World, seed int64) *FactionDamageBonusSystem {
	logger := world.GetLogger()
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system_name", "faction_damage_bonus")
	}

	return &FactionDamageBonusSystem{
		world:             world,
		rng:               rand.New(rand.NewSource(seed)),
		damageMultipliers: make(map[uint64]map[string]float64),
		logger:            logEntry,
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,
			"scifi":     1.1,
			"horror":    0.8,
			"cyberpunk": 1.15,
			"postapoc":  0.9,
		},
	}
}

// SetFactionSystem sets the faction system reference for relationship lookups.
func (s *FactionDamageBonusSystem) SetFactionSystem(fs *FactionSystem) {
	s.factionSystem = fs
}

// SetGenre sets the genre for genre-aware damage bonus scaling.
func (s *FactionDamageBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for damage bonuses")
	}
}

// Update recalculates damage multipliers for all player entities.
// Called each frame to refresh faction-based damage bonuses.
func (s *FactionDamageBonusSystem) Update(entities []*Entity, deltaTime float64) {
	// Clear cache each frame
	for k := range s.damageMultipliers {
		delete(s.damageMultipliers, k)
	}

	if s.factionSystem == nil {
		return
	}

	// Only process player entities
	for _, entity := range entities {
		if _, ok := entity.GetComponent("input"); !ok {
			continue
		}

		// Pre-calculate multipliers for this player against all enemy factions
		s.damageMultipliers[entity.ID] = s.calculateFactionMultipliers()
	}
}

// calculateFactionMultipliers computes damage bonuses against all enemy factions.
func (s *FactionDamageBonusSystem) calculateFactionMultipliers() map[string]float64 {
	multipliers := make(map[string]float64)

	if s.factionSystem == nil {
		return multipliers
	}

	// For each faction, check if player has standing that grants damage bonus
	for factionID, faction := range s.factionSystem.Factions {
		playerRep := s.factionSystem.GetPlayerReputation(factionID)
		bonusPercent := s.calculateBonusPercent(playerRep)

		if bonusPercent <= 0 {
			continue
		}

		// Apply genre multiplier
		genreMult := s.genreMultipliers[s.genreID]
		if genreMult == 0 {
			genreMult = 1.0
		}
		finalBonus := bonusPercent * genreMult

		// Store bonus for enemies of this faction
		// Check relationships to find enemies (relationship <= -50)
		for otherFactionID := range s.factionSystem.Factions {
			if faction.IsEnemy(otherFactionID) {
				if existing, ok := multipliers[otherFactionID]; ok {
					// Take the higher bonus if multiple factions hate this enemy
					if finalBonus > existing {
						multipliers[otherFactionID] = finalBonus
					}
				} else {
					multipliers[otherFactionID] = finalBonus
				}
			}
		}
	}

	return multipliers
}

// calculateBonusPercent returns the damage bonus percentage based on reputation.
// Returns 0.0 for hostile/suspicious, 0.03 for neutral, 0.05-0.15 for friendly.
func (s *FactionDamageBonusSystem) calculateBonusPercent(reputation int) float64 {
	if reputation <= 0 {
		return 0.0 // Hostile or suspicious - no bonus
	}

	if reputation <= 50 {
		return 0.03 // Neutral - 3% bonus
	}

	// Friendly: scales from 5% at 51 rep to 15% at 100 rep
	// Formula: 0.05 + (rep - 50) / 50 * 0.10
	scaledBonus := 0.05 + float64(reputation-50)/50.0*0.10
	return scaledBonus
}

// GetDamageMultiplier returns the damage multiplier for an attacker against a target.
// Returns 1.0 (no bonus) if no faction relationship applies.
// The multiplier is additive: 1.0 + bonus (e.g., 1.10 for 10% bonus).
func (s *FactionDamageBonusSystem) GetDamageMultiplier(attackerID uint64, target *Entity) float64 {
	if target == nil {
		return 1.0
	}

	// Get target's faction
	factionComp, ok := target.GetComponent("faction")
	if !ok {
		return 1.0
	}

	fc := factionComp.(*FactionComponent)
	if fc.IsPlayerFaction {
		return 1.0 // Don't boost damage against player-aligned entities
	}

	// Look up cached multiplier for this attacker
	attackerMults, ok := s.damageMultipliers[attackerID]
	if !ok {
		return 1.0
	}

	bonus, ok := attackerMults[fc.FactionID]
	if !ok {
		return 1.0
	}

	return 1.0 + bonus
}

// OnAttack should be called when a player attacks a target.
// It logs the faction damage bonus applied (if any) for debugging.
func (s *FactionDamageBonusSystem) OnAttack(attacker, target *Entity, baseDamage float64) float64 {
	if attacker == nil || target == nil {
		return baseDamage
	}

	multiplier := s.GetDamageMultiplier(attacker.ID, target)
	if multiplier == 1.0 {
		return baseDamage // No bonus
	}

	finalDamage := baseDamage * multiplier

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		factionComp, _ := target.GetComponent("faction")
		fc := factionComp.(*FactionComponent)
		s.logger.WithFields(logrus.Fields{
			"attackerID":    attacker.ID,
			"targetID":      target.ID,
			"targetFaction": fc.FactionID,
			"multiplier":    multiplier,
			"baseDamage":    baseDamage,
			"finalDamage":   finalDamage,
			"genre":         s.genreID,
		}).Debug("faction damage bonus applied")
	}

	return finalDamage
}

// GetBonusPercent returns the current damage bonus percentage against a faction.
// Useful for UI display showing potential bonus damage against faction enemies.
func (s *FactionDamageBonusSystem) GetBonusPercent(factionID string) float64 {
	if s.factionSystem == nil {
		return 0.0
	}
	rep := s.factionSystem.GetPlayerReputation(factionID)
	baseBonus := s.calculateBonusPercent(rep)

	// Apply genre multiplier
	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	return baseBonus * genreMult
}

// GetGenreMultiplier returns the genre-specific faction damage multiplier.
// Used for UI display and testing.
func (s *FactionDamageBonusSystem) GetGenreMultiplier() float64 {
	mult := s.genreMultipliers[s.genreID]
	if mult == 0 {
		return 1.0
	}
	return mult
}
