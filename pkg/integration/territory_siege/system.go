package territory_siege

import (
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/opd-ai/venture/pkg/world"
	"github.com/sirupsen/logrus"
)

// System wraps SiegeManager as an ECS system for integration into the game world.
type System struct {
	world   *engine.World
	manager *SiegeManager
	logger  *logrus.Entry
}

// NewSystem creates a new territory siege system.
func NewSystem(world *engine.World, tm *world.TerritoryManager, ps *engine.PoliticsSystem, gm *guild.Manager) *System {
	logger := logrus.WithField("system", "territory_siege")
	return &System{
		world:   world,
		manager: NewSiegeManager(world, tm, ps, gm),
		logger:  logger,
	}
}

// Update processes all active sieges and entities with siege participant components.
func (s *System) Update(entities []*engine.Entity, deltaTime float64) {
	// Update all active sieges (phase transitions, victory checks)
	s.manager.Update(deltaTime)

	// Process entities involved in sieges
	for _, entity := range entities {
		if !entity.HasComponent("siege_participant") {
			continue
		}

		comp, ok := entity.GetComponent("siege_participant")
		if !ok {
			continue
		}

		siegeComp, ok := comp.(*SiegeParticipantComponent)
		if !ok {
			s.logger.WithFields(logrus.Fields{
				"entityID":       entity.ID,
				"component_type": "siege_participant",
			}).Warn("Component type assertion failed")
			continue
		}

		// Update participant status in active siege
		if siegeComp.IsActive {
			s.updateParticipant(entity, siegeComp)
		}
	}
}

// updateParticipant updates a siege participant's status and interactions.
func (s *System) updateParticipant(entity *engine.Entity, comp *SiegeParticipantComponent) {
	siege, err := s.manager.GetActiveSiege(comp.SiegeID)
	if err != nil {
		// Siege no longer active
		comp.IsActive = false
		return
	}

	// Track last seen time for attacker elimination check
	comp.LastSeenTime = siege.LastUpdate

	// Apply siege phase effects
	switch siege.CurrentPhase {
	case PhasePreparation:
		// During preparation, participants can call reinforcements
		// (handled by UI/command system, not automatic)

	case PhaseAssault:
		// During assault, participants can damage structures
		// (handled by combat system when attacks hit structures)

	case PhaseResolution:
		// Siege ended, deactivate participant
		comp.IsActive = false
	}
}

// GetManager returns the underlying siege manager for external system access.
func (s *System) GetManager() *SiegeManager {
	return s.manager
}
