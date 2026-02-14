// Package engine provides the TimeOfDaySpellDamageSystem which bridges time-of-day
// cycles with spell damage based on elemental affinity. Fire and Light spells
// deal bonus damage during day, Dark and Shadow spells deal bonus damage at night,
// and Nature (Earth, Wind) spells are strongest at dawn/dusk transitions.
//
// This creates tactical depth: players casting fire spells benefit from daytime
// activity, while necromancers and shadow mages prefer night operations.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDaySpellDamageSystem modifies spell damage based on time of day and
// spell element. This connects TimeOfDayLightingSystem with SpellCastingSystem.
//
// Element-Time damage multipliers:
//   - Fire, Light: Day 1.15, Dawn 1.05, Dusk 0.95, Night 0.85
//   - Dark, Arcane: Night 1.15, Dusk 1.05, Dawn 0.95, Day 0.85
//   - Earth, Wind: Dawn 1.10, Dusk 1.10, Day 1.0, Night 1.0
//   - Ice: Night 1.10, Day 0.95, Dawn 1.0, Dusk 1.0
//   - Lightning: Day 1.05, Night 1.05, Dawn 1.0, Dusk 1.0
type TimeOfDaySpellDamageSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Reference to lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// Update interval to avoid per-frame recalculation
	updateInterval float64
	timeSinceCheck float64

	// Cache multipliers to avoid redundant calculations
	modCache map[timeElementKey]float64

	// Last known time of day for cache invalidation
	lastTimeOfDay palette.TimeOfDay

	// Genre affects element-time relationships
	genreID string

	// Base multipliers: element -> timeOfDay -> multiplier
	baseMultipliers map[magic.ElementType]map[palette.TimeOfDay]float64

	// Genre modifiers (additive adjustment to base multipliers)
	genreModifiers map[string]map[magic.ElementType]float64
}

// timeElementKey combines time of day and element for cache lookup.
type timeElementKey struct {
	timeOfDay palette.TimeOfDay
	element   magic.ElementType
}

// NewTimeOfDaySpellDamageSystem creates a new time-of-day spell damage system.
func NewTimeOfDaySpellDamageSystem(world *World, seed int64) *TimeOfDaySpellDamageSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_spell_damage")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDaySpellDamageSystem created")
	}

	return &TimeOfDaySpellDamageSystem{
		world:           world,
		logger:          logEntry,
		rng:             rand.New(rand.NewSource(seed)),
		updateInterval:  0.5, // Check twice per second
		modCache:        make(map[timeElementKey]float64, 32),
		lastTimeOfDay:   palette.TimeOfDayDay,
		genreID:         "fantasy",
		baseMultipliers: initSpellDamageMultipliers(),
		genreModifiers:  initSpellDamageGenreModifiers(),
	}
}

// initSpellDamageMultipliers sets up the element-time damage relationships.
func initSpellDamageMultipliers() map[magic.ElementType]map[palette.TimeOfDay]float64 {
	return map[magic.ElementType]map[palette.TimeOfDay]float64{
		// Fire/Light: Empowered by sunlight
		magic.ElementFire: {
			palette.TimeOfDayDay:   1.15, // Strong sunlight empowers fire
			palette.TimeOfDayDawn:  1.05, // Rising sun
			palette.TimeOfDayDusk:  0.95, // Fading light
			palette.TimeOfDayNight: 0.85, // Darkness suppresses fire magic
		},
		magic.ElementLight: {
			palette.TimeOfDayDay:   1.20, // Holy light at peak
			palette.TimeOfDayDawn:  1.10, // New day's blessing
			palette.TimeOfDayDusk:  0.90, // Light fading
			palette.TimeOfDayNight: 0.80, // Light struggles in darkness
		},
		// Dark/Arcane: Empowered by darkness
		magic.ElementDark: {
			palette.TimeOfDayNight: 1.20, // Shadows at full strength
			palette.TimeOfDayDusk:  1.10, // Growing darkness
			palette.TimeOfDayDawn:  0.90, // Light returns
			palette.TimeOfDayDay:   0.80, // Sunlight weakens shadows
		},
		magic.ElementArcane: {
			palette.TimeOfDayNight: 1.10, // Stars empower arcane
			palette.TimeOfDayDusk:  1.05, // Mystical twilight
			palette.TimeOfDayDawn:  1.00, // Neutral
			palette.TimeOfDayDay:   0.95, // Slightly harder to focus
		},
		// Nature elements: Empowered by transitions
		magic.ElementEarth: {
			palette.TimeOfDayDawn:  1.10, // Morning dew, fresh earth
			palette.TimeOfDayDusk:  1.10, // Evening calm
			palette.TimeOfDayDay:   1.00, // Neutral
			palette.TimeOfDayNight: 1.00, // Neutral
		},
		magic.ElementWind: {
			palette.TimeOfDayDawn:  1.10, // Morning breeze
			palette.TimeOfDayDusk:  1.10, // Evening wind
			palette.TimeOfDayDay:   1.00, // Neutral
			palette.TimeOfDayNight: 1.00, // Neutral
		},
		// Ice: Cold thrives at night
		magic.ElementIce: {
			palette.TimeOfDayNight: 1.10, // Cold darkness
			palette.TimeOfDayDay:   0.95, // Sun weakens ice slightly
			palette.TimeOfDayDawn:  1.00, // Neutral
			palette.TimeOfDayDusk:  1.00, // Neutral
		},
		// Lightning: Storm activity varies
		magic.ElementLightning: {
			palette.TimeOfDayDay:   1.05, // Storm season activity
			palette.TimeOfDayNight: 1.05, // Night storms
			palette.TimeOfDayDawn:  1.00, // Neutral
			palette.TimeOfDayDusk:  1.00, // Neutral
		},
		// No element: Always neutral
		magic.ElementNone: {
			palette.TimeOfDayDay:   1.00,
			palette.TimeOfDayDawn:  1.00,
			palette.TimeOfDayDusk:  1.00,
			palette.TimeOfDayNight: 1.00,
		},
	}
}

// initSpellDamageGenreModifiers sets up genre-specific adjustments.
func initSpellDamageGenreModifiers() map[string]map[magic.ElementType]float64 {
	return map[string]map[magic.ElementType]float64{
		"fantasy": {
			// Traditional magic: balanced
			magic.ElementFire:  0.00,
			magic.ElementLight: 0.00,
			magic.ElementDark:  0.00,
		},
		"horror": {
			// Dark magic dominant: dark stronger, light weaker
			magic.ElementDark:  0.10,
			magic.ElementLight: -0.10,
		},
		"scifi": {
			// Tech-enhanced: lightning/arcane stronger
			magic.ElementLightning: 0.05,
			magic.ElementArcane:    0.05,
		},
		"cyberpunk": {
			// Neon nights: dark and lightning stronger at night
			magic.ElementDark:      0.05,
			magic.ElementLightning: 0.05,
		},
		"postapoc": {
			// Harsh sun: fire stronger during day, earth stronger overall
			magic.ElementFire:  0.05,
			magic.ElementEarth: 0.05,
		},
	}
}

// SetGenre sets the genre for genre-aware spell damage modifiers.
func (s *TimeOfDaySpellDamageSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.modCache = make(map[timeElementKey]float64, 32) // Clear cache on genre change
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDaySpellDamageSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// Update checks time of day and invalidates cache when time changes.
func (s *TimeOfDaySpellDamageSystem) Update(entities []*Entity, deltaTime float64) {
	if s.lightingSystem == nil {
		return
	}

	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	currentTime := s.lightingSystem.GetCurrentTimeOfDay()

	// Clear cache on time change
	if currentTime != s.lastTimeOfDay {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"time_of_day": currentTime.String(),
			}).Debug("time of day changed, clearing spell damage cache")
		}
		s.lastTimeOfDay = currentTime
		s.modCache = make(map[timeElementKey]float64, 32)
	}
}

// GetDamageModifier returns the damage multiplier for a spell element at current time.
// Returns 1.0 if no lighting system is linked.
// Values > 1.0 indicate bonus damage, < 1.0 indicate reduced damage.
func (s *TimeOfDaySpellDamageSystem) GetDamageModifier(element magic.ElementType) float64 {
	if s.lightingSystem == nil {
		return 1.0
	}

	currentTime := s.lightingSystem.GetCurrentTimeOfDay()
	key := timeElementKey{timeOfDay: currentTime, element: element}

	if cached, ok := s.modCache[key]; ok {
		return cached
	}

	modifier := s.calculateModifier(element, currentTime)
	s.modCache[key] = modifier

	if s.logger != nil && modifier != 1.0 {
		s.logger.WithFields(logrus.Fields{
			"time_of_day": currentTime.String(),
			"element":     element.String(),
			"modifier":    modifier,
		}).Debug("time-of-day spell damage modifier calculated")
	}

	return modifier
}

// calculateModifier computes the damage modifier for an element at a time of day.
func (s *TimeOfDaySpellDamageSystem) calculateModifier(element magic.ElementType, timeOfDay palette.TimeOfDay) float64 {
	// Get base multiplier
	multiplier := 1.0
	if timeMap, ok := s.baseMultipliers[element]; ok {
		if mult, ok := timeMap[timeOfDay]; ok {
			multiplier = mult
		}
	}

	// Apply genre modifier
	if genreMods, ok := s.genreModifiers[s.genreID]; ok {
		if mod, ok := genreMods[element]; ok {
			multiplier += mod
		}
	}

	// Clamp to reasonable bounds
	if multiplier < 0.5 {
		multiplier = 0.5
	}
	if multiplier > 1.5 {
		multiplier = 1.5
	}

	return multiplier
}

// GetCurrentTimeOfDay returns the current time of day for external queries.
func (s *TimeOfDaySpellDamageSystem) GetCurrentTimeOfDay() (palette.TimeOfDay, bool) {
	if s.lightingSystem == nil {
		return palette.TimeOfDayDay, false
	}
	return s.lightingSystem.GetCurrentTimeOfDay(), true
}

// GetBonusDescription returns a human-readable description of time-based damage bonuses.
func (s *TimeOfDaySpellDamageSystem) GetBonusDescription() string {
	if s.lightingSystem == nil {
		return ""
	}

	timeOfDay := s.lightingSystem.GetCurrentTimeOfDay()

	// Find the most beneficial element at current time
	bestElement := magic.ElementNone
	bestBonus := 0.0

	for elem := magic.ElementFire; elem <= magic.ElementArcane; elem++ {
		mult := s.calculateModifier(elem, timeOfDay)
		if mult-1.0 > bestBonus {
			bestBonus = mult - 1.0
			bestElement = elem
		}
	}

	if bestBonus <= 0 {
		return ""
	}

	bonus := int(bestBonus * 100)
	return timeOfDay.String() + ": " + bestElement.String() + " spells +" + intToStr(bonus) + "% damage"
}

// ClearCache clears the modifier cache, useful when time changes abruptly.
func (s *TimeOfDaySpellDamageSystem) ClearCache() {
	s.modCache = make(map[timeElementKey]float64, 32)
}
