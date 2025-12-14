package engine

import (
	"testing"
	"time"
)

func TestNewAvailabilitySystem(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)

	sys := NewAvailabilitySystem(world, clock)
	if sys == nil {
		t.Fatal("NewAvailabilitySystem() returned nil")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.clock != clock {
		t.Error("clock not set correctly")
	}
}

func TestAvailabilitySystem_IsEntityOpen_NoOperatingHours(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	sys := NewAvailabilitySystem(world, clock)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	// Entity without operating hours should always be open
	if !sys.IsEntityOpen(entity) {
		t.Error("Entity without operating hours should be open")
	}
}

func TestAvailabilitySystem_IsEntityOpen_WithOperatingHours(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	// Set clock to Monday at 10:00 AM
	clock.Reset(time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC)) // Jan 6, 2025 is Monday
	sys := NewAvailabilitySystem(world, clock)

	entity := world.CreateEntity()
	hours := NewOperatingHoursComponent()
	hours.SetHours(8, 18) // Open 8 AM - 6 PM
	entity.AddComponent(hours)

	// Should be open at 10 AM on Monday
	if !sys.IsEntityOpen(entity) {
		t.Error("Entity should be open at 10 AM on Monday")
	}

	// Move clock to 6 AM (before opening)
	clock.Reset(time.Date(2025, 1, 6, 6, 0, 0, 0, time.UTC))
	if sys.IsEntityOpen(entity) {
		t.Error("Entity should be closed at 6 AM")
	}

	// Move clock to Sunday (closed day)
	clock.Reset(time.Date(2025, 1, 5, 12, 0, 0, 0, time.UTC)) // Sunday at noon
	if sys.IsEntityOpen(entity) {
		t.Error("Entity should be closed on Sunday")
	}
}

func TestAvailabilitySystem_IsEntityAvailable(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC))
	sys := NewAvailabilitySystem(world, clock)

	entity := world.CreateEntity()
	hours := NewOperatingHoursComponent()
	entity.AddComponent(hours)

	// Flush pending entities
	world.Update(0)

	// Should be available
	if !sys.IsEntityAvailable(entity.ID) {
		t.Error("Entity should be available")
	}

	// Non-existent entity
	if sys.IsEntityAvailable(99999) {
		t.Error("Non-existent entity should not be available")
	}
}

func TestAvailabilitySystem_GetEntityStatus(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC))
	sys := NewAvailabilitySystem(world, clock)

	// Entity without operating hours
	entity1 := world.CreateEntity()
	if status := sys.GetEntityStatus(entity1); status != "Open" {
		t.Errorf("GetEntityStatus() for entity without hours = %q, want %q", status, "Open")
	}

	// Entity with operating hours (open)
	entity2 := world.CreateEntity()
	hours := NewOperatingHoursComponent()
	hours.ClosedMessage = "Gone to lunch"
	entity2.AddComponent(hours)

	if status := sys.GetEntityStatus(entity2); status != "Open" {
		t.Errorf("GetEntityStatus() for open entity = %q, want %q", status, "Open")
	}

	// Move to closed time
	clock.Reset(time.Date(2025, 1, 6, 20, 0, 0, 0, time.UTC)) // 8 PM
	if status := sys.GetEntityStatus(entity2); status != "Gone to lunch" {
		t.Errorf("GetEntityStatus() for closed entity = %q, want %q", status, "Gone to lunch")
	}
}

func TestAvailabilitySystem_GetNextOpenTime(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 6, 6, 0, 0, 0, time.UTC)) // Monday 6 AM
	sys := NewAvailabilitySystem(world, clock)

	entity := world.CreateEntity()
	hours := NewOperatingHoursComponent()
	hours.SetHours(8, 18)
	entity.AddComponent(hours)

	hour, day, hasHours := sys.GetNextOpenTime(entity)
	if !hasHours {
		t.Error("Entity should have operating hours")
	}
	if hour != 8 {
		t.Errorf("Next open hour = %d, want 8", hour)
	}
	if day != Monday {
		t.Errorf("Next open day = %s, want Monday", day.String())
	}

	// Entity without hours
	entity2 := world.CreateEntity()
	_, _, hasHours2 := sys.GetNextOpenTime(entity2)
	if hasHours2 {
		t.Error("Entity without hours should return hasHours=false")
	}
}

func TestAvailabilitySystem_CheckInteractionAllowed(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC)) // Monday 10 AM
	sys := NewAvailabilitySystem(world, clock)

	entity := world.CreateEntity()
	hours := NewOperatingHoursComponent()
	hours.ClosedMessage = "We're closed!"
	entity.AddComponent(hours)

	// Flush pending entities
	world.Update(0)

	// Should be allowed when open
	allowed, msg := sys.CheckInteractionAllowed(entity.ID)
	if !allowed {
		t.Errorf("Interaction should be allowed when open, msg: %s", msg)
	}
	if msg != "" {
		t.Errorf("Message should be empty when allowed, got: %q", msg)
	}

	// Move to closed time
	clock.Reset(time.Date(2025, 1, 6, 20, 0, 0, 0, time.UTC)) // 8 PM
	allowed, msg = sys.CheckInteractionAllowed(entity.ID)
	if allowed {
		t.Error("Interaction should not be allowed when closed")
	}
	if msg == "" {
		t.Error("Message should explain why interaction is blocked")
	}

	// Non-existent entity
	allowed, msg = sys.CheckInteractionAllowed(99999)
	if allowed {
		t.Error("Interaction with non-existent entity should not be allowed")
	}
}

func TestAvailabilitySystem_Update(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC)) // Monday 10 AM
	sys := NewAvailabilitySystem(world, clock)

	entity := world.CreateEntity()
	hours := NewOperatingHoursComponent()
	entity.AddComponent(hours)

	// Add dialog component with shop option
	dialog := &DialogComponent{
		Options: []DialogOption{
			{Text: "Browse your wares", Action: ActionOpenShop, Enabled: true},
			{Text: "Goodbye", Action: ActionCloseDialog, Enabled: true},
		},
	}
	entity.AddComponent(dialog)

	// Update should enable shop option when open
	sys.Update([]*Entity{entity}, 0.016)

	if !dialog.Options[0].Enabled {
		t.Error("Shop option should be enabled when open")
	}
	if dialog.Options[0].Text != "Browse your wares" {
		t.Errorf("Shop option text = %q, want %q", dialog.Options[0].Text, "Browse your wares")
	}

	// Move to closed time and update
	clock.Reset(time.Date(2025, 1, 6, 20, 0, 0, 0, time.UTC)) // 8 PM
	sys.Update([]*Entity{entity}, 0.016)

	if dialog.Options[0].Enabled {
		t.Error("Shop option should be disabled when closed")
	}
	if dialog.Options[0].Text != "Shop is closed" {
		t.Errorf("Shop option text = %q, want %q", dialog.Options[0].Text, "Shop is closed")
	}
}

func TestAvailabilitySystem_NilClock(t *testing.T) {
	world := NewWorld()
	sys := NewAvailabilitySystem(world, nil)

	entity := world.CreateEntity()
	hours := NewOperatingHoursComponent()
	entity.AddComponent(hours)

	// Should default to available when no clock
	if !sys.IsEntityOpen(entity) {
		t.Error("Entity should be considered open when no clock is set")
	}

	// Update should not panic with nil clock
	sys.Update([]*Entity{entity}, 0.016) // Should not panic
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		hour int
		want string
	}{
		{0, "12:00 AM"},
		{1, "01:00 AM"},
		{11, "11:00 AM"},
		{12, "12:00 PM"},
		{13, "01:00 PM"},
		{23, "11:00 PM"},
	}

	for _, tt := range tests {
		if got := formatTime(tt.hour); got != tt.want {
			t.Errorf("formatTime(%d) = %q, want %q", tt.hour, got, tt.want)
		}
	}
}

func TestAvailabilitySystem_AlwaysOpenEntity(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 5, 3, 0, 0, 0, time.UTC)) // Sunday 3 AM
	sys := NewAvailabilitySystem(world, clock)

	entity := world.CreateEntity()
	hours := NewAlwaysOpenComponent("inn")
	entity.AddComponent(hours)

	// Flush pending entities
	world.Update(0)

	// Should be open at any time
	if !sys.IsEntityOpen(entity) {
		t.Error("Always-open entity should be open at 3 AM Sunday")
	}

	allowed, _ := sys.CheckInteractionAllowed(entity.ID)
	if !allowed {
		t.Error("Interaction with always-open entity should be allowed")
	}
}
