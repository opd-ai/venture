// Package engine provides the ReputationMovementSpeedSystem which bridges faction
// reputation with entity movement speed. Players with high standing in allied
// factions move faster through familiar territory.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ReputationMovementSpeedSystem connects faction reputation with movement speed.
// When a player has positive standing with factions, they gain a passive movement
// speed bonus proportional to their highest faction reputation.
//
// Integration Points:
// - Reads from: FactionSystem (player reputation with factions)
// - Reads from: VelocityComponent (current velocity for multiplier application)
// - Modifies: VelocityComponent VX/VY via cached multiplier
//
// Speed Bonus Tiers (based on highest allied faction reputation):
// - Hostile (-100 to -50): No bonus
// - Suspicious (-49 to 0): No bonus
// - Neutral (1 to 50): +3% speed
// - Friendly (51 to 75): +6% speed
// - Honored (76 to 100): +10% speed
//
// Genre Modifiers:
// - Fantasy: Base multipliers (familiar roads guide the faithful)
// - Scifi: +10% bonus (allied transport network access)
// - Horror: -25% bonus (dread slows even the favored)
// - Cyberpunk: +15% bonus (gang territory fast-lanes)
// - PostApoc: -10% bonus (roads are unreliable regardless)
type ReputationMovementSpeedSystem struct {
	world         *World
	rng           *rand.Rand
	factionSystem *FactionSystem
	logger        *logrus.Entry
	genreID       string

	// updateInterval controls how frequently speed is recalculated (seconds)
	updateInterval float64
	timeSinceCheck float64

	// appliedMultipliers caches per-entity speed multipliers to reverse/reapply
	appliedMultipliers map[uint64]float64

	// speedBonuses caches per-entity bonus percentages for UI/query
	speedBonuses map[uint64]float64

	// genreMultipliers scales faction speed bonuses by genre
	genreMultipliers map[string]float64
}

// NewReputationMovementSpeedSystem creates a new reputation movement speed system.
func NewReputationMovementSpeedSystem(world *World, seed int64) *ReputationMovementSpeedSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_movement_speed")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("ReputationMovementSpeedSystem created")
		}
	}

	return &ReputationMovementSpeedSystem{
		world:              world,
		rng:                rand.New(rand.NewSource(seed)),
		updateInterval:     0.5,
		appliedMultipliers: make(map[uint64]float64),
		speedBonuses:       make(map[uint64]float64),
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
func (s *ReputationMovementSpeedSystem) SetFactionSystem(fs *FactionSystem) {
	s.factionSystem = fs
}

// SetGenre sets the genre for genre-aware speed bonus scaling.
func (s *ReputationMovementSpeedSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for reputation movement speed")
	}
}

// Update applies movement speed bonuses to player entities based on faction reputation.
func (s *ReputationMovementSpeedSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	if s.factionSystem == nil {
		return
	}

	newMultiplier := s.calculateSpeedMultiplier()

	for _, entity := range entities {
		s.processEntity(entity, newMultiplier)
	}
}

// processEntity applies reputation-based speed bonus to a single entity.
func (s *ReputationMovementSpeedSystem) processEntity(entity *Entity, newMultiplier float64) {
	if _, ok := entity.GetComponent("input"); !ok {
		return
	}

	vel := entity.GetVelocity()
	if vel == nil {
		return
	}

	currentMultiplier, hasMultiplier := s.appliedMultipliers[entity.ID]

	// Skip if no change needed
	if hasMultiplier && currentMultiplier == newMultiplier {
		return
	}

	// Restore original velocity if we had a previous multiplier
	if hasMultiplier && currentMultiplier != 1.0 {
		vel.VX /= currentMultiplier
		vel.VY /= currentMultiplier
	}

	// Apply new multiplier
	if newMultiplier != 1.0 {
		vel.VX *= newMultiplier
		vel.VY *= newMultiplier
	}

	s.appliedMultipliers[entity.ID] = newMultiplier
	s.speedBonuses[entity.ID] = (newMultiplier - 1.0) * 100.0

	if s.logger != nil && newMultiplier != 1.0 && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"multiplier": newMultiplier,
			"bonus_pct":  (newMultiplier - 1.0) * 100.0,
			"genre":      s.genreID,
		}).Debug("reputation movement speed bonus applied")
	}
}

// calculateSpeedMultiplier computes the speed multiplier from the player's highest faction reputation.
func (s *ReputationMovementSpeedSystem) calculateSpeedMultiplier() float64 {
	if s.factionSystem == nil {
		return 1.0
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
		return 1.0
	}

	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	return 1.0 + baseBonus*genreMult
}

// bonusForReputation returns the base speed bonus fraction for a given reputation value.
func (s *ReputationMovementSpeedSystem) bonusForReputation(reputation int) float64 {
	if reputation <= 0 {
		return 0.0
	}
	if reputation <= 50 {
		return 0.03 // Neutral: +3% speed
	}
	if reputation <= 75 {
		return 0.06 // Friendly: +6% speed
	}
	return 0.10 // Honored: +10% speed
}

// GetSpeedBonus returns the current reputation-based speed bonus percentage for an entity.
func (s *ReputationMovementSpeedSystem) GetSpeedBonus(entityID uint64) float64 {
	return s.speedBonuses[entityID]
}

// GetSpeedMultiplier returns the current reputation-based speed multiplier for an entity.
func (s *ReputationMovementSpeedSystem) GetSpeedMultiplier(entityID uint64) float64 {
	mult, ok := s.appliedMultipliers[entityID]
	if !ok {
		return 1.0
	}
	return mult
}

// GetGenreMultiplier returns the genre-specific speed multiplier.
func (s *ReputationMovementSpeedSystem) GetGenreMultiplier() float64 {
	mult := s.genreMultipliers[s.genreID]
	if mult == 0 {
		return 1.0
	}
	return mult
}
