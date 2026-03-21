// Package engine provides the TimeOfDayEvasionSystem which bridges time-of-day
// lighting cycles with evasion modifiers. Night enhances dodge ability through
// shadows and concealment, while daytime impairs it (harder to evade when visible).
// Genre-specific modifiers further tune these bonuses for thematic gameplay
// (fantasy rogues excel at night, sci-fi sensors track evasion, etc).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayEvasionSystem modifies evasion chance based on time of day.
// This connects the TimeOfDayLightingSystem with StatsComponent.Evasion.
//
// Evasion modifiers by time of day (additive bonuses):
//   - Dawn: +0.01 (transitional light, some cover)
//   - Day: -0.02 (bright visibility, hard to evade)
//   - Dusk: +0.03 (lengthening shadows aid dodging)
//   - Night: +0.05 (darkness provides cover for evasion)
//
// Genre modifiers adjust these base values:
//   - Fantasy: Night +0.03 (rogues thrive in shadows)
//   - Scifi: Night -0.03 (sensors track movement regardless)
//   - Horror: Night +0.04 (survival instincts peak)
//   - Cyberpunk: Dusk/Night +0.02 (urban shadows)
//   - Postapoc: Night +0.03 (predator instincts)
type TimeOfDayEvasionSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Reference to lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// updateInterval controls how often we check time (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Track original evasion values per entity for restoration
	originalEvasion map[uint64]float64

	// Last known time of day for change detection
	lastTimeOfDay palette.TimeOfDay

	// Genre affects which times boost evasion
	genreID string

	// Base evasion bonuses by time of day
	baseModifiers map[palette.TimeOfDay]float64

	// Genre-specific modifiers (added to base)
	genreModifiers map[string]map[palette.TimeOfDay]float64
}

// NewTimeOfDayEvasionSystem creates a new time-of-day evasion system.
func NewTimeOfDayEvasionSystem(world *World, seed int64) *TimeOfDayEvasionSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_evasion")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayEvasionSystem created")
	}

	return &TimeOfDayEvasionSystem{
		world:           world,
		logger:          logEntry,
		rng:             rand.New(rand.NewSource(seed)),
		updateInterval:  1.0, // Check once per second (time changes slowly)
		originalEvasion: make(map[uint64]float64, 64),
		lastTimeOfDay:   palette.TimeOfDayDay,
		genreID:         "fantasy",
		baseModifiers: map[palette.TimeOfDay]float64{
			palette.TimeOfDayDawn:  0.01,  // Transitional light
			palette.TimeOfDayDay:   -0.02, // Bright visibility impairs evasion
			palette.TimeOfDayDusk:  0.03,  // Lengthening shadows
			palette.TimeOfDayNight: 0.05,  // Darkness aids evasion
		},
		genreModifiers: map[string]map[palette.TimeOfDay]float64{
			"fantasy": {
				palette.TimeOfDayNight: 0.03, // Rogues thrive in shadows
				palette.TimeOfDayDusk:  0.01, // Twilight ambushes
			},
			"scifi": {
				palette.TimeOfDayNight: -0.03, // Sensors track movement
				palette.TimeOfDayDay:   0.01,  // Baseline tech advantage
			},
			"horror": {
				palette.TimeOfDayNight: 0.04, // Survival instincts peak
				palette.TimeOfDayDusk:  0.02, // Fear heightens senses
			},
			"cyberpunk": {
				palette.TimeOfDayNight: 0.02, // Urban shadows
				palette.TimeOfDayDusk:  0.02, // Neon distractions
			},
			"postapoc": {
				palette.TimeOfDayNight: 0.03, // Predator instincts
				palette.TimeOfDayDawn:  0.01, // Early survivor reflexes
			},
		},
	}
}

// SetGenre sets the genre for genre-aware evasion modifiers.
func (s *TimeOfDayEvasionSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayEvasionSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// Update checks time of day and updates evasion modifiers for combatants.
func (s *TimeOfDayEvasionSystem) Update(entities []*Entity, deltaTime float64) {
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
				"modifier":    s.getEvasionModifier(currentTime),
			}).Debug("time of day changed, updating evasion modifiers")
		}
		s.lastTimeOfDay = currentTime
	}

	// Update evasion for all entities with stats
	modifier := s.getEvasionModifier(currentTime)
	ApplyTimeOfDayStatModifier(entities, modifier, s.originalEvasion, EvasionAccessor)
}

// getEvasionModifier calculates the total evasion modifier for the current time.
func (s *TimeOfDayEvasionSystem) getEvasionModifier(timeOfDay palette.TimeOfDay) float64 {
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

// GetCurrentModifier returns the current evasion modifier based on time of day.
// This is useful for UI display of active bonuses.
func (s *TimeOfDayEvasionSystem) GetCurrentModifier() float64 {
	if s.lightingSystem == nil {
		return 0.0
	}
	return s.getEvasionModifier(s.lightingSystem.GetCurrentTimeOfDay())
}

// GetOriginalEvasion returns the stored original evasion for an entity.
func (s *TimeOfDayEvasionSystem) GetOriginalEvasion(entityID uint64) (float64, bool) {
	e, ok := s.originalEvasion[entityID]
	return e, ok
}

// GetBonusDescription returns a human-readable description of the current bonus.
func (s *TimeOfDayEvasionSystem) GetBonusDescription() string {
	if s.lightingSystem == nil {
		return ""
	}

	timeOfDay := s.lightingSystem.GetCurrentTimeOfDay()
	mod := s.getEvasionModifier(timeOfDay)

	if mod > 0 {
		bonus := int(mod * 100)
		return timeOfDay.String() + " Agility: +" + evasionItoa(bonus) + "% Evasion"
	}
	if mod < 0 {
		penalty := int(-mod * 100)
		return timeOfDay.String() + " Exposed: -" + evasionItoa(penalty) + "% Evasion"
	}
	return ""
}

// evasionItoa converts an int to string without importing strconv.
func evasionItoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + evasionItoa(-n)
	}
	digits := ""
	for n > 0 {
		digits = string('0'+byte(n%10)) + digits
		n /= 10
	}
	return digits
}
