// Package engine provides the TimeOfDayXPBonusSystem which bridges time-of-day
// cycles with experience point gains. Players earn bonus XP during specific times
// based on genre: night activity in horror/postapoc is rewarded, while sci-fi/cyberpunk
// favor daytime productivity. Dawn and dusk provide smaller bonuses to encourage
// strategic timing of combat and exploration.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayXPBonusSystem modifies XP gains based on time of day.
// This connects the TimeOfDayLightingSystem with the ProgressionSystem.
//
// XP multipliers by time of day (base values before genre modifiers):
//   - Dawn: 1.05 (early bird bonus)
//   - Day: 1.0 (baseline)
//   - Dusk: 1.05 (transitional activity bonus)
//   - Night: 1.15 (risk/reward for dangerous night activity)
//
// Genre modifiers adjust these base values:
//   - Fantasy: Night +5% (quests often involve night work)
//   - Scifi: Day +10%, Night -5% (efficient work schedules)
//   - Horror: Night +15%, Day -5% (darkness is the domain)
//   - Cyberpunk: Dusk/Night +10% (neon-lit night life)
//   - Postapoc: Night +10% (safer to move at night, avoid heat)
type TimeOfDayXPBonusSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Reference to lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// progressionSystem for XP callback integration
	progressionSystem *ProgressionSystem

	// updateInterval controls how often we check time (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Track active XP multiplier per entity
	activeMultipliers map[uint64]float64

	// Last known time of day for change detection
	lastTimeOfDay palette.TimeOfDay

	// Genre affects which times boost XP
	genreID string

	// Base XP multipliers by time of day
	baseMultipliers map[palette.TimeOfDay]float64

	// Genre-specific modifiers (added to base)
	genreModifiers map[string]map[palette.TimeOfDay]float64
}

// NewTimeOfDayXPBonusSystem creates a new time-of-day XP bonus system.
func NewTimeOfDayXPBonusSystem(world *World, seed int64) *TimeOfDayXPBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_xp_bonus")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayXPBonusSystem created")
	}

	return &TimeOfDayXPBonusSystem{
		world:             world,
		logger:            logEntry,
		rng:               rand.New(rand.NewSource(seed)),
		updateInterval:    1.0, // Check once per second (time changes slowly)
		activeMultipliers: make(map[uint64]float64, 32),
		lastTimeOfDay:     palette.TimeOfDayDay,
		genreID:           "fantasy",
		baseMultipliers: map[palette.TimeOfDay]float64{
			palette.TimeOfDayDawn:  1.05, // Early bird bonus
			palette.TimeOfDayDay:   1.0,  // Baseline
			palette.TimeOfDayDusk:  1.05, // Transitional bonus
			palette.TimeOfDayNight: 1.15, // Night risk/reward
		},
		genreModifiers: map[string]map[palette.TimeOfDay]float64{
			"fantasy": {
				palette.TimeOfDayNight: 0.05, // Quest givers active at night
			},
			"scifi": {
				palette.TimeOfDayDay:   0.10,  // Optimal work hours
				palette.TimeOfDayNight: -0.05, // Reduced efficiency
			},
			"horror": {
				palette.TimeOfDayDay:   -0.05, // Less horror in daylight
				palette.TimeOfDayNight: 0.15,  // Darkness rewards courage
			},
			"cyberpunk": {
				palette.TimeOfDayDusk:  0.10, // City comes alive
				palette.TimeOfDayNight: 0.10, // Neon-lit activity
			},
			"postapoc": {
				palette.TimeOfDayDawn:  0.05, // Cooler temperatures
				palette.TimeOfDayNight: 0.10, // Avoid daytime heat/patrols
			},
		},
	}
}

// SetGenre sets the genre for genre-aware XP modifiers.
func (s *TimeOfDayXPBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayXPBonusSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// SetProgressionSystem sets the progression system for XP callback integration.
func (s *TimeOfDayXPBonusSystem) SetProgressionSystem(ps *ProgressionSystem) {
	s.progressionSystem = ps
}

// Update checks time of day and updates active XP multipliers for players.
func (s *TimeOfDayXPBonusSystem) Update(entities []*Entity, deltaTime float64) {
	if s.lightingSystem == nil {
		return
	}

	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	// Get current time of day from lighting system
	currentTime := s.lightingSystem.GetCurrentTimeOfDay()

	// If time changed, log it
	if currentTime != s.lastTimeOfDay {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"time_of_day": currentTime.String(),
				"multiplier":  s.getXPMultiplier(currentTime),
			}).Debug("time of day changed, updating XP multipliers")
		}
		s.lastTimeOfDay = currentTime
	}

	// Update multipliers for all player entities
	multiplier := s.getXPMultiplier(currentTime)
	for _, entity := range entities {
		// Only apply to player entities (have input component)
		if _, ok := entity.GetComponent("input"); !ok {
			continue
		}

		// Store the active multiplier for this entity
		s.activeMultipliers[entity.ID] = multiplier
	}
}

// getXPMultiplier calculates the total XP multiplier for the current time of day.
func (s *TimeOfDayXPBonusSystem) getXPMultiplier(timeOfDay palette.TimeOfDay) float64 {
	// Start with base multiplier
	multiplier := s.baseMultipliers[timeOfDay]
	if multiplier == 0 {
		multiplier = 1.0
	}

	// Apply genre-specific modifier
	if genreMods, ok := s.genreModifiers[s.genreID]; ok {
		if mod, ok := genreMods[timeOfDay]; ok {
			multiplier += mod
		}
	}

	// Clamp to reasonable range
	if multiplier < 0.5 {
		multiplier = 0.5
	}
	if multiplier > 2.0 {
		multiplier = 2.0
	}

	return multiplier
}

// GetActiveMultiplier returns the current XP multiplier for an entity.
// Returns 1.0 if no multiplier is set.
func (s *TimeOfDayXPBonusSystem) GetActiveMultiplier(entityID uint64) float64 {
	if mult, ok := s.activeMultipliers[entityID]; ok {
		return mult
	}
	return 1.0
}

// GetCurrentMultiplier returns the current XP multiplier based on time of day.
// This is useful for UI display of active bonuses.
func (s *TimeOfDayXPBonusSystem) GetCurrentMultiplier() float64 {
	if s.lightingSystem == nil {
		return 1.0
	}
	return s.getXPMultiplier(s.lightingSystem.GetCurrentTimeOfDay())
}

// GetBonusDescription returns a human-readable description of the current bonus.
func (s *TimeOfDayXPBonusSystem) GetBonusDescription() string {
	if s.lightingSystem == nil {
		return ""
	}

	timeOfDay := s.lightingSystem.GetCurrentTimeOfDay()
	mult := s.getXPMultiplier(timeOfDay)

	if mult > 1.0 {
		bonus := int((mult - 1.0) * 100)
		return timeOfDay.String() + " Activity: +" + timeOfDayItoa(bonus) + "% XP"
	}
	if mult < 1.0 {
		penalty := int((1.0 - mult) * 100)
		return timeOfDay.String() + " Penalty: -" + timeOfDayItoa(penalty) + "% XP"
	}
	return ""
}

// timeOfDayItoa converts an int to string without importing strconv.
func timeOfDayItoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + timeOfDayItoa(-n)
	}
	digits := ""
	for n > 0 {
		digits = string('0'+byte(n%10)) + digits
		n /= 10
	}
	return digits
}
