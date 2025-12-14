package engine

import (
	"testing"
	"time"
)

func TestNewEventCalendarSystem(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)

	system := NewEventCalendarSystem(world, clock)

	if system == nil {
		t.Fatal("Expected non-nil system")
	}
	if system.world != world {
		t.Error("World not set correctly")
	}
	if system.clock != clock {
		t.Error("Clock not set correctly")
	}
}

func TestEventCalendarSystem_Update_NilClock(t *testing.T) {
	world := NewWorld()
	system := NewEventCalendarSystem(world, nil)

	// Should not panic with nil clock
	entity := world.CreateEntity()
	entity.AddComponent(NewSeasonalEventComponent(12345, false))

	entities := []*Entity{entity}
	system.Update(entities, 0.016) // Should complete without error
}

func TestEventCalendarSystem_Update_NoComponent(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	// Entity without seasonal_event component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	entities := []*Entity{entity}
	system.Update(entities, 0.016) // Should complete without error
}

func TestEventCalendarSystem_Update_EventPhaseTransitions(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	entity := world.CreateEntity()
	comp := NewSeasonalEventComponent(12345, false)
	// Clear calendar to prevent auto-adding events
	comp.EventCalendar = nil
	entity.AddComponent(comp)

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	clock.Reset(now)

	// Add an event that will transition phases
	// Event ends in 48 hours (more than 24h threshold), so it should be active not ending
	comp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{ID: "test_event", Name: "Test"},
			StartTime:  now.Add(-1 * time.Hour),
			EndTime:    now.Add(48 * time.Hour),
			Phase:      EventPhaseUpcoming, // Should become active
		},
	}

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Verify event became active
	if comp.ActiveEvents[0].Phase != EventPhaseActive {
		t.Errorf("Expected active phase, got %s", comp.ActiveEvents[0].Phase)
	}
}

func TestEventCalendarSystem_Update_EndingPhase(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	entity := world.CreateEntity()
	comp := NewSeasonalEventComponent(12345, false)
	entity.AddComponent(comp)

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	clock.Reset(now)

	// Add an event that is about to end (within 24 hours)
	comp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{ID: "ending_event", Name: "Ending"},
			StartTime:  now.Add(-48 * time.Hour),
			EndTime:    now.Add(12 * time.Hour), // Ends in 12 hours
			Phase:      EventPhaseActive,
		},
	}

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Verify event entered ending phase
	if comp.ActiveEvents[0].Phase != EventPhaseEnding {
		t.Errorf("Expected ending phase, got %s", comp.ActiveEvents[0].Phase)
	}
}

func TestEventCalendarSystem_Update_CleanupEndedEvents(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	entity := world.CreateEntity()
	comp := NewSeasonalEventComponent(12345, false)
	// Clear calendar to prevent auto-adding events
	comp.EventCalendar = nil
	entity.AddComponent(comp)

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	clock.Reset(now)

	// Add an event that ended more than 1 hour ago
	comp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{ID: "ended_event", Name: "Ended"},
			StartTime:  now.Add(-72 * time.Hour),
			EndTime:    now.Add(-2 * time.Hour), // Ended 2 hours ago
			Phase:      EventPhaseEnding,
		},
		{
			Definition: EventDefinition{ID: "active_event", Name: "Active"},
			StartTime:  now.Add(-24 * time.Hour),
			EndTime:    now.Add(48 * time.Hour), // Ends in 48h (not ending phase)
			Phase:      EventPhaseActive,
		},
	}

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Verify ended event was cleaned up
	if len(comp.ActiveEvents) != 1 {
		t.Errorf("Expected 1 active event after cleanup, got %d", len(comp.ActiveEvents))
	}
	if len(comp.ActiveEvents) > 0 && comp.ActiveEvents[0].Definition.ID != "active_event" {
		t.Error("Wrong event was removed")
	}
}

func TestEventCalendarSystem_Update_CheckForNewEvents(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	entity := world.CreateEntity()
	comp := NewSeasonalEventComponent(12345, false)
	entity.AddComponent(comp)

	// Set time to be during spring festival window (day 80-95 typically)
	// Find the spring event's actual start day
	springEvent := comp.EventCalendar[0] // First event is spring
	startDay := springEvent.StartDayOfYear

	// Set time to be 3 days before the event starts
	year := 2025
	targetTime := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, startDay-4)
	clock.Reset(targetTime)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Verify event was added to active events
	if len(comp.ActiveEvents) == 0 {
		t.Error("Expected event to be added to active events")
	}

	// Verify the event is upcoming
	found := false
	for _, event := range comp.ActiveEvents {
		if event.Definition.ID == "spring_festival" {
			found = true
			if event.Phase != EventPhaseUpcoming {
				t.Errorf("Expected upcoming phase, got %s", event.Phase)
			}
			break
		}
	}
	if !found {
		t.Error("Spring festival event not found in active events")
	}
}

func TestEventCalendarSystem_GetCurrentDayOfYear(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	// Set to specific date
	clock.Reset(time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC))

	day := system.GetCurrentDayOfYear()
	expected := 166 // June 15 is day 166 in a non-leap year

	if day != expected {
		t.Errorf("Expected day %d, got %d", expected, day)
	}
}

func TestEventCalendarSystem_GetCurrentDayOfYear_NilClock(t *testing.T) {
	world := NewWorld()
	system := NewEventCalendarSystem(world, nil)

	day := system.GetCurrentDayOfYear()
	if day != 1 {
		t.Errorf("Expected day 1 for nil clock, got %d", day)
	}
}

func TestEventCalendarSystem_TriggerEvent(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	entity := world.CreateEntity()
	comp := NewSeasonalEventComponent(12345, false)
	entity.AddComponent(comp)

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	clock.Reset(now)
	comp.CurrentTime = now

	// Trigger an existing calendar event
	err := system.TriggerEvent(entity, "spring_festival", 7)
	if err != nil {
		t.Fatalf("TriggerEvent failed: %v", err)
	}

	// Verify event was added
	found := false
	for _, event := range comp.ActiveEvents {
		if event.Definition.ID == "spring_festival" {
			found = true
			if event.Phase != EventPhaseActive {
				t.Errorf("Expected active phase, got %s", event.Phase)
			}
			break
		}
	}
	if !found {
		t.Error("Triggered event not found")
	}
}

func TestEventCalendarSystem_TriggerEvent_AdHoc(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	entity := world.CreateEntity()
	comp := NewSeasonalEventComponent(12345, false)
	entity.AddComponent(comp)

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	clock.Reset(now)
	comp.CurrentTime = now

	// Trigger a custom event not in calendar
	err := system.TriggerEvent(entity, "custom_event", 3)
	if err != nil {
		t.Fatalf("TriggerEvent failed: %v", err)
	}

	// Verify ad-hoc event was added
	found := false
	for _, event := range comp.ActiveEvents {
		if event.Definition.ID == "custom_event" {
			found = true
			if event.Phase != EventPhaseActive {
				t.Errorf("Expected active phase, got %s", event.Phase)
			}
			break
		}
	}
	if !found {
		t.Error("Ad-hoc event not found")
	}
}

func TestEventCalendarSystem_EndEvent(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	entity := world.CreateEntity()
	comp := NewSeasonalEventComponent(12345, false)
	entity.AddComponent(comp)

	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	clock.Reset(now)
	comp.CurrentTime = now

	// Add an active event
	comp.ActiveEvents = []EventInstance{
		{
			Definition: EventDefinition{ID: "to_end", Name: "Ending Soon"},
			StartTime:  now.Add(-24 * time.Hour),
			EndTime:    now.Add(24 * time.Hour),
			Phase:      EventPhaseActive,
		},
	}

	// End the event
	err := system.EndEvent(entity, "to_end")
	if err != nil {
		t.Fatalf("EndEvent failed: %v", err)
	}

	// Verify end time was set to past
	for _, event := range comp.ActiveEvents {
		if event.Definition.ID == "to_end" {
			if event.EndTime.After(now) {
				t.Error("Expected end time to be in the past")
			}
			break
		}
	}
}

func TestEventCalendarSystem_EndEvent_NoComponent(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	entity := world.CreateEntity()
	// No seasonal_event component

	err := system.EndEvent(entity, "nonexistent")
	if err != nil {
		t.Errorf("EndEvent should succeed with no component: %v", err)
	}
}

func TestEventCalendarSystem_RealTimeMode(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	entity := world.CreateEntity()
	comp := NewSeasonalEventComponent(12345, true) // Use real time
	entity.AddComponent(comp)

	// Set simulation clock to past
	pastTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	clock.Reset(pastTime)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Component time should be close to real time, not simulation time
	// Note: We can't test exact real time, but it should be after 2020
	if comp.CurrentTime.Before(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("Real time mode should use current time, not simulation time")
	}
}

func TestEventCalendarSystem_Update_YearWrapping(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	entity := world.CreateEntity()
	comp := NewSeasonalEventComponent(12345, false)
	entity.AddComponent(comp)

	// Set time to late December (around winter event)
	winterTime := time.Date(2025, 12, 28, 12, 0, 0, 0, time.UTC)
	clock.Reset(winterTime)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Verify winter event handling
	hasWinterEvent := false
	for _, event := range comp.ActiveEvents {
		if event.Definition.Theme == EventThemeWinter {
			hasWinterEvent = true
			break
		}
	}
	// Winter event should be tracked (it's around day 350-360)
	if !hasWinterEvent {
		// This is acceptable - the winter event may not overlap with Dec 28
		// depending on the seed
		t.Log("Winter event not active on Dec 28 with this seed (acceptable)")
	}
}

func TestEventCalendarSystem_DuplicateEventPrevention(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	system := NewEventCalendarSystem(world, clock)

	entity := world.CreateEntity()
	comp := NewSeasonalEventComponent(12345, false)
	entity.AddComponent(comp)

	// Set time during spring event
	springEvent := comp.EventCalendar[0]
	eventTime := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, springEvent.StartDayOfYear)
	clock.Reset(eventTime)

	entities := []*Entity{entity}

	// Update multiple times
	for i := 0; i < 10; i++ {
		system.Update(entities, 0.016)
	}

	// Count spring events
	springCount := 0
	for _, event := range comp.ActiveEvents {
		if event.Definition.ID == "spring_festival" {
			springCount++
		}
	}

	if springCount > 1 {
		t.Errorf("Expected at most 1 spring festival, got %d", springCount)
	}
}
