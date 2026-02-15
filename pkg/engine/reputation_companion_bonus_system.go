// Package engine provides the ReputationCompanionBonusSystem which bridges faction
// reputation with companion combat statistics. Players with high faction reputation
// passively boost their companion's attack, defense, and speed.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ReputationCompanionBonusSystem connects faction reputation with companion stats.
// When a player has positive standing with factions, their companions receive
// stat bonuses proportional to the owner's highest faction reputation.
//
// Integration Points:
// - Reads from: FactionSystem (player reputation with factions)
// - Reads from: CompanionComponent (owner ID, companion type)
// - Modifies: CompanionStatsComponent (Attack, Defense, Speed)
//
// Bonus Tiers (based on highest allied faction reputation):
// - Hostile (-100 to 0): No bonus
// - Neutral (1 to 50): +5% attack, +5% defense, +3% speed
// - Friendly (51 to 75): +10% attack, +10% defense, +7% speed
// - Honored (76 to 100): +18% attack, +15% defense, +12% speed
//
// Genre Modifiers:
// - Fantasy: Base multipliers (blessed companions fight harder)
// - Scifi: +10% bonus (enhanced AI subroutines from allied corps)
// - Horror: -25% bonus (companions are fearful regardless)
// - Cyberpunk: +15% bonus (street-modded companion upgrades)
// - PostApoc: -10% bonus (scarce resources for companion upkeep)
type ReputationCompanionBonusSystem struct {
	world         *World
	rng           *rand.Rand
	factionSystem *FactionSystem
	logger        *logrus.Entry
	genreID       string

	// updateInterval controls how frequently bonuses recalculate (seconds)
	updateInterval float64
	timeSinceCheck float64

	// activeBonuses tracks current bonus multipliers per companion entity
	activeBonuses map[uint64]*reputationCompanionBonus

	// genreMultipliers scales faction companion bonuses by genre
	genreMultipliers map[string]float64
}

// reputationCompanionBonus stores applied bonus values for reversal
type reputationCompanionBonus struct {
	attackMult  float64
	defenseMult float64
	speedMult   float64
	repTier     int // cached tier for change detection
}

// NewReputationCompanionBonusSystem creates a new reputation companion bonus system.
func NewReputationCompanionBonusSystem(world *World, seed int64) *ReputationCompanionBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "reputation_companion_bonus")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("ReputationCompanionBonusSystem created")
		}
	}

	return &ReputationCompanionBonusSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		updateInterval: 1.0,
		activeBonuses:  make(map[uint64]*reputationCompanionBonus),
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
func (s *ReputationCompanionBonusSystem) SetFactionSystem(fs *FactionSystem) {
	s.factionSystem = fs
}

// SetGenre sets the genre for genre-aware companion bonus scaling.
func (s *ReputationCompanionBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithField("genre", genreID).Debug("genre set for reputation companion bonus")
	}
}

// Update applies companion stat bonuses based on owner faction reputation.
func (s *ReputationCompanionBonusSystem) Update(entities []*Entity, deltaTime float64) {
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

// processEntity applies reputation-based stat bonuses to a single companion.
func (s *ReputationCompanionBonusSystem) processEntity(entity *Entity) {
	compComp, ok := entity.GetComponent("companion")
	if !ok {
		return
	}
	companion, ok := compComp.(*CompanionComponent)
	if !ok {
		return
	}

	statsComp, ok := entity.GetComponent("companionstats")
	if !ok {
		return
	}
	stats, ok := statsComp.(*CompanionStatsComponent)
	if !ok {
		return
	}

	// Get owner's best reputation
	bestRep := s.getBestOwnerReputation(companion.OwnerID)
	tier := s.reputationTier(bestRep)

	// Check if bonus changed
	existing, hasExisting := s.activeBonuses[entity.ID]
	if hasExisting && existing.repTier == tier {
		return // No change
	}

	// Reverse old bonus
	if hasExisting {
		s.reverseBonus(stats, existing)
	}

	// Calculate and apply new bonus
	bonus := s.calculateBonus(tier)
	if bonus.attackMult != 1.0 || bonus.defenseMult != 1.0 || bonus.speedMult != 1.0 {
		stats.Attack *= bonus.attackMult
		stats.Defense *= bonus.defenseMult
		stats.Speed *= bonus.speedMult
		s.activeBonuses[entity.ID] = bonus

		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"entity_id":    entity.ID,
				"owner_id":     companion.OwnerID,
				"tier":         tier,
				"attack_mult":  bonus.attackMult,
				"defense_mult": bonus.defenseMult,
				"speed_mult":   bonus.speedMult,
				"genre":        s.genreID,
			}).Debug("reputation companion bonus applied")
		}
	} else {
		s.activeBonuses[entity.ID] = bonus
	}
}

// reverseBonus undoes a previously applied bonus.
func (s *ReputationCompanionBonusSystem) reverseBonus(stats *CompanionStatsComponent, bonus *reputationCompanionBonus) {
	if bonus.attackMult != 0 {
		stats.Attack /= bonus.attackMult
	}
	if bonus.defenseMult != 0 {
		stats.Defense /= bonus.defenseMult
	}
	if bonus.speedMult != 0 {
		stats.Speed /= bonus.speedMult
	}
}

// getBestOwnerReputation finds the highest faction reputation for the owner entity.
func (s *ReputationCompanionBonusSystem) getBestOwnerReputation(ownerID uint64) int {
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
	return bestRep
}

// reputationTier converts a raw reputation value to a tier index.
func (s *ReputationCompanionBonusSystem) reputationTier(reputation int) int {
	if reputation <= 0 {
		return 0
	}
	if reputation <= 50 {
		return 1
	}
	if reputation <= 75 {
		return 2
	}
	return 3
}

// calculateBonus computes stat multipliers for a reputation tier with genre scaling.
func (s *ReputationCompanionBonusSystem) calculateBonus(tier int) *reputationCompanionBonus {
	var baseAttack, baseDefense, baseSpeed float64

	switch tier {
	case 0: // Hostile/No reputation
		return &reputationCompanionBonus{attackMult: 1.0, defenseMult: 1.0, speedMult: 1.0, repTier: 0}
	case 1: // Neutral (1-50)
		baseAttack = 0.05
		baseDefense = 0.05
		baseSpeed = 0.03
	case 2: // Friendly (51-75)
		baseAttack = 0.10
		baseDefense = 0.10
		baseSpeed = 0.07
	case 3: // Honored (76-100)
		baseAttack = 0.18
		baseDefense = 0.15
		baseSpeed = 0.12
	}

	genreMult := s.genreMultipliers[s.genreID]
	if genreMult == 0 {
		genreMult = 1.0
	}

	return &reputationCompanionBonus{
		attackMult:  1.0 + baseAttack*genreMult,
		defenseMult: 1.0 + baseDefense*genreMult,
		speedMult:   1.0 + baseSpeed*genreMult,
		repTier:     tier,
	}
}

// GetCompanionBonus returns the current bonus multipliers for a companion entity.
func (s *ReputationCompanionBonusSystem) GetCompanionBonus(entityID uint64) (attack, defense, speed float64) {
	bonus, ok := s.activeBonuses[entityID]
	if !ok {
		return 1.0, 1.0, 1.0
	}
	return bonus.attackMult, bonus.defenseMult, bonus.speedMult
}

// GetGenreMultiplier returns the genre-specific bonus multiplier.
func (s *ReputationCompanionBonusSystem) GetGenreMultiplier() float64 {
	mult := s.genreMultipliers[s.genreID]
	if mult == 0 {
		return 1.0
	}
	return mult
}

// HasActiveBonus returns whether a companion entity currently has a reputation bonus.
func (s *ReputationCompanionBonusSystem) HasActiveBonus(entityID uint64) bool {
	bonus, ok := s.activeBonuses[entityID]
	return ok && bonus.repTier > 0
}
