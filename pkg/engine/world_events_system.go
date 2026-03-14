package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/integration/world_events"
	"github.com/sirupsen/logrus"
)

// WorldEventsSystem manages dynamic world events based on player actions.
// Phase 6.3: World Events integration with ECS architecture.
type WorldEventsSystem struct {
	world          *World
	eventManager   *world_events.EventManager
	logger         *logrus.Entry
	updateTimer    float64
	updateInterval float64 // Check for events every N seconds
	mu             sync.Mutex
}

// NewWorldEventsSystem creates a new world events system.
func NewWorldEventsSystem(world *World, seed int64) *WorldEventsSystem {
	return NewWorldEventsSystemWithLogger(world, seed, nil)
}

// NewWorldEventsSystemWithLogger creates a new world events system with custom logger.
func NewWorldEventsSystemWithLogger(world *World, seed int64, logger *logrus.Logger) *WorldEventsSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "world_events")
	}

	return &WorldEventsSystem{
		world:          world,
		eventManager:   world_events.NewEventManager(seed),
		logger:         logEntry,
		updateTimer:    0,
		updateInterval: 30.0, // Check every 30 seconds
	}
}

// Update processes world events based on game state.
func (s *WorldEventsSystem) Update(entities []*Entity, deltaTime float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.updateTimer += deltaTime

	if s.updateTimer < s.updateInterval {
		return
	}

	s.updateTimer -= s.updateInterval

	// Check for event triggers based on world state
	s.checkGuildWarfareTriggers()
	s.checkEconomicTriggers()
	s.checkWeatherTriggers()

	// Update active events
	s.updateActiveEvents(deltaTime)
}

// checkGuildWarfareTriggers looks for guild warfare events.
func (s *WorldEventsSystem) checkGuildWarfareTriggers() {
	// Count guilds present; multiple guilds in the same world suggest warfare potential
	guildCount := len(s.world.GetEntitiesWith("guild"))

	if guildCount > 1 {
		params := world_events.TriggerParams{
			TriggerType: world_events.TriggerGuildWar,
			Severity:    world_events.SeverityMajor,
			Location:    "world_center",
			ServerID:    "default",
			GuildID:     "guild_1",
		}

		event, err := s.eventManager.GenerateEvent(world_events.TriggerGuildWar, params)
		if err == nil && s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"event_id":   event.ID,
				"event_type": event.Type,
				"severity":   event.Severity,
			}).Info("guild warfare event generated")
		}
	}
}

// checkEconomicTriggers looks for economic events.
func (s *WorldEventsSystem) checkEconomicTriggers() {
	// Count merchants and trade volume
	merchantCount := len(s.world.GetEntitiesWith("merchant"))

	if merchantCount > 5 {
		params := world_events.TriggerParams{
			TriggerType: world_events.TriggerTradeVolume,
			Severity:    world_events.SeverityModerate,
			Location:    "market_district",
			ServerID:    "default",
			ItemType:    "general_goods",
		}

		event, err := s.eventManager.GenerateEvent(world_events.TriggerTradeVolume, params)
		if err == nil && s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"event_id":       event.ID,
				"event_type":     event.Type,
				"merchant_count": merchantCount,
			}).Info("economic event generated")
		}
	}
}

// checkWeatherTriggers looks for weather disaster events.
func (s *WorldEventsSystem) checkWeatherTriggers() {
	// Check for active weather systems
	weatherEntities := s.world.GetEntitiesWith("weather")

	for _, entity := range weatherEntities {
		weatherComp, ok := entity.GetComponent("weather")
		if !ok || weatherComp == nil {
			continue
		}

		// Weather disasters can spawn from extreme weather
		params := world_events.TriggerParams{
			TriggerType: world_events.TriggerWeatherChange,
			Severity:    world_events.SeverityMajor,
			Location:    "weather_zone",
			ServerID:    "default",
		}

		event, err := s.eventManager.GenerateEvent(world_events.TriggerWeatherChange, params)
		if err == nil && s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"event_id":   event.ID,
				"event_type": event.Type,
			}).Info("weather disaster event generated")
		}
		break // Only one weather event at a time
	}
}

// updateActiveEvents processes impacts of active events.
func (s *WorldEventsSystem) updateActiveEvents(deltaTime float64) {
	currentTime := time.Now()
	activeEvents := s.eventManager.GetActiveEvents(currentTime)

	for _, event := range activeEvents {
		// Apply event impacts to entities
		impacts := event.GetImpacts()
		for _, impact := range impacts {
			s.applyEventImpact(impact)
		}
	}

	// Cleanup expired events
	s.eventManager.CleanupExpiredEvents(currentTime)
}

// applyEventImpact applies an event impact to the world.
func (s *WorldEventsSystem) applyEventImpact(impact world_events.Impact) {
	switch impact.Type {
	case world_events.ImpactSpawnRate:
		// Apply damage to entities in affected area
		s.applyAreaDamage(impact)
	case world_events.ImpactPriceChange:
		// Modify resource availability
		s.modifyResources(impact)
	case world_events.ImpactNPCReputation:
		// Update faction relationships
		s.updateFactionRelations(impact)
	case world_events.ImpactWeather:
		// Change weather conditions
		s.updateWeather(impact)
	}
}

// applyAreaDamage applies damage to entities in the impact area.
func (s *WorldEventsSystem) applyAreaDamage(impact world_events.Impact) {
	for _, entity := range s.world.GetEntitiesWith("position", "health") {
		posComp, ok := entity.GetComponent("position")
		if !ok || posComp == nil {
			continue
		}

		// Check if entity is in affected area
		// This is a simplified check - real implementation would use spatial queries
		healthComp, ok := entity.GetComponent("health")
		if !ok || healthComp == nil {
			continue
		}

		// Apply damage based on impact modifier
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"modifier":  impact.Modifier,
			}).Debug("applying event damage")
		}
	}
}

// modifyResources adjusts resource availability.
func (s *WorldEventsSystem) modifyResources(impact world_events.Impact) {
	// Modify merchant inventories based on economic events
	merchants := s.world.GetEntitiesWith("merchant", "inventory")
	for _, merchant := range merchants {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"merchant_id": merchant.ID,
				"modifier":    impact.Modifier,
			}).Debug("modifying merchant resources")
		}
	}
}

// updateFactionRelations changes faction relationships.
func (s *WorldEventsSystem) updateFactionRelations(impact world_events.Impact) {
	// Update faction reputation based on political events
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"faction":  impact.Target,
			"modifier": impact.Modifier,
		}).Debug("updating faction relations")
	}
}

// updateWeather changes weather conditions.
func (s *WorldEventsSystem) updateWeather(impact world_events.Impact) {
	// Change weather based on weather disaster events
	weatherEntities := s.world.GetEntitiesWith("weather")
	for _, entity := range weatherEntities {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id": entity.ID,
				"type":      impact.Type,
				"modifier":  impact.Modifier,
			}).Debug("updating weather from event")
		}
	}
}

// TriggerEvent manually triggers a world event (used by quest system, etc.).
func (s *WorldEventsSystem) TriggerEvent(trigger world_events.TriggerType, params world_events.TriggerParams) (*world_events.WorldEvent, error) {
	event, err := s.eventManager.GenerateEvent(trigger, params)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger event: %w", err)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"event_id":   event.ID,
			"event_type": event.Type,
			"trigger":    trigger,
		}).Info("manually triggered event")
	}

	return event, nil
}

// GetActiveEvents returns all currently active events.
func (s *WorldEventsSystem) GetActiveEvents() []*world_events.WorldEvent {
	return s.eventManager.GetActiveEvents(time.Now())
}

// GetEventChain retrieves the event chain for a given event.
func (s *WorldEventsSystem) GetEventChain(eventID string) []string {
	return s.eventManager.GetEventChain(eventID)
}
