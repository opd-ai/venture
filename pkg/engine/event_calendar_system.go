// Package engine provides the event calendar system for seasonal events.
// EventCalendarSystem manages the lifecycle of seasonal events, transitioning
// them between upcoming, active, and ending phases based on game time.
package engine

import (
	"time"

	"github.com/sirupsen/logrus"
)

// EventCalendarSystem manages the lifecycle of seasonal events.
// It processes the world entity's SeasonalEventComponent to trigger,
// update, and end events based on the current game or real time.
type EventCalendarSystem struct {
	world *World
	clock GameClock
}

// NewEventCalendarSystem creates a new event calendar system.
func NewEventCalendarSystem(world *World, clock GameClock) *EventCalendarSystem {
	logrus.WithFields(logrus.Fields{
		"system_name": "event_calendar",
	}).Debug("Creating event calendar system")

	return &EventCalendarSystem{
		world: world,
		clock: clock,
	}
}

// Update processes all entities with seasonal event components.
// It transitions events between phases and activates/deactivates events.
func (s *EventCalendarSystem) Update(entities []*Entity, deltaTime float64) {
	if s.clock == nil {
		return
	}

	currentTime := s.clock.Now()

	for _, entity := range entities {
		if !entity.HasComponent("seasonal_event") {
			continue
		}

		comp, ok := entity.GetComponent("seasonal_event")
		if !ok || comp == nil {
			continue
		}
		eventComp := comp.(*SeasonalEventComponent)

		// Update current time based on mode
		if eventComp.UseRealTime {
			eventComp.CurrentTime = time.Now()
		} else {
			eventComp.CurrentTime = currentTime
		}

		// Process event lifecycle
		s.updateEventInstances(eventComp)
		s.checkForNewEvents(eventComp)
		s.cleanupEndedEvents(eventComp)

		eventComp.LastUpdateTime = eventComp.CurrentTime
	}
}

// updateEventInstances updates the phase of all tracked event instances.
func (s *EventCalendarSystem) updateEventInstances(comp *SeasonalEventComponent) {
	now := comp.CurrentTime
	const upcomingThreshold = 24 * time.Hour
	const endingThreshold = 24 * time.Hour

	for i := range comp.ActiveEvents {
		event := &comp.ActiveEvents[i]

		// Determine new phase based on current time
		switch {
		case now.Before(event.StartTime):
			// Event hasn't started yet
			if event.StartTime.Sub(now) <= upcomingThreshold {
				event.Phase = EventPhaseUpcoming
			}
		case now.After(event.EndTime):
			// Event has ended - will be cleaned up later
			continue
		case event.EndTime.Sub(now) <= endingThreshold:
			// Event is ending soon
			if event.Phase != EventPhaseEnding {
				logrus.WithFields(logrus.Fields{
					"system_name": "event_calendar",
					"event_id":    event.Definition.ID,
					"event_name":  event.Definition.Name,
					"ends_in":     event.EndTime.Sub(now).String(),
				}).Info("Event entering ending phase")
			}
			event.Phase = EventPhaseEnding
		default:
			// Event is active
			if event.Phase != EventPhaseActive {
				logrus.WithFields(logrus.Fields{
					"system_name": "event_calendar",
					"event_id":    event.Definition.ID,
					"event_name":  event.Definition.Name,
					"theme":       event.Definition.Theme,
				}).Info("Event is now active")
			}
			event.Phase = EventPhaseActive
		}
	}
}

// checkForNewEvents looks for calendar events that should be activated.
func (s *EventCalendarSystem) checkForNewEvents(comp *SeasonalEventComponent) {
	now := comp.CurrentTime
	dayOfYear := now.YearDay()
	year := now.Year()

	for _, def := range comp.EventCalendar {
		// Check if event is already tracked
		if s.isEventTracked(comp, def.ID, year) {
			continue
		}

		// Calculate event start and end times for this year
		startTime := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, def.StartDayOfYear-1)
		endTime := startTime.AddDate(0, 0, def.DurationDays)

		// Handle year-end wrapping (e.g., event starts Dec 25, ends Jan 5)
		if def.StartDayOfYear+def.DurationDays > 365 {
			endTime = time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, (def.StartDayOfYear+def.DurationDays)-365-1)
		}

		// Check if we should track this event (within 7 days of start)
		const trackingWindow = 7 * 24 * time.Hour
		if now.After(startTime.Add(-trackingWindow)) && now.Before(endTime) {
			// Add to active events
			instance := EventInstance{
				Definition: def,
				StartTime:  startTime,
				EndTime:    endTime,
				Phase:      EventPhaseUpcoming,
			}

			// Determine initial phase
			if dayOfYear >= def.StartDayOfYear && dayOfYear < def.StartDayOfYear+def.DurationDays {
				instance.Phase = EventPhaseActive
			}

			comp.ActiveEvents = append(comp.ActiveEvents, instance)

			logrus.WithFields(logrus.Fields{
				"system_name":   "event_calendar",
				"event_id":      def.ID,
				"event_name":    def.Name,
				"start_time":    startTime.Format(time.RFC3339),
				"end_time":      endTime.Format(time.RFC3339),
				"initial_phase": instance.Phase,
			}).Info("Tracking new seasonal event")
		}
	}
}

// isEventTracked checks if an event is already being tracked for the given year.
func (s *EventCalendarSystem) isEventTracked(comp *SeasonalEventComponent, eventID string, year int) bool {
	for _, event := range comp.ActiveEvents {
		if event.Definition.ID == eventID && event.StartTime.Year() == year {
			return true
		}
	}
	return false
}

// cleanupEndedEvents removes events that have completely ended.
func (s *EventCalendarSystem) cleanupEndedEvents(comp *SeasonalEventComponent) {
	now := comp.CurrentTime
	const gracePeriod = 1 * time.Hour // Keep event data for 1 hour after end

	activeEvents := make([]EventInstance, 0, len(comp.ActiveEvents))
	for _, event := range comp.ActiveEvents {
		if now.After(event.EndTime.Add(gracePeriod)) {
			logrus.WithFields(logrus.Fields{
				"system_name": "event_calendar",
				"event_id":    event.Definition.ID,
				"event_name":  event.Definition.Name,
			}).Info("Removing ended seasonal event")
			continue
		}
		activeEvents = append(activeEvents, event)
	}
	comp.ActiveEvents = activeEvents
}

// GetCurrentDayOfYear returns the day of year from the clock.
func (s *EventCalendarSystem) GetCurrentDayOfYear() int {
	if s.clock == nil {
		return 1
	}
	return s.clock.Now().YearDay()
}

// TriggerEvent manually activates an event (for testing or special occasions).
func (s *EventCalendarSystem) TriggerEvent(entity *Entity, eventID string, durationDays int) error {
	comp, ok := entity.GetComponent("seasonal_event")
	if !ok || comp == nil {
		return nil
	}
	eventComp := comp.(*SeasonalEventComponent)

	now := eventComp.CurrentTime
	startTime := now
	endTime := now.AddDate(0, 0, durationDays)

	// Find the event definition
	var def *EventDefinition
	for i := range eventComp.EventCalendar {
		if eventComp.EventCalendar[i].ID == eventID {
			def = &eventComp.EventCalendar[i]
			break
		}
	}

	if def == nil {
		// Create an ad-hoc event
		def = &EventDefinition{
			ID:           eventID,
			Name:         eventID,
			Theme:        EventThemeSpring,
			DurationDays: durationDays,
			Seed:         now.UnixNano(),
		}
	}

	instance := EventInstance{
		Definition: *def,
		StartTime:  startTime,
		EndTime:    endTime,
		Phase:      EventPhaseActive,
	}

	eventComp.ActiveEvents = append(eventComp.ActiveEvents, instance)

	logrus.WithFields(logrus.Fields{
		"system_name": "event_calendar",
		"event_id":    eventID,
		"duration":    durationDays,
	}).Info("Manually triggered seasonal event")

	return nil
}

// EndEvent manually ends an event (for testing or admin control).
func (s *EventCalendarSystem) EndEvent(entity *Entity, eventID string) error {
	comp, ok := entity.GetComponent("seasonal_event")
	if !ok || comp == nil {
		return nil
	}
	eventComp := comp.(*SeasonalEventComponent)

	for i := range eventComp.ActiveEvents {
		if eventComp.ActiveEvents[i].Definition.ID == eventID {
			// Set end time to now to trigger cleanup
			eventComp.ActiveEvents[i].EndTime = eventComp.CurrentTime.Add(-time.Hour)
			logrus.WithFields(logrus.Fields{
				"system_name": "event_calendar",
				"event_id":    eventID,
			}).Info("Manually ended seasonal event")
			return nil
		}
	}

	return nil
}
