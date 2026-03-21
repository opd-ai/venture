// Package engine provides the TimeOfDayBlockChanceSystem which bridges time-of-day
// lighting cycles with block chance modifiers. Daytime visibility aids defensive
// reactions, while night reduces blocking effectiveness. Genre-specific modifiers
// further tune these bonuses for thematic gameplay (fantasy shields work better
// in daylight, cyberpunk implants maintain night effectiveness, etc).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayBlockChanceSystem modifies block chance based on time of day.
// This connects the TimeOfDayLightingSystem with StatsComponent.BlockChance.
//
// Block chance modifiers by time of day (additive bonuses):
//   - Dawn: +0.03 (alertness after rest, good visibility)
//   - Day: +0.05 (optimal visibility for defensive reactions)
//   - Dusk: +0.01 (fading light reduces reaction time)
//   - Night: -0.03 (darkness impairs defensive timing)
//
// Genre modifiers adjust these base values:
//   - Fantasy: Day +0.03 (shield techniques require light)
//   - Scifi: Night +0.04 (sensor-assisted blocking)
//   - Horror: Night -0.03 (panic reduces effectiveness)
//   - Cyberpunk: Night +0.02 (cyber-enhanced reflexes)
//   - Postapoc: Dawn +0.02 (survival instincts peak)
type TimeOfDayBlockChanceSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Reference to lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// updateInterval controls how often we check time (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Track original block chance values per entity for restoration
	originalBlockChance map[uint64]float64

	// Last known time of day for change detection
	lastTimeOfDay palette.TimeOfDay

	// Genre affects which times boost block chance
	genreID string

	// Base block chance bonuses by time of day
	baseModifiers map[palette.TimeOfDay]float64

	// Genre-specific modifiers (added to base)
	genreModifiers map[string]map[palette.TimeOfDay]float64
}

// NewTimeOfDayBlockChanceSystem creates a new time-of-day block chance system.
func NewTimeOfDayBlockChanceSystem(world *World, seed int64) *TimeOfDayBlockChanceSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_block_chance")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayBlockChanceSystem created")
	}

	return &TimeOfDayBlockChanceSystem{
		world:               world,
		logger:              logEntry,
		rng:                 rand.New(rand.NewSource(seed)),
		updateInterval:      1.0, // Check once per second (time changes slowly)
		originalBlockChance: make(map[uint64]float64, 64),
		lastTimeOfDay:       palette.TimeOfDayDay,
		genreID:             "fantasy",
		baseModifiers: map[palette.TimeOfDay]float64{
			palette.TimeOfDayDawn:  0.03,  // Morning alertness
			palette.TimeOfDayDay:   0.05,  // Optimal visibility
			palette.TimeOfDayDusk:  0.01,  // Fading light
			palette.TimeOfDayNight: -0.03, // Darkness impairs
		},
		genreModifiers: map[string]map[palette.TimeOfDay]float64{
			"fantasy": {
				palette.TimeOfDayDay: 0.03, // Shield techniques need light
			},
			"scifi": {
				palette.TimeOfDayNight: 0.04, // Sensor-assisted blocking
				palette.TimeOfDayDay:   0.01, // Baseline tech advantage
			},
			"horror": {
				palette.TimeOfDayNight: -0.03, // Panic reduces effectiveness
				palette.TimeOfDayDusk:  -0.02, // Dread at twilight
			},
			"cyberpunk": {
				palette.TimeOfDayNight: 0.02, // Cyber-enhanced reflexes
				palette.TimeOfDayDusk:  0.01, // Neon-lit streets
			},
			"postapoc": {
				palette.TimeOfDayDawn:  0.02,  // Survival instincts peak
				palette.TimeOfDayDay:   0.01,  // Scavenged gear works
				palette.TimeOfDayNight: -0.02, // Predators lurk
			},
		},
	}
}

// SetGenre sets the genre for genre-aware block modifiers.
func (s *TimeOfDayBlockChanceSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayBlockChanceSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// Update checks time of day and updates block chance modifiers for combatants.
func (s *TimeOfDayBlockChanceSystem) Update(entities []*Entity, deltaTime float64) {
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
				"modifier":    s.getBlockModifier(currentTime),
			}).Debug("time of day changed, updating block chance modifiers")
		}
		s.lastTimeOfDay = currentTime
	}

	// Update block chance for all entities with stats
	modifier := s.getBlockModifier(currentTime)
	ApplyTimeOfDayStatModifier(entities, modifier, s.originalBlockChance, BlockChanceAccessor)
}

// getBlockModifier calculates the total block chance modifier for the current time.
func (s *TimeOfDayBlockChanceSystem) getBlockModifier(timeOfDay palette.TimeOfDay) float64 {
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

// GetCurrentModifier returns the current block modifier based on time of day.
// This is useful for UI display of active bonuses.
func (s *TimeOfDayBlockChanceSystem) GetCurrentModifier() float64 {
	if s.lightingSystem == nil {
		return 0.0
	}
	return s.getBlockModifier(s.lightingSystem.GetCurrentTimeOfDay())
}

// GetOriginalBlockChance returns the stored original block chance for an entity.
func (s *TimeOfDayBlockChanceSystem) GetOriginalBlockChance(entityID uint64) (float64, bool) {
	c, ok := s.originalBlockChance[entityID]
	return c, ok
}

// GetBonusDescription returns a human-readable description of the current bonus.
func (s *TimeOfDayBlockChanceSystem) GetBonusDescription() string {
	if s.lightingSystem == nil {
		return ""
	}

	timeOfDay := s.lightingSystem.GetCurrentTimeOfDay()
	mod := s.getBlockModifier(timeOfDay)

	if mod > 0 {
		bonus := int(mod * 100)
		return timeOfDay.String() + " Defense: +" + blockItoa(bonus) + "% Block"
	}
	if mod < 0 {
		penalty := int(-mod * 100)
		return timeOfDay.String() + " Impaired: -" + blockItoa(penalty) + "% Block"
	}
	return ""
}

// blockItoa converts an int to string without importing strconv.
func blockItoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + blockItoa(-n)
	}
	digits := ""
	for n > 0 {
		digits = string('0'+byte(n%10)) + digits
		n /= 10
	}
	return digits
}
