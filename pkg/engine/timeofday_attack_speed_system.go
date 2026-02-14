// Package engine provides the TimeOfDayAttackSpeedSystem which bridges time-of-day
// lighting cycles with melee attack speed (cooldown modifiers). Night enables faster
// attacks through cover and concealment, while daytime visibility slightly increases
// reaction time for opponents. Genre-specific modifiers tune these bonuses for
// thematic gameplay (fantasy rogues strike faster at night, sci-fi sensors compensate).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayAttackSpeedSystem modifies melee attack cooldowns based on time of day.
// This connects the TimeOfDayLightingSystem with AttackComponent.Cooldown.
//
// Cooldown multipliers by time of day (lower = faster attacks):
//   - Dawn: 0.97 (-3% cooldown, transitional advantage)
//   - Day: 1.03 (+3% cooldown, visibility aids defense)
//   - Dusk: 0.95 (-5% cooldown, lengthening shadows enable quick strikes)
//   - Night: 0.90 (-10% cooldown, darkness covers rapid attacks)
//
// Genre modifiers adjust these base values:
//   - Fantasy: Night 0.88 (rogues excel in shadows)
//   - Scifi: Night 0.95 (sensors partially compensate for darkness)
//   - Horror: Day 1.05 (survival horror - enemies react faster in light)
//   - Cyberpunk: Dusk/Night 0.92 (urban shadows favor quick strikes)
//   - Postapoc: Night 0.89 (predator instincts)
type TimeOfDayAttackSpeedSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// Reference to lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// updateInterval controls how often we check time (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// Track original cooldown values per entity for restoration
	originalCooldown map[uint64]float64

	// Last known time of day for change detection
	lastTimeOfDay palette.TimeOfDay

	// Genre affects which times boost attack speed
	genreID string

	// Base cooldown multipliers by time of day (lower = faster)
	baseMultipliers map[palette.TimeOfDay]float64

	// Genre-specific multipliers (replace base when genre matches)
	genreMultipliers map[string]map[palette.TimeOfDay]float64
}

// NewTimeOfDayAttackSpeedSystem creates a new time-of-day attack speed system.
func NewTimeOfDayAttackSpeedSystem(world *World, seed int64) *TimeOfDayAttackSpeedSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_attack_speed")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayAttackSpeedSystem created")
	}

	return &TimeOfDayAttackSpeedSystem{
		world:            world,
		logger:           logEntry,
		rng:              rand.New(rand.NewSource(seed)),
		updateInterval:   1.0, // Check once per second (time changes slowly)
		originalCooldown: make(map[uint64]float64, 64),
		lastTimeOfDay:    palette.TimeOfDayDay,
		genreID:          "fantasy",
		baseMultipliers: map[palette.TimeOfDay]float64{
			palette.TimeOfDayDawn:  0.97, // Transitional advantage
			palette.TimeOfDayDay:   1.03, // Visibility aids defense
			palette.TimeOfDayDusk:  0.95, // Lengthening shadows
			palette.TimeOfDayNight: 0.90, // Darkness covers rapid attacks
		},
		genreMultipliers: map[string]map[palette.TimeOfDay]float64{
			"fantasy": {
				palette.TimeOfDayNight: 0.88, // Rogues excel in shadows
				palette.TimeOfDayDusk:  0.93, // Twilight ambushes
			},
			"scifi": {
				palette.TimeOfDayNight: 0.95, // Sensors partially compensate
				palette.TimeOfDayDay:   1.00, // Tech maintains baseline
			},
			"horror": {
				palette.TimeOfDayDay:   1.05, // Enemies more reactive in light
				palette.TimeOfDayNight: 0.92, // Panic enables quick strikes
			},
			"cyberpunk": {
				palette.TimeOfDayNight: 0.92, // Urban shadows
				palette.TimeOfDayDusk:  0.92, // Neon distractions aid attackers
			},
			"postapoc": {
				palette.TimeOfDayNight: 0.89, // Predator instincts
				palette.TimeOfDayDawn:  0.95, // Early survivor reflexes
			},
		},
	}
}

// SetGenre sets the genre for genre-aware attack speed modifiers.
func (s *TimeOfDayAttackSpeedSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayAttackSpeedSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// Update checks time of day and updates attack cooldown multipliers for combatants.
func (s *TimeOfDayAttackSpeedSystem) Update(entities []*Entity, deltaTime float64) {
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
	timeChanged := currentTime != s.lastTimeOfDay
	if timeChanged {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"time_of_day": currentTime.String(),
				"multiplier":  s.getCooldownMultiplier(currentTime),
			}).Debug("time of day changed, updating attack speed modifiers")
		}
		s.lastTimeOfDay = currentTime
	}

	// Update cooldown for all entities with attack component
	multiplier := s.getCooldownMultiplier(currentTime)
	for _, entity := range entities {
		if !entity.HasComponent("attack") {
			continue
		}

		attackComp, ok := entity.GetComponent("attack")
		if !ok {
			continue
		}

		attack, ok := attackComp.(*AttackComponent)
		if !ok {
			continue
		}

		// Store original cooldown if not already stored
		if _, exists := s.originalCooldown[entity.ID]; !exists {
			s.originalCooldown[entity.ID] = attack.Cooldown
		}

		// Apply time-of-day modified cooldown
		originalCooldown := s.originalCooldown[entity.ID]
		newCooldown := originalCooldown * multiplier

		// Clamp to reasonable range (min 0.1s, max 10x original)
		if newCooldown < 0.1 {
			newCooldown = 0.1
		}
		maxCooldown := originalCooldown * 2.0
		if newCooldown > maxCooldown {
			newCooldown = maxCooldown
		}

		attack.Cooldown = newCooldown
	}
}

// getCooldownMultiplier calculates the cooldown multiplier for the current time.
func (s *TimeOfDayAttackSpeedSystem) getCooldownMultiplier(timeOfDay palette.TimeOfDay) float64 {
	// Check for genre-specific override first
	if genreMults, ok := s.genreMultipliers[s.genreID]; ok {
		if mult, ok := genreMults[timeOfDay]; ok {
			return mult
		}
	}

	// Fall back to base multiplier
	if mult, ok := s.baseMultipliers[timeOfDay]; ok {
		return mult
	}

	// Default to no modifier
	return 1.0
}

// GetCurrentMultiplier returns the current cooldown multiplier based on time of day.
// Values < 1.0 mean faster attacks, > 1.0 means slower attacks.
func (s *TimeOfDayAttackSpeedSystem) GetCurrentMultiplier() float64 {
	if s.lightingSystem == nil {
		return 1.0
	}
	return s.getCooldownMultiplier(s.lightingSystem.GetCurrentTimeOfDay())
}

// GetOriginalCooldown returns the stored original cooldown for an entity.
func (s *TimeOfDayAttackSpeedSystem) GetOriginalCooldown(entityID uint64) (float64, bool) {
	c, ok := s.originalCooldown[entityID]
	return c, ok
}

// GetBonusDescription returns a human-readable description of the current bonus.
func (s *TimeOfDayAttackSpeedSystem) GetBonusDescription() string {
	if s.lightingSystem == nil {
		return ""
	}

	timeOfDay := s.lightingSystem.GetCurrentTimeOfDay()
	mult := s.getCooldownMultiplier(timeOfDay)

	if mult < 1.0 {
		bonus := int((1.0 - mult) * 100)
		return timeOfDay.String() + " Swiftness: +" + attackSpeedItoa(bonus) + "% Attack Speed"
	}
	if mult > 1.0 {
		penalty := int((mult - 1.0) * 100)
		return timeOfDay.String() + " Sluggish: -" + attackSpeedItoa(penalty) + "% Attack Speed"
	}
	return ""
}

// attackSpeedItoa converts an int to string without importing strconv.
func attackSpeedItoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + attackSpeedItoa(-n)
	}
	digits := ""
	for n > 0 {
		digits = string('0'+byte(n%10)) + digits
		n /= 10
	}
	return digits
}
