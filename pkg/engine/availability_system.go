// Package engine provides the availability system for time-based service access.
// AvailabilitySystem checks operating hours before allowing interactions with
// shops, NPCs, and services, creating a living world with realistic schedules.
package engine

import (
	"github.com/sirupsen/logrus"
)

// AvailabilitySystem manages service availability based on operating hours.
// It integrates with the game clock to determine if entities are available
// for interaction at the current game time.
type AvailabilitySystem struct {
	world  *World
	clock  GameClock
	logger *logrus.Entry
}

// NewAvailabilitySystem creates a new availability system.
func NewAvailabilitySystem(world *World, clock GameClock) *AvailabilitySystem {
	logrus.WithFields(logrus.Fields{
		"system_name": "availability",
	}).Debug("Creating availability system")

	var logger *logrus.Entry
	if world != nil {
		logger = world.GetLogger()
	}

	return &AvailabilitySystem{
		world:  world,
		clock:  clock,
		logger: logger,
	}
}

// Update processes availability state for all entities with operating hours.
// This can be used to trigger state changes when shops open/close.
func (s *AvailabilitySystem) Update(entities []*Entity, deltaTime float64) {
	if s.clock == nil {
		return
	}

	currentTime := s.clock.Now()
	currentHour := currentTime.Hour()
	currentDay := DayOfWeek(currentTime.Weekday())

	for _, entity := range entities {
		if !entity.HasComponent("operatingHours") {
			continue
		}

		hoursComp, ok := entity.GetComponent("operatingHours")
		if !ok || hoursComp == nil {
			continue
		}
		hours := hoursComp.(*OperatingHoursComponent)

		isOpen := hours.IsOpenAt(currentHour, currentDay)

		// Update any dependent state (e.g., dialog availability)
		s.updateEntityAvailability(entity, hours, isOpen)
	}
}

// updateEntityAvailability updates entity state based on availability.
// Disables shop dialogs when closed, updates visual state, etc.
func (s *AvailabilitySystem) updateEntityAvailability(entity *Entity, hours *OperatingHoursComponent, isOpen bool) {
	// Update dialog options if entity has a dialog component
	if dialogComp, ok := entity.GetComponent("dialog"); ok {
		if dialog, ok := dialogComp.(*DialogComponent); ok {
			s.updateDialogAvailability(dialog, hours, isOpen)
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"system_name":  "availability",
			"entity_id":    entity.ID,
			"service_type": hours.ServiceType,
			"is_open":      isOpen,
		}).Trace("Updated entity availability")
	}
}

// updateDialogAvailability modifies dialog options based on operating status.
func (s *AvailabilitySystem) updateDialogAvailability(dialog *DialogComponent, hours *OperatingHoursComponent, isOpen bool) {
	// If using MerchantDialogProvider, update shop option availability
	for i := range dialog.Options {
		if dialog.Options[i].Action == ActionOpenShop {
			dialog.Options[i].Enabled = isOpen
			if !isOpen {
				dialog.Options[i].Text = "Shop is closed"
			} else {
				dialog.Options[i].Text = "Browse your wares"
			}
		}
	}
}

// IsEntityAvailable checks if an entity is currently available for interaction.
// Returns true if entity has no operating hours (always available) or is within hours.
func (s *AvailabilitySystem) IsEntityAvailable(entityID uint64) bool {
	if s.clock == nil || s.world == nil {
		return true // Default to available if system not configured
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return false
	}

	return s.IsEntityOpen(entity)
}

// IsEntityOpen checks if an entity with operating hours is currently open.
// Entities without operating hours are considered always open.
func (s *AvailabilitySystem) IsEntityOpen(entity *Entity) bool {
	if s.clock == nil {
		return true
	}

	hoursComp, ok := entity.GetComponent("operatingHours")
	if !ok {
		return true // No operating hours means always available
	}

	hours, ok := hoursComp.(*OperatingHoursComponent)
	if !ok {
		return true
	}

	currentTime := s.clock.Now()
	currentHour := currentTime.Hour()
	currentDay := DayOfWeek(currentTime.Weekday())

	return hours.IsOpenAt(currentHour, currentDay)
}

// GetEntityStatus returns the availability status message for an entity.
// Returns "Open", "Closed", or custom closed message.
func (s *AvailabilitySystem) GetEntityStatus(entity *Entity) string {
	hoursComp, ok := entity.GetComponent("operatingHours")
	if !ok {
		return "Open"
	}

	hours, ok := hoursComp.(*OperatingHoursComponent)
	if !ok {
		return "Open"
	}

	isOpen := s.IsEntityOpen(entity)
	return hours.GetStatusMessage(isOpen)
}

// GetNextOpenTime returns when the entity will next be available.
// Returns hour, day, and whether the entity has operating hours.
func (s *AvailabilitySystem) GetNextOpenTime(entity *Entity) (hour int, day DayOfWeek, hasHours bool) {
	if s.clock == nil {
		return 0, Sunday, false
	}

	hoursComp, ok := entity.GetComponent("operatingHours")
	if !ok {
		return 0, Sunday, false
	}

	hours, ok := hoursComp.(*OperatingHoursComponent)
	if !ok {
		return 0, Sunday, false
	}

	currentTime := s.clock.Now()
	currentHour := currentTime.Hour()
	currentDay := DayOfWeek(currentTime.Weekday())

	nextHour, nextDay := hours.GetNextOpenTime(currentHour, currentDay)
	return nextHour, nextDay, true
}

// CheckInteractionAllowed validates if interaction with an entity is allowed.
// Returns allowed status and a message explaining why if not allowed.
func (s *AvailabilitySystem) CheckInteractionAllowed(entityID uint64) (bool, string) {
	if s.world == nil {
		return true, ""
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok {
		return false, "Entity not found"
	}

	hoursComp, ok := entity.GetComponent("operatingHours")
	if !ok {
		return true, "" // No operating hours = always available
	}

	hours, ok := hoursComp.(*OperatingHoursComponent)
	if !ok {
		return true, ""
	}

	if s.IsEntityOpen(entity) {
		return true, ""
	}

	// Return closed message
	msg := hours.GetStatusMessage(false)
	if nextHour, nextDay, hasHours := s.GetNextOpenTime(entity); hasHours && nextHour >= 0 {
		msg = msg + " Opens at " + formatTime(nextHour) + " on " + nextDay.String()
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"system_name":  "availability",
			"entity_id":    entityID,
			"service_type": hours.ServiceType,
			"message":      msg,
		}).Debug("Interaction blocked - service closed")
	}

	return false, msg
}

// formatTime formats an hour as a 12-hour time string.
func formatTime(hour int) string {
	if hour == 0 {
		return "12:00 AM"
	}
	if hour == 12 {
		return "12:00 PM"
	}
	if hour < 12 {
		return string(rune('0'+hour/10)) + string(rune('0'+hour%10)) + ":00 AM"
	}
	hour -= 12
	return string(rune('0'+hour/10)) + string(rune('0'+hour%10)) + ":00 PM"
}
