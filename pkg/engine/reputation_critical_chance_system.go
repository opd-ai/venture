// Package engine provides the ReputationCriticalChanceBonusSystem which bridges
// faction reputation with critical hit chance. Players with high standing among
// allied factions gain tactical insight that increases their critical strike rate.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ReputationCriticalChanceBonusSystem connects faction reputation with critical
// hit chance. When a player has positive standing with factions, they gain a
// passive critical chance bonus proportional to their highest faction reputation.
//
// Integration Points:
// - Reads from: FactionSystem (player reputation with factions)
// - Reads from: StatsComponent (base CritChance value)
// - Modifies: StatsComponent.CritChance via additive bonus
//
// Crit Chance Bonus Tiers (based on highest allied faction reputation):
// - Hostile (-100 to -50): No bonus
// - Suspicious (-49 to 0): No bonus
// - Neutral (1 to 50): +2% crit chance
// - Friendly (51 to 75): +4% crit chance
// - Honored (76 to 100): +7% crit chance
//
// Genre Modifiers:
// - Fantasy: Base multipliers (divine favor guides the blade)
// - Scifi: +10% bonus (targeting computer upgrades from allies)
// - Horror: -25% bonus (fear dulls precision)
// - Cyberpunk: +15% bonus (black-market implant sharpening)
// - PostApoc: -10% bonus (salvaged gear is imprecise)
type ReputationCriticalChanceBonusSystem struct {
	world         *World
	rng           *rand.Rand
	factionSystem *FactionSystem
	logger        *logrus.Entry
	genreID       string

	// updateInterval controls how frequently crit chance is recalculated (seconds)
	updateInterval float64
	timeSinceCheck float64

	// appliedBonuses caches per-entity crit chance bonus to reverse/reapply
	appliedBonuses map[uint64]float64

	// genreMultipliers scales faction crit bonuses by genre
	genreMultipliers map[string]float64
}

// NewReputationCriticalChanceBonusSystem creates a new reputation critical chance bonus system.
func NewReputationCriticalChanceBonusSystem(world *World, seed int64) *ReputationCriticalChanceBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_critical_chance")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("ReputationCriticalChanceBonusSystem created")
		}
	}

	return &ReputationCriticalChanceBonusSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 0.5,
		appliedBonuses: make(map[uint64]float64),
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,
			"scifi":     1.1,
			"horror":    0.75,
			"cyberpunk": 1.15,
			"postapoc":  0.9,
		},
	}
}

// SetFactionSystem sets the faction system reference for reputation lookups.
func (s *ReputationCriticalChanceBonusSystem) SetFactionSystem(fs *FactionSystem) {
	s.factionSystem = fs
}

// SetGenre sets the genre for genre-aware crit bonus scaling.
func (s *ReputationCriticalChanceBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for reputation critical chance")
	}
}

// Update applies critical chance bonuses to player entities based on faction reputation.
func (s *ReputationCriticalChanceBonusSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	if s.factionSystem == nil {
		return
	}

	newBonus := s.calculateCritBonus()

	for _, entity := range entities {
		s.processEntity(entity, newBonus)
	}
}

// processEntity applies reputation-based crit chance bonus to a single entity.
func (s *ReputationCriticalChanceBonusSystem) processEntity(entity *Entity, newBonus float64) {
	if _, ok := entity.GetComponent("input"); !ok {
		return
	}

	statsComp, ok := entity.GetComponent("stats")
	if !ok {
		return
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return
	}

	currentBonus := s.appliedBonuses[entity.ID]

	if currentBonus == newBonus {
		return
	}

	// Remove old bonus, apply new one
	stats.CritChance -= currentBonus
	stats.CritChance += newBonus

	// Clamp to [0.0, 1.0]
	if stats.CritChance < 0.0 {
		stats.CritChance = 0.0
	} else if stats.CritChance > 1.0 {
		stats.CritChance = 1.0
	}

	s.appliedBonuses[entity.ID] = newBonus

	if s.logger != nil && newBonus != 0.0 && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"bonus":     newBonus,
			"bonus_pct": newBonus * 100.0,
			"genre":     s.genreID,
		}).Debug("reputation critical chance bonus applied")
	}
}

// calculateCritBonus computes the crit chance bonus from the player's highest faction reputation.
func (s *ReputationCriticalChanceBonusSystem) calculateCritBonus() float64 {
	if s.factionSystem == nil {
		return 0.0
	}

	bestRep := 0
	for factionID := range s.factionSystem.Factions {
		rep := s.factionSystem.GetPlayerReputation(factionID)
		if rep > bestRep {
			bestRep = rep
		}
	}

	baseBonus := s.bonusForReputation(bestRep)
	if baseBonus <= 0.0 {
		return 0.0
	}

	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	return baseBonus * genreMult
}

// bonusForReputation returns the base crit chance bonus for a given reputation value.
func (s *ReputationCriticalChanceBonusSystem) bonusForReputation(reputation int) float64 {
	if reputation <= 0 {
		return 0.0
	}
	if reputation <= 50 {
		return 0.02 // Neutral: +2% crit
	}
	if reputation <= 75 {
		return 0.04 // Friendly: +4% crit
	}
	return 0.07 // Honored: +7% crit
}

// GetCritBonus returns the current reputation-based crit bonus for an entity (0.0-1.0).
func (s *ReputationCriticalChanceBonusSystem) GetCritBonus(entityID uint64) float64 {
	return s.appliedBonuses[entityID]
}

// GetCritBonusPercent returns the current reputation-based crit bonus as a percentage.
func (s *ReputationCriticalChanceBonusSystem) GetCritBonusPercent(entityID uint64) float64 {
	return s.appliedBonuses[entityID] * 100.0
}

// GetGenreMultiplier returns the genre-specific crit bonus multiplier.
func (s *ReputationCriticalChanceBonusSystem) GetGenreMultiplier() float64 {
	mult := s.genreMultipliers[s.genreID]
	if mult == 0 {
		return 1.0
	}
	return mult
}
