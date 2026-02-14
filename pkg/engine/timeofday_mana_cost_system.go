// Package engine provides the TimeOfDayManaCostSystem which bridges time-of-day
// cycles with spell mana costs based on elemental affinity. Fire and Light spells
// are cheaper during day, Dark and Shadow spells are cheaper at night, and Nature
// (Earth, Wind) spells are cheaper at dawn/dusk transitions.
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

// TimeOfDayManaCostSystem modifies spell mana costs based on time of day and
// spell element. This connects TimeOfDayLightingSystem with SpellCastingSystem.
//
// Element-Time cost multipliers:
//   - Fire, Light: Day 0.85, Dawn 0.95, Dusk 1.05, Night 1.15
//   - Dark, Arcane: Night 0.85, Dusk 0.95, Dawn 1.05, Day 1.15
//   - Earth, Wind: Dawn 0.90, Dusk 0.90, Day 1.0, Night 1.0
//   - Ice, Lightning: Day 0.95, Night 0.95, Dawn 1.0, Dusk 1.0
//   - None: No time-based modification (1.0 always)
type TimeOfDayManaCostSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Reference to lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// Update interval to avoid per-frame recalculation
	updateInterval float64
	timeSinceCheck float64

	// Cache multipliers per entity to avoid redundant lookups
	// Maps entityID -> map of spell element -> cost multiplier
	elementMultiplierCache map[uint64]map[magic.ElementType]float64

	// Last known time of day for change detection
	lastTimeOfDay palette.TimeOfDay

	// Genre affects element-time relationships
	genreID string

	// Base multipliers: element -> timeOfDay -> multiplier
	baseMultipliers map[magic.ElementType]map[palette.TimeOfDay]float64

	// Genre modifiers (additive adjustment to base multipliers)
	genreModifiers map[string]map[magic.ElementType]float64
}

// NewTimeOfDayManaCostSystem creates a new time-of-day mana cost system.
func NewTimeOfDayManaCostSystem(world *World, seed int64) *TimeOfDayManaCostSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_mana_cost")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayManaCostSystem created")
	}

	return &TimeOfDayManaCostSystem{
		world:                  world,
		logger:                 logEntry,
		rng:                    rand.New(rand.NewSource(seed)),
		updateInterval:         0.5, // Check twice per second
		elementMultiplierCache: make(map[uint64]map[magic.ElementType]float64, 32),
		lastTimeOfDay:          palette.TimeOfDayDay,
		genreID:                "fantasy",
		baseMultipliers:        initBaseMultipliers(),
		genreModifiers:         initGenreModifiers(),
	}
}

// initBaseMultipliers sets up the element-time cost relationships.
func initBaseMultipliers() map[magic.ElementType]map[palette.TimeOfDay]float64 {
	return map[magic.ElementType]map[palette.TimeOfDay]float64{
		// Fire/Light: Empowered by sunlight
		magic.ElementFire: {
			palette.TimeOfDayDay:   0.85, // Strong sunlight empowers fire
			palette.TimeOfDayDawn:  0.95, // Rising sun
			palette.TimeOfDayDusk:  1.05, // Fading light
			palette.TimeOfDayNight: 1.15, // Darkness suppresses fire magic
		},
		magic.ElementLight: {
			palette.TimeOfDayDay:   0.85, // Holy light at peak
			palette.TimeOfDayDawn:  0.90, // New day's blessing
			palette.TimeOfDayDusk:  1.05, // Light fading
			palette.TimeOfDayNight: 1.20, // Light struggles in darkness
		},
		// Dark/Arcane: Empowered by darkness
		magic.ElementDark: {
			palette.TimeOfDayNight: 0.85, // Shadows at full strength
			palette.TimeOfDayDusk:  0.95, // Growing darkness
			palette.TimeOfDayDawn:  1.05, // Light returns
			palette.TimeOfDayDay:   1.15, // Sunlight weakens shadows
		},
		magic.ElementArcane: {
			palette.TimeOfDayNight: 0.90, // Stars empower arcane
			palette.TimeOfDayDusk:  0.95, // Mystical twilight
			palette.TimeOfDayDawn:  1.00, // Neutral
			palette.TimeOfDayDay:   1.05, // Slightly harder to focus
		},
		// Nature elements: Empowered by transitions
		magic.ElementEarth: {
			palette.TimeOfDayDawn:  0.90, // Morning dew, fresh earth
			palette.TimeOfDayDusk:  0.90, // Evening calm
			palette.TimeOfDayDay:   1.00, // Neutral
			palette.TimeOfDayNight: 1.00, // Neutral
		},
		magic.ElementWind: {
			palette.TimeOfDayDawn:  0.90, // Morning breeze
			palette.TimeOfDayDusk:  0.90, // Evening wind
			palette.TimeOfDayDay:   1.00, // Neutral
			palette.TimeOfDayNight: 1.00, // Neutral
		},
		// Elemental extremes: Slight bonuses at opposite times
		magic.ElementIce: {
			palette.TimeOfDayNight: 0.95, // Cold darkness
			palette.TimeOfDayDay:   0.95, // Sun doesn't melt magic ice
			palette.TimeOfDayDawn:  1.00, // Neutral
			palette.TimeOfDayDusk:  1.00, // Neutral
		},
		magic.ElementLightning: {
			palette.TimeOfDayDay:   0.95, // Storm season
			palette.TimeOfDayNight: 0.95, // Night storms
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

// initGenreModifiers sets up genre-specific adjustments.
func initGenreModifiers() map[string]map[magic.ElementType]float64 {
	return map[string]map[magic.ElementType]float64{
		"fantasy": {
			// Traditional magic: slight boost to all elements
			magic.ElementFire:  -0.02,
			magic.ElementLight: -0.02,
			magic.ElementDark:  -0.02,
		},
		"horror": {
			// Dark magic dominant: dark cheaper, light more expensive
			magic.ElementDark:  -0.10,
			magic.ElementLight: 0.10,
		},
		"scifi": {
			// Tech-enhanced: lightning/arcane cheaper
			magic.ElementLightning: -0.05,
			magic.ElementArcane:    -0.05,
		},
		"cyberpunk": {
			// Neon nights: dark and lightning cheaper
			magic.ElementDark:      -0.05,
			magic.ElementLightning: -0.05,
		},
		"postapoc": {
			// Harsh sun: fire more expensive, earth cheaper
			magic.ElementFire:  0.05,
			magic.ElementEarth: -0.05,
		},
	}
}

// SetGenre sets the genre for genre-aware mana cost modifiers.
func (s *TimeOfDayManaCostSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayManaCostSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// Update checks time of day and caches element-based mana cost multipliers.
func (s *TimeOfDayManaCostSystem) Update(entities []*Entity, deltaTime float64) {
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
			}).Debug("time of day changed, clearing mana cost cache")
		}
		s.lastTimeOfDay = currentTime
		// Clear the cache
		for k := range s.elementMultiplierCache {
			delete(s.elementMultiplierCache, k)
		}
	}

	// Pre-compute multipliers for spellcasting entities
	for _, entity := range entities {
		if !entity.HasComponent("spell_slots") {
			continue
		}

		// Build element multiplier map for this entity
		elemMap := make(map[magic.ElementType]float64, 9)
		for elem := magic.ElementNone; elem <= magic.ElementArcane; elem++ {
			elemMap[elem] = s.getMultiplier(elem, currentTime)
		}
		s.elementMultiplierCache[entity.ID] = elemMap
	}
}

// getMultiplier calculates the mana cost multiplier for an element at current time.
func (s *TimeOfDayManaCostSystem) getMultiplier(element magic.ElementType, timeOfDay palette.TimeOfDay) float64 {
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

// GetElementMultiplier returns the mana cost multiplier for a specific element
// and entity. Returns 1.0 if no modifier is cached.
func (s *TimeOfDayManaCostSystem) GetElementMultiplier(entityID uint64, element magic.ElementType) float64 {
	if elemMap, ok := s.elementMultiplierCache[entityID]; ok {
		if mult, ok := elemMap[element]; ok {
			return mult
		}
	}
	// Fallback: compute directly if not cached
	if s.lightingSystem != nil {
		return s.getMultiplier(element, s.lightingSystem.GetCurrentTimeOfDay())
	}
	return 1.0
}

// GetEffectiveManaCost calculates the effective mana cost for a spell,
// applying time-of-day element modifiers. Returns the modified cost.
func (s *TimeOfDayManaCostSystem) GetEffectiveManaCost(entityID uint64, spell *magic.Spell) int {
	if spell == nil {
		return 0
	}

	baseCost := spell.Stats.ManaCost
	multiplier := s.GetElementMultiplier(entityID, spell.Element)
	effectiveCost := float64(baseCost) * multiplier

	// Round to nearest integer, minimum 1 mana
	result := int(effectiveCost + 0.5)
	if result < 1 {
		result = 1
	}

	return result
}

// GetCurrentMultiplierForElement returns the current multiplier for an element
// based on time of day. Useful for UI display.
func (s *TimeOfDayManaCostSystem) GetCurrentMultiplierForElement(element magic.ElementType) float64 {
	if s.lightingSystem == nil {
		return 1.0
	}
	return s.getMultiplier(element, s.lightingSystem.GetCurrentTimeOfDay())
}

// GetBonusDescription returns a human-readable description of time-based bonuses.
func (s *TimeOfDayManaCostSystem) GetBonusDescription() string {
	if s.lightingSystem == nil {
		return ""
	}

	timeOfDay := s.lightingSystem.GetCurrentTimeOfDay()

	// Find the most beneficial element at current time
	bestElement := magic.ElementNone
	bestBonus := 0.0

	for elem := magic.ElementFire; elem <= magic.ElementArcane; elem++ {
		mult := s.getMultiplier(elem, timeOfDay)
		if 1.0-mult > bestBonus {
			bestBonus = 1.0 - mult
			bestElement = elem
		}
	}

	if bestBonus <= 0 {
		return ""
	}

	bonus := int(bestBonus * 100)
	return timeOfDay.String() + ": " + bestElement.String() + " spells -" + intToStr(bonus) + "% mana"
}

// intToStr converts an int to string without importing strconv.
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToStr(-n)
	}
	digits := ""
	for n > 0 {
		digits = string('0'+byte(n%10)) + digits
		n /= 10
	}
	return digits
}
