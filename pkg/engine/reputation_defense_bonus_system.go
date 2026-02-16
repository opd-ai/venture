// Package engine provides the ReputationDefenseBonusSystem which bridges faction
// reputation with defensive damage reduction in combat.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ReputationDefenseBonusSystem connects faction reputation with combat defense.
// When the player is attacked by an enemy of a faction they have good standing with,
// they receive a defense bonus (damage reduction) based on their reputation level.
// This is the defensive counterpart to FactionDamageBonusSystem.
//
// Integration Points:
// - Reads from: FactionComponent (faction membership for attacker identification)
// - Reads from: FactionSystem (player reputation with factions, faction relationships)
// - Reads from: StatsComponent (base Defense value)
// - Modifies: Effective defense via GetDefenseBonus() called by combat resolver
//
// Defense Bonus Tiers (based on reputation with factions who consider attacker enemy):
// - Hostile (-100 to -50): No bonus
// - Suspicious (-49 to 0): No bonus
// - Neutral (1 to 50): +2% defense bonus
// - Friendly (51 to 100): +4% to +12% defense bonus (scales with reputation)
//
// Genre Modifiers:
// - Fantasy: Base multipliers (honor shields the faithful)
// - Scifi: +10% faction bonuses (corporate armor tech access)
// - Horror: -25% faction bonuses (no one can truly protect you)
// - Cyberpunk: +15% faction bonuses (gang armor upgrades)
// - PostApoc: -10% faction bonuses (scavenged protection is unreliable)
type ReputationDefenseBonusSystem struct {
	world         *World
	rng           *rand.Rand
	factionSystem *FactionSystem
	logger        *logrus.Entry
	genreID       string

	// defenseMultipliers caches defense bonus multipliers per entity per frame.
	// Maps defenderID -> attacker factionID -> multiplier
	defenseMultipliers map[uint64]map[string]float64

	// genreMultipliers scales faction defense bonuses by genre
	genreMultipliers map[string]float64
}

// NewReputationDefenseBonusSystem creates a new reputation defense bonus system.
func NewReputationDefenseBonusSystem(world *World, seed int64) *ReputationDefenseBonusSystem {
	logger := world.GetLogger()
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system_name", "reputation_defense_bonus")
	}

	return &ReputationDefenseBonusSystem{
		world:              world,
		rng:                rand.New(rand.NewSource(seed)),
		defenseMultipliers: make(map[uint64]map[string]float64),
		logger:             logEntry,
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,
			"scifi":     1.1,
			"horror":    0.75,
			"cyberpunk": 1.15,
			"postapoc":  0.9,
		},
	}
}

// SetFactionSystem sets the faction system reference for relationship lookups.
func (s *ReputationDefenseBonusSystem) SetFactionSystem(fs *FactionSystem) {
	s.factionSystem = fs
}

// SetGenre sets the genre for genre-aware defense bonus scaling.
func (s *ReputationDefenseBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for defense bonuses")
	}
}

// Update recalculates defense multipliers for all player entities.
func (s *ReputationDefenseBonusSystem) Update(entities []*Entity, deltaTime float64) {
	// Clear cache each frame
	for k := range s.defenseMultipliers {
		delete(s.defenseMultipliers, k)
	}

	if s.factionSystem == nil {
		return
	}

	for _, entity := range entities {
		if _, ok := entity.GetComponent("input"); !ok {
			continue
		}

		s.defenseMultipliers[entity.ID] = s.calculateFactionDefenseMultipliers()
	}
}

// calculateFactionDefenseMultipliers computes defense bonuses against all enemy factions.
func (s *ReputationDefenseBonusSystem) calculateFactionDefenseMultipliers() map[string]float64 {
	multipliers := make(map[string]float64)

	if s.factionSystem == nil {
		return multipliers
	}

	for factionID, faction := range s.factionSystem.Factions {
		playerRep := s.factionSystem.GetPlayerReputation(factionID)
		bonusPercent := s.calculateBonusPercent(playerRep)

		if bonusPercent <= 0 {
			continue
		}

		genreMult := s.genreMultipliers[s.genreID]
		if genreMult == 0 {
			genreMult = 1.0
		}
		finalBonus := bonusPercent * genreMult

		// Store bonus for enemies of this faction
		for otherFactionID := range s.factionSystem.Factions {
			if faction.IsEnemy(otherFactionID) {
				if existing, ok := multipliers[otherFactionID]; ok {
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

// calculateBonusPercent returns the defense bonus percentage based on reputation.
// Returns 0.0 for hostile/suspicious, 0.02 for neutral, 0.04-0.12 for friendly.
func (s *ReputationDefenseBonusSystem) calculateBonusPercent(reputation int) float64 {
	if reputation <= 0 {
		return 0.0
	}

	if reputation <= 50 {
		return 0.02 // Neutral - 2% bonus
	}

	// Friendly: scales from 4% at 51 rep to 12% at 100 rep
	scaledBonus := 0.04 + float64(reputation-50)/50.0*0.08
	return scaledBonus
}

// GetDefenseMultiplier returns the defense multiplier for a defender against an attacker.
// Returns 1.0 (no bonus) if no faction relationship applies.
// The multiplier is additive: 1.0 + bonus (e.g., 1.08 for 8% bonus defense).
func (s *ReputationDefenseBonusSystem) GetDefenseMultiplier(defenderID uint64, attacker *Entity) float64 {
	if attacker == nil {
		return 1.0
	}

	factionComp, ok := attacker.GetComponent("faction")
	if !ok {
		return 1.0
	}

	fc := factionComp.(*FactionComponent)
	if fc.IsPlayerFaction {
		return 1.0
	}

	defenderMults, ok := s.defenseMultipliers[defenderID]
	if !ok {
		return 1.0
	}

	bonus, ok := defenderMults[fc.FactionID]
	if !ok {
		return 1.0
	}

	return 1.0 + bonus
}

// OnDefend should be called when a player is attacked.
// Returns the modified damage after applying reputation-based defense bonus.
func (s *ReputationDefenseBonusSystem) OnDefend(defender, attacker *Entity, incomingDamage float64) float64 {
	if defender == nil || attacker == nil {
		return incomingDamage
	}

	multiplier := s.GetDefenseMultiplier(defender.ID, attacker)
	if multiplier == 1.0 {
		return incomingDamage
	}

	// Defense multiplier reduces damage: damage / multiplier
	reducedDamage := incomingDamage / multiplier

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		factionComp, _ := attacker.GetComponent("faction")
		fc := factionComp.(*FactionComponent)
		s.logger.WithFields(logrus.Fields{
			"defenderID":      defender.ID,
			"attackerID":      attacker.ID,
			"attackerFaction": fc.FactionID,
			"multiplier":      multiplier,
			"incomingDamage":  incomingDamage,
			"reducedDamage":   reducedDamage,
			"genre":           s.genreID,
		}).Debug("reputation defense bonus applied")
	}

	return reducedDamage
}

// GetBonusPercent returns the current defense bonus percentage against a faction.
func (s *ReputationDefenseBonusSystem) GetBonusPercent(factionID string) float64 {
	if s.factionSystem == nil {
		return 0.0
	}
	rep := s.factionSystem.GetPlayerReputation(factionID)
	baseBonus := s.calculateBonusPercent(rep)

	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	return baseBonus * genreMult
}

// GetGenreMultiplier returns the genre-specific faction defense multiplier.
func (s *ReputationDefenseBonusSystem) GetGenreMultiplier() float64 {
	mult := s.genreMultipliers[s.genreID]
	if mult == 0 {
		return 1.0
	}
	return mult
}
