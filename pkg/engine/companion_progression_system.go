package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// CompanionProgressionSystem handles companion leveling and stat scaling
type CompanionProgressionSystem struct {
	world            *World
	levelUpCallbacks []CompanionLevelUpCallback
	logger           *logrus.Entry
}

// NewCompanionProgressionSystem creates a new companion progression system
func NewCompanionProgressionSystem(world *World) *CompanionProgressionSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "companion_progression")
	}
	return &CompanionProgressionSystem{
		world:            world,
		levelUpCallbacks: make([]CompanionLevelUpCallback, 0),
		logger:           logEntry,
	}
}

// AddLevelUpCallback registers a callback for companion level-up events.
func (s *CompanionProgressionSystem) AddLevelUpCallback(callback CompanionLevelUpCallback) {
	if callback != nil {
		s.levelUpCallbacks = append(s.levelUpCallbacks, callback)
	}
}

// Update processes companion XP gain and leveling
func (s *CompanionProgressionSystem) Update(deltaTime float64) {
	entities := s.world.GetEntitiesWith("companion", "companionstats")

	for _, entity := range entities {
		companionCompRaw, ok := entity.GetComponent("companion")
		if !ok {
			continue
		}
		companionComp := companionCompRaw.(*CompanionComponent)

		statsCompRaw, ok := entity.GetComponent("companionstats")
		if !ok {
			continue
		}
		statsComp := statsCompRaw.(*CompanionStatsComponent)

		// Check for level up
		xpNeeded := s.calculateXPForLevel(companionComp.Level + 1)
		if companionComp.Experience >= xpNeeded {
			companionComp.Level++
			companionComp.Experience -= xpNeeded
			s.applyLevelUpBonus(companionComp, statsComp)
			s.notifyLevelUp(entity, companionComp.Level)
		}
	}
}

// notifyLevelUp calls all registered level-up callbacks.
func (s *CompanionProgressionSystem) notifyLevelUp(entity *Entity, newLevel int) {
	for _, callback := range s.levelUpCallbacks {
		callback(entity, newLevel)
	}
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"new_level": newLevel,
		}).Debug("companion leveled up")
	}
}

// GainExperience adds experience to a companion
func (s *CompanionProgressionSystem) GainExperience(companionID uint64, amount float64) {
	entity, ok := s.world.GetEntity(companionID)
	if !ok || entity == nil {
		return
	}

	companionCompRaw, ok := entity.GetComponent("companion")
	if !ok {
		return
	}

	comp := companionCompRaw.(*CompanionComponent)
	comp.Experience += amount
}

func (s *CompanionProgressionSystem) calculateXPForLevel(level int) float64 {
	return 100.0 * math.Pow(1.5, float64(level-1))
}

func (s *CompanionProgressionSystem) applyLevelUpBonus(companion *CompanionComponent, stats *CompanionStatsComponent) {
	// Scale stats based on level
	stats.Attack *= 1.15
	stats.Defense *= 1.15
	stats.MaxHP *= 1.2
	stats.HP = stats.MaxHP
	stats.Speed *= 1.05

	// Increase loyalty
	companion.Loyalty = math.Min(100.0, companion.Loyalty+5.0)
}
