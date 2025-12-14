// Package engine provides the seasonal event component for time-limited events.
// SeasonalEventComponent enables the world to host recurring seasonal festivals
// and holidays with procedurally generated content.
package engine

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// EventTheme defines the thematic category of a seasonal event.
type EventTheme string

const (
	// EventThemeSpring represents spring festivals with renewal themes.
	EventThemeSpring EventTheme = "spring"
	// EventThemeSummer represents summer festivals with abundance themes.
	EventThemeSummer EventTheme = "summer"
	// EventThemeAutumn represents autumn festivals with harvest themes.
	EventThemeAutumn EventTheme = "autumn"
	// EventThemeWinter represents winter festivals with celebration themes.
	EventThemeWinter EventTheme = "winter"
)

// EventPhase defines the lifecycle phase of a running event.
type EventPhase string

const (
	// EventPhaseUpcoming indicates an event will start soon (within 24 hours).
	EventPhaseUpcoming EventPhase = "upcoming"
	// EventPhaseActive indicates an event is currently running.
	EventPhaseActive EventPhase = "active"
	// EventPhaseEnding indicates an event will end soon (within 24 hours).
	EventPhaseEnding EventPhase = "ending"
)

// EventDefinition defines the properties of a seasonal event.
type EventDefinition struct {
	// ID is the unique identifier for this event
	ID string `json:"id"`
	// Name is the display name of the event
	Name string `json:"name"`
	// Theme is the thematic category (spring, summer, autumn, winter)
	Theme EventTheme `json:"theme"`
	// DurationDays is how long the event runs
	DurationDays int `json:"duration_days"`
	// StartDayOfYear is when the event starts (1-365)
	StartDayOfYear int `json:"start_day_of_year"`
	// Seed is used for deterministic content generation
	Seed int64 `json:"seed"`
	// Description is a short description of the event
	Description string `json:"description"`
}

// EventInstance represents a running or scheduled event instance.
type EventInstance struct {
	// Definition contains the event configuration
	Definition EventDefinition `json:"definition"`
	// StartTime is when this instance began/will begin
	StartTime time.Time `json:"start_time"`
	// EndTime is when this instance ends/will end
	EndTime time.Time `json:"end_time"`
	// Phase is the current lifecycle phase
	Phase EventPhase `json:"phase"`
}

// SeasonalEventComponent tracks seasonal events for the world.
// It maintains the event calendar and tracks active event instances.
type SeasonalEventComponent struct {
	// ActiveEvents contains currently active event instances
	ActiveEvents []EventInstance `json:"active_events"`
	// EventCalendar contains the yearly event schedule
	EventCalendar []EventDefinition `json:"event_calendar"`
	// CurrentTime is the current game or real time
	CurrentTime time.Time `json:"current_time"`
	// UseRealTime determines whether to use real-world time
	UseRealTime bool `json:"use_real_time"`
	// LastUpdateTime tracks when events were last checked
	LastUpdateTime time.Time `json:"last_update_time"`
	// CalendarSeed is the seed used to generate the calendar
	CalendarSeed int64 `json:"calendar_seed"`
}

// NewSeasonalEventComponent creates a new seasonal event component.
func NewSeasonalEventComponent(calendarSeed int64, useRealTime bool) *SeasonalEventComponent {
	logrus.WithFields(logrus.Fields{
		"component_type": "seasonal_event",
		"calendar_seed":  calendarSeed,
		"use_real_time":  useRealTime,
	}).Debug("Creating seasonal event component")

	comp := &SeasonalEventComponent{
		ActiveEvents:   make([]EventInstance, 0),
		EventCalendar:  make([]EventDefinition, 0),
		CurrentTime:    time.Unix(0, 0),
		UseRealTime:    useRealTime,
		LastUpdateTime: time.Unix(0, 0),
		CalendarSeed:   calendarSeed,
	}

	// Generate the default event calendar
	comp.EventCalendar = GenerateEventCalendar(calendarSeed)

	return comp
}

// Type returns the component type identifier.
func (s *SeasonalEventComponent) Type() string {
	return "seasonal_event"
}

// GetActiveEvents returns all currently active events.
func (s *SeasonalEventComponent) GetActiveEvents() []EventInstance {
	active := make([]EventInstance, 0)
	for _, event := range s.ActiveEvents {
		if event.Phase == EventPhaseActive {
			active = append(active, event)
		}
	}
	return active
}

// GetUpcomingEvents returns events that will start within the given hours.
func (s *SeasonalEventComponent) GetUpcomingEvents(withinHours int) []EventInstance {
	upcoming := make([]EventInstance, 0)
	threshold := s.CurrentTime.Add(time.Duration(withinHours) * time.Hour)

	for _, event := range s.ActiveEvents {
		if event.Phase == EventPhaseUpcoming && event.StartTime.Before(threshold) {
			upcoming = append(upcoming, event)
		}
	}
	return upcoming
}

// IsEventActive returns true if the specified event is currently active.
func (s *SeasonalEventComponent) IsEventActive(eventID string) bool {
	for _, event := range s.ActiveEvents {
		if event.Definition.ID == eventID && event.Phase == EventPhaseActive {
			return true
		}
	}
	return false
}

// GetEventByID returns the event instance for the given ID, if active.
func (s *SeasonalEventComponent) GetEventByID(eventID string) *EventInstance {
	for i := range s.ActiveEvents {
		if s.ActiveEvents[i].Definition.ID == eventID {
			return &s.ActiveEvents[i]
		}
	}
	return nil
}

// HasActiveEventWithTheme returns true if any active event has the specified theme.
func (s *SeasonalEventComponent) HasActiveEventWithTheme(theme EventTheme) bool {
	for _, event := range s.ActiveEvents {
		if event.Phase == EventPhaseActive && event.Definition.Theme == theme {
			return true
		}
	}
	return false
}

// Serialize encodes the component to bytes for persistence.
func (s *SeasonalEventComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type": "seasonal_event",
		"active_events":  len(s.ActiveEvents),
		"calendar_size":  len(s.EventCalendar),
	}).Debug("Serializing seasonal event component")

	data, err := json.Marshal(s)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "seasonal_event",
			"error":          err.Error(),
		}).Error("Failed to serialize seasonal event component")
		return nil, err
	}
	return data, nil
}

// Deserialize decodes the component from bytes.
func (s *SeasonalEventComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "seasonal_event",
		"bytes":          len(data),
	}).Debug("Deserializing seasonal event component")

	if err := json.Unmarshal(data, s); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "seasonal_event",
			"error":          err.Error(),
		}).Error("Failed to deserialize seasonal event component")
		return err
	}
	return nil
}

// GenerateEventCalendar generates a deterministic yearly event calendar.
// Uses the seed to create consistent events across sessions.
func GenerateEventCalendar(seed int64) []EventDefinition {
	rng := rand.New(rand.NewSource(seed))

	// Base event templates for each season
	events := []EventDefinition{
		// Spring Festival - renewal, planting
		{
			ID:             "spring_festival",
			Name:           generateEventName(rng, EventThemeSpring),
			Theme:          EventThemeSpring,
			DurationDays:   7,
			StartDayOfYear: 80 + rng.Intn(15), // Late March
			Seed:           rng.Int63(),
			Description:    "A celebration of renewal and new beginnings.",
		},
		// Summer Solstice - abundance, light
		{
			ID:             "summer_solstice",
			Name:           generateEventName(rng, EventThemeSummer),
			Theme:          EventThemeSummer,
			DurationDays:   5,
			StartDayOfYear: 170 + rng.Intn(10), // Late June
			Seed:           rng.Int63(),
			Description:    "The longest day brings great festivities.",
		},
		// Harvest Festival - autumn, gathering
		{
			ID:             "harvest_festival",
			Name:           generateEventName(rng, EventThemeAutumn),
			Theme:          EventThemeAutumn,
			DurationDays:   10,
			StartDayOfYear: 265 + rng.Intn(20), // Late September
			Seed:           rng.Int63(),
			Description:    "A time to celebrate the year's bounty.",
		},
		// Winter Celebration - warmth, community
		{
			ID:             "winter_celebration",
			Name:           generateEventName(rng, EventThemeWinter),
			Theme:          EventThemeWinter,
			DurationDays:   14,
			StartDayOfYear: 350 + rng.Intn(10), // Late December
			Seed:           rng.Int63(),
			Description:    "Warmth and cheer in the coldest days.",
		},
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "seasonal_event",
		"seed":           seed,
		"events_count":   len(events),
	}).Debug("Generated event calendar")

	return events
}

// generateEventName creates a procedural event name based on theme.
func generateEventName(rng *rand.Rand, theme EventTheme) string {
	prefixes := map[EventTheme][]string{
		EventThemeSpring: {"Bloom", "Awakening", "Renewal", "Verdant"},
		EventThemeSummer: {"Sunfire", "Radiant", "Golden", "Midsummer"},
		EventThemeAutumn: {"Harvest", "Amber", "Twilight", "Crimson"},
		EventThemeWinter: {"Frost", "Starlight", "Evergreen", "Hearthfire"},
	}

	suffixes := []string{"Festival", "Celebration", "Gathering", "Rite"}

	prefix := prefixes[theme][rng.Intn(len(prefixes[theme]))]
	suffix := suffixes[rng.Intn(len(suffixes))]

	return prefix + " " + suffix
}
