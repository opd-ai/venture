package engine

import (
	"testing"
	"time"
)

func TestSeasonalEventComponent_Type(t *testing.T) {
	comp := NewSeasonalEventComponent(12345, false)
	if comp.Type() != "seasonal_event" {
		t.Errorf("Expected type 'seasonal_event', got '%s'", comp.Type())
	}
}

func TestSeasonalEventComponent_NewSeasonalEventComponent(t *testing.T) {
	comp := NewSeasonalEventComponent(12345, true)

	if comp.CalendarSeed != 12345 {
		t.Errorf("Expected CalendarSeed 12345, got %d", comp.CalendarSeed)
	}
	if !comp.UseRealTime {
		t.Error("Expected UseRealTime to be true")
	}
	if len(comp.EventCalendar) != 4 {
		t.Errorf("Expected 4 events in calendar, got %d", len(comp.EventCalendar))
	}
	if len(comp.ActiveEvents) != 0 {
		t.Errorf("Expected empty active events, got %d", len(comp.ActiveEvents))
	}
}

func TestSeasonalEventComponent_GetActiveEvents(t *testing.T) {
	comp := NewSeasonalEventComponent(12345, false)

	// Add test events with different phases
	comp.ActiveEvents = []EventInstance{
		{Definition: EventDefinition{ID: "active1"}, Phase: EventPhaseActive},
		{Definition: EventDefinition{ID: "upcoming1"}, Phase: EventPhaseUpcoming},
		{Definition: EventDefinition{ID: "active2"}, Phase: EventPhaseActive},
		{Definition: EventDefinition{ID: "ending1"}, Phase: EventPhaseEnding},
	}

	active := comp.GetActiveEvents()
	if len(active) != 2 {
		t.Errorf("Expected 2 active events, got %d", len(active))
	}

	// Verify only active events are returned
	for _, event := range active {
		if event.Phase != EventPhaseActive {
			t.Errorf("Expected active phase, got %s", event.Phase)
		}
	}
}

func TestSeasonalEventComponent_GetUpcomingEvents(t *testing.T) {
	comp := NewSeasonalEventComponent(12345, false)
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	comp.CurrentTime = now

	// Add upcoming events at different times
	comp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{ID: "soon"},
			Phase:      EventPhaseUpcoming,
			StartTime:  now.Add(6 * time.Hour),
		},
		{
			Definition: EventDefinition{ID: "later"},
			Phase:      EventPhaseUpcoming,
			StartTime:  now.Add(48 * time.Hour),
		},
		{
			Definition: EventDefinition{ID: "active"},
			Phase:      EventPhaseActive,
			StartTime:  now.Add(-1 * time.Hour),
		},
	}

	upcoming := comp.GetUpcomingEvents(24)
	if len(upcoming) != 1 {
		t.Errorf("Expected 1 upcoming event within 24 hours, got %d", len(upcoming))
	}
	if len(upcoming) > 0 && upcoming[0].Definition.ID != "soon" {
		t.Errorf("Expected 'soon' event, got '%s'", upcoming[0].Definition.ID)
	}
}

func TestSeasonalEventComponent_IsEventActive(t *testing.T) {
	comp := NewSeasonalEventComponent(12345, false)

	comp.ActiveEvents = []EventInstance{
		{Definition: EventDefinition{ID: "active"}, Phase: EventPhaseActive},
		{Definition: EventDefinition{ID: "upcoming"}, Phase: EventPhaseUpcoming},
	}

	tests := []struct {
		eventID  string
		expected bool
	}{
		{"active", true},
		{"upcoming", false},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.eventID, func(t *testing.T) {
			result := comp.IsEventActive(tt.eventID)
			if result != tt.expected {
				t.Errorf("IsEventActive(%s) = %v, expected %v", tt.eventID, result, tt.expected)
			}
		})
	}
}

func TestSeasonalEventComponent_GetEventByID(t *testing.T) {
	comp := NewSeasonalEventComponent(12345, false)

	comp.ActiveEvents = []EventInstance{
		{Definition: EventDefinition{ID: "event1", Name: "First Event"}, Phase: EventPhaseActive},
		{Definition: EventDefinition{ID: "event2", Name: "Second Event"}, Phase: EventPhaseUpcoming},
	}

	// Test existing event
	event := comp.GetEventByID("event1")
	if event == nil {
		t.Fatal("Expected to find event1")
	}
	if event.Definition.Name != "First Event" {
		t.Errorf("Expected 'First Event', got '%s'", event.Definition.Name)
	}

	// Test non-existent event
	event = comp.GetEventByID("nonexistent")
	if event != nil {
		t.Error("Expected nil for non-existent event")
	}
}

func TestSeasonalEventComponent_HasActiveEventWithTheme(t *testing.T) {
	comp := NewSeasonalEventComponent(12345, false)

	comp.ActiveEvents = []EventInstance{
		{Definition: EventDefinition{ID: "spring", Theme: EventThemeSpring}, Phase: EventPhaseActive},
		{Definition: EventDefinition{ID: "winter", Theme: EventThemeWinter}, Phase: EventPhaseUpcoming},
	}

	tests := []struct {
		theme    EventTheme
		expected bool
	}{
		{EventThemeSpring, true},  // Active
		{EventThemeWinter, false}, // Not active (upcoming)
		{EventThemeSummer, false}, // Not present
		{EventThemeAutumn, false}, // Not present
	}

	for _, tt := range tests {
		t.Run(string(tt.theme), func(t *testing.T) {
			result := comp.HasActiveEventWithTheme(tt.theme)
			if result != tt.expected {
				t.Errorf("HasActiveEventWithTheme(%s) = %v, expected %v", tt.theme, result, tt.expected)
			}
		})
	}
}

func TestSeasonalEventComponent_Serialize_Deserialize(t *testing.T) {
	original := NewSeasonalEventComponent(12345, true)
	original.CurrentTime = time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	original.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{
				ID:             "test_event",
				Name:           "Test Festival",
				Theme:          EventThemeSummer,
				DurationDays:   7,
				StartDayOfYear: 166,
				Seed:           99999,
			},
			StartTime: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 6, 22, 0, 0, 0, 0, time.UTC),
			Phase:     EventPhaseActive,
		},
	}

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Serialize returned empty data")
	}

	// Deserialize
	restored := &SeasonalEventComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify values
	if restored.CalendarSeed != original.CalendarSeed {
		t.Errorf("CalendarSeed mismatch: %d vs %d", restored.CalendarSeed, original.CalendarSeed)
	}
	if restored.UseRealTime != original.UseRealTime {
		t.Errorf("UseRealTime mismatch: %v vs %v", restored.UseRealTime, original.UseRealTime)
	}
	if len(restored.ActiveEvents) != len(original.ActiveEvents) {
		t.Errorf("ActiveEvents count mismatch: %d vs %d", len(restored.ActiveEvents), len(original.ActiveEvents))
	}
	if len(restored.EventCalendar) != len(original.EventCalendar) {
		t.Errorf("EventCalendar count mismatch: %d vs %d", len(restored.EventCalendar), len(original.EventCalendar))
	}

	// Verify event instance
	if len(restored.ActiveEvents) > 0 {
		if restored.ActiveEvents[0].Definition.ID != "test_event" {
			t.Errorf("Event ID mismatch: %s", restored.ActiveEvents[0].Definition.ID)
		}
		if restored.ActiveEvents[0].Phase != EventPhaseActive {
			t.Errorf("Event phase mismatch: %s", restored.ActiveEvents[0].Phase)
		}
	}
}

func TestGenerateEventCalendar_Determinism(t *testing.T) {
	seed := int64(12345)

	// Generate twice with same seed
	calendar1 := GenerateEventCalendar(seed)
	calendar2 := GenerateEventCalendar(seed)

	if len(calendar1) != len(calendar2) {
		t.Fatalf("Calendar lengths differ: %d vs %d", len(calendar1), len(calendar2))
	}

	for i := range calendar1 {
		if calendar1[i].ID != calendar2[i].ID {
			t.Errorf("Event %d ID mismatch: %s vs %s", i, calendar1[i].ID, calendar2[i].ID)
		}
		if calendar1[i].Name != calendar2[i].Name {
			t.Errorf("Event %d Name mismatch: %s vs %s", i, calendar1[i].Name, calendar2[i].Name)
		}
		if calendar1[i].StartDayOfYear != calendar2[i].StartDayOfYear {
			t.Errorf("Event %d StartDayOfYear mismatch: %d vs %d", i, calendar1[i].StartDayOfYear, calendar2[i].StartDayOfYear)
		}
		if calendar1[i].Seed != calendar2[i].Seed {
			t.Errorf("Event %d Seed mismatch: %d vs %d", i, calendar1[i].Seed, calendar2[i].Seed)
		}
	}
}

func TestGenerateEventCalendar_Content(t *testing.T) {
	calendar := GenerateEventCalendar(42)

	if len(calendar) != 4 {
		t.Fatalf("Expected 4 events, got %d", len(calendar))
	}

	// Verify we have one event for each theme
	themes := make(map[EventTheme]bool)
	for _, event := range calendar {
		themes[event.Theme] = true

		// Verify event properties
		if event.ID == "" {
			t.Error("Event has empty ID")
		}
		if event.Name == "" {
			t.Error("Event has empty Name")
		}
		if event.DurationDays < 1 {
			t.Errorf("Event %s has invalid duration: %d", event.ID, event.DurationDays)
		}
		if event.StartDayOfYear < 1 || event.StartDayOfYear > 365 {
			t.Errorf("Event %s has invalid start day: %d", event.ID, event.StartDayOfYear)
		}
	}

	// Verify all themes present
	if !themes[EventThemeSpring] {
		t.Error("Missing spring event")
	}
	if !themes[EventThemeSummer] {
		t.Error("Missing summer event")
	}
	if !themes[EventThemeAutumn] {
		t.Error("Missing autumn event")
	}
	if !themes[EventThemeWinter] {
		t.Error("Missing winter event")
	}
}

func TestEventTheme_Constants(t *testing.T) {
	themes := []EventTheme{
		EventThemeSpring,
		EventThemeSummer,
		EventThemeAutumn,
		EventThemeWinter,
	}

	for _, theme := range themes {
		if theme == "" {
			t.Error("Event theme constant is empty")
		}
	}
}

func TestEventPhase_Constants(t *testing.T) {
	phases := []EventPhase{
		EventPhaseUpcoming,
		EventPhaseActive,
		EventPhaseEnding,
	}

	for _, phase := range phases {
		if phase == "" {
			t.Error("Event phase constant is empty")
		}
	}
}

func TestGenerateEventCalendar_DifferentSeeds(t *testing.T) {
	calendar1 := GenerateEventCalendar(111)
	calendar2 := GenerateEventCalendar(222)

	// Events should be different with different seeds
	differences := 0
	for i := range calendar1 {
		if calendar1[i].Name != calendar2[i].Name {
			differences++
		}
		if calendar1[i].StartDayOfYear != calendar2[i].StartDayOfYear {
			differences++
		}
	}

	if differences == 0 {
		t.Error("Different seeds should produce different calendars")
	}
}
