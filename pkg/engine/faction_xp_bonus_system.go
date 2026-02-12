package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// FactionXPBonusSystem connects faction reputation with XP rewards.
// When the player kills an enemy of a faction they have good standing with,
// they receive bonus XP based on their reputation level with that faction.
//
// Integration Points:
// - Reads from: FactionComponent (player faction reputation)
// - Reads from: FactionSystem (faction relationships - who is enemy of whom)
// - Modifies: ExperienceComponent (awards bonus XP via ProgressionSystem)
//
// XP Bonus Tiers (based on reputation with the faction whose enemy was killed):
// - Hostile (-100 to -50): No bonus (they're your enemies too)
// - Suspicious (-49 to 0): No bonus
// - Neutral (1 to 50): +5% XP bonus
// - Friendly (51 to 100): +10% to +25% XP bonus (scales with reputation)
type FactionXPBonusSystem struct {
	world             *World
	rng               *rand.Rand
	factionSystem     *FactionSystem
	progressionSystem *ProgressionSystem
	logger            *logrus.Entry

	// pendingBonuses stores XP bonuses to apply next frame
	pendingBonuses []xpBonus
}

// xpBonus represents a pending XP bonus to award
type xpBonus struct {
	entityID  uint64
	bonusXP   int
	factionID string
	reason    string
}

// NewFactionXPBonusSystem creates a new faction XP bonus system.
func NewFactionXPBonusSystem(world *World, seed int64) *FactionXPBonusSystem {
	logger := world.GetLogger()
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system_name", "faction_xp_bonus")
	}

	return &FactionXPBonusSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		pendingBonuses: make([]xpBonus, 0, 16),
		logger:         logEntry,
	}
}

// SetFactionSystem sets the faction system reference for relationship lookups.
func (s *FactionXPBonusSystem) SetFactionSystem(fs *FactionSystem) {
	s.factionSystem = fs
}

// SetProgressionSystem sets the progression system reference for awarding XP.
func (s *FactionXPBonusSystem) SetProgressionSystem(ps *ProgressionSystem) {
	s.progressionSystem = ps
}

// Update processes pending XP bonuses.
func (s *FactionXPBonusSystem) Update(entities []*Entity, deltaTime float64) {
	if len(s.pendingBonuses) == 0 {
		return
	}

	for _, bonus := range s.pendingBonuses {
		entity, ok := s.world.GetEntity(bonus.entityID)
		if !ok || entity == nil {
			continue
		}

		if s.progressionSystem != nil {
			err := s.progressionSystem.AwardXP(entity, bonus.bonusXP)
			if err == nil && s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
				s.logger.WithFields(logrus.Fields{
					"entityID":  bonus.entityID,
					"bonusXP":   bonus.bonusXP,
					"factionID": bonus.factionID,
					"reason":    bonus.reason,
				}).Debug("awarded faction XP bonus")
			}
		}
	}

	// Clear processed bonuses
	s.pendingBonuses = s.pendingBonuses[:0]
}

// OnEnemyKilled should be called when an entity kills another entity.
// It calculates and queues faction XP bonuses based on faction relationships.
func (s *FactionXPBonusSystem) OnEnemyKilled(killer, victim *Entity, baseXP int) {
	if killer == nil || victim == nil || s.factionSystem == nil {
		return
	}

	// Only award bonuses to player entities
	if _, ok := killer.GetComponent("input"); !ok {
		return
	}

	// Get victim's faction
	victimFactionComp, ok := victim.GetComponent("faction")
	if !ok {
		return
	}

	victimFaction := victimFactionComp.(FactionComponent)
	if victimFaction.IsPlayerFaction {
		return // Victim is player - no faction bonus
	}

	// Check all factions for enemy relationships with the victim's faction
	for factionID, faction := range s.factionSystem.Factions {
		if !faction.IsEnemy(victimFaction.FactionID) {
			continue // This faction is not an enemy of the victim's faction
		}

		// Get player's reputation with this faction
		playerRep := s.factionSystem.GetPlayerReputation(factionID)

		// Calculate XP bonus based on player's standing with this faction
		bonusPercent := s.calculateBonusPercent(playerRep)
		if bonusPercent <= 0 {
			continue
		}

		bonusXP := int(float64(baseXP) * bonusPercent)
		if bonusXP > 0 {
			s.pendingBonuses = append(s.pendingBonuses, xpBonus{
				entityID:  killer.ID,
				bonusXP:   bonusXP,
				factionID: factionID,
				reason:    "Killed enemy of " + faction.Name,
			})

			if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
				s.logger.WithFields(logrus.Fields{
					"killerID":      killer.ID,
					"victimFaction": victimFaction.FactionID,
					"allyFaction":   factionID,
					"playerRep":     playerRep,
					"bonusPercent":  bonusPercent,
					"bonusXP":       bonusXP,
				}).Debug("queued faction XP bonus")
			}
		}
	}
}

// calculateBonusPercent returns the XP bonus percentage based on reputation.
// Returns 0.0 for hostile/suspicious, 0.05 for neutral, 0.10-0.25 for friendly.
func (s *FactionXPBonusSystem) calculateBonusPercent(reputation int) float64 {
	if reputation <= 0 {
		return 0.0 // Hostile or suspicious - no bonus
	}

	if reputation <= 50 {
		return 0.05 // Neutral - 5% bonus
	}

	// Friendly: scales from 10% at 51 rep to 25% at 100 rep
	// Formula: 0.10 + (rep - 50) / 50 * 0.15
	scaledBonus := 0.10 + float64(reputation-50)/50.0*0.15
	return scaledBonus
}

// GetBonusPercent returns the current XP bonus percentage for a faction.
// Useful for UI display.
func (s *FactionXPBonusSystem) GetBonusPercent(factionID string) float64 {
	if s.factionSystem == nil {
		return 0.0
	}
	rep := s.factionSystem.GetPlayerReputation(factionID)
	return s.calculateBonusPercent(rep)
}

// GetPendingBonusCount returns the number of pending XP bonuses.
// Useful for testing.
func (s *FactionXPBonusSystem) GetPendingBonusCount() int {
	return len(s.pendingBonuses)
}
