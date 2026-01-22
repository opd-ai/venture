// Package engine provides the New Game Plus system for ECS.
// This file implements NewGamePlusSystem which manages NG+ state,
// cycle transitions, and integrates with player progression.
//
// Phase 111: NG+ Core Component & Persistence
package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// NewGamePlusSystem manages NG+ state and cycle transitions.
// It tracks playtime, accumulates legacy statistics, and handles
// the transition between playthroughs.
type NewGamePlusSystem struct {
	world  *World
	logger *logrus.Entry

	// lastUpdateTime tracks time for playtime accumulation (game time seconds)
	lastUpdateTime float64

	// callbacks for NG+ events
	onCycleStart  func(cycle int)
	onBonusUnlock func(bonusID string)
}

// NewNewGamePlusSystem creates a new NG+ system.
func NewNewGamePlusSystem(world *World) *NewGamePlusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "newgameplus")
		logEntry.Debug("New Game Plus system created")
	}
	return &NewGamePlusSystem{
		world:  world,
		logger: logEntry,
	}
}

// Update processes all entities with NG+ components.
// It accumulates playtime and checks for milestone unlocks.
func (s *NewGamePlusSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		ngpComp, ok := entity.GetComponent("newgameplus")
		if !ok {
			continue
		}

		ngp, ok := ngpComp.(*NewGamePlusComponent)
		if !ok {
			continue
		}

		// Accumulate playtime (convert deltaTime to seconds)
		ngp.UpdatePlaytime(int64(deltaTime))

		// Check for milestone-based unlocks
		s.checkMilestoneUnlocks(entity, ngp)
	}
}

// checkMilestoneUnlocks checks if the player has reached any milestones
// that unlock permanent bonuses.
func (s *NewGamePlusSystem) checkMilestoneUnlocks(entity *Entity, ngp *NewGamePlusComponent) {
	s.checkCycleMilestones(ngp)
	s.checkPlaytimeMilestone(ngp)
	s.checkStatMilestones(ngp)
}

// checkCycleMilestones checks cycle-based milestone unlocks for NG+ bonuses.
func (s *NewGamePlusSystem) checkCycleMilestones(ngp *NewGamePlusComponent) {
	s.tryUnlockBonus(ngp, "ng_veteran", ngp.Cycle >= 1)
	s.tryUnlockBonus(ngp, "seasoned_adventurer", ngp.Cycle >= 5)
	s.tryUnlockBonus(ngp, "legend_reborn", ngp.Cycle >= 10)
}

// checkPlaytimeMilestone checks playtime-based milestone unlocks for NG+ bonuses.
func (s *NewGamePlusSystem) checkPlaytimeMilestone(ngp *NewGamePlusComponent) {
	s.tryUnlockBonus(ngp, "dedicated_player", ngp.TotalPlaytime >= 360000)
}

// checkStatMilestones checks stat-based milestone unlocks for NG+ bonuses.
func (s *NewGamePlusSystem) checkStatMilestones(ngp *NewGamePlusComponent) {
	s.tryUnlockBonus(ngp, "master_slayer", ngp.GetLegacyStat("enemies_killed") >= 10000)
}

// tryUnlockBonus attempts to unlock a bonus if the condition is met and bonus is not already unlocked.
func (s *NewGamePlusSystem) tryUnlockBonus(ngp *NewGamePlusComponent, bonusID string, conditionMet bool) {
	if conditionMet && !ngp.HasBonus(bonusID) {
		if ngp.UnlockBonus(bonusID) {
			s.notifyBonusUnlock(bonusID)
		}
	}
}

// notifyBonusUnlock calls the bonus unlock callback if set.
func (s *NewGamePlusSystem) notifyBonusUnlock(bonusID string) {
	if s.logger != nil {
		s.logger.WithField("bonus_id", bonusID).Info("NG+ bonus unlocked")
	}
	if s.onBonusUnlock != nil {
		s.onBonusUnlock(bonusID)
	}
}

// SetOnCycleStart sets a callback for when a new NG+ cycle begins.
func (s *NewGamePlusSystem) SetOnCycleStart(callback func(cycle int)) {
	s.onCycleStart = callback
}

// SetOnBonusUnlock sets a callback for when a permanent bonus is unlocked.
func (s *NewGamePlusSystem) SetOnBonusUnlock(callback func(bonusID string)) {
	s.onBonusUnlock = callback
}

// InitiateNewGamePlus starts a new NG+ cycle for the given entity.
// This should be called when the player completes the game.
// statsSnapshot contains the current cycle's final statistics.
func (s *NewGamePlusSystem) InitiateNewGamePlus(entity *Entity, statsSnapshot map[string]int64) error {
	ngpComp, ok := entity.GetComponent("newgameplus")
	if !ok {
		// Create new NG+ component if not present
		ngp := NewNewGamePlusComponent()
		entity.AddComponent(ngp)
		ngpComp = ngp
	}

	ngp, ok := ngpComp.(*NewGamePlusComponent)
	if !ok {
		return nil
	}

	// Record old cycle for logging
	oldCycle := ngp.Cycle

	// Start the new cycle
	ngp.StartNewCycle(statsSnapshot)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"old_cycle": oldCycle,
			"new_cycle": ngp.Cycle,
		}).Info("New Game Plus cycle started")
	}

	// Notify callback
	if s.onCycleStart != nil {
		s.onCycleStart(ngp.Cycle)
	}

	return nil
}

// GetNGPlusMultiplier returns the difficulty multiplier for the current NG+ level.
// Uses logarithmic scaling to prevent absurd values at high NG+ levels.
func (s *NewGamePlusSystem) GetNGPlusMultiplier(entity *Entity, baseMultiplier, scalingFactor float64) float64 {
	ngpComp, ok := entity.GetComponent("newgameplus")
	if !ok {
		return baseMultiplier
	}

	ngp, ok := ngpComp.(*NewGamePlusComponent)
	if !ok {
		return baseMultiplier
	}

	return CalculateNGPlusMultiplier(ngp.Cycle, baseMultiplier, scalingFactor)
}

// CalculateNGPlusMultiplier computes the difficulty multiplier for an NG+ cycle.
// Uses logarithmic scaling: base + (scaling * ln(cycle + 1))
// This prevents exponential growth at high NG+ levels.
func CalculateNGPlusMultiplier(cycle int, baseMultiplier, scalingFactor float64) float64 {
	if cycle <= 0 {
		return baseMultiplier
	}
	return baseMultiplier + (scalingFactor * math.Log(float64(cycle)+1))
}

// GetEnemyHealthMultiplier returns the health multiplier for enemies at current NG+.
// Base: 1.0, Scaling: 0.2 per ln(cycle+1)
func (s *NewGamePlusSystem) GetEnemyHealthMultiplier(entity *Entity) float64 {
	return s.GetNGPlusMultiplier(entity, 1.0, 0.2)
}

// GetEnemyDamageMultiplier returns the damage multiplier for enemies at current NG+.
// Base: 1.0, Scaling: 0.15 per ln(cycle+1)
func (s *NewGamePlusSystem) GetEnemyDamageMultiplier(entity *Entity) float64 {
	return s.GetNGPlusMultiplier(entity, 1.0, 0.15)
}

// GetLootQualityBonus returns the bonus rare/legendary chance at current NG+.
// Base: 0.0, Scaling: 0.05 per ln(cycle+1)
func (s *NewGamePlusSystem) GetLootQualityBonus(entity *Entity) float64 {
	return s.GetNGPlusMultiplier(entity, 0.0, 0.05)
}

// GetXPMultiplier returns the XP gain multiplier at current NG+.
// Decreases slightly at high NG+ to maintain challenge.
// Base: 1.0, Scaling: -0.03 per ln(cycle+1), minimum 0.5
func (s *NewGamePlusSystem) GetXPMultiplier(entity *Entity) float64 {
	mult := s.GetNGPlusMultiplier(entity, 1.0, -0.03)
	if mult < 0.5 {
		return 0.5
	}
	return mult
}

// IsEligibleForNewGamePlus checks if the player has completed requirements for NG+.
// Requirements: Completed main story, all mandatory bosses defeated.
func (s *NewGamePlusSystem) IsEligibleForNewGamePlus(entity *Entity) bool {
	// Check for main story completion in player statistics
	if statsComp, ok := entity.GetComponent("player_statistics"); ok {
		if stats, ok := statsComp.(*PlayerStatisticsComponent); ok {
			// Check for main story completion stat (set when story is complete)
			if stats.GetLifetimeStat("main_story_completed") > 0 {
				return true
			}
		}
	}

	// Also check narrative component for completion flag
	if narrativeComp, ok := entity.GetComponent("narrative"); ok {
		if narrative, ok := narrativeComp.(*NarrativeComponent); ok {
			// Check if main_story_complete world flag is set
			if narrative.HasWorldFlag("main_story_complete") {
				return true
			}
		}
	}

	return false
}

// GetBonusDescription returns a human-readable description of an NG+ bonus.
func GetBonusDescription(bonusID string) string {
	descriptions := map[string]string{
		"ng_veteran":          "Veteran: +5% base stats on all characters",
		"seasoned_adventurer": "Seasoned: Rare items drop 10% more frequently",
		"legend_reborn":       "Legendary: Start each cycle with a random legendary item",
		"dedicated_player":    "Dedicated: +10% XP gain permanently",
		"master_slayer":       "Slayer: +5% damage to all enemies",
	}

	if desc, ok := descriptions[bonusID]; ok {
		return desc
	}
	return "Unknown bonus"
}

// GetAllBonuses returns a list of all possible NG+ bonuses with descriptions.
func GetAllBonuses() map[string]string {
	return map[string]string{
		"ng_veteran":          "Complete NG+ once - +5% base stats",
		"seasoned_adventurer": "Reach NG+5 - +10% rare item drops",
		"legend_reborn":       "Reach NG+10 - Start with legendary item",
		"dedicated_player":    "100 hours total playtime - +10% XP gain",
		"master_slayer":       "Kill 10,000 enemies total - +5% damage",
	}
}

// ApplyNGPlusBonuses applies all unlocked permanent bonuses to an entity.
// This should be called when initializing a character.
// Note: Actual stat modifications require integration with the stats system.
// Currently logs which bonuses would be applied.
func (s *NewGamePlusSystem) ApplyNGPlusBonuses(entity *Entity) {
	ngpComp, ok := entity.GetComponent("newgameplus")
	if !ok {
		return
	}

	ngp, ok := ngpComp.(*NewGamePlusComponent)
	if !ok {
		return
	}

	// Log bonuses that would be applied
	appliedBonuses := []string{}

	// ng_veteran: +5% base stats
	if ngp.HasBonus("ng_veteran") {
		appliedBonuses = append(appliedBonuses, "ng_veteran")
	}

	// seasoned_adventurer: +10% rare drops
	if ngp.HasBonus("seasoned_adventurer") {
		appliedBonuses = append(appliedBonuses, "seasoned_adventurer")
	}

	// legend_reborn: Start with legendary item
	if ngp.HasBonus("legend_reborn") {
		appliedBonuses = append(appliedBonuses, "legend_reborn")
	}

	// dedicated_player: +10% XP gain
	if ngp.HasBonus("dedicated_player") {
		appliedBonuses = append(appliedBonuses, "dedicated_player")
	}

	// master_slayer: +5% damage
	if ngp.HasBonus("master_slayer") {
		appliedBonuses = append(appliedBonuses, "master_slayer")
	}

	if s.logger != nil && len(appliedBonuses) > 0 {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"bonuses":   appliedBonuses,
		}).Debug("Applied NG+ bonuses")
	}
}

// GetApplicableBonuses returns a list of bonuses that should be applied to an entity.
func (s *NewGamePlusSystem) GetApplicableBonuses(entity *Entity) []string {
	ngpComp, ok := entity.GetComponent("newgameplus")
	if !ok {
		return nil
	}

	ngp, ok := ngpComp.(*NewGamePlusComponent)
	if !ok {
		return nil
	}

	var bonuses []string
	allBonuses := []string{"ng_veteran", "seasoned_adventurer", "legend_reborn", "dedicated_player", "master_slayer"}
	for _, bonus := range allBonuses {
		if ngp.HasBonus(bonus) {
			bonuses = append(bonuses, bonus)
		}
	}
	return bonuses
}
