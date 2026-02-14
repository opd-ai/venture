// Package engine provides the TimeOfDayCompanionBonusSystem which bridges
// time-of-day cycles with companion combat statistics based on companion type.
// This creates tactical depth where daytime and nighttime affect companion
// effectiveness differently (e.g., undead companions gain bonuses at night).
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/sirupsen/logrus"
)

// TimeOfDayCompanionBonusSystem modifies companion stats based on time of day
// and companion type. Different companion types benefit from different times:
//   - Undead companions: bonuses at night, penalties during day
//   - Spirit companions: bonuses at dawn/dusk (liminal times), neutral otherwise
//   - Pet companions: bonuses during day, penalties at night
//   - Elemental companions: mostly neutral with slight dawn/dusk variance
//   - Robot companions: consistent (unaffected by day/night)
//   - Insect companions: bonuses during day, penalties at night
//
// Genre-specific modifiers adjust bonus strength.
type TimeOfDayCompanionBonusSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry
	genre  string

	// Reference to time-of-day lighting system for current time state
	lightingSystem *TimeOfDayLightingSystem

	// Update throttling
	updateInterval float64
	timeSinceCheck float64

	// Cache for time-of-day bonuses to avoid recalculating each frame
	bonusCache map[uint64]*TimeOfDayCompanionBonusComponent

	// Last known time of day for change detection
	lastTimeOfDay palette.TimeOfDay

	// Genre multipliers for bonus scaling
	genreMultipliers map[string]float64

	// Time of day to companion type bonus mappings
	timeBonuses map[palette.TimeOfDay]map[CompanionType]timeCompanionBonus
}

// timeCompanionBonus defines stat bonuses for a time/companion combination
type timeCompanionBonus struct {
	attackMult  float64 // Multiplier (1.0 = no change, 1.2 = +20%)
	defenseMult float64
	speedMult   float64
}

// TimeOfDayCompanionBonusComponent stores time-based combat modifiers for companions.
// This is a transient component recalculated when time of day changes.
type TimeOfDayCompanionBonusComponent struct {
	// AttackBonus multiplier (1.0 = no bonus, 1.2 = +20%)
	AttackBonus float64

	// DefenseBonus multiplier (1.0 = no bonus, 0.8 = -20%)
	DefenseBonus float64

	// SpeedBonus multiplier (1.0 = no bonus, 1.15 = +15%)
	SpeedBonus float64

	// TimeOfDay is the current time providing the bonus
	TimeOfDay string

	// CompanionTypeName is the companion type receiving the bonus
	CompanionTypeName string
}

// Type returns the component type identifier.
func (c *TimeOfDayCompanionBonusComponent) Type() string {
	return "timeofday_companion_bonus"
}

// NewTimeOfDayCompanionBonusSystem creates a new time-of-day companion bonus system.
func NewTimeOfDayCompanionBonusSystem(world *World, seed int64) *TimeOfDayCompanionBonusSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "timeofday_companion_bonus")
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("TimeOfDayCompanionBonusSystem created")
	}

	s := &TimeOfDayCompanionBonusSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		logger:         logEntry,
		updateInterval: 1.0, // Check once per second (time changes slowly)
		bonusCache:     make(map[uint64]*TimeOfDayCompanionBonusComponent, 32),
		lastTimeOfDay:  palette.TimeOfDayDay,
		genreMultipliers: map[string]float64{
			"fantasy":   1.0, // Standard day/night effects
			"scifi":     0.5, // Artificial lighting reduces natural effects
			"horror":    1.4, // Strong night/darkness influence
			"cyberpunk": 0.4, // Constant artificial lighting
			"postapoc":  1.2, // Strong natural cycles
		},
	}

	s.initTimeBonuses()
	return s
}

// initTimeBonuses initializes the time-to-companion bonus mappings.
func (s *TimeOfDayCompanionBonusSystem) initTimeBonuses() {
	s.timeBonuses = make(map[palette.TimeOfDay]map[CompanionType]timeCompanionBonus)

	// Dawn bonuses - liminal time, spirits empowered
	s.timeBonuses[palette.TimeOfDayDawn] = map[CompanionType]timeCompanionBonus{
		CompanionTypeSpirit:    {attackMult: 1.20, defenseMult: 1.15, speedMult: 1.10}, // Liminal power
		CompanionTypeElemental: {attackMult: 1.10, defenseMult: 1.05, speedMult: 1.05}, // Elemental awakening
		CompanionTypeUndead:    {attackMult: 0.90, defenseMult: 0.95, speedMult: 0.95}, // Weakening as sun rises
		CompanionTypePet:       {attackMult: 1.05, defenseMult: 1.05, speedMult: 1.10}, // Morning energy
		CompanionTypeInsect:    {attackMult: 1.10, defenseMult: 1.05, speedMult: 1.05}, // Active at dawn
		CompanionTypeRobot:     {attackMult: 1.00, defenseMult: 1.00, speedMult: 1.00}, // Unaffected
	}

	// Day bonuses - full light, living creatures thrive
	s.timeBonuses[palette.TimeOfDayDay] = map[CompanionType]timeCompanionBonus{
		CompanionTypePet:       {attackMult: 1.15, defenseMult: 1.10, speedMult: 1.15}, // Energetic in daylight
		CompanionTypeInsect:    {attackMult: 1.15, defenseMult: 1.10, speedMult: 1.10}, // Most insects active
		CompanionTypeElemental: {attackMult: 1.05, defenseMult: 1.05, speedMult: 1.00}, // Solar energy (fire/light)
		CompanionTypeSpirit:    {attackMult: 0.95, defenseMult: 0.95, speedMult: 1.00}, // Diminished presence
		CompanionTypeUndead:    {attackMult: 0.75, defenseMult: 0.80, speedMult: 0.85}, // Sunlight weakens
		CompanionTypeRobot:     {attackMult: 1.00, defenseMult: 1.00, speedMult: 1.00}, // Unaffected
	}

	// Dusk bonuses - liminal time, spirits empowered
	s.timeBonuses[palette.TimeOfDayDusk] = map[CompanionType]timeCompanionBonus{
		CompanionTypeSpirit:    {attackMult: 1.25, defenseMult: 1.20, speedMult: 1.15}, // Peak liminal power
		CompanionTypeUndead:    {attackMult: 1.10, defenseMult: 1.10, speedMult: 1.05}, // Awakening as sun sets
		CompanionTypeElemental: {attackMult: 1.05, defenseMult: 1.05, speedMult: 1.00}, // Transition energy
		CompanionTypePet:       {attackMult: 0.95, defenseMult: 1.00, speedMult: 0.95}, // Winding down
		CompanionTypeInsect:    {attackMult: 1.00, defenseMult: 1.00, speedMult: 0.95}, // Some still active
		CompanionTypeRobot:     {attackMult: 1.00, defenseMult: 1.00, speedMult: 1.00}, // Unaffected
	}

	// Night bonuses - darkness empowers undead and spirits
	s.timeBonuses[palette.TimeOfDayNight] = map[CompanionType]timeCompanionBonus{
		CompanionTypeUndead:    {attackMult: 1.35, defenseMult: 1.25, speedMult: 1.15}, // Full dark power
		CompanionTypeSpirit:    {attackMult: 1.15, defenseMult: 1.10, speedMult: 1.10}, // Nocturnal spirits
		CompanionTypeElemental: {attackMult: 1.00, defenseMult: 1.00, speedMult: 0.95}, // Neutral (varies by element)
		CompanionTypePet:       {attackMult: 0.85, defenseMult: 0.90, speedMult: 0.85}, // Less active at night
		CompanionTypeInsect:    {attackMult: 0.80, defenseMult: 0.85, speedMult: 0.80}, // Most insects dormant
		CompanionTypeRobot:     {attackMult: 1.00, defenseMult: 1.00, speedMult: 1.00}, // Unaffected
	}
}

// SetLightingSystem sets the time-of-day lighting system reference.
func (s *TimeOfDayCompanionBonusSystem) SetLightingSystem(lightingSystem *TimeOfDayLightingSystem) {
	s.lightingSystem = lightingSystem
	if s.logger != nil {
		s.logger.Debug("lighting system linked")
	}
}

// SetGenre sets the genre for genre-specific time modifiers.
func (s *TimeOfDayCompanionBonusSystem) SetGenre(genreID string) {
	s.genre = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set")
	}
}

// Update processes all companion entities and applies time-of-day bonuses.
func (s *TimeOfDayCompanionBonusSystem) Update(entities []*Entity, deltaTime float64) {
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

	// If time hasn't changed, nothing to update
	if currentTime == s.lastTimeOfDay {
		return
	}

	// Time changed - update all companion bonuses
	s.updateAllCompanionBonuses(entities, currentTime)
	s.lastTimeOfDay = currentTime

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"time_of_day":     currentTime.String(),
			"companion_count": len(s.bonusCache),
		}).Debug("time of day changed, updating companion bonuses")
	}
}

// updateAllCompanionBonuses recalculates bonuses for all companions when time changes.
func (s *TimeOfDayCompanionBonusSystem) updateAllCompanionBonuses(entities []*Entity, timeOfDay palette.TimeOfDay) {
	for _, entity := range entities {
		// Only process companions
		compComp, hasCompanion := entity.GetComponent("companion")
		if !hasCompanion {
			continue
		}

		companion, ok := compComp.(*CompanionComponent)
		if !ok {
			continue
		}

		s.applyTimeBonus(entity, companion, timeOfDay)
	}
}

// applyTimeBonus applies time-based bonuses to a companion entity.
func (s *TimeOfDayCompanionBonusSystem) applyTimeBonus(entity *Entity, companion *CompanionComponent, timeOfDay palette.TimeOfDay) {
	bonus := s.calculateTimeBonus(companion.CompanionType, timeOfDay)
	if bonus == nil {
		s.removeTimeBonus(entity)
		return
	}

	// Reverse old bonus before applying new
	if existingComp, ok := entity.GetComponent("timeofday_companion_bonus"); ok {
		if existingBonus, ok := existingComp.(*TimeOfDayCompanionBonusComponent); ok {
			s.reverseStatBonus(entity, existingBonus)
		}
	}

	// Apply new bonus to companion stats
	s.applyBonusToStats(entity, bonus)

	// Update or add component
	if existing, hasExisting := entity.GetComponent("timeofday_companion_bonus"); hasExisting {
		if existingBonus, ok := existing.(*TimeOfDayCompanionBonusComponent); ok {
			*existingBonus = *bonus
		}
	} else {
		entity.AddComponent(bonus)
	}

	s.bonusCache[entity.ID] = bonus
	s.logBonusApplication(entity, bonus, companion.CompanionType, timeOfDay)
}

// removeTimeBonus removes time bonus from a companion entity.
func (s *TimeOfDayCompanionBonusSystem) removeTimeBonus(entity *Entity) {
	if existingComp, ok := entity.GetComponent("timeofday_companion_bonus"); ok {
		if bonus, ok := existingComp.(*TimeOfDayCompanionBonusComponent); ok {
			s.reverseStatBonus(entity, bonus)
		}
		entity.RemoveComponent("timeofday_companion_bonus")
	}
	delete(s.bonusCache, entity.ID)
}

// calculateTimeBonus computes bonuses for a time/companion type combination.
func (s *TimeOfDayCompanionBonusSystem) calculateTimeBonus(compType CompanionType, timeOfDay palette.TimeOfDay) *TimeOfDayCompanionBonusComponent {
	timeMap, hasTime := s.timeBonuses[timeOfDay]
	if !hasTime {
		return nil
	}

	bonusData, hasBonus := timeMap[compType]
	if !hasBonus {
		return nil
	}

	// Apply genre multiplier
	genreMult := s.genreMultipliers[s.genre]
	if genreMult == 0 {
		genreMult = 1.0
	}

	// Scale bonus away from 1.0 by genre multiplier
	attackBonus := 1.0 + (bonusData.attackMult-1.0)*genreMult
	defenseBonus := 1.0 + (bonusData.defenseMult-1.0)*genreMult
	speedBonus := 1.0 + (bonusData.speedMult-1.0)*genreMult

	return &TimeOfDayCompanionBonusComponent{
		AttackBonus:       attackBonus,
		DefenseBonus:      defenseBonus,
		SpeedBonus:        speedBonus,
		TimeOfDay:         timeOfDay.String(),
		CompanionTypeName: s.companionTypeName(compType),
	}
}

// applyBonusToStats applies the time bonus to companion stats.
func (s *TimeOfDayCompanionBonusSystem) applyBonusToStats(entity *Entity, bonus *TimeOfDayCompanionBonusComponent) {
	statsComp, ok := entity.GetComponent("companionstats")
	if !ok {
		return
	}

	stats, ok := statsComp.(*CompanionStatsComponent)
	if !ok {
		return
	}

	stats.Attack *= bonus.AttackBonus
	stats.Defense *= bonus.DefenseBonus
	stats.Speed *= bonus.SpeedBonus
}

// reverseStatBonus reverses previously applied bonus from companion stats.
func (s *TimeOfDayCompanionBonusSystem) reverseStatBonus(entity *Entity, bonus *TimeOfDayCompanionBonusComponent) {
	statsComp, ok := entity.GetComponent("companionstats")
	if !ok {
		return
	}

	stats, ok := statsComp.(*CompanionStatsComponent)
	if !ok {
		return
	}

	if bonus.AttackBonus != 0 {
		stats.Attack /= bonus.AttackBonus
	}
	if bonus.DefenseBonus != 0 {
		stats.Defense /= bonus.DefenseBonus
	}
	if bonus.SpeedBonus != 0 {
		stats.Speed /= bonus.SpeedBonus
	}
}

// companionTypeName returns a string name for the companion type.
func (s *TimeOfDayCompanionBonusSystem) companionTypeName(compType CompanionType) string {
	switch compType {
	case CompanionTypePet:
		return "Pet"
	case CompanionTypeSummon:
		return "Summon"
	case CompanionTypeHireling:
		return "Hireling"
	case CompanionTypeElemental:
		return "Elemental"
	case CompanionTypeUndead:
		return "Undead"
	case CompanionTypeRobot:
		return "Robot"
	case CompanionTypeSpirit:
		return "Spirit"
	case CompanionTypeInsect:
		return "Insect"
	default:
		return "Unknown"
	}
}

// logBonusApplication logs when a time bonus is applied.
func (s *TimeOfDayCompanionBonusSystem) logBonusApplication(entity *Entity, bonus *TimeOfDayCompanionBonusComponent, compType CompanionType, timeOfDay palette.TimeOfDay) {
	if s.logger == nil || s.logger.Logger.GetLevel() < logrus.DebugLevel {
		return
	}

	s.logger.WithFields(logrus.Fields{
		"entity_id":      entity.ID,
		"time_of_day":    timeOfDay.String(),
		"companion_type": s.companionTypeName(compType),
		"attack_bonus":   bonus.AttackBonus,
		"defense_bonus":  bonus.DefenseBonus,
		"speed_bonus":    bonus.SpeedBonus,
		"genre":          s.genre,
	}).Debug("time-of-day companion bonus applied")
}

// HasActiveBonus returns whether a companion has an active time-of-day bonus.
func (s *TimeOfDayCompanionBonusSystem) HasActiveBonus(companionID uint64) bool {
	_, exists := s.bonusCache[companionID]
	return exists
}

// GetBonusCount returns the number of companions with active time-of-day bonuses.
func (s *TimeOfDayCompanionBonusSystem) GetBonusCount() int {
	return len(s.bonusCache)
}

// GetCurrentTimeOfDay returns the current time of day from the lighting system.
func (s *TimeOfDayCompanionBonusSystem) GetCurrentTimeOfDay() palette.TimeOfDay {
	if s.lightingSystem == nil {
		return palette.TimeOfDayDay
	}
	return s.lightingSystem.GetCurrentTimeOfDay()
}

// GetLastTimeOfDay returns the last processed time of day.
func (s *TimeOfDayCompanionBonusSystem) GetLastTimeOfDay() palette.TimeOfDay {
	return s.lastTimeOfDay
}
