// Package engine provides the TerrainAmbushCritSystem which bridges
// terrain concealment with critical hit bonuses for ambush-style combat.
// Entities hiding in good terrain cover receive bonus critical chance.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// TerrainAmbushCritSystem modifies entity critical hit chance based on
// terrain concealment. This connects TerrainStealthSystem with combat
// to reward tactical positioning with ambush damage bonuses.
//
// Crit bonus calculation:
//   - Stealth multiplier < 0.7: +10-15% crit chance (heavy cover)
//   - Stealth multiplier < 0.85: +5-8% crit chance (moderate cover)
//   - Stealth multiplier < 0.95: +2-4% crit chance (light cover)
//   - Stealth multiplier >= 0.95: no bonus (exposed)
//
// Genre-specific scaling:
//   - Fantasy: +10% bonus (ambush tactics in magical forests)
//   - Scifi: -10% bonus (thermal sensors reduce ambush effectiveness)
//   - Horror: +20% bonus (fear amplifies surprise attacks)
//   - Cyberpunk: +15% bonus (hacking + ambush synergy)
//   - Postapoc: +5% bonus (scavenger tactics)
type TerrainAmbushCritSystem struct {
	world                *World
	terrainStealthSystem *TerrainStealthSystem
	rng                  *rand.Rand
	logger               *logrus.Entry
	genreID              string

	// Update timing
	updateInterval float64
	timeSinceCheck float64

	// Crit bonus tracking
	critBonuses map[uint64]float64 // entityID -> current crit bonus applied

	// Genre multipliers for ambush effectiveness
	genreMultipliers map[string]float64
}

// NewTerrainAmbushCritSystem creates a new terrain ambush crit system.
func NewTerrainAmbushCritSystem(world *World, seed int64) *TerrainAmbushCritSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "terrain_ambush_crit")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("terrain ambush crit system created")
		}
	}

	return &TerrainAmbushCritSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		updateInterval: 0.25, // 4 times per second
		critBonuses:    make(map[uint64]float64, 64),
		genreMultipliers: map[string]float64{
			"fantasy":   1.10, // Magical forests favor ambush
			"scifi":     0.90, // Thermal sensors reduce effectiveness
			"horror":    1.20, // Fear amplifies surprise attacks
			"cyberpunk": 1.15, // Hacking + ambush synergy
			"postapoc":  1.05, // Scavenger tactics
		},
	}
}

// SetTerrainStealthSystem sets the stealth system used for concealment lookups.
func (s *TerrainAmbushCritSystem) SetTerrainStealthSystem(tss *TerrainStealthSystem) {
	s.terrainStealthSystem = tss
	// Clear bonuses when stealth system changes
	s.critBonuses = make(map[uint64]float64, 64)
	if s.logger != nil {
		s.logger.Debug("terrain stealth system linked")
	}
}

// SetGenre sets the genre for genre-specific ambush effectiveness.
func (s *TerrainAmbushCritSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes entities and applies ambush crit bonuses based on terrain concealment.
func (s *TerrainAmbushCritSystem) Update(entities []*Entity, deltaTime float64) {
	if s.terrainStealthSystem == nil {
		return
	}

	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	for _, entity := range entities {
		s.updateEntityAmbushCrit(entity)
	}
}

// updateEntityAmbushCrit calculates and applies ambush crit bonus for an entity.
func (s *TerrainAmbushCritSystem) updateEntityAmbushCrit(entity *Entity) {
	// Only apply to entities with stats
	stats := entity.GetStats()
	if stats == nil {
		s.removeBonus(entity.ID, nil)
		return
	}

	// Get terrain stealth multiplier
	stealthMult := s.terrainStealthSystem.GetStealthMultiplier(entity.ID)

	// Calculate new crit bonus based on concealment
	newBonus := s.calculateCritBonus(stealthMult)

	// Apply genre scaling
	genreMult := s.getGenreMultiplier()
	newBonus *= genreMult

	// Get current bonus to calculate delta
	currentBonus := s.critBonuses[entity.ID]
	if newBonus == currentBonus {
		return // No change
	}

	// Remove old bonus and apply new
	stats.CritChance -= currentBonus
	stats.CritChance += newBonus

	// Clamp crit chance
	if stats.CritChance < 0 {
		stats.CritChance = 0
	} else if stats.CritChance > 1.0 {
		stats.CritChance = 1.0
	}

	// Track new bonus
	if newBonus > 0 {
		s.critBonuses[entity.ID] = newBonus
	} else {
		delete(s.critBonuses, entity.ID)
	}

	s.logCritBonusChange(entity, currentBonus, newBonus, stealthMult)
}

// calculateCritBonus determines crit bonus based on stealth multiplier.
// Lower stealth multiplier = better concealment = higher crit bonus.
func (s *TerrainAmbushCritSystem) calculateCritBonus(stealthMult float64) float64 {
	switch {
	case stealthMult < 0.5:
		// Exceptional concealment (secret door, heavy cover)
		return 0.18 // +18% crit chance
	case stealthMult < 0.7:
		// Heavy cover (trees, structures)
		return 0.12 // +12% crit chance
	case stealthMult < 0.85:
		// Moderate cover
		return 0.07 // +7% crit chance
	case stealthMult < 0.95:
		// Light cover
		return 0.03 // +3% crit chance
	default:
		// Exposed position
		return 0
	}
}

// getGenreMultiplier returns the genre-specific ambush effectiveness multiplier.
func (s *TerrainAmbushCritSystem) getGenreMultiplier() float64 {
	if mult, ok := s.genreMultipliers[s.genreID]; ok {
		return mult
	}
	return 1.0
}

// removeBonus removes any applied crit bonus from an entity.
func (s *TerrainAmbushCritSystem) removeBonus(entityID uint64, stats *StatsComponent) {
	bonus, exists := s.critBonuses[entityID]
	if !exists {
		return
	}

	if stats != nil {
		stats.CritChance -= bonus
		if stats.CritChance < 0 {
			stats.CritChance = 0
		}
	}
	delete(s.critBonuses, entityID)
}

// GetCritBonus returns the current terrain ambush crit bonus for an entity.
func (s *TerrainAmbushCritSystem) GetCritBonus(entityID uint64) float64 {
	return s.critBonuses[entityID]
}

// SetUpdateInterval configures how often ambush crits are recalculated.
func (s *TerrainAmbushCritSystem) SetUpdateInterval(seconds float64) {
	if seconds > 0 {
		s.updateInterval = seconds
	}
}

// logCritBonusChange logs when an entity's ambush crit bonus changes.
func (s *TerrainAmbushCritSystem) logCritBonusChange(entity *Entity, oldBonus, newBonus, stealthMult float64) {
	if s.logger == nil || s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}

	s.logger.WithFields(logrus.Fields{
		"entity_id":     entity.ID,
		"old_bonus":     oldBonus,
		"new_bonus":     newBonus,
		"stealth_mult":  stealthMult,
		"genre":         s.genreID,
		"genre_scaling": s.getGenreMultiplier(),
	}).Debug("terrain ambush crit bonus updated")
}
