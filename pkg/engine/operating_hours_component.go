// Package engine provides the operating hours component for time-based service availability.
// OperatingHoursComponent enables shops, services, and NPCs to have scheduled operating
// hours, creating a living world where not everything is available 24/7.
package engine

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// DayOfWeek represents a day of the week for scheduling.
type DayOfWeek int

const (
	Sunday DayOfWeek = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// String returns the day name.
func (d DayOfWeek) String() string {
	switch d {
	case Sunday:
		return "Sunday"
	case Monday:
		return "Monday"
	case Tuesday:
		return "Tuesday"
	case Wednesday:
		return "Wednesday"
	case Thursday:
		return "Thursday"
	case Friday:
		return "Friday"
	case Saturday:
		return "Saturday"
	default:
		return "Unknown"
	}
}

// OperatingHoursComponent tracks when a service/shop is available for interaction.
// Used by AvailabilitySystem to check if interactions should be allowed.
type OperatingHoursComponent struct {
	// OpenHour is when the service opens (0-23)
	OpenHour int `json:"open_hour"`
	// CloseHour is when the service closes (0-23)
	CloseHour int `json:"close_hour"`
	// DaysOpen indicates which days the service is open (indexed 0-6, Sunday-Saturday)
	DaysOpen [7]bool `json:"days_open"`
	// IsAlwaysOpen bypasses hour/day checks when true
	IsAlwaysOpen bool `json:"is_always_open"`
	// ClosedMessage is displayed when service is unavailable
	ClosedMessage string `json:"closed_message"`
	// ServiceType describes what kind of service this is (for logging/display)
	ServiceType string `json:"service_type"`
}

// NewOperatingHoursComponent creates an operating hours component with default business hours.
// Default: Open 8:00-18:00, Monday-Saturday.
func NewOperatingHoursComponent() *OperatingHoursComponent {
	return &OperatingHoursComponent{
		OpenHour:      8,
		CloseHour:     18,
		DaysOpen:      [7]bool{false, true, true, true, true, true, true}, // Mon-Sat
		IsAlwaysOpen:  false,
		ClosedMessage: "This service is currently closed.",
		ServiceType:   "shop",
	}
}

// NewAlwaysOpenComponent creates an operating hours component that's always available.
// Use for services like inns, guard posts, or emergency services.
func NewAlwaysOpenComponent(serviceType string) *OperatingHoursComponent {
	return &OperatingHoursComponent{
		OpenHour:      0,
		CloseHour:     24,
		DaysOpen:      [7]bool{true, true, true, true, true, true, true},
		IsAlwaysOpen:  true,
		ClosedMessage: "",
		ServiceType:   serviceType,
	}
}

// Type returns the component type identifier.
func (o *OperatingHoursComponent) Type() string {
	return "operatingHours"
}

// SetHours sets the open and close hours.
// Hours use 24-hour format (0-23).
func (o *OperatingHoursComponent) SetHours(openHour, closeHour int) {
	o.OpenHour = clampHour(openHour)
	o.CloseHour = clampHour(closeHour)
}

// SetDaysOpen sets which days the service is open.
// days is a 7-element slice indexed 0-6 (Sunday-Saturday).
func (o *OperatingHoursComponent) SetDaysOpen(days [7]bool) {
	o.DaysOpen = days
}

// SetOpenOnDay sets whether the service is open on a specific day.
func (o *OperatingHoursComponent) SetOpenOnDay(day DayOfWeek, isOpen bool) {
	if day >= 0 && day <= 6 {
		o.DaysOpen[day] = isOpen
	}
}

// IsOpenAt checks if the service is open at the given hour and day.
// Returns true if service is available at that time.
func (o *OperatingHoursComponent) IsOpenAt(hour int, day DayOfWeek) bool {
	if o.IsAlwaysOpen {
		return true
	}

	// Check day first
	if day < 0 || day > 6 || !o.DaysOpen[day] {
		return false
	}

	// Check hours (handle overnight hours like 22:00-06:00)
	return o.isWithinHours(hour)
}

// isWithinHours checks if the hour is within operating hours.
// Handles both normal hours (8-18) and overnight hours (22-6).
func (o *OperatingHoursComponent) isWithinHours(hour int) bool {
	hour = clampHour(hour)

	if o.OpenHour <= o.CloseHour {
		// Normal hours (e.g., 8:00 - 18:00)
		return hour >= o.OpenHour && hour < o.CloseHour
	}
	// Overnight hours (e.g., 22:00 - 06:00)
	return hour >= o.OpenHour || hour < o.CloseHour
}

// GetNextOpenTime returns the next hour when the service will be open.
// Returns -1 if service is always open or never opens.
func (o *OperatingHoursComponent) GetNextOpenTime(currentHour int, currentDay DayOfWeek) (int, DayOfWeek) {
	if o.IsAlwaysOpen {
		return currentHour, currentDay
	}

	// Check current day first
	if o.DaysOpen[currentDay] {
		if currentHour < o.OpenHour {
			return o.OpenHour, currentDay
		}
		if o.OpenHour > o.CloseHour && currentHour >= o.OpenHour {
			// Currently within overnight hours
			return currentHour, currentDay
		}
	}

	// Search upcoming days
	for i := 1; i <= 7; i++ {
		nextDay := DayOfWeek((int(currentDay) + i) % 7)
		if o.DaysOpen[nextDay] {
			return o.OpenHour, nextDay
		}
	}

	// Never opens (all days closed)
	return -1, currentDay
}

// GetStatusMessage returns a message describing availability status.
func (o *OperatingHoursComponent) GetStatusMessage(isOpen bool) string {
	if isOpen {
		return "Open"
	}
	if o.ClosedMessage != "" {
		return o.ClosedMessage
	}
	return "Closed"
}

// Serialize encodes the component to bytes for persistence.
func (o *OperatingHoursComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type": "operatingHours",
		"service_type":   o.ServiceType,
	}).Debug("Serializing operating hours component")

	data, err := json.Marshal(o)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "operatingHours",
			"error":          err.Error(),
		}).Error("Failed to serialize operating hours component")
		return nil, err
	}
	return data, nil
}

// Deserialize decodes the component from bytes.
func (o *OperatingHoursComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "operatingHours",
		"bytes":          len(data),
	}).Debug("Deserializing operating hours component")

	if err := json.Unmarshal(data, o); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "operatingHours",
			"error":          err.Error(),
		}).Error("Failed to deserialize operating hours component")
		return err
	}
	return nil
}

// clampHour ensures hour is within valid range 0-23.
func clampHour(hour int) int {
	if hour < 0 {
		return 0
	}
	if hour > 23 {
		return 23
	}
	return hour
}
