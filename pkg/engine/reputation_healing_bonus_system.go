// Package engine provides the ReputationHealingBonusSystem which bridges faction
// reputation with health regeneration. Players with high reputation in allied
// factions passively regenerate health faster.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ReputationHealingBonusSystem connects faction reputation with health regeneration.
// When a player has positive standing with factions, they gain a passive health
// regeneration bonus proportional to their highest faction reputation.
//
// Integration Points:
// - Reads from: FactionSystem (player reputation with factions)
// - Reads from: HealthComponent (current/max health)
// - Modifies: HealthComponent.Current via Heal()
//
// Healing Bonus Tiers (based on highest allied faction reputation):
// - Hostile (-100 to -50): No bonus
// - Suspicious (-49 to 0): No bonus
// - Neutral (1 to 50): +0.5 HP/sec
// - Friendly (51 to 75): +1.0 HP/sec
// - Honored (76 to 100): +2.0 HP/sec
//
// Genre Modifiers:
// - Fantasy: Base multipliers (divine blessing heals the faithful)
// - Scifi: +10% regen (advanced med-tech from allied corps)
// - Horror: -30% regen (wounds fester regardless of allies)
// - Cyberpunk: +15% regen (street doc network access)
// - PostApoc: -15% regen (limited medical supplies)
type ReputationHealingBonusSystem struct {
	world         *World
	rng           *rand.Rand
	factionSystem *FactionSystem
	logger        *logrus.Entry
	genreID       string

	// updateInterval controls how frequently healing ticks (seconds)
	updateInterval float64
	timeSinceCheck float64

	// regenRates caches per-entity regen rates for UI/query
	regenRates map[uint64]float64

	// genreMultipliers scales faction healing bonuses by genre
	genreMultipliers map[string]float64
}

// NewReputationHealingBonusSystem creates a new reputation healing bonus system.
func NewReputationHealingBonusSystem(world *World, seed int64) *ReputationHealingBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_healing_bonus")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("ReputationHealingBonusSystem created")
		}
	}

	return &ReputationHealingBonusSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 0.5,
		regenRates:     make(map[uint64]float64),
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,
			"scifi":     1.1,
			"horror":    0.7,
			"cyberpunk": 1.15,
			"postapoc":  0.85,
		},
	}
}

// SetFactionSystem sets the faction system reference for reputation lookups.
func (s *ReputationHealingBonusSystem) SetFactionSystem(fs *FactionSystem) {
	s.factionSystem = fs
}

// SetGenre sets the genre for genre-aware healing bonus scaling.
func (s *ReputationHealingBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for reputation healing")
	}
}

// Update applies health regeneration bonuses to player entities based on faction reputation.
func (s *ReputationHealingBonusSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	elapsed := s.timeSinceCheck
	s.timeSinceCheck = 0

	if s.factionSystem == nil {
		return
	}

	for _, entity := range entities {
		s.processEntity(entity, elapsed)
	}
}

// processEntity applies reputation-based healing to a single entity.
func (s *ReputationHealingBonusSystem) processEntity(entity *Entity, elapsed float64) {
	// Only apply to player-controlled entities
	if _, ok := entity.GetComponent("input"); !ok {
		delete(s.regenRates, entity.ID)
		return
	}

	healthComp, ok := entity.GetComponent("health")
	if !ok {
		delete(s.regenRates, entity.ID)
		return
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok || health.IsDead() {
		delete(s.regenRates, entity.ID)
		return
	}

	// Don't regen if at full health
	if health.Current >= health.Max {
		s.regenRates[entity.ID] = 0
		return
	}

	// Calculate regen rate from best faction reputation
	regenRate := s.calculateRegenRate()
	s.regenRates[entity.ID] = regenRate

	if regenRate <= 0 {
		return
	}

	regenAmount := regenRate * elapsed
	oldHealth := health.Current
	health.Heal(regenAmount)

	if s.logger != nil && health.Current != oldHealth && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"regen_rate":  regenRate,
			"regen_amount": regenAmount,
			"old_health":  oldHealth,
			"new_health":  health.Current,
			"max_health":  health.Max,
			"genre":       s.genreID,
		}).Debug("reputation healing bonus applied")
	}
}

// calculateRegenRate computes the health regen rate from the player's highest faction reputation.
func (s *ReputationHealingBonusSystem) calculateRegenRate() float64 {
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

	baseRegen := s.regenForReputation(bestRep)
	if baseRegen <= 0 {
		return 0
	}

	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	return baseRegen * genreMult
}

// regenForReputation returns the base HP/sec regen rate for a given reputation value.
func (s *ReputationHealingBonusSystem) regenForReputation(reputation int) float64 {
	if reputation <= 0 {
		return 0.0
	}
	if reputation <= 50 {
		return 0.5 // Neutral: 0.5 HP/sec
	}
	if reputation <= 75 {
		return 1.0 // Friendly: 1.0 HP/sec
	}
	return 2.0 // Honored: 2.0 HP/sec
}

// GetRegenRate returns the current reputation-based regen rate for an entity.
func (s *ReputationHealingBonusSystem) GetRegenRate(entityID uint64) float64 {
	return s.regenRates[entityID]
}

// GetGenreMultiplier returns the genre-specific healing multiplier.
func (s *ReputationHealingBonusSystem) GetGenreMultiplier() float64 {
	mult := s.genreMultipliers[s.genreID]
	if mult == 0 {
		return 1.0
	}
	return mult
}
