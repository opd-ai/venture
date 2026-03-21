// Package engine provides the TimeOfDayCriticalChanceSystem which bridges time-of-day
// lighting cycles with critical hit chance modifiers. Night activity rewards ambush-style
// attacks with bonus crit chance, while daytime provides clarity bonuses for certain genres.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayCriticalChanceSystem modifies critical hit chance based on time of day.
// This connects the TimeOfDayLightingSystem with StatsComponent.CritChance.
//
// Crit chance modifiers by time of day (additive bonuses):
//   - Dawn: +0.02 (fresh start, clear vision)
//   - Day: +0.00 (baseline)
//   - Dusk: +0.03 (shadows begin, ambush advantage)
//   - Night: +0.05 (darkness enables surprise attacks)
//
// Genre modifiers adjust these base values:
//   - Fantasy: Night +0.02 (rogues thrive in shadows)
//   - Scifi: Day +0.03, Night -0.02 (targeting systems need light)
//   - Horror: Night +0.05 (violence in darkness)
//   - Cyberpunk: Dusk/Night +0.03 (neon-lit assassinations)
//   - Postapoc: Night +0.03 (ambush predators)
type TimeOfDayCriticalChanceSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Reference to lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// updateInterval controls how often we check time (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Track original crit chance values per entity for restoration
	originalCritChance map[uint64]float64

	// Last known time of day for change detection
	lastTimeOfDay palette.TimeOfDay

	// Genre affects which times boost crit chance
	genreID string

	// Base crit chance bonuses by time of day
	baseModifiers map[palette.TimeOfDay]float64

	// Genre-specific modifiers (added to base)
	genreModifiers map[string]map[palette.TimeOfDay]float64
}

// NewTimeOfDayCriticalChanceSystem creates a new time-of-day critical chance system.
func NewTimeOfDayCriticalChanceSystem(world *World, seed int64) *TimeOfDayCriticalChanceSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_crit_chance")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayCriticalChanceSystem created")
	}

	return &TimeOfDayCriticalChanceSystem{
		world:              world,
		logger:             logEntry,
		rng:                rand.New(rand.NewSource(seed)),
		updateInterval:     1.0, // Check once per second (time changes slowly)
		originalCritChance: make(map[uint64]float64, 64),
		lastTimeOfDay:      palette.TimeOfDayDay,
		genreID:            "fantasy",
		baseModifiers: map[palette.TimeOfDay]float64{
			palette.TimeOfDayDawn:  0.02, // Early bird accuracy
			palette.TimeOfDayDay:   0.00, // Baseline
			palette.TimeOfDayDusk:  0.03, // Shadows aid ambush
			palette.TimeOfDayNight: 0.05, // Darkness enables surprise
		},
		genreModifiers: map[string]map[palette.TimeOfDay]float64{
			"fantasy": {
				palette.TimeOfDayNight: 0.02, // Rogues thrive in shadows
			},
			"scifi": {
				palette.TimeOfDayDay:   0.03,  // Targeting systems optimal
				palette.TimeOfDayNight: -0.02, // Poor visibility
			},
			"horror": {
				palette.TimeOfDayNight: 0.05, // Violence in darkness
				palette.TimeOfDayDusk:  0.02, // Fear builds at twilight
			},
			"cyberpunk": {
				palette.TimeOfDayDusk:  0.03, // City comes alive
				palette.TimeOfDayNight: 0.03, // Neon-lit assassinations
			},
			"postapoc": {
				palette.TimeOfDayDawn:  0.02, // Cooler temps, clear shots
				palette.TimeOfDayNight: 0.03, // Ambush predators
			},
		},
	}
}

// SetGenre sets the genre for genre-aware crit modifiers.
func (s *TimeOfDayCriticalChanceSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayCriticalChanceSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// Update checks time of day and updates crit chance modifiers for combatants.
func (s *TimeOfDayCriticalChanceSystem) Update(entities []*Entity, deltaTime float64) {
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

	// If time changed, log it and update all entities
	timeChanged := currentTime != s.lastTimeOfDay
	if timeChanged {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"time_of_day": currentTime.String(),
				"modifier":    s.getCritModifier(currentTime),
			}).Debug("time of day changed, updating crit chance modifiers")
		}
		s.lastTimeOfDay = currentTime
	}

	// Update crit chance for all entities with stats
	modifier := s.getCritModifier(currentTime)
	ApplyTimeOfDayStatModifier(entities, modifier, s.originalCritChance, CritChanceAccessor)
}

// getCritModifier calculates the total crit chance modifier for the current time.
func (s *TimeOfDayCriticalChanceSystem) getCritModifier(timeOfDay palette.TimeOfDay) float64 {
	// Start with base modifier
	modifier := s.baseModifiers[timeOfDay]

	// Apply genre-specific modifier
	if genreMods, ok := s.genreModifiers[s.genreID]; ok {
		if mod, ok := genreMods[timeOfDay]; ok {
			modifier += mod
		}
	}

	// Clamp total modifier to reasonable range (-0.10 to +0.15)
	if modifier < -0.10 {
		modifier = -0.10
	}
	if modifier > 0.15 {
		modifier = 0.15
	}

	return modifier
}

// GetCurrentModifier returns the current crit modifier based on time of day.
// This is useful for UI display of active bonuses.
func (s *TimeOfDayCriticalChanceSystem) GetCurrentModifier() float64 {
	if s.lightingSystem == nil {
		return 0.0
	}
	return s.getCritModifier(s.lightingSystem.GetCurrentTimeOfDay())
}

// GetOriginalCritChance returns the stored original crit chance for an entity.
func (s *TimeOfDayCriticalChanceSystem) GetOriginalCritChance(entityID uint64) (float64, bool) {
	c, ok := s.originalCritChance[entityID]
	return c, ok
}

// GetBonusDescription returns a human-readable description of the current bonus.
func (s *TimeOfDayCriticalChanceSystem) GetBonusDescription() string {
	if s.lightingSystem == nil {
		return ""
	}

	timeOfDay := s.lightingSystem.GetCurrentTimeOfDay()
	mod := s.getCritModifier(timeOfDay)

	if mod > 0 {
		bonus := int(mod * 100)
		return timeOfDay.String() + " Precision: +" + critItoa(bonus) + "% Crit"
	}
	if mod < 0 {
		penalty := int(-mod * 100)
		return timeOfDay.String() + " Obscured: -" + critItoa(penalty) + "% Crit"
	}
	return ""
}

// critItoa converts an int to string without importing strconv.
func critItoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + critItoa(-n)
	}
	digits := ""
	for n > 0 {
		digits = string('0'+byte(n%10)) + digits
		n /= 10
	}
	return digits
}
