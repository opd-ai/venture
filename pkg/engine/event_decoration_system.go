// Package engine provides the event decoration system for seasonal events.
// EventDecorationSystem applies and removes decorations from entities based on
// active seasonal events in the world, integrating with the Event Calendar.
package engine

import (
	"github.com/sirupsen/logrus"
)

// DecorationTransitionSpeed is the rate at which decorations appear/disappear.
// Value represents progress per second (1.0 = full transition in 1 second).
const DecorationTransitionSpeed = 0.5

// EventDecorationSystem manages entity decorations based on seasonal events.
// It monitors SeasonalEventComponent to trigger decoration changes and
// handles smooth transitions when events start or end.
type EventDecorationSystem struct {
	world *World
}

// NewEventDecorationSystem creates a new event decoration system.
func NewEventDecorationSystem(world *World) *EventDecorationSystem {
	logrus.WithFields(logrus.Fields{
		"system_name": "event_decoration",
	}).Debug("Creating event decoration system")

	return &EventDecorationSystem{
		world: world,
	}
}

// Update processes all entities with event decoration components.
// It applies decorations when events start and removes them when events end.
func (s *EventDecorationSystem) Update(entities []*Entity, deltaTime float64) {
	// Find the world entity with seasonal events
	var eventComp *SeasonalEventComponent
	for _, entity := range entities {
		if comp, ok := entity.GetComponent("seasonal_event"); ok {
			eventComp = comp.(*SeasonalEventComponent)
			break
		}
	}

	// Get active event info
	var activeEvent *EventInstance
	if eventComp != nil {
		activeEvents := eventComp.GetActiveEvents()
		if len(activeEvents) > 0 {
			activeEvent = &activeEvents[0] // Use first active event
		}
	}

	// Process all decoratable entities
	for _, entity := range entities {
		if !entity.HasComponent("event_decoration") {
			continue
		}

		comp, ok := entity.GetComponent("event_decoration")
		if !ok || comp == nil {
			continue
		}
		decoComp := comp.(*EventDecorationComponent)

		s.updateDecoration(entity, decoComp, activeEvent, deltaTime)
	}
}

// updateDecoration updates a single entity's decoration state.
func (s *EventDecorationSystem) updateDecoration(entity *Entity, decoComp *EventDecorationComponent, activeEvent *EventInstance, deltaTime float64) {
	if activeEvent != nil {
		s.handleActiveEvent(entity, decoComp, activeEvent, deltaTime)
	} else {
		s.handleNoEvent(entity, decoComp, deltaTime)
	}

	// Process transition animation
	if decoComp.IsTransitioning {
		s.processTransition(decoComp, deltaTime)
	}

	// Update city visuals if entity has both components
	s.updateCityVisuals(entity, decoComp)
}

// handleActiveEvent processes decoration when an event is active.
func (s *EventDecorationSystem) handleActiveEvent(entity *Entity, decoComp *EventDecorationComponent, activeEvent *EventInstance, deltaTime float64) {
	targetTheme := ThemeFromEventTheme(activeEvent.Definition.Theme)

	// Check if we need to apply new decorations
	if decoComp.EventID != activeEvent.Definition.ID {
		s.startApplyingDecorations(entity, decoComp, activeEvent, targetTheme)
	}

	// Calculate decoration level based on event phase
	targetLevel := s.calculateDecorationLevel(activeEvent)
	if decoComp.DecorationLevel != targetLevel && !decoComp.IsTransitioning {
		decoComp.DecorationLevel = targetLevel
	}
}

// handleNoEvent processes decoration when no event is active.
func (s *EventDecorationSystem) handleNoEvent(entity *Entity, decoComp *EventDecorationComponent, deltaTime float64) {
	// Start removing decorations if any are present
	if decoComp.HasDecorations() && !decoComp.IsTransitioning {
		s.startRemovingDecorations(decoComp)
	}
}

// startApplyingDecorations initiates decoration application for a new event.
func (s *EventDecorationSystem) startApplyingDecorations(entity *Entity, decoComp *EventDecorationComponent, event *EventInstance, theme DecorationTheme) {
	// Clear any existing decorations first
	if decoComp.HasDecorations() {
		decoComp.ClearDecorations()
	}

	// Set up for new event
	decoComp.EventID = event.Definition.ID
	decoComp.GenerateDecorations(theme, 1.0)

	// Start apply transition
	decoComp.IsTransitioning = true
	decoComp.TransitionDirection = 1
	decoComp.TransitionProgress = 0.0

	logrus.WithFields(logrus.Fields{
		"system_name": "event_decoration",
		"entity_id":   entity.ID,
		"event_id":    event.Definition.ID,
		"theme":       theme,
	}).Info("Starting decoration application")
}

// startRemovingDecorations initiates decoration removal.
func (s *EventDecorationSystem) startRemovingDecorations(decoComp *EventDecorationComponent) {
	decoComp.IsTransitioning = true
	decoComp.TransitionDirection = -1

	logrus.WithFields(logrus.Fields{
		"system_name": "event_decoration",
		"event_id":    decoComp.EventID,
		"theme":       decoComp.ActiveTheme,
	}).Info("Starting decoration removal")
}

// processTransition advances the decoration transition animation.
func (s *EventDecorationSystem) processTransition(decoComp *EventDecorationComponent, deltaTime float64) {
	decoComp.TransitionProgress += float64(decoComp.TransitionDirection) * DecorationTransitionSpeed * deltaTime

	// Check for transition completion
	if decoComp.TransitionDirection > 0 && decoComp.TransitionProgress >= 1.0 {
		// Finished applying
		decoComp.TransitionProgress = 1.0
		decoComp.IsTransitioning = false
		decoComp.TransitionDirection = 0
	} else if decoComp.TransitionDirection < 0 && decoComp.TransitionProgress <= 0.0 {
		// Finished removing
		decoComp.ClearDecorations()
	}
}

// calculateDecorationLevel determines target decoration level from event phase.
func (s *EventDecorationSystem) calculateDecorationLevel(event *EventInstance) float64 {
	switch event.Phase {
	case EventPhaseUpcoming:
		return 0.3 // Light decorations during build-up
	case EventPhaseActive:
		return 1.0 // Full decorations during event
	case EventPhaseEnding:
		return 0.6 // Reduced decorations as event winds down
	default:
		return 0.0
	}
}

// updateCityVisuals applies decoration effects to city visual components.
func (s *EventDecorationSystem) updateCityVisuals(entity *Entity, decoComp *EventDecorationComponent) {
	cityVisual, ok := entity.GetComponent("city_visual")
	if !ok || cityVisual == nil {
		return
	}

	cv := cityVisual.(*CityVisualComponent)

	// Boost decoration density during events
	if decoComp.HasDecorations() {
		baseDecoration := cv.DecorationDensity
		eventBoost := decoComp.DecorationLevel * decoComp.TransitionProgress * 0.3
		cv.DecorationDensity = clampFloat(baseDecoration+eventBoost, 0.0, 1.0)

		// Boost lighting during winter events
		if decoComp.ActiveTheme == DecorationThemeWinter {
			cv.LightingLevel = clampFloat(cv.LightingLevel+0.15, 0.0, 1.0)
		}

		// Boost market activity during harvest
		if decoComp.ActiveTheme == DecorationThemeAutumn {
			cv.MarketActivity = clampFloat(cv.MarketActivity+0.2, 0.0, 1.0)
		}
	}
}

// ApplyDecorationsToEntity manually applies decorations to a specific entity.
// This is useful for testing or special cases where automatic detection isn't used.
func (s *EventDecorationSystem) ApplyDecorationsToEntity(entity *Entity, theme DecorationTheme, level float64) error {
	decoComp, ok := entity.GetComponent("event_decoration")
	if !ok || decoComp == nil {
		return nil
	}

	dc := decoComp.(*EventDecorationComponent)
	dc.GenerateDecorations(theme, level)
	dc.TransitionProgress = 1.0
	dc.IsTransitioning = false

	logrus.WithFields(logrus.Fields{
		"system_name": "event_decoration",
		"entity_id":   entity.ID,
		"theme":       theme,
		"level":       level,
	}).Info("Manually applied decorations")

	return nil
}

// RemoveDecorationsFromEntity manually removes decorations from a specific entity.
func (s *EventDecorationSystem) RemoveDecorationsFromEntity(entity *Entity) error {
	decoComp, ok := entity.GetComponent("event_decoration")
	if !ok || decoComp == nil {
		return nil
	}

	dc := decoComp.(*EventDecorationComponent)
	dc.ClearDecorations()

	logrus.WithFields(logrus.Fields{
		"system_name": "event_decoration",
		"entity_id":   entity.ID,
	}).Info("Manually removed decorations")

	return nil
}

// GetDecoratedEntities returns all entities with active decorations.
func (s *EventDecorationSystem) GetDecoratedEntities(entities []*Entity) []*Entity {
	result := make([]*Entity, 0)
	for _, entity := range entities {
		decoComp, ok := entity.GetComponent("event_decoration")
		if !ok || decoComp == nil {
			continue
		}
		dc := decoComp.(*EventDecorationComponent)
		if dc.HasDecorations() {
			result = append(result, entity)
		}
	}
	return result
}

// GetDecorationStats returns statistics about current decorations.
func (s *EventDecorationSystem) GetDecorationStats(entities []*Entity) DecorationStats {
	stats := DecorationStats{
		ThemeCounts: make(map[DecorationTheme]int),
	}

	for _, entity := range entities {
		decoComp, ok := entity.GetComponent("event_decoration")
		if !ok || decoComp == nil {
			continue
		}
		dc := decoComp.(*EventDecorationComponent)

		if dc.HasDecorations() {
			stats.DecoratedEntities++
			stats.ThemeCounts[dc.ActiveTheme]++
			stats.TotalElements += len(dc.Elements)
			stats.TotalParticleEffects += len(dc.ParticleEffects)
		}

		if dc.IsTransitioning {
			stats.TransitioningEntities++
		}
	}

	return stats
}

// DecorationStats contains aggregate decoration information.
type DecorationStats struct {
	// DecoratedEntities is the count of entities with active decorations
	DecoratedEntities int
	// TransitioningEntities is the count of entities currently transitioning
	TransitioningEntities int
	// ThemeCounts maps themes to entity counts
	ThemeCounts map[DecorationTheme]int
	// TotalElements is the total number of decoration elements
	TotalElements int
	// TotalParticleEffects is the total number of active particle effects
	TotalParticleEffects int
}
