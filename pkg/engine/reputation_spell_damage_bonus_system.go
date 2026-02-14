// Package engine provides the ReputationSpellDamageBonusSystem which bridges faction
// reputation with spell damage output. Players with high reputation in allied
// factions gain a passive MagicPower bonus, rewarding diplomatic engagement.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ReputationSpellDamageBonusSystem connects faction reputation with spell damage.
// When a player has positive standing with factions, they gain a passive MagicPower
// bonus proportional to their highest faction reputation.
//
// Integration Points:
// - Reads from: FactionSystem (player reputation with factions)
// - Reads from: StatsComponent (MagicPower)
// - Modifies: StatsComponent.MagicPower (additive bonus)
//
// Spell Damage Bonus Tiers (based on highest allied faction reputation):
// - Hostile (-100 to -50): No bonus
// - Suspicious (-49 to 0): No bonus
// - Neutral (1 to 50): +2 MagicPower
// - Friendly (51 to 75): +5 MagicPower
// - Honored (76 to 100): +10 MagicPower
//
// Genre Modifiers:
// - Fantasy: Base multipliers (arcane knowledge from allied mages)
// - Scifi: +15% bonus (shared tech enhances energy weapons)
// - Horror: -25% bonus (eldritch power resists augmentation)
// - Cyberpunk: +10% bonus (netrunner network access)
// - PostApoc: -10% bonus (limited magical knowledge)
type ReputationSpellDamageBonusSystem struct {
	world         *World
	rng           *rand.Rand
	factionSystem *FactionSystem
	logger        *logrus.Entry
	genreID       string

	// updateInterval controls how frequently bonuses are recalculated (seconds)
	updateInterval float64
	timeSinceCheck float64

	// appliedBonuses tracks the bonus currently applied per entity to allow removal
	appliedBonuses map[uint64]float64

	// genreMultipliers scales spell damage bonuses by genre
	genreMultipliers map[string]float64
}

// NewReputationSpellDamageBonusSystem creates a new reputation spell damage bonus system.
func NewReputationSpellDamageBonusSystem(world *World, seed int64) *ReputationSpellDamageBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_spell_damage_bonus")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("ReputationSpellDamageBonusSystem created")
		}
	}

	return &ReputationSpellDamageBonusSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 1.0,
		appliedBonuses: make(map[uint64]float64),
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,
			"scifi":     1.15,
			"horror":    0.75,
			"cyberpunk": 1.10,
			"postapoc":  0.90,
		},
	}
}

// SetFactionSystem sets the faction system reference for reputation lookups.
func (s *ReputationSpellDamageBonusSystem) SetFactionSystem(fs *FactionSystem) {
	s.factionSystem = fs
}

// SetGenre sets the genre for genre-aware spell damage bonus scaling.
func (s *ReputationSpellDamageBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for reputation spell damage")
	}
}

// Update applies MagicPower bonuses to player entities based on faction reputation.
func (s *ReputationSpellDamageBonusSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	if s.factionSystem == nil {
		return
	}

	for _, entity := range entities {
		s.processEntity(entity)
	}
}

// processEntity applies or removes reputation-based spell damage bonus.
func (s *ReputationSpellDamageBonusSystem) processEntity(entity *Entity) {
	// Only apply to player-controlled entities
	if _, ok := entity.GetComponent("input"); !ok {
		s.removeBonus(entity)
		return
	}

	statsComp, ok := entity.GetComponent("stats")
	if !ok {
		s.removeBonus(entity)
		return
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		s.removeBonus(entity)
		return
	}

	// Calculate new bonus from best faction reputation
	newBonus := s.calculateBonus()

	// Remove old bonus and apply new one
	oldBonus := s.appliedBonuses[entity.ID]
	if oldBonus != newBonus {
		stats.MagicPower -= oldBonus
		stats.MagicPower += newBonus
		s.appliedBonuses[entity.ID] = newBonus

		if s.logger != nil && newBonus != oldBonus && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":    entity.ID,
				"old_bonus":    oldBonus,
				"new_bonus":    newBonus,
				"magic_power":  stats.MagicPower,
				"genre":        s.genreID,
			}).Debug("reputation spell damage bonus updated")
		}
	}
}

// removeBonus removes the previously applied bonus from an entity.
func (s *ReputationSpellDamageBonusSystem) removeBonus(entity *Entity) {
	oldBonus, exists := s.appliedBonuses[entity.ID]
	if !exists || oldBonus == 0 {
		delete(s.appliedBonuses, entity.ID)
		return
	}

	statsComp, ok := entity.GetComponent("stats")
	if ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			stats.MagicPower -= oldBonus
		}
	}
	delete(s.appliedBonuses, entity.ID)
}

// calculateBonus computes the MagicPower bonus from the player's highest faction reputation.
func (s *ReputationSpellDamageBonusSystem) calculateBonus() float64 {
	if s.factionSystem == nil {
		return 0
	}

	bestRep := 0
	for factionID := range s.factionSystem.Factions {
		rep := s.factionSystem.GetPlayerReputation(factionID)
		if rep > bestRep {
			bestRep = rep
		}
	}

	baseBonus := s.bonusForReputation(bestRep)
	if baseBonus <= 0 {
		return 0
	}

	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	return baseBonus * genreMult
}

// bonusForReputation returns the base MagicPower bonus for a given reputation value.
func (s *ReputationSpellDamageBonusSystem) bonusForReputation(reputation int) float64 {
	if reputation <= 0 {
		return 0.0
	}
	if reputation <= 50 {
		return 2.0 // Neutral: +2 MagicPower
	}
	if reputation <= 75 {
		return 5.0 // Friendly: +5 MagicPower
	}
	return 10.0 // Honored: +10 MagicPower
}

// GetBonus returns the current reputation-based spell damage bonus for an entity.
func (s *ReputationSpellDamageBonusSystem) GetBonus(entityID uint64) float64 {
	return s.appliedBonuses[entityID]
}

// GetGenreMultiplier returns the genre-specific spell damage multiplier.
func (s *ReputationSpellDamageBonusSystem) GetGenreMultiplier() float64 {
	mult := s.genreMultipliers[s.genreID]
	if mult == 0 {
		return 1.0
	}
	return mult
}
